package main

import (
	"bufio"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type HTTPSynchronizer struct {
	filename string   `validate:"required"`
	urls     []string `validate:"required,url"`
}

func NewHTTPSynchronizer(filename string, url string) (*HTTPSynchronizer, error) {
	downloadURLS := make([]string, 0)
	if url != "" {
		downloadURLS = append([]string{url}, ZoneDownloadURL...)
	} else {
		downloadURLS = ZoneDownloadURL
	}
	synchronizer := &HTTPSynchronizer{filename: filename, urls: downloadURLS}
	err := validator.New().Struct(synchronizer)
	if err != nil {
		return nil, err
	}
	return synchronizer, nil
}

func (synchronizer *HTTPSynchronizer) Download() (*ZoneStore, error) {
	var response *http.Response
	var err error
	for _, url := range synchronizer.urls {
		log.Debugf("download zone file from %s start", url)
		response, err = http.Get(url)
		if err != nil {
			log.Errorf("download zone file from %s fail:%s", url, err)
			continue
		}
	}

	if response == nil || err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(response.Body)

	rrs := make([]dns.RR, 0)
	for scanner.Scan() {
		rr, err := dns.NewRR(scanner.Text())
		if err != nil {
			continue
		}
		rrs = append(rrs, rr)
	}
	zoneStore := NewZoneStoreFromRRSet(rrs)
	if zoneStore == nil {
		return nil, errors.New("zone store not create success")
	}
	return zoneStore, nil
}

func (synchronizer *HTTPSynchronizer) SyncToFile(store *ZoneStore) error {
	return store.ToFile(synchronizer.filename)
}

func (synchronizer *HTTPSynchronizer) SyncFromFile() (*ZoneStore, error) {
	return NewZoneStoreFromFile(synchronizer.filename)
}
