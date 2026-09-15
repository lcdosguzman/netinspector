package scanner

func DefaultDiscoveryPorts() []int {
	return []int{22, 53, 80, 139, 443, 445, 631, 1900, 3389, 5900, 8000, 8080, 8443}
}

func serviceName(port int) string {
	names := map[int]string{
		22:   "ssh",
		53:   "dns",
		80:   "http",
		139:  "netbios",
		443:  "https",
		445:  "smb",
		631:  "ipp",
		1900: "ssdp",
		3389: "rdp",
		5900: "vnc",
		8000: "http-alt",
		8080: "http-proxy",
		8443: "https-alt",
	}
	if name, ok := names[port]; ok {
		return name
	}
	return "unknown"
}
