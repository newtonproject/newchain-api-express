package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/newtonproject/newchain-api-express/newtonclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (cli *CLI) buildAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [new|list|balance]",
		Short: fmt.Sprintf("Manage %s accounts", cli.blockchain.String()),
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			return
		},
	}

	cmd.AddCommand(cli.buildAccountNewCmd())
	cmd.AddCommand(cli.buildAccountBalanceCmd())
	cmd.AddCommand(cli.buildAccountListCmd())
	cmd.AddCommand(cli.buildAccountUpdateCmd())
	cmd.AddCommand(cli.buildAccountImportCmd())
	cmd.AddCommand(cli.buildAccountExportCmd())

	if cli.blockchain == NewChain {
		cmd.AddCommand(cli.buildAccountConvertCmd())
	}

	return cmd
}

func (cli *CLI) buildAccountNewCmd() *cobra.Command {
	accountNewCmd := &cobra.Command{
		Use:                   "new [-n number] [--faucet] [-s] [-l]",
		Short:                 "create a new account",
		Args:                  cobra.MinimumNArgs(0),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			var wallet *keystore.KeyStore
			if cmd.Flags().Changed("light") {
				light, _ := cmd.Flags().GetBool("light")
				standard, _ := cmd.Flags().GetBool("standard")
				if light && !standard {
					wallet = keystore.NewKeyStore(cli.walletPath,
						keystore.LightScryptN, keystore.LightScryptP)
				}
			}
			if wallet == nil {
				wallet = keystore.NewKeyStore(cli.walletPath,
					keystore.StandardScryptN, keystore.StandardScryptP)
			}

			numOfNew, err := cmd.Flags().GetInt("numOfNew")
			if err != nil {
				numOfNew = viper.GetInt("account.numOfNew")
			}

			faucet, _ := cmd.Flags().GetBool("faucet")

			aList, err := cli.createAccount(wallet, numOfNew)
			if err != nil {
				fmt.Println(err)
				return
			}

			for _, a := range aList {
				fmt.Println(a.String())
				if faucet {
					getFaucet(cli.rpcURL, a.String())
				}
			}

		},
	}

	accountNewCmd.Flags().IntP("numOfNew", "n", 1, "number of the new account")
	accountNewCmd.Flags().Bool("faucet", false, "get faucet for new account")
	accountNewCmd.Flags().BoolP("standard", "s", false, "use the standard scrypt for keystore")
	accountNewCmd.Flags().BoolP("light", "l", false, "use the light scrypt for keystore")
	return accountNewCmd
}

func (cli *CLI) createAccount(wallet *keystore.KeyStore, numOfNew int) ([]common.Address, error) {
	walletPassword, err := getPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.", true)
	if err != nil {
		return nil, err
	}

	if numOfNew <= 0 {
		fmt.Printf("number[%d] of new account less then 1\n", numOfNew)
		numOfNew = 1
	}

	aList := make([]common.Address, 0)
	for i := 0; i < numOfNew; i++ {
		account, err := wallet.NewAccount(walletPassword)
		if err != nil {
			return nil, fmt.Errorf("account error: %v", err)
		}

		aList = append(aList, account.Address)
	}

	return aList, nil
}

func (cli *CLI) buildAccountListCmd() *cobra.Command {
	accountListCmd := &cobra.Command{
		Use:                   "list",
		Short:                 "list all accounts in the wallet path",
		Args:                  cobra.MinimumNArgs(0),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			wallet := keystore.NewKeyStore(cli.walletPath,
				keystore.LightScryptN, keystore.LightScryptP)

			if len(wallet.Accounts()) == 0 {
				fmt.Println("empty wallet, create account first")
				return
			}

			for _, account := range wallet.Accounts() {
				fmt.Println(account.Address.Hex())
			}
		},
	}

	return accountListCmd
}

