package main

import "testing"

func TestGetTLDFromDomain(t *testing.T){
	for s, expect := range map[string]string{
		"":                  ".",
		".": ".",
		"com": "com.",
		"com.":"com.",
		".net.":"net.",
		"www.google.com":"com.",
		"www.google.com.":"com.",
	} {
		if got := getTLDFromDomain(s); got != expect {
			t.Errorf("getTLDFromDomain(%s) = %s, expected %s", s, got, expect)
		}
	}
}


func TestQueryAXFR(t *testing.T){
	rootData, err := queryAXFR(".", DefaultAXFRRootList[0])
	if  err !=nil {
		t.Errorf("expect root transfer success got data but got err:%s", err)
	}
	if len(rootData) == 0 {
		t.Error("expect root transfer success got data but got zero")
	}
}