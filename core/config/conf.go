package config

import (
	"fmt"
	"gfa/common/kafka"
	"gfa/common/log"
	"github.com/BurntSushi/toml"
	"os"
)

var (
	CoreConf *config
)

func Init(conf string) {
	_, err := toml.DecodeFile(conf, &CoreConf)
	if err != nil {
		fmt.Printf("Err %v", err)
		os.Exit(1)
	}
}

type config struct {
	Log     log.Config
	Server  Server
	Traffic Traffic
	Kafka   kafka.Config
}

type Server struct {
	Port       int
	Mode       string
	Size       int64
	MaxWorkers int
	MaxQueue   int
}

type Traffic struct {
	Path        string
	NetworkCard string
	Interval    string
	Size        string
	Workers     int
	Location    string
}
