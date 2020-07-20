package cli

import (
	"github.com/spf13/cobra"
)

func (cli *CLI) buildRootCmd() {
	if cli.rootCmd != nil {
		cli.rootCmd.ResetFlags()
		cli.rootCmd.ResetCommands()
	}

	rootCmd := &cobra.Command{
		Use:              cli.name,
		Short:            cli.name + " is a commandline example for exchange",
		Run:              cli.help,
		PersistentPreRun: cli.setup,
	}
	cli.rootCmd = rootCmd

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cli.config, "config", "c", defaultConfigFile, "the `path` to config file")
	rootCmd.PersistentFlags().StringP("rpcURL", "i", defaultRPCURL, "NewChain json rpc or ipc `url`")
	rootCmd.PersistentFlags().StringP("host", "H", "127.0.0.1:8888", "the `host` of the server, [bind_address]:port")

	// Basic commands
	rootCmd.AddCommand(cli.buildVersionCmd()) // version

	// server
	rootCmd.AddCommand(cli.buildServerCmd()) // NewChainAPIExpress server

	// client
	rootCmd.AddCommand(cli.buildAccountCmd()) // account
	rootCmd.AddCommand(cli.buildPayCmd())     // pay
	rootCmd.AddCommand(cli.buildInfoCmd())    // info

}
