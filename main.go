package main

import (
	"flag"
	"fmt"
	"gfa/common/kafka"
	"gfa/common/log"
	"gfa/core/analyzer"
	"gfa/core/config"
	"gfa/core/router"
	"gfa/core/server"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"os"
	"runtime"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		cfg string
	)
	flag.StringVar(&cfg, "conf", "", "server config [toml]")
	flag.Parse()
	if len(cfg) == 0 {
		fmt.Println("config is empty")
		os.Exit(0)
	}
	config.Init(cfg)
	conf := config.CoreConf
	log.Init(&conf.Log)
	gin.SetMode(conf.Server.Mode)
	kafka.C = conf.Kafka
	if err := kafka.Init(); err != nil {
		log.Fatal("Init kafka failed", zap.Error(err))
	}
	analyzer.Init()
	if err := server.Run(router.NewHttpRouter()); nil != err {
		log.Error("server run error", zap.Error(err))
	}
}
