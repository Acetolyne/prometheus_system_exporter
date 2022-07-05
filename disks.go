//go:build !windows

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

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

	//Disk percentage collector
	if setting_disk_percentage {
		fmt.Println("Disk space percentage collector enabled")

		out, err := exec.Command("df", "-h").Output()
		if err != nil {
			fmt.Println(err)
		}
		dfout := strings.Split(string(out[:]), "\n")
		for _, d := range dfout {
			//@todo make prefix /dev/X come from config file and be a list
			//@todo do below for each value in list
			if strings.HasPrefix(d, "/dev/vda") {
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
				err = prometheus.DefaultRegisterer.Register(tempcollector)
				if err != nil {
					fmt.Println("RegistrationError", err)
				}
				percent, err := strconv.ParseFloat(hd.percentage, 32)
				if err != nil {
					fmt.Println("Error converting disk used percentage to float", err)
				}
				tempcollector.Set(percent)
				//for thisdev := range prometheus
			}
			//fmt.Println(d)
		}
	}
}