package main

import (
	"bufio"
	"errors"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"os"
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

var ZoneDownloadURL = []string{
	"https://www.internic.net/domain/root.zone",
}

type ZoneData struct {
	NS         []dns.RR
	Additional []dns.RR
}
type ZoneStore struct {
	data map[string]map[uint16][]dns.RR
	zone map[string]*ZoneData
	rrsigs map[string]map[uint16]dns.RR
}

func NewZoneStoreFromFile(filename string) (*ZoneStore, error) {
	isExist := fileExists(filename)
	if isExist != true {
		return nil, errors.New("file not exist")
	}
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
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

func NewZoneStoreFromRRSet(data []dns.RR) *ZoneStore {
	if len(data) == 0 {
		return nil
	}
	results := make(map[string]map[uint16][]dns.RR)
	rrsigs := make(map[string]map[uint16]dns.RR)
	for _, rr := range data {
		domain := rr.Header().Name
		qType := rr.Header().Rrtype
		_, ok := results[domain]
		if ok != true {
			results[domain] = make(map[uint16][]dns.RR)
		}
		_, ok = results[domain][qType]
		if ok != true {
			results[domain][qType] = make([]dns.RR, 0)
		}
		results[domain][qType] = append(results[domain][qType], rr)
		if qType == dns.TypeRRSIG{
			_, ok := rrsigs[domain]
			if ok == false {
				rrsigs[domain] = make(map[uint16]dns.RR)
			}
			rrsigs[domain][qType] = rr
		}
	}
	zone := make(map[string]*ZoneData)
	for domain, store := range results {
		nsdata, ok := store[dns.TypeNS]
		if ok != true {
			continue
		}
		zoneItme := ZoneData{
			NS:         make([]dns.RR, 0),
			Additional: make([]dns.RR, 0),
		}
		zoneItme.NS = append(zoneItme.NS, store[dns.TypeNS]...)
		for _, rr := range nsdata {
			casted, ok := rr.(*dns.NS)
			if ok == false {
				continue
			}
			if z, ok := results[casted.Ns]; ok == true {
				if aData, ok := z[dns.TypeA]; ok == true {
					zoneItme.Additional = append(zoneItme.Additional, aData...)
				}
				if aaaaData, ok := z[dns.TypeAAAA]; ok == true {
					zoneItme.Additional = append(zoneItme.Additional, aaaaData...)
				}
			}
		}
		zone[domain] = &zoneItme
	}
	if len(results) == 0 || len(zone) == 0 {
		return nil
	}
	zoneStore := &ZoneStore{data: results, zone: zone, rrsigs: rrsigs}
	return zoneStore
}

func (store *ZoneStore) ToFile(filename string) error {
	err := fileCreateIfNotExists(filename)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, typeWithRRs := range store.data {
		for _, rrs := range typeWithRRs {
			for _, rr := range rrs {
				_, err := file.WriteString(rr.String() + "\n")
				if err != nil {
					log.Errorf(err.Error())
				}
			}
		}
	}
	return nil
}
func (store *ZoneStore) Query(domain string, qType uint16, do bool) (answer []dns.RR, ns []dns.RR, additional []dns.RR, aa bool) {
	domain = dns.Fqdn(domain)
	if domain == "." {
		if data, ok := store.data[domain]; ok {
			if typeData, ok := data[qType]; ok {
				answer = typeData
				if qType == dns.TypeNS {
					additional = store.zone[domain].Additional
				}
			} else {
				if soa, ok := data[dns.TypeSOA]; ok {
					ns = soa
				}
			}
		}
		aa = true
	} else {
		tld := getTLDFromDomain(domain)
		if data, ok := store.data[tld]; ok {
			if typeData, ok := data[dns.TypeNS]; ok {
				ns = typeData
				additional = store.zone[tld].Additional
			}
		}
	}
	if do == true{
		//Todo: append rrsig for each response type
	}
	return
}
