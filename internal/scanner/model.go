package scanner

import "time"

type ScanResult struct {
	ID        string      `json:"id"`
	Mode      string      `json:"mode"`
	Network   string      `json:"network"`
	StartedAt time.Time   `json:"startedAt"`
	EndedAt   time.Time   `json:"endedAt"`
	Devices   []Device    `json:"devices"`
	Events    []ScanEvent `json:"events"`
}

type Device struct {
	IP        string     `json:"ip"`
	Hostname  string     `json:"hostname,omitempty"`
	MAC       string     `json:"mac,omitempty"`
	Vendor    string     `json:"vendor,omitempty"`
	Type      DeviceType `json:"type"`
	Hints     []string   `json:"hints,omitempty"`
	LatencyMS int64      `json:"latencyMs"`
	IsActive  bool       `json:"isActive"`
	Ports     []Port     `json:"ports,omitempty"`
}

type DeviceType string

const (
	DeviceRouter  DeviceType = "ROUTER"
	DeviceDesktop DeviceType = "DESKTOP"
	DeviceMobile  DeviceType = "MOBILE"
	DeviceTV      DeviceType = "TV"
	DevicePrinter DeviceType = "PRINTER"
	DeviceIoT     DeviceType = "IOT"
	DeviceUnknown DeviceType = "UNKNOWN"
)

type Port struct {
	Number      int    `json:"number"`
	Protocol    string `json:"protocol"`
	ServiceName string `json:"serviceName"`
}

type ScanEvent struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	DeviceIP  string    `json:"deviceIp,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
