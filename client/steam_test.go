package main

import (
	"errors"
	"log"
	"testing"
)

func TestGetJoinLink(t *testing.T) {
	got, err := getJoinLink()

	if err != nil {
		if errors.Is(err, errNoConnectString) {
			log.Println("connect string is empty")
		} else {
			t.Error(err)
		}
	} else {
		log.Println(got)
	}
}
