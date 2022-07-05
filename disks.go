//go:build !windows

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type harddrive struct {
	device     string
	size       string
	used       string
	available  string
	percentage string
	mountpoint string
}

func ExportDisks() {
	setting_disk_percentage, _ := Cfg.Section("collectors").Key("disk_space_percentage").Bool()
	setting_disk_prefixes := Cfg.Section("drives").Key("allowed_prefixes").String()

	//Disk percentage collector
	if setting_disk_percentage {
		fmt.Println("Disk space percentage collector enabled")

		//@todo make the infinite loop a go routine that passes in the current tempcollector object
		out, err := exec.Command("df", "-h").Output()
		if err != nil {
			fmt.Println(err)
		}
		dfout := strings.Split(string(out[:]), "\n")
		prefixes := strings.Split(string(setting_disk_prefixes[:]), ",")
		for _, d := range dfout {
			for _, pre := range prefixes {
				//@todo make prefix /dev/X come from config file and be a list
				//@todo do below for each value in list
				if strings.HasPrefix(d, strings.TrimSpace(pre)) {
					curdev := strings.Fields(d)
					hd := harddrive{
						curdev[0],                               //device ex /dev/sda1
						curdev[1],                               //size ex 20G or 2000MB
						curdev[2],                               //used size ex 13G or 4000MB
						curdev[3],                               //available size ex 13G or 4000MB
						strings.Replace(curdev[4], "%", "", -1), //percentage used but we strip the percent sign 20% > 20
						curdev[5],                               //mount point ex /boot/efi
					}

					tempcollector := prometheus.NewGauge(prometheus.GaugeOpts{
						Name: "hdstats_" + hd.device[strings.LastIndex(hd.device, "/")+1:],
						Help: "percentage for mountpoint " + hd.device,
					})
					//@todo construct prometheus stat name from either device or mountpoint based on the configuration file
					//register if not already registered
					err := prometheus.DefaultRegisterer.Register(tempcollector)
					if err != nil {
						fmt.Println("RegistrationError", err)
					}
					//Start go routine to keep sending stats for each device sending in current tempcollector object and hd.device
					go func(collector prometheus.Collector, dev string) {
						for {
							out, err := exec.Command("df", "-h", dev).Output()
							if err != nil {
								fmt.Println(err)
							}
							dfout := strings.Split(string(out[:]), "\n")
							cur := dfout[1]
							curdev = strings.Fields(cur)
							hd := harddrive{
								curdev[0],                               //device ex /dev/sda1
								curdev[1],                               //size ex 20G or 2000MB
								curdev[2],                               //used size ex 13G or 4000MB
								curdev[3],                               //available size ex 13G or 4000MB
								strings.Replace(curdev[4], "%", "", -1), //percentage used but we strip the percent sign 20% > 20
								curdev[5],                               //mount point ex /boot/efi
							}
							//for _, d := range dfout {
							//	fmt.Println("\n", d)
							percent, _ := strconv.ParseFloat(hd.percentage, 32)
							// if err != nil {
							// 	fmt.Println("Error converting disk used percentage to float", err)
							// }
							tempcollector.Set(percent)
							time.Sleep(time.Duration(Delay) * time.Second)
						}
					}(tempcollector, hd.device)

				}
			}
		}
	}
}
