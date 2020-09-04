package main

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

//  None of the root services are guaranteed to be available.
//  It is possible that ICANN or some of the root server operators will turn off
//	the AXFR capability on the servers.
var DefaultAXFRRootList = []string{
	"k.root-servers.net:53",
	"b.root-servers.net:53",
	"c.root-servers.net:53",
	"d.root-servers.net:53",
	"g.root-servers.net:53",
	"lax.xfr.dns.icann.org:53",
	"iad.xfr.dns.icann.org:53",
}

var IANAZoneDownloadURL = "https://www.internic.net/domain/root.zone"

type ZoneData struct {
	NS         []dns.RR
	Additional []dns.RR
	Extra      []dns.RR
}
type ZoneStore struct {
	data map[string]map[uint16][]dns.RR
	zone map[string]*ZoneData
}

func (query *ZoneStore) Query(domain string, qType uint16) (answer []dns.RR, ns []dns.RR, extra []dns.RR) {
	//if data, ok := query.data[domain];ok{
	//	if res, ok := data[qType];ok{
	//		return answer,
	//	}
	//}
	return []dns.RR{}, []dns.RR{}, []dns.RR{}
}

type ZoneSynchronizer interface {
	Download() (*ZoneStore, error)
}

type AxfrSynchronizer struct {
	axfrServers []string `validate:"required,hostname_port"`
}

func NewAXFRSynchronizer(server string) (*AxfrSynchronizer, error) {
	var validate = validator.New()
	axfrServer := DefaultAXFRRootList
	if server != "" {
		axfrServer = append([]string{server}, DefaultAXFRRootList...)
	}
	syncer := &AxfrSynchronizer{axfrServers: axfrServer}
	err := validate.Struct(syncer)
	if err != nil {
		return nil, err
	}
	return syncer, nil
}

func (syncer *AxfrSynchronizer) Download() (*ZoneStore, error) {
	for _, server := range syncer.axfrServers {
		data, err := queryAXFR(".", server)
		if err != nil {
			log.Errorf("send axfr to server : %s error : %s", server, err)
			continue
		}
		return syncer.parser(data), nil
	}
	return nil, errors.New("send axfr to all servers failed")
}

func (syncer *AxfrSynchronizer) parser(data []*dns.Envelope) *ZoneStore {
	results := make(map[string]map[uint16][]dns.RR)
	for _, envelope := range data {
		if len(envelope.RR) == 0 {
			continue
		}

		for _, rr := range envelope.RR {
			domain := rr.Header().Name
			qtype := rr.Header().Rrtype
			_, ok := results[domain]
			if ok != true {
				results[domain] = make(map[uint16][]dns.RR)
			}
			_, ok = results[domain][qtype]
			if ok != true {
				results[domain][qtype] = make([]dns.RR, 0)
			}
			results[domain][qtype] = append(results[domain][qtype], rr)
		}
	}
	zone := make(map[string]*ZoneData)
	for domain, store := range results{
		zoneItme := ZoneData{
			NS:         make([]dns.RR,0),
			Additional: make([]dns.RR,0),
			Extra:      make([]dns.RR,0),
		}
		zoneItme.NS = append(zoneItme.NS, store[dns.TypeNS]...)
		for _, rr:= range store[dns.TypeNS]{
			nsServer := rr.Header().Name
			if z, ok := results[nsServer];ok == true{
				if aData, ok := z[dns.TypeA]; ok == true{
					zoneItme.Additional = append(zoneItme.Additional, aData...)
				}
				if aaaaData, ok := z[dns.TypeAAAA]; ok == true{
					zoneItme.Additional = append(zoneItme.Additional, aaaaData...)
				}
			}
		}
		zone[domain] = &zoneItme
	}

	return &ZoneStore{data: results}
}

type HTTPSynchronizer struct {
	url string `validate:"required,url"`
}

func NewHTTPSynchronizer(url string) (*HTTPSynchronizer, error) {
	downloadURL := IANAZoneDownloadURL
	if url != "" {
		downloadURL = url
	}
	syncer := &HTTPSynchronizer{url: downloadURL}
	err := validator.New().Struct(syncer)
	if err != nil {
		return nil, err
	}
	return syncer, nil
}

func (syncer *HTTPSynchronizer) Download() (*ZoneStore, error) {
	return nil, nil
}
