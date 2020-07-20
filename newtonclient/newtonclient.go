// Package newtonclient provides a client for the NewChain RPC API.
package newtonclient

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/newtonproject/newchain-api-express/rpc"
)

// Client defines typed wrappers for the Ethereum RPC API.
type Client struct {
	c *rpc.Client
}

// Dial connects a client to the given URL.
func Dial(rawurl string) (*Client, error) {
	return DialContext(context.Background(), rawurl)
}

func DialContext(ctx context.Context, rawurl string) (*Client, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

// NewClient creates a client that uses the given RPC client.
func NewClient(c *rpc.Client) *Client {
	return &Client{c}
}

func (ec *Client) Close() {
	ec.c.Close()
}

type BaseInfo struct {
	GasPrice     *big.Int `json:"gasPrice"`
	NetworkID    uint64   `json:"networkID"`
	NonceLatest  uint64   `json:"nonceLatest"`
	NoncePending uint64   `json:"noncePending"`
	Balance      *big.Int `json:"balance"`
}

// NetworkID returns the network ID (also known as the chain ID) for this chain.
func (ec *Client) GetBaseInfo(ctx context.Context, account common.Address) (*BaseInfo, error) {
	var args = struct {
		Address common.Address `json:"address"`
	}{
		Address: account,
	}

	var info struct {
		NonceLatest  *hexutil.Uint64 `json:"nonceLatest"`
		NoncePending *hexutil.Uint64 `json:"noncePending"`
		GasPrice     *hexutil.Big    `json:"gasPrice"`
		NetworkID    uint64          `json:"networkID"`
		Balance      *hexutil.Big    `json:"balance"`
	}
	if err := ec.c.CallObjectContext(ctx, &info, "newton_getBaseInfo", args); err != nil {
		return nil, err
	}

	gasPrice := big.NewInt(1)
	if info.GasPrice != nil {
		gasPrice = gasPrice.Set(info.GasPrice.ToInt())
	}

	networkID := info.NetworkID

	nonceLatest := uint64(0)
	if info.NonceLatest != nil {
		nonceLatest = uint64(*info.NonceLatest)
	}
	noncePending := uint64(0)
	if info.NoncePending != nil {
		noncePending = uint64(*info.NoncePending)
	}

	balance := big.NewInt(0)
	if info.Balance != nil {
		balance = balance.Set(info.Balance.ToInt())
	}

	return &BaseInfo{
		NonceLatest:  nonceLatest,
		NoncePending: noncePending,
		GasPrice:     gasPrice,
		NetworkID:    networkID,
		Balance:      balance,
	}, nil
}

// SendTransaction injects a signed transaction into the pending pool for execution.
func (ec *Client) SendTransaction(ctx context.Context, rlpTx, signature []byte, from common.Address, wait uint64) (common.Hash, error) {
	var hash common.Hash

	var tx = struct {
		From      common.Address `json:"from"`
		RlpTx     hexutil.Bytes  `json:"tx"`
		Signature hexutil.Bytes  `json:"signature"`
		Wait      uint64         `json:"wait"`
	}{
		From:      from,
		RlpTx:     rlpTx,
		Signature: signature,
		Wait:      wait,
	}

	err := ec.c.CallObjectContext(ctx, &hash, "newton_sendTransaction", tx)
	if err != nil {
		return common.Hash{}, err
	}
	return hash, err
}

// SendRawTransaction injects a signed transaction into the pending pool for execution.
func (ec *Client) SendRawTransaction(ctx context.Context, tx *types.Transaction, wait uint64) (common.Hash, error) {
	var hash common.Hash

	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return common.Hash{}, err
	}

	var sendTx = struct {
		Tx   hexutil.Bytes `json:"tx"`
		Wait uint64        `json:"wait"`
	}{
		Tx:   data,
		Wait: wait,
	}

	err = ec.c.CallObjectContext(ctx, &hash, "newton_sendRawTransaction", sendTx)
	if err != nil {
		return common.Hash{}, err
	}

	return hash, nil
}
