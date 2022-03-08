package command

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/cake/go-project-template/apiserver"
	"gitlab.com/cake/golibs/intercom"
)

func NewAPICmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "api",
		Short: "Print the gin api list",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			viper.Set("http.mode", "release")
			router, err := apiserver.GinRouter()
			if err != nil {
				panic(err)
			}
			intercom.PrintGinRouteInfo(router.Routes())
		},
	}
	return c
}
