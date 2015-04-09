package broadcast

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProcessingController_Show(t *testing.T) {
	processing := &Processing{}
	processing.Setup(&ProcessingConfig{Amplification: 3})
	controller := NewProcessingController(processing)

	request, _ := http.NewRequest("GET", "http://localhost:9000/processing.json", nil)

	response := httptest.NewRecorder()
	controller.ServeHTTP(response, request)

	if response.Code != 200 {
		t.Errorf("Wrong response code :\n got: %v\nwant: %v", response.Code, 200)
	}

	body := response.Body.String()
	if expected := "{\"Amplification\":3}"; body != expected {
		t.Errorf("Wrong processing json config :\n got: %v\nwant: %v", body, expected)
	}
}

func TestProcessingController_Update(t *testing.T) {
	processing := &Processing{}
	processing.Setup(&ProcessingConfig{Amplification: 3})
	controller := NewProcessingController(processing)

	request, _ := http.NewRequest("PUT", "http://localhost:9000/processing.json", strings.NewReader("{\"Amplification\":0}"))

	response := httptest.NewRecorder()
	controller.ServeHTTP(response, request)

	if response.Code != 200 {
		body, _ := ioutil.ReadAll(response.Body)
		t.Fatalf("Wrong response code (%s):\n got: %v\nwant: %v", body, response.Code, 200)
	}

	if processing.Config().Amplification != 0 {
		t.Errorf(" :\n got: %v\nwant: %v", processing.Config().Amplification, 0)
	}
}
