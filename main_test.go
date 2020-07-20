package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/newtonproject/newchain-api-express/cli"
)

func getTempFile() (string, func()) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}

	file := dir + string(os.PathSeparator) + "lumen-integration-test.json"

	return file, func() {
		logrus.Debugf("cleaning up temp file: %s", file)
		os.RemoveAll(dir)
	}
}

func run(cli *cli.CLI, command string) string {
	fmt.Printf("$ ./NewExchange %s\n", command)
	got := cli.TestCommand(command)
	fmt.Printf("%s\n", got)
	return strings.TrimSpace(got)
}

func runArgs(cli *cli.CLI, args ...string) string {
	fmt.Printf("$ ./NewExchange %s\n", strings.Join(args, " "))
	got := cli.Embeddable().Run(args...)
	fmt.Printf("%s\n", got)
	return strings.TrimSpace(got)
}

func expectOutput(t *testing.T, cli *cli.CLI, want string, command string) {
	got := run(cli, command)

	if got != want {
		t.Errorf("(%s) wrong output: want %v, got %v", command, want, got)
	}
}

func newCLI() (*cli.CLI, func()) {
	_, cleanupFunc := getTempFile()

	glumen := cli.NewCLI()
	glumen.TestCommand("version")
	run(glumen, fmt.Sprintf("version"))

	return glumen, cleanupFunc
}
