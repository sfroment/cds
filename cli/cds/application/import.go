package application

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ovh/cds/cli"
	"github.com/ovh/cds/sdk"
	"github.com/ovh/cds/sdk/exportentities"
	"github.com/spf13/cobra"
)

var (
	importFormat, importURL string
	importForce             bool
)

func importCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "cds application import <projectKey> [file] [--url <url> --format json|yaml] [--force]",
		Long:  "See documentation on https://github.com/ovh/cds/tree/master/doc/tutorials",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				sdk.Exit("Wrong usage: see %s\n", cmd.Short)
			}

			projectKey := args[0]
			msg := []string{}
			btes := []byte{}

			if len(args) == 2 {
				name := args[1]
				importFormat = "yaml"
				if strings.HasSuffix(name, ".json") {
					importFormat = "json"
				} else if strings.HasSuffix(name, ".hcl") {
					importFormat = "hcl"
				}
				var err error
				btes, _, err = exportentities.ReadFile(name)
				if err != nil {
					sdk.Exit("Error: %s\n", err)
				}
			} else if importURL != "" {
				var err error
				btes, _, err = exportentities.ReadURL(importURL, importFormat)
				if err != nil {
					sdk.Exit("Error: %s\n", err)
				}
			} else {
				sdk.Exit("Wrong usage: see %s\n", cmd.Short)
			}

			var url string
			url = fmt.Sprintf("/project/%s/application/import?format=%s", projectKey, importFormat)

			if importForce {
				url += "&forceUpdate=true"
			}

			data, code, err := sdk.Request("POST", url, btes)
			if sdk.ErrorIs(err, sdk.ErrPipelineAlreadyExists) {
				fmt.Print("Pipline already exists. ")
				if cli.AskForConfirmation("Do you want to override ?") {
					url = fmt.Sprintf("/project/%s/application/import?format=%s&forceUpdate=true", projectKey, importFormat)
					data, code, err = sdk.Request("POST", url, btes)
				} else {
					sdk.Exit("Aborted\n")
				}
			}

			if code > 400 {
				sdk.Exit("Error: %d - %s\n", code, err)
			}
			if err != nil {
				sdk.Exit("Error: %s\n", err)
			}

			if err := json.Unmarshal(data, &msg); err != nil {
				sdk.Exit("Error: %s\n", err)
			}

			for _, s := range msg {
				fmt.Println(s)
			}

			if code == 400 {
				sdk.Exit("Error while importing application\n")
			}
		},
	}

	cmd.Flags().StringVarP(&importURL, "url", "", "", "Import application from an URL")
	cmd.Flags().StringVarP(&importFormat, "format", "", "yaml", "Configuration file format")
	cmd.Flags().BoolVarP(&importForce, "force", "", false, "Use force flag to update your application")

	return cmd
}
