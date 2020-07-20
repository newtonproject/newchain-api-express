package api

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"
)

type NotifyConfig struct {
	Server      string
	Username    string
	Password    string
	ClientID    string
	QoS         byte
	PrefixTopic string // topic = <PrefixTopic>/<address>/<confirmedBlock>
}

func getPublishClient(n *NotifyConfig) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().AddBroker(n.Server).SetClientID(n.ClientID)
	opts.SetUsername(n.Username)
	opts.SetPassword(n.Password)
	c := mqtt.NewClient(opts)

	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return c, nil
}

func (s *Server) sendNotify(tx *TransferTx, confirmed int64) {

	payload, err := json.Marshal(tx)
	if err != nil {
		log.Error(err)
		return
	}
	var topic string
	if tx.To == nil {
		topic = fmt.Sprintf("%s/ContractCreate", s.notify.PrefixTopic)
	} else {
		topic = fmt.Sprintf("%s/%s/%d", s.notify.PrefixTopic, strings.ToLower(tx.To.String()[2:]), confirmed)
	}

	log.WithFields(logrus.Fields{
		"publish": topic,
	}).Info(string(payload))

	s.nc.Publish(topic, s.notify.QoS, false, string(payload))
}

type TransferTx struct {
	From        common.Address  `json:"from"`
	To          *common.Address `json:"to"`
	Value       *big.Int        `json:"value"`
	Hash        common.Hash     `json:"hash"`
	Data        []byte          `json:"data"`
	BlockNumber *big.Int        `json:"blockNumber"`
}

// UnmarshalJSON decodes from json format to a TransferTx.
func (c *TransferTx) UnmarshalJSON(data []byte) error {
	type Tx struct {
		From  common.Address  `json:"from"`
		To    *common.Address `json:"to"`
		Value string          `json:"value"`
		Hash  common.Hash     `json:"hash"`
	}
	var tx Tx
	err := json.Unmarshal(data, &tx)
	if err != nil {
		return err
	}
	c.From = tx.From
	c.To = tx.To
	value, err := hexutil.DecodeBig(tx.Value)
	if err != nil {
		return err
	}
	c.Value = value
	c.Hash = tx.Hash

	return nil
}

// MarshalJSON encodes to json format.
func (c *TransferTx) MarshalJSON() ([]byte, error) {
	type Tx struct {
		From        common.Address  `json:"from"`
		To          *common.Address `json:"to"`
		Value       *hexutil.Big    `json:"value"`
		Hash        common.Hash     `json:"hash"`
		Data        hexutil.Bytes   `json:"data"`
		BlockNumber *hexutil.Big    `json:"blockNumber"`
	}

	enc := &Tx{
		From:        c.From,
		To:          c.To,
		Value:       (*hexutil.Big)(c.Value),
		Hash:        c.Hash,
		Data:        c.Data,
		BlockNumber: (*hexutil.Big)(c.BlockNumber),
	}

	return json.Marshal(&enc)
}
