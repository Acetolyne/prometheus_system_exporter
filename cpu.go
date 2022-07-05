//go:build !windows

package main

import (
	"fmt"
)

// type memory struct {
// 	total     string
// 	used      string
// 	free      string
// 	shared    string
// 	cache     string
// 	available string
// }

func ExportCPU() {
	//@todo figure out where to get the stats for CPU  /proc/stat?
	setting_cpu_percentage, _ := Cfg.Section("collectors").Key("cpu_percentage").Bool()

	//Disk percentage collector
	if setting_cpu_percentage {
		fmt.Println("CPU percentage collector enabled")

		// out, err := exec.Command("free").Output()
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// dfout := strings.Split(string(out[:]), "\n")
		// for _, d := range dfout {
		// 	if strings.HasPrefix(d, "Mem:") {
		// 		curdev := strings.Fields(d)
		// mem := memory{
		// 	curdev[1], //toal memory
		// 	curdev[2], //used memory
		// 	curdev[3], //free memory
		// 	curdev[4], //shared memory
		// 	curdev[5], //buffer/cache
		// 	curdev[6], //available memory
		// }
		// tempcollector := prometheus.NewGauge(prometheus.GaugeOpts{
		// 	Name: "cpu_used_percentage",
		// 	Help: "current cpu used in percentage",
		// })
		//register if not already registered
		// err = prometheus.DefaultRegisterer.Register(tempcollector)
		// if err != nil {
		// 	fmt.Println("RegistrationError", err)
		// }
		// used, _ := strconv.ParseFloat(mem.used, 64)
		// total, _ := strconv.ParseFloat(mem.total, 64)
		// large_percent := used / total * 10000
		// fmt.Println(large_percent)
		// small_percent := int(math.Round(large_percent))
		// percent := float64(small_percent) / 100
		// fmt.Println(percent)
		// tempcollector.Set(percent)
	}
}

//}
//}
