package broadcast

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strings"
)

type CommandConfig struct {
	File string `json:"-"`

	HttpServer HttpServerConfig
	Log        LogConfig
	Metrics    MetricsConfig
}

func LoadConfig(file string, config interface{}) error {
	if file == "" {
		return nil
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		Log.Printf("Config file not found: %s", file)
		return nil
	}

	Log.Printf("Read config file: %s", file)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, config)
	if err != nil {
		return err
	}

	return nil
}

func SaveConfig(file string, config interface{}) error {
	if file == "" {
		return nil
	}

	Log.Printf("Save config file: %s", file)

	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, jsonBytes, 0640)
	if err != nil {
		return err
	}

	return nil
}

func (config *CommandConfig) BaseFlags(flags *flag.FlagSet) {
	flags.StringVar(&config.File, "config", "", "The config file to be loaded on startup")

	config.HttpServer.Flags(flags, "http")
	config.Log.Flags(flags, "log")
	config.Metrics.Flags(flags, "metrics")
}

func (config *CommandConfig) BaseApply(httpServer *HttpServer) {
	config.HttpServer.Apply(httpServer)
	config.Log.Apply()
	config.Metrics.Apply()
}

type UDPClientConfig struct {
	CommandConfig

	Alsa AlsaInputConfig
	Udp  UDPOutputConfig
}

func (config *UDPClientConfig) Flags(flags *flag.FlagSet) {
	config.BaseFlags(flags)

	config.Alsa.Flags(flags, "alsa")
	config.Udp.Flags(flags, "udp")
}

func (config *UDPClientConfig) Apply(alsaInput *AlsaInput, udpOutput *UDPOutput, httpServer *HttpServer) {
	config.BaseApply(httpServer)

	config.Alsa.Apply(alsaInput)
	config.Udp.Apply(udpOutput)
}

type UDPServerConfig struct {
	CommandConfig

	Alsa AlsaOutputConfig
	Udp  UDPInputConfig
}

func (config *UDPServerConfig) Flags(flags *flag.FlagSet) {
	config.BaseFlags(flags)

	config.Alsa.Flags(flags, "alsa")
	config.Udp.Flags(flags, "udp")
}

func (config *UDPServerConfig) Apply(alsaOutput *AlsaOutput, udpInput *UDPInput, httpServer *HttpServer) {
	config.BaseApply(httpServer)

	config.Alsa.Apply(alsaOutput)
	config.Udp.Apply(udpInput)
}

type LogConfig struct {
	Debug  bool
	Syslog bool
}

func (config *LogConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.BoolVar(&config.Debug, strings.Join([]string{prefix, "debug"}, "-"), false, "Enable debug messages")
	flags.BoolVar(&config.Syslog, strings.Join([]string{prefix, "syslog"}, "-"), false, "Redirect messages to syslog")
}

func (config *LogConfig) Apply() {
	Log.Debug = config.Debug
	Log.Syslog = config.Syslog
}

type BackupConfig struct {
	CommandConfig

	Alsa  AlsaInputConfig
	Files TimedFileOutputConfig
}

func (config *BackupConfig) Flags(flags *flag.FlagSet) {
	config.BaseFlags(flags)

	config.Alsa.Flags(flags, "alsa")
	config.Files.Flags(flags, "files")
}

func (config *BackupConfig) Apply(alsaInput *AlsaInput, timedFileOutput *TimedFileOutput, httpServer *HttpServer) {
	config.BaseApply(httpServer)

	config.Alsa.Apply(alsaInput)
	config.Files.Apply(timedFileOutput)

	timedFileOutput.SetSampleRate(alsaInput.SampleRate)
	timedFileOutput.SetChannelCount(alsaInput.Channels)
}
