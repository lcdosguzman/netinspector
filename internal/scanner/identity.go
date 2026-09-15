package scanner

import (
	"fmt"
	"strconv"
	"strings"
)

var ouiVendors = map[string]string{
	"3C:06:30": "Apple, Inc.",
	"78:66:9D": "Hui Zhou Gaoshengda Technology Co., Ltd.",
	"C8:94:02": "Chongqing Fugui Electronics Co., Ltd.",

	// Common consumer device vendors. This is intentionally a seed table; a full
	// IEEE OUI registry import can replace it later.
	"00:17:F2": "Apple, Inc.",
	"00:1C:B3": "Apple, Inc.",
	"3C:22:FB": "Apple, Inc.",
	"70:70:0D": "Apple, Inc.",
	"A4:83:E7": "Apple, Inc.",
	"AC:BC:32": "Apple, Inc.",
	"F0:DB:E2": "Apple, Inc.",
	"00:1A:11": "Google, Inc.",
	"3C:5A:B4": "Google, Inc.",
	"F4:F5:D8": "Google, Inc.",
	"00:1A:79": "Samsung Electronics Co., Ltd.",
	"F8:4E:73": "Samsung Electronics Co., Ltd.",
	"00:09:DF": "Vestel Elektronik San ve Tic. A.S.",
	"00:0C:E7": "MediaTek Inc.",
	"FC:F1:52": "Sony Interactive Entertainment Inc.",
	"B8:27:EB": "Raspberry Pi Foundation",
	"DC:A6:32": "Raspberry Pi Trading Ltd.",
	"24:77:03": "Intel Corporate",
	"84:15:D3": "Network equipment vendor",
}

func enrichDevice(device Device) Device {
	device.Vendor = vendorName(device.MAC, device.Vendor)
	device.Hints = appendUnique(device.Hints, identificationHints(device)...)

	inferredType := inferDeviceType(device)
	if inferredType != DeviceUnknown {
		device.Type = inferredType
	} else if device.Type == "" {
		device.Type = DeviceUnknown
	}

	return device
}

func vendorName(mac string, fallback string) string {
	if fallback != "" {
		return fallback
	}
	if mac == "" {
		return ""
	}
	if isLocallyAdministered(mac) {
		return "Private/randomized MAC"
	}

	oui := macOUI(mac)
	if vendor, ok := ouiVendors[oui]; ok {
		return vendor
	}
	if oui != "" {
		return fmt.Sprintf("Unknown vendor (%s)", oui)
	}
	return ""
}

func identificationHints(device Device) []string {
	hints := make([]string, 0, 3)

	if device.MAC != "" && isLocallyAdministered(device.MAC) {
		hints = append(hints, "La MAC es privada/randomizada; suele pasar en teléfonos o laptops con privacidad Wi-Fi.")
	}
	if device.Vendor != "" && !strings.HasPrefix(device.Vendor, "Unknown vendor") && device.Vendor != "Private/randomized MAC" {
		hints = append(hints, "Fabricante estimado por OUI de la MAC.")
	}
	if strings.Contains(strings.ToLower(device.Vendor), "gaoshengda") {
		hints = append(hints, "El fabricante suele aparecer como módulo Wi-Fi en TVs, IoT o dispositivos multimedia.")
	}
	if hasPort(device, 8443) || hasPort(device, 8000) || hasPort(device, 8008) || hasPort(device, 8060) {
		hints = append(hints, "Puertos compatibles con dispositivos multimedia, TV, cast o panel web embebido.")
	}
	if hasPort(device, 631) {
		hints = append(hints, "Puerto IPP detectado; posible impresora.")
	}
	if hasPort(device, 53) && hasPort(device, 80) {
		hints = append(hints, "DNS y HTTP abiertos; posible router o gateway.")
	}

	return hints
}

func inferDeviceType(device Device) DeviceType {
	vendor := strings.ToLower(device.Vendor)
	hostname := strings.ToLower(device.Hostname)

	if hasPort(device, 53) && hasPort(device, 80) {
		return DeviceRouter
	}
	if hasPort(device, 631) {
		return DevicePrinter
	}
	if strings.Contains(hostname, "tv") {
		return DeviceTV
	}
	if strings.Contains(hostname, "iphone") || strings.Contains(hostname, "ipad") || strings.Contains(hostname, "android") {
		return DeviceMobile
	}
	if strings.Contains(vendor, "apple") || strings.Contains(vendor, "intel") || strings.Contains(hostname, "macbook") || strings.Contains(hostname, "desktop") {
		return DeviceDesktop
	}
	if strings.Contains(vendor, "samsung") || strings.Contains(vendor, "sony") || strings.Contains(vendor, "vestel") || hasPort(device, 8000) || hasPort(device, 8008) || hasPort(device, 8060) || hasPort(device, 8443) {
		return DeviceTV
	}
	if strings.Contains(vendor, "gaoshengda") {
		return DeviceIoT
	}
	if strings.Contains(vendor, "private/randomized") {
		return DeviceMobile
	}

	return DeviceUnknown
}

func macOUI(mac string) string {
	normalized := normalizedMACParts(mac)
	if len(normalized) < 3 {
		return ""
	}
	return strings.Join(normalized[:3], ":")
}

func isLocallyAdministered(mac string) bool {
	parts := normalizedMACParts(mac)
	if len(parts) == 0 {
		return false
	}

	firstOctet, err := strconv.ParseUint(parts[0], 16, 8)
	if err != nil {
		return false
	}
	return firstOctet&0x02 != 0
}

func normalizedMACParts(mac string) []string {
	fields := strings.FieldsFunc(mac, func(r rune) bool {
		return r == ':' || r == '-' || r == '.'
	})
	if len(fields) == 1 && len(fields[0]) >= 6 {
		value := strings.ToUpper(fields[0])
		return []string{value[0:2], value[2:4], value[4:6]}
	}
	if len(fields) < 3 {
		return nil
	}

	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) == 1 {
			field = "0" + field
		}
		if len(field) != 2 {
			return nil
		}
		parts = append(parts, strings.ToUpper(field))
	}
	return parts
}

func hasPort(device Device, portNumber int) bool {
	for _, port := range device.Ports {
		if port.Number == portNumber {
			return true
		}
	}
	return false
}
