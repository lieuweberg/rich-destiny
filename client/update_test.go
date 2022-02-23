package main

import (
	"log"
	"testing"
)

func TestGetNewReleases(t *testing.T) {
	// Modify these values as appropriate
	version = "v0.2.6"
	updatedVersion = "v0.2.6"
	want := 0
	storage = &storageStruct{
		Prereleases: false,
	}

	got, _ := getNewReleases()

	for _, r := range got {
		log.Print(r.Name)
	}

	log.Println("Don't worry if this is incorrect. You'll have to change the version number and 'want' manually to test it.")
	if len(got) != want {
		t.Errorf("getNewReleases was incorrect, got: %d, want %d, struct: %+v", len(got), want, got)
	}
}

func TestGetChecksumFromBody(t *testing.T) {
	got := getChecksumFromBody("SHA256: 5E4FAB223ED4C78DF989B290E53B50A8B39F7B2F32E8DD23B2C21335FD48A526\r\n\r\nChanges:\r\n - New [website](https://lieuweberg.com/rich-destiny) :D")
	expected := "5E4FAB223ED4C78DF989B290E53B50A8B39F7B2F32E8DD23B2C21335FD48A526"
	if got != expected {
		t.Errorf("getChecksumFromBody was incorrect, got: %s, want: %s", got, expected)
	}
}
