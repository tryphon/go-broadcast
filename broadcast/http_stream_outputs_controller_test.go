package broadcast

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/jsonq"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testHttpStreamOutputsController() *HttpStreamOutputsController {
	config := &HttpStreamOutputsConfig{}
	LoadConfig("testdata/http_stream_outputs_config.json", config)

	outputs := NewHttpStreamOutputs()
	outputs.Setup(config)

	return &HttpStreamOutputsController{outputs: outputs}
}

func jsonQuery(text string) *jsonq.JsonQuery {
	data := map[string]interface{}{}
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.Decode(&data)

	return jsonq.NewQuery(data)
}

func TestHttpStreamOutputsController_Index(t *testing.T) {
	controller := testHttpStreamOutputsController()

	request, _ := http.NewRequest("GET", "http://localhost:9000/streams.json", nil)

	response := httptest.NewRecorder()
	controller.ServeHTTP(response, request)

	if response.Code != 200 {
		t.Errorf("Wrong response code :\n got: %v (%s)\nwant: %v", response.Code, response.Body.String(), 200)
	}

	data := make([]map[string]interface{}, 0)
	decoder := json.NewDecoder(strings.NewReader(response.Body.String()))
	decoder.Decode(&data)

	firstTarget := data[0]["Target"]

	if firstTarget != "http://source:secret@stream-in.tryphon.eu:8000/stagebox1.mp3" {
		t.Errorf(" :\n got: %v\nwant: %v", firstTarget, "http://source:secret@stream-in.tryphon.eu:8000/stagebox1.mp3")
	}
}

func TestHttpStreamOutputsController_Show(t *testing.T) {
	controller := testHttpStreamOutputsController()

	request, _ := http.NewRequest("GET", "http://localhost:9000/streams/ogg.json", nil)

	response := httptest.NewRecorder()
	controller.ServeHTTP(response, request)

	if response.Code != 200 {
		t.Errorf("Wrong response code :\n got: %v\nwant: %v", response.Code, 200)
	}

	jq := jsonQuery(response.Body.String())
	target, err := jq.String("Target")
	if err != nil {
		t.Fatal(err)
	}

	if target != "http://source:secret@stream-in.tryphon.eu:8000/stagebox1.ogg" {
		t.Errorf("JSON response should contain stream attributes :\n got: %v\nwant: %v", target, "http://source:secret@stream-in.tryphon.eu:8000/stagebox1.ogg")
	}
}

func TestHttpStreamOutputsController_Update(t *testing.T) {
	controller := testHttpStreamOutputsController()
	defer controller.outputs.Stop()

	newTarget := "http://source:secret@stream-in.tryphon.eu:8000/change.ogg"
	request, _ := http.NewRequest("PUT", "http://localhost:9000/streams/ogg.json", strings.NewReader(fmt.Sprintf(`{"Target":"%s"}`, newTarget)))

	response := httptest.NewRecorder()
	controller.ServeHTTP(response, request)

	if response.Code != 200 {
		body, _ := ioutil.ReadAll(response.Body)
		t.Fatalf("Wrong response code (%s):\n got: %v\nwant: %v", body, response.Code, 200)
	}

	stream := controller.outputs.Stream("ogg")
	if stream.output.Target != newTarget {
		t.Errorf("Stream.Target should be changed :\n got: %v\nwant: %v", stream.output.Target, newTarget)
	}

	jq := jsonQuery(response.Body.String())
	target, err := jq.String("Target")
	if err != nil {
		t.Fatal(err)
	}

	if target != newTarget {
		t.Errorf("JSON response should contain updated attributes :\n got: %v\nwant: %v", target, newTarget)
	}
}

func TestHttpStreamOutputsController_Destroy(t *testing.T) {
	controller := testHttpStreamOutputsController()

	request, _ := http.NewRequest("DELETE", "http://localhost:9000/streams/ogg.json", nil)

	response := httptest.NewRecorder()
	controller.ServeHTTP(response, request)

	if response.Code != 200 {
		t.Errorf("Wrong response code :\n got: %v\nwant: %v", response.Code, 200)
	}

	stream := controller.outputs.Stream("ogg")
	if stream != nil {
		t.Errorf("Stream should be deleted:\n got: %v", stream)
	}

	jq := jsonQuery(response.Body.String())
	target, err := jq.String("Target")
	if err != nil {
		t.Fatal(err)
	}

	if target != "http://source:secret@stream-in.tryphon.eu:8000/stagebox1.ogg" {
		t.Errorf("JSON response should contain deleted stream attributes :\n got: %v\nwant: %v", target, "http://source:secret@stream-in.tryphon.eu:8000/stagebox1.ogg")
	}
}

func TestHttpStreamOutputsController_Create(t *testing.T) {
	controller := testHttpStreamOutputsController()
	defer controller.outputs.Stop()

	newTarget := "http://source:secret@stream-in.tryphon.eu:8000/change.ogg"
	request, _ := http.NewRequest("POST", "http://localhost:9000/streams.json", strings.NewReader(fmt.Sprintf(`{"Identifier": "new", "Target":"%s"}`, newTarget)))

	response := httptest.NewRecorder()
	controller.ServeHTTP(response, request)

	if response.Code != 200 {
		t.Errorf("Wrong response code :\n got: %v\nwant: %v", response.Code, 200)
	}

	stream := controller.outputs.Stream("new")
	if stream == nil {
		t.Fatal("Stream with identifier 'new' is not found")
	}
	if stream.output.Target != newTarget {
		t.Errorf("Stream.Target should be changed :\n got: %v\nwant: %v", stream.output.Target, newTarget)
	}

	jq := jsonQuery(response.Body.String())
	target, err := jq.String("Target")
	if err != nil {
		t.Fatal(err)
	}

	if target != newTarget {
		t.Errorf("JSON response should contain attributes of created Stream:\n got: %v\nwant: %v", target, newTarget)
	}
}
