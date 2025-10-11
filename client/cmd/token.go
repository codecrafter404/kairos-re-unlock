/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/codecrafter404/go-nodepair"
	"github.com/spf13/cobra"
)

// token represents the vpnToken command
var token = &cobra.Command{
	Use:   "token",
	Short: "Generate an EdgeVPN token",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(nodepair.GenerateToken())
	},
}

func init() {
	rootCmd.AddCommand(token)
}
