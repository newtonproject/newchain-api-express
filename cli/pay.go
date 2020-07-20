package cli

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/newtonproject/newchain-api-express/newtonclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (cli *CLI) buildPayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "pay <to> <amount> <--from address>",
		Short:                 "Pay to address with amount",
		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if !common.IsHexAddress(args[0]) {
				fmt.Println("To address error")
				return
			}
			to := common.HexToAddress(args[0])

			amount, err := getAmountWei(args[1], UnitETH)
			if err != nil {
				fmt.Println("Amount error: ", err)
				return
			}

			fromStr, err := cmd.Flags().GetString("from")
			if err != nil {
				fmt.Println(err)
				return
			}
			if !common.IsHexAddress(fromStr) {
				fmt.Println("From address error")
				return
			}
			from := common.HexToAddress(fromStr)

			wait := int64(1)
			if cmd.Flags().Changed("wait") {
				wait, err = cmd.Flags().GetInt64("wait")
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			nonce := uint64(viper.GetInt64(fmt.Sprintf("Client.%s.NoncePending", from.String())))
			gasPirce, ok := big.NewInt(0).SetString(viper.GetString("Client.GasPrice"), 10)
			if !ok {
				fmt.Println("Get gas price from config error")
				return
			}
			gasLimit := uint64(viper.GetInt64("Client.GasLimit"))
			chainId, ok := big.NewInt(0).SetString(viper.GetString("Client.ChainID"), 10)
			if !ok {
				fmt.Println("Get chainID from config error")
				return
			}

			tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPirce, nil)
			message, err := rlp.EncodeToBytes(tx)
			if err != nil {
				fmt.Println(err)
				return
			}
			signer := types.NewEIP155Signer(chainId)
			messageHash := signer.Hash(tx)

			fmt.Println("The tx is as follow: ")
			fmt.Println("To: ", to.String())
			fmt.Println("Amount: ", getWeiAmountTextByUnit(amount, UnitETH))

			wallet := keystore.NewKeyStore(cli.walletPath, keystore.LightScryptN, keystore.LightScryptP)

			prompt := fmt.Sprintf("Unlocking account %s to sign tx", from.String())
			walletPassword, err := getPassPhrase(prompt, false)

			err = wallet.Unlock(accounts.Account{Address: from}, walletPassword)
			if err != nil {
				fmt.Println(err)
				return
			}

			signature, err := wallet.SignHash(accounts.Account{Address: from}, messageHash.Bytes())
			if err != nil {
				fmt.Println(err)
				return
			}
			if len(signature) != 65 {
				fmt.Println("signature len error")
				return
			}

			rpcurl := viper.GetString("Client.RPCUrl")

			client, err := newtonclient.Dial(rpcurl)
			if err != nil {
				fmt.Println(err)
				return
			}

			ctx := context.Background()

			hash, err := client.SendTransaction(ctx, message, signature[:64], from, uint64(wait))
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Hash: ", hash.String())

			// ok, update nonce
			viper.Set(fmt.Sprintf("Client.%s.Nonce", from.String()), nonce+1)
			err = viper.WriteConfigAs(cli.config)
			if err != nil {
				fmt.Println("WriteConfig:", err)
				return
			}
			fmt.Println("Update nonce to config: ", cli.config)

		},
	}

	cmd.Flags().String("from", "", "the from address")
	cmd.Flags().Int64("wait", 1, "the wait level(0,1,2)")

	return cmd
}
