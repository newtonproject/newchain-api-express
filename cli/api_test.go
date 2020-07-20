package cli

import "testing"

func TestForce(t *testing.T) {
	cli := NewCLI()

	cli.TestCommand("escrow")
}
