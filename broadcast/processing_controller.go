package broadcast

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ProcessingController struct {
	processing *Processing
}

func NewProcessingController(processing *Processing) (controller *ProcessingController) {
	return &ProcessingController{processing: processing}
}

func (controller *ProcessingController) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var body []byte
	if request.Body != nil {
		var err error
		body, err = ioutil.ReadAll(request.Body)
		if err != nil {
			controller.fatal(response, err)
			return
		}
	}

	switch request.Method {
	case "GET":
		controller.Show(response)
	case "PUT":
		controller.Update(response, body)
	}
}

func (controller *ProcessingController) Show(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")

	Log.Debugf("Retrieve processing")

	jsonBytes, err := json.Marshal(controller.processing.Config())
	if err == nil {
		response.Write(jsonBytes)
	} else {
		controller.fatal(response, err)
	}
}

func (controller *ProcessingController) Update(response http.ResponseWriter, body []byte) {
	response.Header().Set("Content-Type", "application/json")

	Log.Debugf("Update processing %s", string(body))

	config := controller.processing.Config()

	err := json.Unmarshal(body, &config)
	if err != nil {
		controller.fatal(response, err)
		return
	}

	controller.processing.Setup(config)

	jsonBytes, _ := json.Marshal(config)
	response.Write(jsonBytes)
}

func (controller *ProcessingController) fatal(response http.ResponseWriter, err error) {
	http.Error(response, fmt.Sprintf("Unknown error: %v", err), 500)
}
