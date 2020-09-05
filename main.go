package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"time"
)

var listenAt string
var syncDuration time.Duration
var zoneFileName string
var prefer string
var syncMethod string
var debug bool

func init() {
	flag.StringVar(&syncMethod, "type", "axfr", "sync method for zone file only support axfr and http")
	flag.StringVar(&prefer, "prefer", "", "custom prefer root servers or url for sync data")
	flag.StringVar(&zoneFileName, "file", "root.zone", "local root zone file name")
	flag.StringVar(&listenAt, "listen", "0.0.0.0:53", "root dns server listen port")
	flag.DurationVar(&syncDuration, "interval", time.Minute, "sync original root zone file from upstream server")
	flag.BoolVar(&debug, "debug", false, "enable debug level log output")
}

func main() {
	flag.Parse()
	// 日志设置：如果不设置级别，默认为warning
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	if debug == true {
		log.SetLevel(log.DebugLevel)
		log.Debug("set application log level to debug")
	} else {
		log.SetLevel(log.InfoLevel)

	}
	manager, err := NewManager(zoneFileName, syncDuration, syncMethod, prefer)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("start sync from remote dns server")
	err = manager.Sync()
	if err != nil {
		log.Errorf("sync fail:%s", err)
		// using local server file if exist
		err := manager.SyncFromFile()
		if err != nil {
			log.Error(err)
			return
		}
		log.Warning("load local zone file success")
		log.Warning("server will provide dns response using stale zone data")
	}
	log.Infof("ready to serve root dns query")
	log.Panic(manager.Run(listenAt))
}
