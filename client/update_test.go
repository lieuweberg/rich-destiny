package main

import "testing"

func TestFilterReleases(t *testing.T) {
	version = "v0.2.2"
	got := filterReleases(releasesFromGithub{
		releaseElement{
			Name:       "v0.2.4",
			Draft:      false,
			Prerelease: false,
			Assets:     []releaseAsset{},
		},
		releaseElement{
			Name:       "v0.2.4-1",
			Draft:      false,
			Prerelease: true,
			Assets:     []releaseAsset{},
		},
		releaseElement{
			Name:       "v0.2.3",
			Draft:      false,
			Prerelease: false,
			Assets:     []releaseAsset{},
		},
		releaseElement{
			Name:       "v0.2.2",
			Draft:      false,
			Prerelease: false,
			Assets:     []releaseAsset{},
		},
		releaseElement{
			Name:       "v0.2.1",
			Draft:      false,
			Prerelease: false,
			Assets:     []releaseAsset{},
		},
	})

	if len(got) != 3 {
		t.Errorf("filterReleases was incorrect, got: %d, want %d, struct: %+v", len(got), 3, got)
	}
}

func TestGetChecksumFromBody(t *testing.T) {
	got := getChecksumFromBody("SHA256: 5E4FAB223ED4C78DF989B290E53B50A8B39F7B2F32E8DD23B2C21335FD48A526\r\n\r\nChanges:\r\n - New [website](https://lieuweberg.com/rich-destiny) :D")
	expected := "5E4FAB223ED4C78DF989B290E53B50A8B39F7B2F32E8DD23B2C21335FD48A526"
	if got != expected {
		t.Errorf("getChecksumFromBody was incorrect, got: %s, want: %s", got, expected)
	}
}
