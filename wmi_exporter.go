package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	pushGatewayURL = "http://127.0.0.1:9091/metrics/job/aida64_sensors"
	wmicNamespace  = `\\root\WMI`
	wmicPath       = `AIDA64_SensorValues`
)

func init() {
	file, err := os.OpenFile("wexpo.log", syscall.O_CREAT|syscall.O_RDWR|syscall.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
}

func main() {
	for {
		metrics, err := getWMICMetrics()
		if err != nil {
			log.Printf("Error occur while getting WMIC metrics: %v\n", err)
		} else {
			err = pushToPushGateway(metrics)
			if err != nil {
				log.Printf("Error occur while pushing WMIC metrics: %v\n", err)
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func pushToPushGateway(metrics string) error {
	resp, err := http.Post(pushGatewayURL, "text/plain", strings.NewReader(metrics))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Panicf("response status is: %s\n", resp.Status)
	} else {
		log.Printf("Metrics pushed to pushgateway successfully")
	}
	return nil
}

func getWMICMetrics() (string, error) {
	cmd := exec.Command("wmic", "/namespace:"+wmicNamespace, "path", wmicPath, "GET", "/FORMAT:csv")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var output bytes.Buffer
	cmd.Stdout = &output
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return parseWMICMetrics(output.String()), nil
}

func parseWMICMetrics(rawStr string) string {
	scanner := bufio.NewScanner(strings.NewReader(rawStr))
	var metrics strings.Builder
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	scanner.Scan()
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")
		fieldLen := len(fields)
		if fieldLen >= 5 {
			value := strings.TrimSpace(fields[4])
			// check if the value is float
			_, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			label := formatLabel(fields[2], fields[3], re)
			metric := fmt.Sprintf("%s %s\n", label, value)
			metrics.WriteString(metric)
		}
	}
	return metrics.String()
}

// T: Temperatures
// V: Voltage
// S: System
// F: Fan
// P: Power
// D: ?
func formatLabel(label string, stype string, re *regexp.Regexp) string {
	prefix := strings.ToLower(wmicPath + "_" + strings.Trim(re.ReplaceAllString(label, "_"), "_"))
	switch stype {
	case "T":
		return prefix + "_temperature"
	case "V":
		return prefix + "_voltage"
	case "S":
		return prefix + "_system"
	case "F":
		return prefix + "_fan"
	case "P":
		return prefix + "_power"
	default:
		return prefix
	}
}
