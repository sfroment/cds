package main

import (
	"reflect"

	"github.com/spf13/cobra"

	"github.com/ovh/cds/cli"
)

var (
	tokenCmd = cli.Command{
		Name:  "token",
		Short: "Manage CDS token",
	}

	token = cli.NewCommand(tokenCmd, nil,
		[]*cobra.Command{
			cli.NewListCommand(tokenListCmd, tokenListRun, nil),
			cli.NewListCommand(tokenGenerateCmd, tokenGenerateRun, nil),
		})
)

var tokenListCmd = cli.Command{
	Name:  "list",
	Short: "List CDS Tokens",
	Args:  []cli.Arg{},
	Flags: []cli.Flag{
		{
			Name:      "group",
			Kind:      reflect.String,
			ShortHand: "g",
			Usage:     "filter token on a group",
		},
	},
}

func tokenListRun(v cli.Values) (cli.ListResult, error) {
	return nil, nil
}

var tokenGenerateCmd = cli.Command{
	Name:  "generate",
	Short: "Generate CDS Token",
	Args: []cli.Arg{
		{
			Name:   "group",
			Weight: 0,
		},
		{
			Name:   "expiration",
			Weight: 1,
		},
	},
}

func tokenGenerateRun(v cli.Values) (cli.ListResult, error) {
	return nil, nil
}