func (cli *CLI) buildAccountConvertCmd() *cobra.Command {
	accountListCmd := &cobra.Command{
		Use:                   "convert",
		Short:                 "convert address to NewChainAddress",
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Run: func(cmd *cobra.Command, args []string) {

			client, err := ethclient.Dial(cli.rpcURL)
			if err != nil {
				fmt.Printf("Error: build client error(%v)\n", err)
				return
			}

			chainID, err := client.NetworkID(context.Background())
			if err != nil {
				fmt.Printf("Error: get chainID error(%v)\n", err)
				return
			}

			for _, addressStr := range args {
				if common.IsHexAddress(addressStr) {
					address := common.HexToAddress(addressStr)
					fmt.Println(address.String(), addressToNew(chainID.Bytes(), address))
					continue
				}

				address, err := newToAddress(chainID.Bytes(), addressStr)
				if err != nil {
					fmt.Println(err, addressStr)
					continue
				}
				fmt.Println(address.String(), addressStr)
			}

		},
	}

	return accountListCmd
}

func (cli *CLI) buildAccountUpdateCmd() *cobra.Command {
	accountNewCmd := &cobra.Command{
		Use:                   "update <address> [-s]",
		Short:                 "Update an existing account",
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Run: func(cmd *cobra.Command, args []string) {
			var wallet *keystore.KeyStore
			standard, _ := cmd.Flags().GetBool("standard")
			if standard {
				wallet = keystore.NewKeyStore(cli.walletPath,
					keystore.StandardScryptN, keystore.StandardScryptP)
			} else {
				wallet = keystore.NewKeyStore(cli.walletPath,
					keystore.LightScryptN, keystore.LightScryptP)
			}

			addrStr := args[0]
			var address common.Address
			if !common.IsHexAddress(addrStr) {
				fmt.Println("Error: No accounts specified to update")
				return
			}
			address = common.HexToAddress(addrStr)
			account := accounts.Account{Address: address}

			if account.Address == (common.Address{}) {
				fmt.Println("Error: ", errRequiredFromAddress)
				return
			}
			if _, err := wallet.Find(account); err != nil {
				fmt.Printf("Error: %v (%s)\n", err, account.Address.String())
				return
			}

			var err error
			walletPassword := ""
			var trials int
			for trials = 0; trials < 3; trials++ {
				prompt := fmt.Sprintf("Unlocking account %s | Attempt %d/%d", account.Address.String(), trials+1, 3)
				if walletPassword == "" {
					walletPassword, _ = getPassPhrase(prompt, false)
				} else {
					fmt.Println(prompt, "\nUse the the password has set")
				}
				err = wallet.Unlock(account, walletPassword)
				if err == nil {
					break
				}
				walletPassword = ""
			}

			if trials >= 3 {
				if err != nil {
					fmt.Println("Error: unlock account error: ", err)
					return
				}
				fmt.Printf("Error: failed to unlock account %s (%v)\n", account.Address.String(), err)
				return
			}

			newWalletPassword, err := getPassPhrase("Please give a new password. Do not forget this password.", true)
			if err != nil {
				fmt.Println("Error: getPassPhrase error: ", err)
				return
			}

			if err := wallet.Update(account, walletPassword, newWalletPassword); err != nil {
				fmt.Println("Error: udpate account error: ", err)
				return
			}

			fmt.Println("Successfully updated the account: ", account.Address.String())
		},
	}

	accountNewCmd.Flags().BoolP("standard", "s", false, "use the standard scrypt for keystore")
	return accountNewCmd
}

func addressToNew(chainID []byte, address common.Address) string {
	input := append(chainID, address.Bytes()...)
	return "NEW" + base58.CheckEncode(input, 0)
}

func newToAddress(chainID []byte, newAddress string) (common.Address, error) {
	if newAddress[:3] != "NEW" {
		return common.Address{}, errors.New("not NEW address")
	}

	decoded, version, err := base58.CheckDecode(newAddress[3:])
	if err != nil {
		return common.Address{}, err
	}
	if version != 0 {
		return common.Address{}, errors.New("illegal version")
	}
	if len(decoded) < 20 {
		return common.Address{}, errors.New("illegal decoded length")
	}
	if !bytes.Equal(decoded[:len(decoded)-20], chainID) {
		return common.Address{}, errors.New("illegal ChainID")
	}

	address := common.BytesToAddress(decoded[len(decoded)-20:])

	return address, nil
}

