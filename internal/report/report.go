package report

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/lcdosguzman/lansweepgo/internal/scanner"
)

func JSON(scan scanner.ScanResult) ([]byte, error) {
	return json.MarshalIndent(scan, "", "  ")
}

func CSV(scan scanner.ScanResult) ([]byte, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	if err := writer.Write([]string{
		"scan_id",
		"mode",
		"network",
		"started_at",
		"ended_at",
		"device_ip",
		"hostname",
		"mac",
		"vendor",
		"type",
		"is_active",
		"latency_ms",
		"ports",
		"hints",
	}); err != nil {
		return nil, err
	}

	for _, device := range scan.Devices {
		if err := writer.Write([]string{
			scan.ID,
			scan.Mode,
			scan.Network,
			scan.StartedAt.Format(timeFormat),
			scan.EndedAt.Format(timeFormat),
			device.IP,
			device.Hostname,
			device.MAC,
			device.Vendor,
			string(device.Type),
			strconv.FormatBool(device.IsActive),
			strconv.FormatInt(device.LatencyMS, 10),
			formatPorts(device.Ports),
			strings.Join(device.Hints, " | "),
		}); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

const timeFormat = "2006-01-02T15:04:05Z07:00"

func formatPorts(ports []scanner.Port) string {
	formatted := make([]string, 0, len(ports))
	for _, port := range ports {
		formatted = append(formatted, port.Protocol+"/"+strconv.Itoa(port.Number)+" "+port.ServiceName)
	}
	return strings.Join(formatted, " | ")
}
