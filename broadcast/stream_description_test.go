package broadcast

import (
	"reflect"
	"testing"
)

func testStreamDescription() *StreamDescription {
	return &StreamDescription{
		BitRate:     96000,
		Public:      true,
		Name:        "GoBroadcast test stream",
		Description: "Test description",
		URL:         "http://projects.tryphon.eu/projects/go-broadcast",
		Genre:       "Test",
	}
}

func TestStreamDescription_IsEmpty(t *testing.T) {
	description := testStreamDescription()
	if description.IsEmpty() {
		t.Errorf("Description should not be empty: %v", description)
	}

	description = &StreamDescription{}
	if !description.IsEmpty() {
		t.Errorf("Description should be empty: %v", description)
	}
}

func TestStreamDescription_IcecastHeaders(t *testing.T) {
	description := testStreamDescription()

	expectedHeaders := map[string]string{
		"ice-bitrate":     "96000",
		"ice-public":      "1",
		"ice-name":        "GoBroadcast test stream",
		"ice-description": "Test description",
		"ice-url":         "http://projects.tryphon.eu/projects/go-broadcast",
		"ice-genre":       "Test",
	}
	if !reflect.DeepEqual(description.IcecastHeaders(), expectedHeaders) {
		t.Errorf("Wrong icecast headers :\n got: %v\nwant: %v", description.IcecastHeaders(), expectedHeaders)
	}

	description = &StreamDescription{}
	expectedHeaders = map[string]string{
		"ice-public": "0",
	}
	if !reflect.DeepEqual(description.IcecastHeaders(), expectedHeaders) {
		t.Errorf("Wrong icecast headers :\n got: %v\nwant: %v", description.IcecastHeaders(), expectedHeaders)
	}
}

func TestStreamDescription_ShoutcastHeaders(t *testing.T) {
	description := testStreamDescription()

	expectedHeaders := map[string]string{
		"icy-bt":    "96000",
		"icy-pub":   "1",
		"icy-name":  "GoBroadcast test stream",
		"icy-url":   "http://projects.tryphon.eu/projects/go-broadcast",
		"icy-genre": "Test",
	}
	if !reflect.DeepEqual(description.ShoutcastHeaders(), expectedHeaders) {
		t.Errorf("Wrong icecast headers :\n got: %v\nwant: %v", description.ShoutcastHeaders(), expectedHeaders)
	}

	description = &StreamDescription{}
	expectedHeaders = map[string]string{
		"icy-bt":  "0",
		"icy-pub": "0",
	}
	if !reflect.DeepEqual(description.ShoutcastHeaders(), expectedHeaders) {
		t.Errorf("Wrong icecast headers :\n got: %v\nwant: %v", description.ShoutcastHeaders(), expectedHeaders)
	}
}

func TestStreamDescription_PublicZeroOrOne(t *testing.T) {
	description := StreamDescription{}

	description.Public = false
	if description.PublicZeroOrOne() != "0" {
		t.Errorf(" :\n got: %v\nwant: %v", description.PublicZeroOrOne(), "0")
	}

	description.Public = true
	if description.PublicZeroOrOne() != "1" {
		t.Errorf(" :\n got: %v\nwant: %v", description.PublicZeroOrOne(), "1")
	}
}
