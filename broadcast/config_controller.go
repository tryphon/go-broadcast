package broadcast

import (
	"net/http"
)

type ConfigManager interface {
	ConfigToJSON() []byte
	SaveConfig() error
}

type ConfigController struct {
	manager ConfigManager
}

func NewConfigController(manager ConfigManager) *ConfigController {
	return &ConfigController{manager}
}

func (controller *ConfigController) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	// Log.Debugf("ConfigController request: %s", request)

	path := request.URL.Path

	switch {
	case path == "/config.json" && request.Method == "GET":
		controller.Show(response)
	case path == "/config/save.json" && request.Method == "POST":
		controller.Save(response)
	}
}

func (controller *ConfigController) Show(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.Write(controller.manager.ConfigToJSON())
}

func (controller *ConfigController) Save(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	controller.manager.SaveConfig()
}
