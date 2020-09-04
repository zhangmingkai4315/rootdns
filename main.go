package main

import (
	"flag"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)



var listenAt string
var syncDuration time.Duration
var zoneFileName string
var upstream string
var syncType string
func init() {
	flag.StringVar(&syncType, "t", "axfr","sync method for zone file[axfr]")
	flag.StringVar(&upstream, "upstream", "", "root servers for sync data")
	flag.StringVar(&zoneFileName, "f", "root.zone", "root zone file name[root.zone]")
	flag.StringVar(&listenAt, "l", "0.0.0.0:53", "root dns server listen port[0.0.0.0:53]")
	flag.DurationVar(&syncDuration, "d", time.Minute, "sync original root zone file from upstream server")
}

type Manager struct {
	sync.RWMutex
	zoneStore         *ZoneStore
	synchronizer    ZoneSynchronizer
	zonefile     string
	syncDuration time.Duration
}

func NewManager(fileName string , duration time.Duration) *Manager{
	manager := Manager{
		zonefile:    fileName,
		syncDuration: duration,
	}
	go func() {
		for range time.NewTicker(duration).C{
			logrus.Debug("sync file from server using axfr")
			err := manager.sync()
			if err != nil{
				logrus.Errorf("sync root zone file fail %s", err)
			}
		}
	}()

	return &manager
}

func (manager *Manager) sync()error{

	data, err:= manager.synchronizer.Download()
	if err != nil{
		return err
	}
	manager.Lock()
	defer manager.Unlock()
	manager.zoneStore = data
	return nil
}


func (manager *Manager)handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	if len(r.Question) == 0{
		w.WriteMsg(m)
		return
	}
	domain := r.Question[0].Name
	qtype := r.Question[0].Qtype

	tld:=getTLDFromDomain(domain)
	manager.RLock()
	defer manager.RUnlock()
	answer, ns, extra := manager.zoneStore.Query(tld,qtype)
	m.Answer = answer
	m.Ns = ns
	m.Extra = extra
	w.WriteMsg(m)
}

func (manager *Manager) Run(listenAt string) error {
	server := dns.Server{Addr: listenAt, Net: "udp"}
	dns.HandleFunc(".", manager.handleRequest)

	logrus.Infof("start dns server at : %s", listenAt)
	err := server.ListenAndServe()
	return err
}

func main(){
	flag.Parse()
	manager := NewManager(zoneFileName, syncDuration)
	logrus.Panic(manager.Run(listenAt))
}