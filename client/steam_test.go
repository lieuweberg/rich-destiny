package main

import (
	"log"
	"testing"
)

func TestGetJoinLink(t *testing.T) {
	got := getJoinLink()

	if got == "" {
		t.Errorf("got empty string")
	} else {
		log.Println(got)
	}
}