package broadcast

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	metrics "github.com/tryphon/go-metrics"
	"net/http"
)

type HttpServer struct {
	Bind                   string
	SoundMeterAudioHandler *SoundMeterAudioHandler
}

func (server *HttpServer) Init() error {
	if server.Bind != "" {
		http.HandleFunc("/metrics.json", server.metricsJSON)
		http.HandleFunc("/soundmeter.json", server.soundMeterJSON)
		http.Handle("/soundmeter.ws", websocket.Handler(server.soundMeterWebSocket))

		go http.ListenAndServe(server.Bind, nil)
	}
	return nil
}

func (server *HttpServer) metricsJSON(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "*")

	jsonBytes, _ := json.Marshal(metrics.DefaultRegistry)
	response.Write(jsonBytes)
}

func (server *HttpServer) soundMeterJSON(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "*")

	jsonBytes, _ := json.Marshal(server.SoundMeterAudioHandler)
	response.Write(jsonBytes)
}

func (server *HttpServer) soundMeterWebSocket(webSocket *websocket.Conn) {
	Log.Debugf("New SoundMeter websocket connection")

	receiver := server.SoundMeterAudioHandler.NewReceiver()
	defer receiver.Close()

	go func() {
		for metrics := range receiver.Channel {
			jsonBytes, _ := json.Marshal(metrics)
			err := websocket.Message.Send(webSocket, string(jsonBytes))
			if err != nil {
				Log.Debugf("Can't send websocket message: %v", err)
				break
			}
		}
	}()

	for {
		var message string
		err := websocket.Message.Receive(webSocket, &message)
		if err != nil {
			break
		}
	}

	Log.Debugf("Close SoundMeter websocket connection")
}