package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"gopkg.in/ini.v1"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// func init() {
// 	// Metrics have to be registered to be exposed:
// 	prometheus.MustRegister(diskPerc)
// }

const configDir = "/etc/prometheus-system-exporter"
const configPath = configDir + "/settings.ini"
const logPath = configDir + "/error.log"

const defaultConfig = `[general]
ip = 0.0.0.0
port = 9091

[collectors]
delay_seconds = 15
disk_space_percentage = true
memory_percentage = true
cpu_percentage = true
cpu_temperature = true
temp_in_fahrenheit = false

[drives]
allowed_prefixes = /dev/sda,/dev/vda
`

var Cfg *ini.File
var Delay int

func init() {
	ensureDefaultConfig()
	setupLogging()

	var err error
	Cfg, err = ini.Load(configPath)
	if err != nil {
		log.Println("Fail to read config file:", err)
		os.Exit(1)
	}

	Delay, _ = Cfg.Section("collectors").Key("delay_seconds").Int()
}

// setupLogging redirects the standard logger's output to the error log file
// so all collector output ends up in one place. Falls back to stderr if the
// file can't be opened, e.g. because ensureDefaultConfig couldn't create configDir.
func setupLogging() {
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Unable to open log file", logPath, err)
		return
	}

	log.SetOutput(logFile)
}

// ensureDefaultConfig creates the config directory and a default settings.ini
// if they don't already exist. Currently only implemented for Linux.
func ensureDefaultConfig() {
	if runtime.GOOS != "linux" {
		return
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Println("Unable to create config directory", configDir, err)
			return
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
			fmt.Println("Unable to create default config file", configPath, err)
		}
	}
}

//@todo show example of adding application as a service for Linux and Windows
//@todo make matching exporter files for windows
//@todo make update argument that will automatically update the agent from master git branch
//@todo auto build binaries with github webhooks
func main() {

	//Get the settings from config file
	//fmt.Println("App Mode:", cfg.Section("").Key("app_mode").String())
	log.Println("Allowed Drive Prefixes", Cfg.Section("drives").Key("allowed_prefixes").String())
	go ExportDisks()
	go ExportMemory()
	go ExportCPU()
	go ExportCPUTemp()

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(Cfg.Section("general").Key("ip").String()+":"+Cfg.Section("general").Key("port").String(), nil))
}
