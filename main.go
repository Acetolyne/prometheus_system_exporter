package main

import (
	"fmt"
	"log"
	"net/http"

	"gopkg.in/ini.v1"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// func init() {
// 	// Metrics have to be registered to be exposed:
// 	prometheus.MustRegister(diskPerc)
// }

//@todo if there is no config file then create one with defaults
var Cfg, _ = ini.Load("/etc/prometheus_system_exporter/settings.ini") // initialize a CFG
var Delay, _ = Cfg.Section("collectors").Key("delay_seconds").Int()

// if err != nil {
// 	fmt.Printf("Fail to read file: %v", err)
// 	os.Exit(1)
// }

//@todo register the binary as a service
//@todo make matching exporter files for windows
//@todo make update argument that will automatically update the agent from master git branch
//@todo auto build binaries with github webhooks
func main() {

	//Get the settings from config file
	//fmt.Println("App Mode:", cfg.Section("").Key("app_mode").String())
	fmt.Println("Allowed Drive Prefixes", Cfg.Section("drives").Key("allowed_prefixes").String())
	go ExportDisks()
	go ExportMemory()

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())
	//@todo make port come from config file
	//@todo make interface also come from port so we can handle locking to ip range in case of multiple interfaces
	log.Fatal(http.ListenAndServe(":9091", nil))
}
