package main

import (
	"fmt"
	"testing"
)

func TestAxfrSynchronizer(t *testing.T) {
	synchronizer, err := NewAXFRSynchronizer("")
	if err != nil{
		t.Errorf("empty server will alway use default and never fail")
		return
	}
	data, err := synchronizer.Download()
	if err != nil{
		t.Errorf("download fail : %s", err)
		return
	}

	fmt.Printf("%v", data)

}