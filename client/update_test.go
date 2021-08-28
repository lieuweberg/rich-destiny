package main

import (
	"log"
	"testing"
)

func TestGetNewReleases(t *testing.T) {
	version = "v0.2.1"
	got, _ := getNewReleases()

	log.Println("Don't worry if this is incorrect. You'll have to change the version number manually to test it.")
	if len(got) != 2 {
		t.Errorf("getNewReleases was incorrect, got: %d, want %d, struct: %+v", len(got), 3, got)
	}
}

func TestGetChecksumFromBody(t *testing.T) {
	got := getChecksumFromBody("SHA256: 5E4FAB223ED4C78DF989B290E53B50A8B39F7B2F32E8DD23B2C21335FD48A526\r\n\r\nChanges:\r\n - New [website](https://lieuweberg.com/rich-destiny) :D")
	expected := "5E4FAB223ED4C78DF989B290E53B50A8B39F7B2F32E8DD23B2C21335FD48A526"
	if got != expected {
		t.Errorf("getChecksumFromBody was incorrect, got: %s, want: %s", got, expected)
	}
}
