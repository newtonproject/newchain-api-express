package api

import (
	"context"
	"crypto/elliptic"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/newtonproject/newchain-api-express/params"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var log *logrus.Logger

type Config struct {
	CandidateFee uint64
	VoteFee      uint64
}

var (
	secp256r1N     = elliptic.P256().Params().N
	secp256r1halfN = new(big.Int).Div(secp256r1N, big.NewInt(2))
)

// Server is used to implement forceproto.ForceServer.
type Server struct {
	logger *logrus.Logger

	rpcURL    string
	networkID uint64
	gasPrice  *big.Int

	txChan          chan interface{}
	txs2Confirm     []*TransferTx
	txs2ConfirmLock sync.Mutex

	// notify
	notify *NotifyConfig
	nc     mqtt.Client
}

func logRequest(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info != nil {
		log.WithField("method", info.FullMethod).Info(req)
	}

	// Continue execution of handler after ensuring a valid token.
	return handler(ctx, req)
}

// NewServer listen and server
func NewServer(rpcURL string, notify *NotifyConfig) (*Server, error) {
	log = logrus.New()
	log.Out = os.Stdout

	if notify == nil {
		return nil, errors.New("not set notify config")
	}
	log.Infoln("Try to connect to MQTT server...")
	nc, err := getPublishClient(notify)
	if err != nil {
		return nil, err
	}
	if nc == nil {
		return nil, errors.New("MQTT client init failed")
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	networkID, err := client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	server := &Server{
		rpcURL:      rpcURL,
		gasPrice:    gasPrice,
		networkID:   networkID.Uint64(),
		txChan:      make(chan interface{}, 1024),
		txs2Confirm: make([]*TransferTx, 0),
		notify:      notify,
		nc:          nc,
	}

	go func() {
		// TODO: update gasPrice hourly
	}()

	go server.handleTxs()
	go server.handleTxs2Confirm()

	return server, nil
}

// GetBaseInfoArgs address
type GetBaseInfoArgs struct {
	Address common.Address `json:"address"`
}

type BaseInfo struct {
	NonceLatest  *hexutil.Uint64 `json:"nonceLatest"`
	NoncePending *hexutil.Uint64 `json:"noncePending"`
	GasPrice     *hexutil.Big    `json:"gasPrice"`
	NetworkID    uint64          `json:"networkID"`
	Balance      *hexutil.Big    `json:"balance"`
}

func (s *Server) GetBaseInfo(ctx context.Context, args GetBaseInfoArgs) (*BaseInfo, error) {
	address := args.Address

	client, err := ethclient.Dial(s.rpcURL)
	if err != nil {
		return nil, err
	}

	nonceLatest, err := client.NonceAt(ctx, address, nil)
	if err != nil {
		return nil, err
	}

	noncePending, err := client.PendingNonceAt(ctx, address)
	if err != nil {
		return nil, err
	}

	balance, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, err
	}

	return &BaseInfo{
		GasPrice:     (*hexutil.Big)(s.gasPrice),
		NetworkID:    s.networkID,
		NonceLatest:  (*hexutil.Uint64)(&nonceLatest),
		NoncePending: (*hexutil.Uint64)(&noncePending),
		Balance:      (*hexutil.Big)(balance),
	}, nil
}

// SendRawTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendRawTxArgs struct {
	Tx   hexutil.Bytes `json:"tx"`
	Wait uint64        `json:"wait"`
}

func (s *Server) SendRawTransaction(ctx context.Context, args SendRawTxArgs) (common.Hash, error) {
	wait := args.Wait
	if wait != params.LevelWaitBroadcast && wait != params.LevelWaitConfirmed {
		wait = params.LevelNoWait
	}

	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(args.Tx, tx); err != nil {
		return common.Hash{}, err
	}

	signer := types.NewEIP155Signer(big.NewInt(0).SetUint64(s.networkID))
	from, err := signer.Sender(tx)
	if err != nil {
		return common.Hash{}, err
	}

	// notify received
	s.txChan <- txNotifyReceived{tx: &TransferTx{
		From:  from,
		To:    tx.To(),
		Value: tx.Value(),
		Hash:  tx.Hash(),
		Data:  tx.Data(),
	}}

	if wait == params.LevelNoWait {
		s.txChan <- tx2Broadcast{tx: tx, from: from}
		return tx.Hash(), nil
	}

	client, err := ethclient.Dial(s.rpcURL)
	if err != nil {
		return common.Hash{}, err
	}

	err = client.SendTransaction(ctx, tx)
	if err != nil {
		return common.Hash{}, err
	}

	// notify broadcast
	s.txChan <- txNotifyBroadcast{tx: &TransferTx{
		From:  from,
		To:    tx.To(),
		Value: tx.Value(),
		Hash:  tx.Hash(),
		Data:  tx.Data(),
	}}

	if wait == params.LevelWaitBroadcast {
		s.txChan <- tx2Confirm{tx: tx, from: from}
		return tx.Hash(), nil
	}

	_, err = bind.WaitMined(ctx, client, tx)
	if err != nil {
		return common.Hash{}, err
	}

	// notify confirmed
	s.txChan <- txNotifyConfirmed{tx: &TransferTx{
		From:  from,
		To:    tx.To(),
		Value: tx.Value(),
		Hash:  tx.Hash(),
		Data:  tx.Data(),
	}}

	return tx.Hash(), err
}

// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendTxArgs struct {
	From      common.Address `json:"from"`
	Tx        hexutil.Bytes  `json:"tx"`
	Signature hexutil.Bytes  `json:"signature"`
	Wait      uint64         `json:"wait"`
}

func (s *Server) SendTransaction(ctx context.Context, args SendTxArgs) (common.Hash, error) {
	wait := args.Wait
	if wait != params.LevelWaitBroadcast && wait != params.LevelWaitConfirmed {
		wait = params.LevelNoWait
	}

	from := args.From

	rlpTx := []byte(args.Tx)
	sign := []byte(args.Signature)
	if len(sign) != 64 {
		return common.Hash{}, errors.New("invalid signature length")
	}

	var tx *types.Transaction
	err := rlp.DecodeBytes(rlpTx, &tx)
	if err != nil {
		return common.Hash{}, err
	}

	signer := types.NewEIP155Signer(big.NewInt(0).SetUint64(s.networkID))
	sHash := signer.Hash(tx)

	// sign
	signature := make([]byte, 32*2+1)
	copy(signature[:32], sign) // r

	// check s
	// update upper range of s values (ECDSA malleability)
	// see discussion in secp256k1/libsecp256k1/include/secp256k1.h
	signS := big.NewInt(0).SetBytes(sign[32:64])
	if signS.Cmp(secp256r1halfN) > 0 {
		signS = new(big.Int).Sub(secp256r1N, signS)
	}
	sBytes := signS.Bytes()
	copy(signature[64-len(sBytes):], sBytes) // s

	recId := byte(0)
	for recId = 0; recId < 4; recId++ {
		signature[64] = recId // v
		pk, _ := crypto.SigToPub(sHash.Bytes(), signature)
		if pk != nil && crypto.PubkeyToAddress(*pk) == from {
			break
		}
	}
	if recId == 4 {
		return common.Hash{}, fmt.Errorf("invalid signature, could not construct a recoverable key")
	}

	signTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return common.Hash{}, err
	}

	// ok, tx is ok

	// notify received
	s.txChan <- txNotifyReceived{tx: &TransferTx{
		From:  from,
		To:    signTx.To(),
		Value: signTx.Value(),
		Hash:  signTx.Hash(),
		Data:  signTx.Data(),
	}}

	if wait == params.LevelNoWait {
		s.txChan <- tx2Broadcast{tx: signTx, from: from}
		return signTx.Hash(), nil
	}

	client, err := ethclient.Dial(s.rpcURL)
	if err != nil {
		return common.Hash{}, err
	}
	err = client.SendTransaction(ctx, signTx)
	if err != nil {
		return common.Hash{}, err
	}

	// notify broadcast
	s.txChan <- txNotifyBroadcast{tx: &TransferTx{
		From:  from,
		To:    signTx.To(),
		Value: signTx.Value(),
		Hash:  signTx.Hash(),
		Data:  signTx.Data(),
	}}

	if wait == params.LevelWaitBroadcast {
		s.txChan <- tx2Confirm{tx: signTx, from: from}
		return signTx.Hash(), nil
	}

	_, err = bind.WaitMined(ctx, client, signTx)
	if err != nil {
		return common.Hash{}, err
	}

	// notify confirmed
	s.txChan <- txNotifyConfirmed{tx: &TransferTx{
		From:  from,
		To:    signTx.To(),
		Value: signTx.Value(),
		Hash:  signTx.Hash(),
		Data:  signTx.Data(),
	}}

	return signTx.Hash(), err

}
