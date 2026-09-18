package api

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	"github.com/lcdosguzman/lansweepgo/internal/demo"
	networkv1 "github.com/lcdosguzman/lansweepgo/internal/gen/network/v1"
	"github.com/lcdosguzman/lansweepgo/internal/network"
	"github.com/lcdosguzman/lansweepgo/internal/scanner"
	appservice "github.com/lcdosguzman/lansweepgo/internal/service"
	"github.com/lcdosguzman/lansweepgo/internal/storage"
)

type networkService struct {
	repository storage.ScanRepository
	scans      appservice.ScanService
}

func newNetworkService(repository storage.ScanRepository, scans appservice.ScanService) networkService {
	return networkService{
		repository: repository,
		scans:      scans,
	}
}

func (service networkService) GetLocalNetwork(_ context.Context, _ *connect.Request[networkv1.GetLocalNetworkRequest]) (*connect.Response[networkv1.GetLocalNetworkResponse], error) {
	local, err := network.DetectLocalNetwork()
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}

	return connect.NewResponse(&networkv1.GetLocalNetworkResponse{
		InterfaceName: local.InterfaceName,
		Ip:            local.IP,
		Cidr:          local.CIDR,
	}), nil
}

func (service networkService) StartScan(ctx context.Context, request *connect.Request[networkv1.StartScanRequest]) (*connect.Response[networkv1.ScanResult], error) {
	result, err := service.scans.StartScan(ctx, appservice.StartScanRequest{
		Mode: scanModeFromProto(request.Msg.Mode),
		CIDR: request.Msg.Cidr,
	})
	if err != nil {
		return nil, connectScanServiceError(err)
	}
	return connect.NewResponse(scanResultToProto(result)), nil
}

