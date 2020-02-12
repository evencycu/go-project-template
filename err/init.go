package err

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gitlab.com/cake/go-project-template/gpt"
	"gitlab.com/cake/intercom"
)

func AddErrorEndpoint(rootGroup *gin.RouterGroup) {
	// settings
	apiTimeout := viper.GetDuration("http.api_timeout")

	ErrorGroup := rootGroup.Group(gpt.APIErrorPath,
		intercom.AccessMiddleware(apiTimeout, gpt.GetNamespace()),
	)
	{
		ErrorGroup.GET("", randomErr)
		ErrorGroup.GET("/upstream", upstreamErr)
	}
}
