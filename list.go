package main

import (
	"fmt"
	"log"

	"github.com/Loyalsoldier/domain-list-custom/lib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().StringP("config", "c", "config.json", "URI of the JSON format config file")
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List available domain lists",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		
		instance, err := lib.NewInstance()
		if err != nil {
			log.Fatal(err)
		}

		if err := instance.InitConfig(configFile); err != nil {
			log.Fatal(err)
		}

		// Process only input to get the list
		for idx, inputConfig := range instance.Config.Input {
			log.Printf("Processing input [%d/%d]: type=%s, action=%s", 
				idx+1, len(instance.Config.Input), inputConfig.Type, inputConfig.Action)
			
			converter, err := inputConfig.GetInputConverter()
			if err != nil {
				log.Fatal(err)
			}

			newContainer, err := converter.Input(instance.Container)
			if err != nil {
				log.Fatal(err)
			}

			if newContainer != nil {
				instance.Container = newContainer
			}
		}

		// List all entries
		fmt.Printf("\nAvailable domain lists (%d total):\n", instance.Container.Len())
		fmt.Println("---")
		
		names := instance.Container.GetNames()
		for _, name := range names {
			entry, found := instance.Container.GetEntry(name)
			if !found {
				continue
			}
			fmt.Printf("  - %s (%d domains)\n", name, len(entry.GetDomains()))
		}
	},
}
