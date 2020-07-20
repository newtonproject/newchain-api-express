package api

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type tx2Broadcast struct {
	tx   *types.Transaction
	from common.Address
}

type tx2Confirm struct {
	tx   *types.Transaction
	from common.Address
}

type txNotifyReceived struct {
	tx *TransferTx
}

type txNotifyBroadcast struct {
	tx *TransferTx
}

type txNotifyConfirmed struct {
	tx *TransferTx
}

func (s *Server) handleTxs() {
	for {
		select {
		case ch := <-s.txChan:
			switch msg := ch.(type) {
			case tx2Broadcast:
				tx := msg.tx
				from := msg.from
				s.handleBroadcastTx(tx, from)
			case tx2Confirm:
				tx := msg.tx
				from := msg.from
				s.txs2ConfirmLock.Lock()
				s.txs2Confirm = append(s.txs2Confirm, &TransferTx{
					From:  from,
					To:    tx.To(),
					Value: tx.Value(),
					Hash:  tx.Hash(),
					Data:  tx.Data(),
				})
				s.txs2ConfirmLock.Unlock()
			case txNotifyReceived:
				s.sendNotify(msg.tx, -1)
			case txNotifyBroadcast:
				s.sendNotify(msg.tx, 0)
			case txNotifyConfirmed:
				s.sendNotify(msg.tx, 1)
			default:
				log.Warningf("Unknown message type sent: %T", msg)
			}
		}
	}
}

func (s *Server) handleBroadcastTx(tx *types.Transaction, from common.Address) {
	client, err := ethclient.Dial(s.rpcURL)
	if err != nil {
		log.Errorf("%s: BroadcastTx Dial error: %v\n", tx.Hash().String(), err)
		return
	}

	err = client.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Errorf("%s: SendTransaction error: %v\n", tx.Hash().String(), err)
		return
	}

	// send notify
	// notify Broadcast
	s.txChan <- txNotifyBroadcast{tx: &TransferTx{
		From:  from,
		To:    tx.To(),
		Value: tx.Value(),
		Hash:  tx.Hash(),
		Data:  tx.Data(),
	}}

	// ok, wait to be mined
	s.txs2ConfirmLock.Lock()
	s.txs2Confirm = append(s.txs2Confirm, &TransferTx{
		From:  from,
		To:    tx.To(),
		Value: tx.Value(),
		Hash:  tx.Hash(),
		Data:  tx.Data(),
	})
	s.txs2ConfirmLock.Unlock()

	return
}

func (s *Server) handleTxs2Confirm() {
	// get block inter
	var blockPeriod int64
	for {
		client, err := ethclient.Dial(s.rpcURL)
		if err != nil {
			log.Errorln(err)
			time.Sleep(time.Second * 3)
			continue
		}
		ctx := context.Background()

		latestBlock, err := client.HeaderByNumber(ctx, nil)
		if err != nil {
			log.Errorln(err)
			time.Sleep(time.Second * 3)
			continue

		}

		parenBlock, err := client.HeaderByNumber(ctx, big.NewInt(0).Sub(latestBlock.Number, big.NewInt(1)))
		if err != nil {
			log.Errorln(err)
			time.Sleep(time.Second * 3)
			continue
		}

		blockPeriod = int64(latestBlock.Time - parenBlock.Time)
		if blockPeriod <= 0 {
			log.Errorln("get block period error", blockPeriod)
			time.Sleep(time.Second * 3)
			continue
		}

		break

	}
	log.Infof("Block Period is: %d second\n", blockPeriod)

	run := func() {
		var txs []*TransferTx
		s.txs2ConfirmLock.Lock()
		if len(s.txs2Confirm) > 0 {
			txs = s.txs2Confirm
			s.txs2Confirm = make([]*TransferTx, 0)
		}
		s.txs2ConfirmLock.Unlock()

		if len(txs) == 0 {
			return
		}

		client, err := ethclient.Dial(s.rpcURL)
		if err != nil {
			log.Errorf("handleTxs2Confirm Dial error: %v\n", err)
			return
		}

		for _, tx := range txs {
			receipt, err := client.TransactionReceipt(context.Background(), tx.Hash)
			if receipt == nil {
				// add tx back to txs
				s.txs2ConfirmLock.Lock()
				s.txs2Confirm = append(s.txs2Confirm, tx)
				s.txs2ConfirmLock.Unlock()

				if err != nil && err != ethereum.NotFound {
					log.Errorln("handleTxs2Confirm TransactionReceipt error: %v\n", err)
				}

				continue
			}

			// ok, found confirmed tx, notify

			// notify confirmed
			s.txChan <- txNotifyConfirmed{tx: tx}

		}

		return

	}

	ticker := time.NewTicker(time.Duration(blockPeriod) * time.Second)
	for {
		select {
		case <-ticker.C:
			log.Infoln("Run Txs2Confirm ...")
			run()
		}
	}
}
