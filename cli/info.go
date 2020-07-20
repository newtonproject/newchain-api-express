package cli

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/newtonproject/newchain-api-express/newtonclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (cli *CLI) buildInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "info <hexaddress> [--update]",
		Short:                 "Get base info from API",
		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !common.IsHexAddress(args[0]) {
				fmt.Println("invalid hex address")
				return
			}
			address := common.HexToAddress(args[0])

			rpcurl := viper.GetString("Client.RPCUrl")
			if rpcurl == "" {
				rpcurl = cli.rpcURL
			}

			client, err := newtonclient.Dial(rpcurl)
			if err != nil {
				fmt.Println(err)
				return
			}

			info, err := client.GetBaseInfo(context.Background(), address)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println("The base info is as follow: ")
			fmt.Println("Address: ", address.String())
			fmt.Println("NonceLatest: ", info.NonceLatest)
			fmt.Println("NoncePending: ", info.NoncePending)
			fmt.Println("Balance: ", getWeiAmountTextByUnit(info.Balance, UnitETH))
			fmt.Println("GasPrice: ", getWeiAmountTextByUnit(info.GasPrice, UnitETH))
			fmt.Println("ChainID: ", info.NetworkID)

			update, _ := cmd.Flags().GetBool("update")
			if update {
				viper.Set("Client.ChainID", info.NetworkID)
				viper.Set("Client.GasPrice", info.GasPrice.String())
				if viper.GetInt64("Client.GasLimit") == 0 {
					viper.Set("Client.GasLimit", 21000)
				}
				if viper.GetString("Client.RPCURL") == "" {
					viper.Set("Client.RPCURL", rpcurl)
				}
				viper.Set(fmt.Sprintf("Client.%s.NonceLatest", address.String()), info.NonceLatest)
				viper.Set(fmt.Sprintf("Client.%s.NoncePending", address.String()), info.NoncePending)
				viper.Set(fmt.Sprintf("Client.%s.Balance", address.String()), info.Balance.String())

				err = viper.WriteConfigAs(cli.config)
				if err != nil {
					fmt.Println("WriteConfig:", err)
					return
				}
				fmt.Println("Update info to config: ", cli.config)
			}

		},
	}

	cmd.Flags().BoolP("update", "u", false, "update local info to config")

	return cmd
}
