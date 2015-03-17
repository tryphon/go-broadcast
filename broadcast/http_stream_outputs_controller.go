package broadcast

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

type HttpStreamOutputsController struct {
	outputs *HttpStreamOutputs
}

func NewHttpStreamOutputsController(outputs *HttpStreamOutputs) (controller *HttpStreamOutputsController) {
	return &HttpStreamOutputsController{outputs: outputs}
}

func (controller *HttpStreamOutputsController) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	// Log.Debugf("HttpStreamOutputsController request: %s", request)

	path := request.URL.Path
	resourcePathPattern := regexp.MustCompile("/streams/([0-9a-zA-Z-]+).json")

	var body []byte
	if request.Body != nil {
		var err error
		body, err = ioutil.ReadAll(request.Body)
		if err != nil {
			controller.fatal(response, err)
			return
		}
	}

	switch {
	case resourcePathPattern.MatchString(path):
		identifier := resourcePathPattern.FindStringSubmatch(path)[1]

		switch {
		case request.Method == "GET":
			controller.Show(response, identifier)
		case request.Method == "DELETE":
			controller.Delete(response, identifier)
		case request.Method == "PUT":
			controller.Update(response, identifier, body)
		}
	case path == "/streams.json":
		switch {
		case request.Method == "GET":
			controller.Index(response)
		case request.Method == "POST":
			controller.Create(response, body)
		}
	case path == "/events.json" && request.Method == "GET":
		controller.IndexEvents(response)
	}
}

func (controller *HttpStreamOutputsController) Config() HttpStreamOutputsConfig {
	return controller.outputs.Config()
}

func (controller *HttpStreamOutputsController) Status() HttpStreamOutputsStatus {
	return controller.outputs.Status()
}

func (controller *HttpStreamOutputsController) Index(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	jsonBytes, err := json.Marshal(controller.Status().Streams)
	if err == nil {
		response.Write(jsonBytes)
	} else {
		controller.fatal(response, err)
	}
}

func (controller *HttpStreamOutputsController) IndexEvents(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	jsonBytes, err := json.Marshal(controller.Status().Events)
	if err == nil {
		response.Write(jsonBytes)
	} else {
		controller.fatal(response, err)
	}
}

func (controller *HttpStreamOutputsController) Show(response http.ResponseWriter, identifier string) {
	response.Header().Set("Content-Type", "application/json")

	Log.Debugf("Retrieve stream : '%s'", identifier)
	stream := controller.outputs.Stream(identifier)

	if stream != nil {
		jsonBytes, err := json.Marshal(stream.Status())
		if err == nil {
			response.Write(jsonBytes)
		} else {
			controller.fatal(response, err)
		}
	} else {
		http.Error(response, fmt.Sprintf("Stream not found: '%s'", identifier), 404)
	}
}

func (controller *HttpStreamOutputsController) Update(response http.ResponseWriter, identifier string, body []byte) {
	response.Header().Set("Content-Type", "application/json")

	stream := controller.outputs.Stream(identifier)

	if stream != nil {
		Log.Debugf("Update stream %s : %s", identifier, string(body))
		config := stream.Config()

		err := json.Unmarshal(body, &config)
		if err != nil {
			controller.fatal(response, err)
			return
		}

		stream.Stop()
		stream.Setup(&config)
		stream.Start()

		jsonBytes, _ := json.Marshal(stream.Config())
		response.Write(jsonBytes)
	} else {
		http.Error(response, fmt.Sprintf("Stream not found: '%s'", identifier), 404)
	}
}

func (controller *HttpStreamOutputsController) Delete(response http.ResponseWriter, identifier string) {
	response.Header().Set("Content-Type", "application/json")

	stream := controller.outputs.Destroy(identifier)

	if stream != nil {
		stream.Stop()

		jsonBytes, _ := json.Marshal(stream.Config())
		response.Write(jsonBytes)
	} else {
		http.Error(response, fmt.Sprintf("Stream not found: '%s'", identifier), 404)
	}
}

func (controller *HttpStreamOutputsController) Create(response http.ResponseWriter, body []byte) {
	response.Header().Set("Content-Type", "application/json")

	Log.Debugf("Create stream : %s", string(body))

	config := NewBufferedHttpStreamOutputConfig()
	err := json.Unmarshal(body, &config)
	if err != nil {
		controller.fatal(response, err)
		return
	}

	stream := controller.outputs.Create(&config)
	if stream != nil {
		stream.Init()
		stream.Start()

		jsonBytes, _ := json.Marshal(stream.Config())
		response.Write(jsonBytes)
	} else {
		http.Error(response, "Can't create Stream", 500)
	}
}

func (controller *HttpStreamOutputsController) fatal(response http.ResponseWriter, err error) {
	http.Error(response, fmt.Sprintf("Unknown error: %v", err), 500)
}
