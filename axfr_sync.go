package main

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type ZoneSynchronizer interface {
	Download() (*ZoneStore, error)
	SyncToFile(data *ZoneStore) error
	SyncFromFile() (*ZoneStore, error)
}

type AxfrSynchronizer struct {
	filename    string   `validate:"required"`
	axfrServers []string `validate:"required,hostname_port"`
}

func NewAXFRSynchronizer(filename string, server string) (*AxfrSynchronizer, error) {
	var validate = validator.New()
	axfrServer := DefaultAXFRRootList
	if server != "" {
		axfrServer = append([]string{server}, DefaultAXFRRootList...)
	}
	synchronizer := &AxfrSynchronizer{
		filename:    filename,
		axfrServers: axfrServer,
	}
	err := validate.Struct(synchronizer)
	if err != nil {
		return nil, err
	}
	return synchronizer, nil
}

func (synchronizer *AxfrSynchronizer) Download() (*ZoneStore, error) {
	for _, server := range synchronizer.axfrServers {
		log.Debugf("start axfr from server: %s", server)
		data, err := queryAXFR(".", server)
		if err != nil {
			log.Errorf("send axfr to server : %s error : %s", server, err)
			continue
		}
		rrs := make([]dns.RR, 0)
		for _, envelope := range data {
			for _, rr := range envelope.RR {
				rrs = append(rrs, rr)
			}
		}
		log.Debugf("axfr transfer from server: %s success", server)
		zoneStore := NewZoneStoreFromRRSet(rrs)
		if zoneStore == nil {
			return nil, errors.New("zone store not create success")
		}
		return zoneStore, nil
	}
	return nil, errors.New("send axfr request to all servers failed")
}

func (synchronizer *AxfrSynchronizer) SyncToFile(store *ZoneStore) error {
	return store.ToFile(synchronizer.filename)
}

func (synchronizer *AxfrSynchronizer) SyncFromFile() (*ZoneStore, error) {
	return NewZoneStoreFromFile(synchronizer.filename)
}
