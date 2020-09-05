package main

import (
	"errors"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Manager struct {
	sync.RWMutex
	zoneStore    *ZoneStore
	synchronizer ZoneSynchronizer
	zoneFile     string
	syncMethod   string
	syncDuration time.Duration
}

func NewManager(fileName string, duration time.Duration, syncMethod string, preferServer string) (*Manager, error) {
	var synchronizer ZoneSynchronizer
	var err error
	if syncMethod == "axfr" {
		synchronizer, err = NewAXFRSynchronizer(fileName, preferServer)
		if err != nil {
			return nil, err
		}
		if preferServer == "" {
			preferServer = DefaultAXFRRootList[0]
		}
		log.Infof("using axfr to sync zone data from [%s,..]", preferServer)
	} else if syncMethod == "http" {
		synchronizer, err = NewHTTPSynchronizer(fileName, preferServer)
		if err != nil {
			return nil, err
		}
		if preferServer == "" {
			preferServer = ZoneDownloadURL[0]
		}
		log.Infof("using axfr to sync zone data from %s", preferServer)
	} else {
		return nil, errors.New("unsupported sync method")
	}
	if duration.Seconds() < 30 {
		return nil, errors.New("sync interval should greater than 30 seconds")
	}
	manager := Manager{
		zoneFile:     fileName,
		syncDuration: duration,
		syncMethod:   syncMethod,
		synchronizer: synchronizer,
	}
	return &manager, nil
}

func (manager *Manager) Sync() error {
	data, err := manager.synchronizer.Download()
	if err != nil {
		return err
	}
	manager.Lock()
	manager.zoneStore = data
	manager.Unlock()
	return manager.synchronizer.SyncToFile(data)
}

func (manager *Manager) SyncFromFile() error {
	data, err := manager.synchronizer.SyncFromFile()
	if err != nil {
		return err
	}
	manager.zoneStore = data
	return nil
}

func (manager *Manager) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	if len(r.Question) == 0 {
		w.WriteMsg(m)
		return
	}
	domain := r.Question[0].Name
	qType := r.Question[0].Qtype
	do := r.IsEdns0().Do()
	manager.RLock()
	defer manager.RUnlock()
	if manager.zoneStore == nil {
		m.Rcode = dns.RcodeServerFailure
		w.WriteMsg(m)
		return
	}
	answer, ns, additional, aa := manager.zoneStore.Query(domain, qType, do)
	m.Answer = answer
	m.Ns = ns
	m.Extra = additional
	m.Authoritative = aa
	m.SetEdns0(4096, false)
	w.WriteMsg(m)
}

func (manager *Manager) Run(listenAt string) error {
	go func() {
		for range time.NewTicker(manager.syncDuration).C {
			err := manager.Sync()
			if err != nil {
				log.Errorf("sync fail: %s ", err)
			}
		}
	}()
	server := dns.Server{Addr: listenAt, Net: "udp"}
	dns.HandleFunc(".", manager.handleRequest)
	log.Infof("start dns server at : %s", listenAt)
	err := server.ListenAndServe()
	return err
}
