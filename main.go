package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var greeting string

	rootCmd := &cobra.Command{
		Use:   "helloworld [name]",
		Short: "Print a friendly greeting",
		Long:  "helloworld prints a greeting for the given name, defaulting to \"world\".",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "world"
			if len(args) > 0 {
				name = args[0]
			}
			fmt.Printf("%s %s\n", greeting, name)
			return nil
		},
	}

	rootCmd.Flags().StringVarP(&greeting, "greeting", "g", "hello", "greeting to use")

	goodbyeCmd := &cobra.Command{
		Use:   "goodbye [name]",
		Short: "Print a friendly farewell",
		Long:  "goodbye prints a farewell for the given name, defaulting to \"world\".",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "world"
			if len(args) > 0 {
				name = args[0]
			}
			fmt.Printf("goodbye %s\n", name)
			return nil
		},
	}

	rootCmd.AddCommand(goodbyeCmd)

	hithereCmd := &cobra.Command{
		Use:   "hithere [name]",
		Short: "Print a casual hello",
		Long:  "hithere prints a casual greeting for the given name, defaulting to \"world\".",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "world"
			if len(args) > 0 {
				name = args[0]
			}
			fmt.Printf("hi there %s\n", name)
			return nil
		},
	}

	rootCmd.AddCommand(hithereCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
