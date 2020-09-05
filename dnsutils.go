package main

import (
	"errors"
	"github.com/miekg/dns"
	"os"
	"strings"
)

func getTLDFromDomain(domain string) string {
	domain = dns.Fqdn(domain)
	if domain == "." {
		return "."
	}
	sLabels := strings.Split(domain, ".")
	return sLabels[len(sLabels)-2] + "."
}

func queryAXFR(zone string, server string) ([]*dns.Envelope, error) {
	t := new(dns.Transfer)
	m := new(dns.Msg)
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{
		Name:   zone,
		Qtype:  dns.TypeAXFR,
		Qclass: dns.ClassINET,
	}
	c, err := t.In(m, server)
	if err != nil {
		return nil, err
	}
	result := make([]*dns.Envelope, 0)
	for r := range c {
		result = append(result, r)
	}
	return result, nil
}

func queryHTTP(server string) ([]*dns.Envelope, error) {
	result := make([]*dns.Envelope, 0)

	//for r := range c {
	//	result = append(result, r)
	//}

	return result, nil
}

// fileCreateIfNotExists checks if a file exists and is not a directory
// create a new file is not exist
func fileCreateIfNotExists(filename string) error {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		emptyFile, err := os.Create(filename)
		if err != nil {
			return err
		}
		emptyFile.Close()
		return nil
	}
	if info.IsDir() == true {
		return errors.New("path is a directory not a file")
	}
	return nil
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
