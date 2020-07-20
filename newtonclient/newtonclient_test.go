package newtonclient

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// Verify that Client implements the ethereum interfaces.
var ()

func TestSendTransaction(t *testing.T) {
	client, err := Dial("http://127.0.0.1:8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	// {
	// 	"from": "0x97549e368acafdcae786bb93d98379f1d1561a29",
	// 	"to": "0x97549e368acafdcae786bb93d98379f1d1561a29",
	// 	"value": "1",
	// 	"data": "",
	// 	"nonce": 1247,
	// 	"gasPrice": 100,
	// 	"gas": 21000,
	// 	"v": "0x0801",
	// 	"r": "0xb81c24deef2dd9cde2328bc697ca67ad04ed5ed1649c76abc059f99a90e13968",
	// 	"s": "0x5de45f1149d600c3dda5390a9efaa6beb9fb78d35e871c67f445986b7ecbb2e9",
	// 	"hash": "0x4eada77aed522c7831abc18b29462bb3ec8e011b3884f73d27293ab064b95d61",
	// 	"chainID": 1007
	// }

	// chainID := big.NewInt(1007)
	// signer := types.NewEIP155Signer(chainID)

	to := common.HexToAddress("0x97549e368acafdcae786bb93d98379f1d1561a29")
	tx := types.NewTransaction(1248, to, big.NewInt(0).Mul(big.NewInt(1), big.NewInt(1e+18)),
		21000, big.NewInt(100), nil)

	// rlpHash([]interface{}{
	// 	tx.data.AccountNonce,
	// 	tx.data.Price,
	// 	tx.data.GasLimit,
	// 	tx.data.Recipient,
	// 	tx.data.Amount,
	// 	tx.data.Payload,
	// 	s.chainId, uint(0), uint(0),
	// })

	// message, _ := rlp.EncodeToBytes([]interface{}{
	// 	tx.Nonce(),
	// 	tx.GasPrice(),
	// 	tx.Gas(),
	// 	*tx.To(),
	// 	tx.Value(),
	// 	tx.Data(),
	// 	chainID, uint(0), uint(0),
	// })

	message, _ := rlp.EncodeToBytes(tx)

	// 	"r": "0xb81c24deef2dd9cde2328bc697ca67ad04ed5ed1649c76abc059f99a90e13968",
	// 	"s": "0x5de45f1149d600c3dda5390a9efaa6beb9fb78d35e871c67f445986b7ecbb2e9",
	signature := common.Hex2Bytes("2bfdd5d619d589e5c3d389affbab514ec3d36fe1e21b42d6e09b059e98d7202a7d3c7a5f0325a72cc17ff5b7d436a6d562f27ff1608bc1df60c7166c81a4a948")

	// "r": "0x2bfdd5d619d589e5c3d389affbab514ec3d36fe1e21b42d6e09b059e98d7202a",
	// 	"s": "0x7d3c7a5f0325a72cc17ff5b7d436a6d562f27ff1608bc1df60c7166c81a4a948",

	hash, err := client.SendTransaction(context.Background(), message, signature, common.HexToAddress("0x97549e368acafdcae786bb93d98379f1d1561a29"), 0)
	fmt.Println(hash.String(), err)

}
