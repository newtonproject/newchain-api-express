package cli

import (
	"fmt"
	"net/http"

	"github.com/newtonproject/newchain-api-express/api"
	"github.com/newtonproject/newchain-api-express/rpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (cli *CLI) buildServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "api",
		Short:                 "Run as NewChain Express API server",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			viper.BindPFlag("rpcURL", cli.rootCmd.PersistentFlags().Lookup("rpcURL"))
			viper.BindPFlag("host", cli.rootCmd.PersistentFlags().Lookup("host"))
			viper.SetDefault("rpcURL", defaultRPCURL)
			if rpcURL := viper.GetString("rpcURL"); rpcURL != "" {
				cli.rpcURL = viper.GetString("rpcURL")
			}

			hostAddress := viper.GetString("host")
			if hostAddress == "" {
				hostAddress = "127.0.0.1:8888"
			}

			log.Printf("Listening at %v...", hostAddress)

			notify, err := loadNotifyConfig()
			if err != nil {
				log.Println(err)
				return
			}

			s, err := api.NewServer(cli.rpcURL, notify)
			if err != nil {
				log.Println(err)
				return
			}

			rpcServer := rpc.NewServer()
			err = rpcServer.RegisterName("newton", s)
			if err != nil {
				log.Println(err)
				return
			}

			err = http.ListenAndServe(hostAddress, rpcServer)
			if err != nil {
				log.Println(err)
				return
			}

			return
		},
	}

	return cmd
}

func loadNotifyConfig() (*api.NotifyConfig, error) {
	p := "Notify"

	server := viper.GetString(p + ".Server")
	if server == "" {
		return nil, fmt.Errorf("%s server is empty", p)
	}
	username := viper.GetString(p + ".Username")
	if username == "" {
		return nil, fmt.Errorf("%s username is empty", p)
	}
	password := viper.GetString(p + ".Password")
	if password == "" {
		return nil, fmt.Errorf("%s password is empty", p)
	}
	clientID := viper.GetString(p + ".ClientID")
	if clientID == "" {
		clientID = fmt.Sprintf("NewChainAPIExpress")
	}
	qos := viper.GetInt(p + ".QoS")
	if !(qos == 0 || qos == 1 || qos == 2) {
		return nil, fmt.Errorf("%s QoS only 0,1,2", p)
	}

	prefixTopic := viper.GetString(p + ".PrefixTopic")

	return &api.NotifyConfig{
		Server:      server,
		Username:    username,
		Password:    password,
		ClientID:    clientID,
		QoS:         byte(qos),
		PrefixTopic: prefixTopic,
	}, nil
}
