//go:build !windows

package main

import (
	"errors"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const hwmonPath = "/sys/class/hwmon"
const procStatPath = "/proc/stat"

type tempSensor struct {
	name  string // hwmon driver name e.g. coretemp, k10temp
	label string // e.g. "Package id 0", "Core 0", or "tempN" if unlabeled
	path  string // path to the tempN_input file
	id    string // hwmon device dir e.g. "hwmon0", used to disambiguate name+label collisions across sockets
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
				id:    hwmon.Name(),
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

// readCPUStat reads /proc/stat and returns the total and idle jiffies from the
// aggregate "cpu" line, along with the number of "cpuN" lines which is the core count.
func readCPUStat() (total float64, idle float64, cores int, err error) {
	raw, err := os.ReadFile(procStatPath)
	if err != nil {
		return 0, 0, 0, err
	}

	foundTotal := false
	for _, line := range strings.Split(string(raw), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || !strings.HasPrefix(fields[0], "cpu") {
			continue
		}

		if fields[0] != "cpu" {
			cores++
			continue
		}

		// user nice system idle iowait irq softirq steal - guest time is already included in user/nice
		for i, field := range fields[1:] {
			if i >= 8 {
				break
			}
			jiffies, err := strconv.ParseFloat(field, 64)
			if err != nil {
				return 0, 0, 0, err
			}
			total += jiffies
			if i == 3 || i == 4 { // idle + iowait
				idle += jiffies
			}
		}
		foundTotal = true
	}

	if !foundTotal {
		return 0, 0, 0, errors.New("no aggregate cpu line found in " + procStatPath)
	}

	return total, idle, cores, nil
}

func ExportCPU() {
	setting_cpu_percentage, _ := Cfg.Section("collectors").Key("cpu_percentage").Bool()

	//CPU percentage collector
	if setting_cpu_percentage {
		log.Println("CPU percentage collector enabled")

		percentcollector := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cpu_percentage_used",
			Help: "current cpu used in percentage",
		})
		corecollector := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cpu_core_count",
			Help: "number of cpu cores",
		})
		//register if not already registered
		for _, collector := range []prometheus.Collector{percentcollector, corecollector} {
			err := prometheus.DefaultRegisterer.Register(collector)
			if err != nil {
				log.Println("RegistrationError", err)
			}
		}

		prevTotal, prevIdle, cores, err := readCPUStat()
		if err != nil {
			log.Println("Error reading cpu stats from", procStatPath, err)
		}
		corecollector.Set(float64(cores))

		for {
			time.Sleep(time.Duration(Delay) * time.Second)

			total, idle, cores, err := readCPUStat()
			if err != nil {
				log.Println("Error reading cpu stats from", procStatPath, err)
				continue
			}
			corecollector.Set(float64(cores))

			//usage is the share of non-idle time since the previous sample
			totalDelta := total - prevTotal
			idleDelta := idle - prevIdle
			if totalDelta > 0 {
				percentcollector.Set(math.Round((totalDelta-idleDelta)/totalDelta*10000) / 100)
			}
			prevTotal, prevIdle = total, idle
		}
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

		seenNames := make(map[string]bool)

		for _, sensor := range sensors {
			metricName := "cpu_temp_" + sanitizeMetricName(sensor.name) + "_" + sanitizeMetricName(sensor.label) + "_" + unit
			if seenNames[metricName] {
				// Multiple hwmon devices can share the same driver name and label,
				// e.g. per-socket "coretemp" devices both reporting "Core 0" -
				// disambiguate using the hwmon device id to avoid a registration collision.
				metricName = metricName + "_" + sanitizeMetricName(sensor.id)
			}
			seenNames[metricName] = true

			tempcollector := prometheus.NewGauge(prometheus.GaugeOpts{
				Name: metricName,
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
