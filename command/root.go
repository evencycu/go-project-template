package command

import (
	"github.com/spf13/cobra"
	"gitlab.com/cake/goctx"
)

var systemCtx goctx.Context

func init() {
	systemCtx = goctx.Background()
	systemCtx.Set(goctx.LogKeyCID, "system init")
}

var rootCmd = &cobra.Command{
	Use:   "go-project-template",
	Short: "go-project-template is a project template for golang backend server development",
	Long: `A Fast and Flexible Static Site Generator built with
				  love by spf13 and friends in Go.
				  Complete documentation is available at http://hugo.spf13.com`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() error {
	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(NewAPICmd())
	rootCmd.AddCommand(NewServerCmd())
	return rootCmd.Execute()
}
