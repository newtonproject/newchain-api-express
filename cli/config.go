package cli

import (
	"os"

	"github.com/spf13/viper"
)

const defaultConfigFile = "./config.toml"
const defaultWalletPath = "./wallet"

func (cli *CLI) defaultConfig() {

	viper.SetDefault("walletPath", defaultWalletPath)
}

func (cli *CLI) setupConfig() error {

	// var ret bool
	var err error

	cli.defaultConfig()

	viper.SetConfigName(defaultConfigFile)
	viper.AddConfigPath(".")
	cfgFile := cli.config
	if cfgFile != "" {
		if _, err = os.Stat(cfgFile); err == nil {
			viper.SetConfigFile(cfgFile)
			err = viper.ReadInConfig()
		} else if cfgFile != defaultConfigFile {
			return err
		}
	} else {
		// The default configuration is enabled.
		err = nil
	}

	if rpcURL := viper.GetString("rpcURL"); rpcURL != "" {
		cli.rpcURL = viper.GetString("rpcURL")
	}
	if walletPath := viper.GetString("walletPath"); walletPath != "" {
		cli.walletPath = viper.GetString("walletPath")
	}
	if log := viper.GetString("log"); log != "" {
		cli.logfile = viper.GetString("log")
	}

	return nil
}