func (cli *CLI) buildAccountImportCmd() *cobra.Command {
	accountListCmd := &cobra.Command{
		Use:                   "import",
		Short:                 "import hex private key to wallet",
		Args:                  cobra.MinimumNArgs(0),
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Run: func(cmd *cobra.Command, args []string) {
			hexkey, err := console.Stdin.PromptPassword("Enter private key: ")
			if err != nil {
				fmt.Println(err)
				return
			}
			hexkeylen := len(hexkey)
			if hexkeylen > 1 {
				if hexkey[0:2] == "0x" || hexkey[0:2] == "0X" {
					hexkey = hexkey[2:]
				}
			}

			pkey, err := crypto.HexToECDSA(hexkey)
			if err != nil {
				fmt.Println(err)
				return
			}

			wallet := keystore.NewKeyStore(cli.walletPath,
				keystore.LightScryptN, keystore.LightScryptP)

			walletPassword, err := getPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.", true)
			if err != nil {
				fmt.Println("Error: ", err)
				return
			}

			a, err := wallet.ImportECDSA(pkey, walletPassword)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(a.Address.String())

		},
	}

	return accountListCmd
}

func (cli *CLI) buildAccountExportCmd() *cobra.Command {
	accountListCmd := &cobra.Command{
		Use:                   "export <hexAddress>",
		Short:                 "export hex private key of the specified address",
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Run: func(cmd *cobra.Command, args []string) {

			hexAddress := args[0]
			if !common.IsHexAddress(hexAddress) {
				fmt.Printf("Error: %s is not valid hex-encoded address\n", hexAddress)
				return
			}

			address := common.HexToAddress(hexAddress)
			account := accounts.Account{Address: address}

			wallet := keystore.NewKeyStore(cli.walletPath,
				keystore.LightScryptN, keystore.LightScryptP)

			if !wallet.HasAddress(address) {
				fmt.Println("The given address is not present")
				return
			}

			var err error

			prompt := fmt.Sprintf("Unlocking account %s", account.Address.String())
			walletPassword, _ := getPassPhrase(prompt, false)
			keyJSON, err := wallet.Export(account, walletPassword, walletPassword)
			if err != nil {
				fmt.Println(err)
				return
			}

			key, err := keystore.DecryptKey(keyJSON, walletPassword)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(common.ToHex(key.PrivateKey.D.Bytes()))

		},
	}

	return accountListCmd
}

func (cli *CLI) buildAccountBalanceCmd() *cobra.Command {
	accountBalanceCmd := &cobra.Command{
		Use:                   "balance [hexAddress]",
		Short:                 "get balance of address",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			cli.showBalance(cmd, args, true)

		},
	}

	return accountBalanceCmd
}

func (cli *CLI) showBalance(cmd *cobra.Command, args []string, showSum bool) {
	var err error

	unit, _ := cmd.Flags().GetString("unit")
	if unit != "" && !stringInSlice(unit, UnitList) {
		fmt.Printf("Unit(%s) for invalid. %s.\n", unit, UnitString)
		fmt.Fprint(os.Stderr, cmd.UsageString())
		return
	}

	var addressList []common.Address

	if len(args) <= 0 {
		wallet := keystore.NewKeyStore(cli.walletPath,
			keystore.LightScryptN, keystore.LightScryptP)

		for _, account := range wallet.Accounts() {
			addressList = append(addressList, account.Address)
		}

	} else {
		for _, addressStr := range args {
			addressList = append(addressList, common.HexToAddress(addressStr))
		}
	}

	client, err := newtonclient.Dial(viper.GetString("Client.RPCUrl"))
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx := context.Background()

	balanceSum := big.NewInt(0)
	for _, address := range addressList {
		info, err := client.GetBaseInfo(ctx, address)
		if err != nil {
			fmt.Println("GetBaseInfo error:", err)
			return
		}
		balance := info.Balance

		balanceSum.Add(balanceSum, balance)

		fmt.Printf("Address[%s] Balance[%s]\n", address.Hex(), getWeiAmountTextUnitByUnit(balance, unit))
	}

	if showSum {
		fmt.Println("Number Of Accounts:", len(addressList))
		fmt.Println("Total Balance:", getWeiAmountTextUnitByUnit(balanceSum, unit))
	}

	return
}
