package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/newtonproject/newchain-api-express/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var log *logrus.Logger

// CLI represents a command-line interface. This class is
// not threadsafe.
type CLI struct {
	name       string
	rootCmd    *cobra.Command
	version    string
	walletPath string
	rpcURL     string
	config     string
	testing    bool
	logfile    string

	blockchain BlockChain
}

// NewCLI returns an initialized CLI
func NewCLI() *CLI {
	bc, err := getBlockChain()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	log = logrus.New()
	log.SetOutput(os.Stdout)

	cli := &CLI{
		name:       "NewChainAPIExpress",
		rootCmd:    nil,
		version:    utils.Version(),
		walletPath: "",
		rpcURL:     "",
		testing:    false,
		config:     "",
		logfile:    "./error.log",

		blockchain: bc,
	}

	cli.buildRootCmd()
	return cli
}

// CopyCLI returns an copy  CLI
func CopyCLI(cli *CLI) *CLI {
	cpy := &CLI{
		rootCmd:    nil,
		version:    cli.version,
		walletPath: cli.walletPath,
		rpcURL:     cli.rpcURL,
		testing:    false,
		config:     cli.config,
		logfile:    cli.logfile,
	}

	cpy.buildRootCmd()
	return cpy
}

// Execute parses the command line and processes it.
func (cli *CLI) Execute() {
	cli.rootCmd.Execute()
}

// setup turns up the CLI environment, and gets called by Cobra before
// a command is executed.
func (cli *CLI) setup(cmd *cobra.Command, args []string) {
	if err := cli.setupConfig(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func (cli *CLI) help(cmd *cobra.Command, args []string) {
	fmt.Fprint(os.Stderr, cmd.UsageString())

	os.Exit(-1)

}

// TestCommand test command
func (cli *CLI) TestCommand(command string) string {
	cli.testing = true
	result := cli.Run(strings.Fields(command)...)
	cli.testing = false
	return result
}

// Run executes CLI with the given arguments. Used for testing. Not thread safe.
func (cli *CLI) Run(args ...string) string {
	oldStdout := os.Stdout

	r, w, _ := os.Pipe()

	os.Stdout = w

	cli.rootCmd.SetArgs(args)
	cli.rootCmd.Execute()
	cli.buildRootCmd()

	w.Close()

	os.Stdout = oldStdout

	var stdOut bytes.Buffer
	io.Copy(&stdOut, r)
	return stdOut.String()
}

// Embeddable returns a CLI that you can embed into your own Go programs. This
// is not thread-safe.
func (cli *CLI) Embeddable() *CLI {
	cli.testing = true
	return cli
}
