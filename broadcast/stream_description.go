package broadcast

import (
	"strconv"
)

type StreamDescription struct {
	BitRate     int
	Public      bool
	Name        string
	Description string
	URL         string
	Genre       string
}

func (description *StreamDescription) IsEmpty() bool {
	return description.BitRate == 0 && description.Name == "" && description.Description == "" && description.URL == "" && description.Genre == ""
}

func (description *StreamDescription) IcecastHeaders() map[string]string {
	headers := map[string]string{}

	headers["ice-public"] = description.PublicZeroOrOne()

	if description.BitRate > 0 {
		headers["ice-bitrate"] = strconv.Itoa(description.BitRate)
	}

	var stringAttributes = []struct {
		name  string
		value string
	}{
		{"ice-name", description.Name},
		{"ice-url", description.URL},
		{"ice-genre", description.Genre},
		{"ice-description", description.Description},
	}
	for _, attribute := range stringAttributes {
		if attribute.value != "" {
			headers[attribute.name] = attribute.value
		}
	}

	return headers
}

func (description *StreamDescription) PublicZeroOrOne() string {
	if description.Public {
		return "1"
	} else {
		return "0"
	}
}
