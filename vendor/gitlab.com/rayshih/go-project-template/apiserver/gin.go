package apiserver

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"gitlab.com/general-backend/goctx"
	"gitlab.com/general-backend/m800log"
	"gitlab.com/general-backend/mgopool"
)

func InitGinServer(ctx goctx.Context, version string) *http.Server {
	// Create gin http server.
	gin.SetMode(viper.GetString("http.mode"))
	router := gin.New()
	router.Use(gin.Recovery())

	// Add gin prometheus metrics
	p := ginprometheus.NewPrometheus("gin")
	p.Use(router)

	// general service for debugging
	router.GET("/health", health)
	router.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"version": version})
	})
	router.GET("/config", appConfig)
	router.GET("/mongo", mongoInfo)
	router.GET("/ready", ready)

	// TODO: Add application routing

	port := viper.GetString("http.port")
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  viper.GetDuration("http.read_timeout"),
		WriteTimeout: viper.GetDuration("http.write_timeout"),
	}

	go func() {
		m800log.Info(ctx, fmt.Sprintf("Server is running and listening port: %s", port))
		// service connections
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	return httpServer
}

func mongoInfo(c *gin.Context) {
	status := map[string]interface{}{}
	status["Len"] = mgopool.Len()
	status["IsAvailable"] = mgopool.IsAvailable()
	status["Cap"] = mgopool.Cap()
	status["Mode"] = mgopool.Mode()
	status["Config"] = mgopool.ShowConfig()
	status["LiveServers"] = mgopool.LiveServers()
	c.JSON(200, status)
}

func appConfig(c *gin.Context) {
	settings := viper.AllSettings()
	delete(settings, "database")
	c.JSON(200, settings)
}

func health(c *gin.Context) {
	err := mgopool.Ping(goctx.Background())
	if err != nil {
		c.JSON(503, gin.H{})
		return
	}
	c.JSON(200, gin.H{})
}

func ready(c *gin.Context) {
	err := mgopool.Ping(goctx.Background())
	if err != nil {
		c.JSON(503, gin.H{})
		return
	}
	c.JSON(200, gin.H{})
}
