package analyzer

import (
	"bytes"
	"fmt"
	"gfa/common/async"
	"gfa/common/log"
	"gfa/common/utils"
	"gfa/core/config"
	"gfa/core/global"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

var (
	pool        *async.WorkerPool
	networkCard string
	dirPath     string
	size        string
	interval    string
	maxWorkers  int
	maxQueue    int
	pcapWorkers int
	alertSize   int64
	location    string
	Command     string
)

const (
	suffix = "+0800.pcap"
)

type analyzer struct {
	Pcap      string `json:"pcap"`
	Location  string `json:"location"`
	TimeStamp string `json:"timestamp"`
}

func Init() {
	networkCard = config.CoreConf.Traffic.NetworkCard
	dirPath = config.CoreConf.Traffic.Path
	size = config.CoreConf.Traffic.Size
	interval = config.CoreConf.Traffic.Interval
	maxWorkers = config.CoreConf.Server.MaxWorkers
	maxQueue = config.CoreConf.Server.MaxQueue
	alertSize = config.CoreConf.Server.Size
	pcapWorkers = config.CoreConf.Traffic.Workers
	location = config.CoreConf.Traffic.Location
	Command = fmt.Sprintf("tcpdump -i %v -s 0 -G %v -C %v -w %v/%%Y-%%m-%%dT%%H:%%M:%%S%v &", networkCard, interval, size, dirPath, suffix)
	pool = async.NewWorkerPool(maxWorkers, maxQueue, zap.NewExample()).Run()
	go fileClear()
	go dispatch()
	go startTcpDump()
}

func startTcpDump() {
	if err := utils.CreateDirectory(dirPath); err != nil {
		panic(err)
	}
	if _, err := execute(Command); err != nil {
		panic(err)
	}
}

func StopTcpDump() {
	if _, err := execute(fmt.Sprintf("pkill -f %s", Command)); err != nil {
		log.Errorf("stop tcpDump err:%v", err)
	}
}

func dispatch() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			job := &traffic{analyzer: analyzer{Pcap: newWireSharkFile()}}
			pool.Add(job)
		case <-global.Ctx.Done():
			return
		}
	}
}

func newWireSharkFile() string {
	lastSecondStr := time.Now().Add(-1 * time.Second).Format(utils.DefaultTimeLayout)
	return fmt.Sprintf("%v%v", strings.Replace(lastSecondStr, " ", "T", -1), suffix)
}

func fileClear() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			dirSize, _ := utils.DirSize(dirPath)
			if dirSize > alertSize {
				dir := utils.FileInfo(dirPath)
				index := len(dir) * 2 / 3
				for _, d := range dir[:index] {
					_ = os.RemoveAll(path.Join([]string{dirPath, d.Name()}...))
				}
			}
		case <-global.Ctx.Done():
			return
		}
	}
}

func execute(command string) (string, error) {
	var (
		out    bytes.Buffer
		stderr bytes.Buffer
	)
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf(stderr.String())
	}
	return out.String(), nil
}

func Close() {
	global.Cancel()
	pool.Close()
}
