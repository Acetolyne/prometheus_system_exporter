package main

import (
	"fmt"
	"log"
	"net/http"

	"gopkg.in/ini.v1"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// var (
// 	diskPerc = prometheus.NewGauge(prometheus.GaugeOpts{
// 		Name: "disk_space_percentage",
// 		Help: "percentage for each disk mount point",
// 	})
// hdFailures = prometheus.NewCounterVec(
// 	prometheus.CounterOpts{
// 		Name: "hd_errors_total",
// 		Help: "Number of hard-disk errors.",
// 	},
// 	[]string{"device"},
// )
// )

// func init() {
// 	// Metrics have to be registered to be exposed:
// 	prometheus.MustRegister(diskPerc)
// }

//@todo if there is no config file then create one with defaults
var Cfg, _ = ini.Load("/etc/prometheus_system_exporter/settings.ini") // initialize a CFG
// if err != nil {
// 	fmt.Printf("Fail to read file: %v", err)
// 	os.Exit(1)
// }

//@todo register the binary as a service
//@todo make matching exporter files for windows
func main() {

	//Get the settings from config file
	//fmt.Println("App Mode:", cfg.Section("").Key("app_mode").String())
	fmt.Println("Allowed Drive Prefixes", Cfg.Section("drives").Key("allowed_prefixes").String())
	//@todo check if these need to be in an infinite loop to update
	go ExportDisks()
	go ExportMemory()

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())
	//@todo make port come from config file
	//@todo make interface also come from port so we can handle locking to ip range in case of multiple interfaces
	log.Fatal(http.ListenAndServe(":9091", nil))
}
