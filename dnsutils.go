package main

import (
	"github.com/miekg/dns"
	"strings"
)

func getTLDFromDomain(domain string) string {
	domain = dns.Fqdn(domain)
	if domain == "."{
		return "."
	}
	sLabels := strings.Split(domain, ".")
	return sLabels[len(sLabels)-2] + "."
}


func queryAXFR(zone string , server string)([]*dns.Envelope, error){
	t := new(dns.Transfer)
	m := new(dns.Msg)
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{
		Name:   zone,
		Qtype:  dns.TypeAXFR,
		Qclass: dns.ClassINET,
	}
	c, err := t.In(m, server)
	if err != nil{
		return nil, err
	}
	result := make([]*dns.Envelope, 0)
	for r := range c {
		result = append(result, r)
	}
	return result, nil
}


func queryHTTP(server string)([]*dns.Envelope, error){
	result := make([]*dns.Envelope, 0)

	//for r := range c {
	//	result = append(result, r)
	//}

	return result, nil
}