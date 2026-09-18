//go:build !windows

package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const hwmonPath = "/sys/class/hwmon"

type tempSensor struct {
	name  string // hwmon driver name e.g. coretemp, k10temp
	label string // e.g. "Package id 0", "Core 0", or "tempN" if unlabeled
	path  string // path to the tempN_input file
}

// discoverTempSensors walks /sys/class/hwmon looking for tempN_input files,
// since that works without requiring lm-sensors to be installed.
func discoverTempSensors() []tempSensor {
	var sensors []tempSensor

	hwmons, err := os.ReadDir(hwmonPath)
	if err != nil {
		log.Println("Unable to read hwmon path", hwmonPath, err)
		return sensors
	}

	for _, hwmon := range hwmons {
		devPath := filepath.Join(hwmonPath, hwmon.Name())

		devName := hwmon.Name()
		if nameBytes, err := os.ReadFile(filepath.Join(devPath, "name")); err == nil {
			if trimmed := strings.TrimSpace(string(nameBytes)); trimmed != "" {
				devName = trimmed
			}
		}

		entries, err := os.ReadDir(devPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name(), "temp") || !strings.HasSuffix(entry.Name(), "_input") {
				continue
			}

			inputPath := filepath.Join(devPath, entry.Name())
			label := strings.TrimSuffix(entry.Name(), "_input")

			labelPath := strings.TrimSuffix(inputPath, "_input") + "_label"
			if labelBytes, err := os.ReadFile(labelPath); err == nil {
				if trimmed := strings.TrimSpace(string(labelBytes)); trimmed != "" {
					label = trimmed
				}
			}

			sensors = append(sensors, tempSensor{
				name:  devName,
				label: label,
				path:  inputPath,
			})
		}
	}

	return sensors
}

// readTempCelsius reads a tempN_input file, which reports millidegrees Celsius.
func readTempCelsius(path string) (float64, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	millideg, err := strconv.ParseFloat(strings.TrimSpace(string(raw)), 64)
	if err != nil {
		return 0, err
	}

	return millideg / 1000, nil
}

// celsiusToFahrenheit converts a Celsius temperature to Fahrenheit.
func celsiusToFahrenheit(celsius float64) float64 {
	return celsius*9/5 + 32
}

// sanitizeMetricName makes a hwmon device/label name safe for use in a prometheus metric name.
func sanitizeMetricName(s string) string {
	s = strings.ToLower(s)
	s = strings.NewReplacer(" ", "_", "-", "_", "/", "_").Replace(s)

	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

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
		log.Println("CPU percentage collector enabled")

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

func ExportCPUTemp() {
	setting_cpu_temperature, _ := Cfg.Section("collectors").Key("cpu_temperature").Bool()
	setting_temp_in_fahrenheit, _ := Cfg.Section("collectors").Key("temp_in_fahrenheit").Bool()

	//CPU temperature collector
	if setting_cpu_temperature {
		sensors := discoverTempSensors()
		if len(sensors) == 0 {
			log.Println("CPU temperature collector enabled but no sensors found under", hwmonPath)
			return
		}

		log.Println("CPU temperature collector enabled")

		unit := "celsius"
		if setting_temp_in_fahrenheit {
			unit = "fahrenheit"
		}

		for _, sensor := range sensors {
			tempcollector := prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "cpu_temp_" + sanitizeMetricName(sensor.name) + "_" + sanitizeMetricName(sensor.label) + "_" + unit,
				Help: "current temperature in " + unit + " for sensor " + sensor.name + " " + sensor.label,
			})
			//register if not already registered
			err := prometheus.DefaultRegisterer.Register(tempcollector)
			if err != nil {
				log.Println("RegistrationError", err)
				continue
			}

			//Start go routine to keep sending stats for each sensor
			go func(collector prometheus.Gauge, path string, fahrenheit bool) {
				for {
					celsius, err := readTempCelsius(path)
					if err != nil {
						log.Println("Error reading temperature from", path, err)
					} else if fahrenheit {
						collector.Set(celsiusToFahrenheit(celsius))
					} else {
						collector.Set(celsius)
					}
					time.Sleep(time.Duration(Delay) * time.Second)
				}
			}(tempcollector, sensor.path, setting_temp_in_fahrenheit)
		}
	}
}
