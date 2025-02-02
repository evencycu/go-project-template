package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"gitlab.com/cake/gopkg"
)

func NewVersionCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Hugo",
		Long:  `All software has versions. This is Hugo's`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%+v\n", gopkg.GetVersion())
		},
	}
	return c
}
