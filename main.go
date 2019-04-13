package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"gitlab.com/general-backend/goctx"
	"gitlab.com/general-backend/gotrace"
	"gitlab.com/general-backend/m800log"
	"gitlab.com/general-backend/mgopool"
	"gitlab.com/rayshih/go-project-template/apiserver"

	"github.com/spf13/viper"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

type options struct {
	version    bool
	prof       bool
	configFile string
}

var opts options
var systemCtx goctx.Context

func init() {
	flag.BoolVar(&opts.version, "build", false, "GoLang build version.")
	flag.BoolVar(&opts.prof, "prof", false, "GoLang profiling function.")
	flag.StringVar(&opts.configFile, "config", "./local.toml", "Path to Config File")

	systemCtx = goctx.Background()
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [arguments] <command> \n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if opts.version {
		fmt.Fprintf(os.Stderr, "%s\n", runtime.Version())
	}
	if opts.prof {
		ActivateProfile()
	}
	viper.AutomaticEnv()
	viper.SetConfigFile(opts.configFile)
	viper.ReadInConfig() // Find and read the config file
	log.SetFlags(log.LstdFlags)

	// Init log
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	var err error
	err = m800log.Initialize(viper.GetString("log.output"), viper.GetString("log.level"))
	if err != nil {
		panic(err)
	}
	m800log.SetM800JSONFormatter(viper.GetString("log.timestamp_format"), appName, version)
	m800log.SetAccessLevel(viper.GetString("log.access_level"))
	// Init tracer
	err = initTracer()
	if err != nil {
		panic(err)
	}

	// Init mongo

	// if connect to multi mongodb cluster, take this pool to use

	// pool, err = mgopool.NewSessionPool(getXXXMongoDBInfo())
	// if err != nil {
	// 	m800log.Error(systemCtx, "mongo connect error:", err, ", config:", viper.AllSettings())
	// }

	err = mgopool.Initialize(getMongoDBInfo())
	if err != nil {
		m800log.Error(systemCtx, "mongo connect error:", err, ", config:", viper.AllSettings())
		panic(err)
	}

	httpServer := apiserver.InitGinServer(systemCtx, version)

	// graceful shutdown
	quit := make(chan os.Signal, 5)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

func initTracer() error {
	if !viper.GetBool("jaeger.enabled") {
		log.Println("Jaeger disabled")
		return nil
	}
	sConf := &jaegercfg.SamplerConfig{
		Type:  jaeger.SamplerTypeRateLimiting,
		Param: viper.GetFloat64("jaeger.sample_rate"),
	}
	rConf := &jaegercfg.ReporterConfig{
		QueueSize:           viper.GetInt("jaeger.queue_size"),
		BufferFlushInterval: viper.GetDuration("jaeger.flush_interval"),
		LocalAgentHostPort:  viper.GetString("jaeger.host"),
		LogSpans:            viper.GetBool("jaeger.log_spans"),
	}
	log.Printf("Sampler Config:%+v\nReporterConfig:%+v\n", sConf, rConf)
	if err := gotrace.InitJaeger(appName, sConf, rConf); err != nil {
		return fmt.Errorf("init tracer error:%s", err.Error())
	}
	return nil
}

func getMongoDBInfo() *mgopool.DBInfo {
	name := viper.GetString("database.mgo.name")
	mgoMaxConn := viper.GetInt("database.mgo.max_conn")
	mgoUser := viper.GetString("database.mgo.user")
	mgoPassword := viper.GetString("database.mgo.password")
	mgoAuthDatabase := viper.GetString("database.mgo.authdatabase")
	mgoTimeout := viper.GetDuration("database.mgo.timeout")
	mgoDirect := viper.GetBool("database.mgo.direct")
	mgoSecondary := viper.GetBool("database.mgo.secondary")
	mgoMonogs := viper.GetBool("database.mgo.mongos")
	mgoAddrs := strings.Split(viper.GetString("database.mgo.hosts"), ";")
	if len(mgoAddrs) == 0 {
		log.Fatal("Config error: no mongo hosts")
	}
	return mgopool.NewDBInfo(name, mgoAddrs, mgoUser, mgoPassword,
		mgoAuthDatabase, mgoTimeout, mgoMaxConn, mgoDirect, mgoSecondary, mgoMonogs)
}