func (service networkService) ListScans(ctx context.Context, request *connect.Request[networkv1.ListScansRequest]) (*connect.Response[networkv1.ListScansResponse], error) {
	scans, err := service.repository.ListScans(ctx, storage.ListScansFilter{
		Query: request.Msg.Query,
		Limit: int(request.Msg.Limit),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &networkv1.ListScansResponse{
		Scans: make([]*networkv1.ScanSummary, 0, len(scans)),
	}
	for _, scan := range scans {
		response.Scans = append(response.Scans, scanSummaryToProto(scan))
	}

	return connect.NewResponse(response), nil
}

func (service networkService) GetScan(ctx context.Context, request *connect.Request[networkv1.GetScanRequest]) (*connect.Response[networkv1.ScanResult], error) {
	result, err := service.repository.GetScan(ctx, request.Msg.Id)
	if err != nil {
		if err == storage.ErrScanNotFound {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(scanResultToProto(result)), nil
}

func (service networkService) StreamScan(ctx context.Context, request *connect.Request[networkv1.StreamScanRequest], stream *connect.ServerStream[networkv1.ScanEvent]) error {
	result := demo.NewScan()
	if request.Msg.ScanId != "" && request.Msg.ScanId != result.ID {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("scan %q not found", request.Msg.ScanId))
	}

	for _, event := range result.Events {
		if err := stream.Send(scanEventToProto(event)); err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	return nil
}

func (service networkService) InspectDevice(_ context.Context, request *connect.Request[networkv1.InspectDeviceRequest]) (*connect.Response[networkv1.InspectDeviceResponse], error) {
	return connect.NewResponse(&networkv1.InspectDeviceResponse{
		Ip:          request.Msg.Ip,
		OpenPorts:   nil,
		EstimatedOs: "Unknown",
	}), nil
}

func scanResultToProto(result scanner.ScanResult) *networkv1.ScanResult {
	devices := make([]*networkv1.Device, 0, len(result.Devices))
	for _, device := range result.Devices {
		devices = append(devices, deviceToProto(device))
	}

	events := make([]*networkv1.ScanEvent, 0, len(result.Events))
	for _, event := range result.Events {
		events = append(events, scanEventToProto(event))
	}

	return &networkv1.ScanResult{
		Id:        result.ID,
		Mode:      scanModeToProto(result.Mode),
		Network:   result.Network,
		StartedAt: result.StartedAt.UnixMilli(),
		EndedAt:   result.EndedAt.UnixMilli(),
		Devices:   devices,
		Events:    events,
	}
}

func scanSummaryToProto(summary storage.ScanSummary) *networkv1.ScanSummary {
	return &networkv1.ScanSummary{
		Id:          summary.ID,
		Mode:        scanModeToProto(summary.Mode),
		Network:     summary.Network,
		StartedAt:   summary.StartedAt.UnixMilli(),
		EndedAt:     summary.EndedAt.UnixMilli(),
		DeviceCount: int32(summary.DeviceCount),
	}
}

func deviceToProto(device scanner.Device) *networkv1.Device {
	ports := make([]*networkv1.PortInfo, 0, len(device.Ports))
	for _, port := range device.Ports {
		ports = append(ports, &networkv1.PortInfo{
			Port:        int32(port.Number),
			Protocol:    port.Protocol,
			ServiceName: port.ServiceName,
		})
	}

	return &networkv1.Device{
		Ip:         device.IP,
		Mac:        device.MAC,
		Vendor:     device.Vendor,
		Hostname:   device.Hostname,
		LatencyMs:  device.LatencyMS,
		IsActive:   device.IsActive,
		DeviceType: deviceTypeToProto(device.Type),
		Ports:      ports,
		Hints:      device.Hints,
	}
}

func scanEventToProto(event scanner.ScanEvent) *networkv1.ScanEvent {
	return &networkv1.ScanEvent{
		EventType: scanEventTypeToProto(event.Type),
		Message:   event.Message,
		Timestamp: event.Timestamp.UnixMilli(),
	}
}

func scanModeToProto(mode string) networkv1.ScanMode {
	switch strings.ToUpper(mode) {
	case "REAL":
		return networkv1.ScanMode_SCAN_MODE_REAL
	case "DEMO":
		return networkv1.ScanMode_SCAN_MODE_DEMO
	default:
		return networkv1.ScanMode_SCAN_MODE_UNSPECIFIED
	}
}

func scanModeFromProto(mode networkv1.ScanMode) string {
	switch mode {
	case networkv1.ScanMode_SCAN_MODE_REAL:
		return appservice.ScanModeReal
	case networkv1.ScanMode_SCAN_MODE_DEMO, networkv1.ScanMode_SCAN_MODE_UNSPECIFIED:
		return appservice.ScanModeDemo
	default:
		return mode.String()
	}
}

func connectScanServiceError(err error) error {
	switch {
	case errors.Is(err, appservice.ErrUnsupportedScanMode):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, appservice.ErrLocalNetworkUnavailable):
		return connect.NewError(connect.CodeUnavailable, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}

func deviceTypeToProto(deviceType scanner.DeviceType) networkv1.DeviceType {
	switch deviceType {
	case scanner.DeviceRouter:
		return networkv1.DeviceType_DEVICE_TYPE_ROUTER
	case scanner.DeviceDesktop:
		return networkv1.DeviceType_DEVICE_TYPE_DESKTOP
	case scanner.DeviceMobile:
		return networkv1.DeviceType_DEVICE_TYPE_MOBILE
	case scanner.DeviceTV:
		return networkv1.DeviceType_DEVICE_TYPE_TV
	case scanner.DevicePrinter:
		return networkv1.DeviceType_DEVICE_TYPE_PRINTER
	case scanner.DeviceIoT:
		return networkv1.DeviceType_DEVICE_TYPE_IOT
	default:
		return networkv1.DeviceType_DEVICE_TYPE_UNKNOWN
	}
}

func scanEventTypeToProto(eventType string) networkv1.ScanEventType {
	switch strings.ToUpper(eventType) {
	case "SCAN_STARTED":
		return networkv1.ScanEventType_SCAN_EVENT_TYPE_SCAN_STARTED
	case "DEVICE_DISCOVERED":
		return networkv1.ScanEventType_SCAN_EVENT_TYPE_DEVICE_DISCOVERED
	case "DEVICE_UPDATED":
		return networkv1.ScanEventType_SCAN_EVENT_TYPE_DEVICE_UPDATED
	case "SCAN_FINISHED":
		return networkv1.ScanEventType_SCAN_EVENT_TYPE_SCAN_FINISHED
	case "SCAN_FAILED":
		return networkv1.ScanEventType_SCAN_EVENT_TYPE_SCAN_FAILED
	default:
		return networkv1.ScanEventType_SCAN_EVENT_TYPE_UNSPECIFIED
	}
}
