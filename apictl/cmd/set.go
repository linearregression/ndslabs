// Copyright © 2016 National Data Service

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// listCmd represents the list command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set optional stack values",
}

func init() {
	RootCmd.AddCommand(setCmd)
	setCmd.AddCommand(setEnvCmd)
}

var setEnvCmd = &cobra.Command{
	Use:    "env [stack service id] [var name] [var value]",
	Short:  "Set stack service environment values",
	PreRun: Connect,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 3 {
			cmd.Usage()
			os.Exit(-1)
		}
		ssid := args[0]
		varName := args[1]
		varValue := args[2]

		if strings.Index(ssid, "-") <= 0 {
			fmt.Printf("Invalid stack service id (looks like a stack Id?): %s\n", ssid)
			return
		}

		sid := ssid[0:strings.Index(ssid, "-")]
		stack, err := client.GetStack(sid)
		if err != nil {
			fmt.Printf("Get stack failed: %s\n", err)
			return
		}

		ssidFound := false
		for i, stackService := range stack.Services {
			if stackService.Id == ssid {
				spec, err := client.GetService(stackService.Service)
				if err != nil {
					fmt.Printf("Error getting service spec %s\n", err.Error)
				}
				if stackService.Config == nil {
					stackService.Config = make(map[string]string)
				}
				found := false
				for _, config := range spec.Config {
					if config.Name == varName {
						if config.CanOverride {
							fmt.Printf("%s %s %t\n", varName, varValue, config.CanOverride)
							stackService.Config[varName] = varValue
							found = true
						} else {
							fmt.Printf("Cannot override variable %s\n", varName)
							return
						}
					}
				}
				if !found {
					// Custom variable
					stackService.Config[varName] = varValue
				}
				stack.Services[i] = stackService
				ssidFound = true
			}
		}
		if !ssidFound {
			fmt.Printf("No such stack service id %s\n", ssid)
		}
		err = client.UpdateStack(stack)
		if err != nil {
			fmt.Printf("Error updating stack: %s\n", err)
			return
		}
		if Verbose {
			data, err := json.MarshalIndent(stack, "", "   ")
			if err != nil {
				fmt.Printf("Error marshalling stack %s\n", err.Error)
				return
			}

			fmt.Println(string(data))
		}

	},
	PostRun: RefreshToken,
}
