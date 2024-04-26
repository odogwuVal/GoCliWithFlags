/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/PacktPublishing/Go-for-DevOps/chapter/6/grpc/client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Fetches a quote of the day from the QOTD server",
	Long: `This command allows you to fetch a quote of the day from our
QOTD server we designed in our chapter on gRPC. This command defaults to a
production server (which doesn't exist). This can be changed to the devlopement
server (which doesn't exist) using --dev or to a specific address with --addr .

Example usage for a random author:
qotd get

Example usage for a specific author:
qotd get --author="mark twain"

Example usage using a 127.0.0.1 for the server:
qotd get -addr=127.0.0.1:80 -author="mark twain"
`,
	Run: func(cmd *cobra.Command, args []string) {
		const devAddr = "127.0.0.1:3450"
		fs := cmd.Flags()
		addr := mustString(fs, "addr")

		if mustBool(fs, "dev") {
			addr = devAddr
		}

		c, err := client.New(addr)
		if err != nil {
			fmt.Println("error: ", err)
			os.Exit(1)
		}

		a, q, err := c.QOTD(cmd.Context(), mustString(fs, "author"))
		if err != nil {
			fmt.Println("error: ", err)
			os.Exit(1)
		}

		switch {
		case mustBool(fs, "json"):
			b, err := json.Marshal(
				struct {
					Author string
					Quote  string
				}{a, q},
			)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s\n", b)
		default:
			fmt.Println("Author: ", a)
			fmt.Println("Quote: ", q)
		}
	},
}

func mustString(fs *pflag.FlagSet, name string) string {
	v, err := fs.GetString(name)
	if err != nil {
		panic(err)
	}
	return v
}

func mustBool(fs *pflag.FlagSet, name string) bool {
	v, err := fs.GetBool(name)
	if err != nil {
		panic(err)
	}
	return v
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// **************************************************************************************
	// Adds a flag called --dev that can be shortened to -d and defaults to false
	// Adds a flag called --addr that defaults to "127.0.0.1:80"
	// Adds a flag called --author that can be shortened to -a
	// Adds a flag called --json that defaults to false
	getCmd.Flags().BoolP("dev", "d", false, "Uses the dev server instead of prod")
	getCmd.Flags().String("addr", "127.0.0.1:80", "Set the QOTD server to use, defaults to production")
	getCmd.Flags().StringP("author", "a", "", "Specify the author to get a quote for")
	getCmd.Flags().Bool("json", false, "Output is in JSON format")
}
