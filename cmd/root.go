package cmd

import (
	"fmt"
	"os"

	"github.com/aneurysm9/challenger/vm"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "challenger",
	Short: "A Synacor Challenge VM",
	Run: func(cmd *cobra.Command, args []string) {
		machine, err := vm.LoadImage("challenge.bin")
		if err != nil {
			fmt.Printf("Error loading image: %s", err)
			os.Exit(1)
		}
		machine.Run()
	},
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// Do stuff here
}

// Execute the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
