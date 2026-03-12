package macos

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/aruis/wifictl/internal/profile"
)

type NetworkService struct {
	Device  string
	Service string
	HardwarePort string
}

type Status struct {
	Service    string
	HardwarePort string
	Device     string
	SSID       string
	AssociationKnown bool
	Associated bool
	Method     string
	IP         string
	Mask       string
	Gateway    string
	DNS        []string
}

type Backend struct{}

func New() *Backend {
	return &Backend{}
}

func (b *Backend) ResolveService(ctx context.Context, serviceName string) (NetworkService, error) {
	services, err := b.ListServices(ctx)
	if err != nil {
		return NetworkService{}, err
	}
	for _, service := range services {
		if service.Service == serviceName {
			return service, nil
		}
	}
	return NetworkService{}, fmt.Errorf("network service %q not found", serviceName)
}

func (b *Backend) CurrentSSID(ctx context.Context, device string) (string, error) {
	out, err := run(ctx, "networksetup", "-getairportnetwork", device)
	if err != nil {
		return "", fmt.Errorf("read current Wi-Fi network: %w", err)
	}

	if strings.Contains(out, "You are not associated with an AirPort network") {
		return "", nil
	}

	_, ssid, found := strings.Cut(strings.TrimSpace(out), ": ")
	if !found {
		return "", fmt.Errorf("unexpected current Wi-Fi output: %s", strings.TrimSpace(out))
	}

	return strings.TrimSpace(ssid), nil
}

func (b *Backend) ApplyDHCP(ctx context.Context, service string) error {
	if _, err := run(ctx, "networksetup", "-setdhcp", service); err != nil {
		return fmt.Errorf("apply DHCP on %q: %w", service, err)
	}

	return nil
}

func (b *Backend) ApplyDNSAuto(ctx context.Context, service string) error {
	if _, err := run(ctx, "networksetup", "-setdnsservers", service, "empty"); err != nil {
		return fmt.Errorf("apply automatic DNS on %q: %w", service, err)
	}

	return nil
}

func (b *Backend) ApplyDNSServers(ctx context.Context, service string, servers []string) error {
	if len(servers) == 0 {
		return fmt.Errorf("at least one DNS server is required")
	}

	args := append([]string{"-setdnsservers", service}, servers...)
	if _, err := run(ctx, "networksetup", args...); err != nil {
		return fmt.Errorf("apply DNS on %q: %w", service, err)
	}

	return nil
}

func (b *Backend) ApplyStatic(ctx context.Context, service string, config profile.Config) error {
	if _, err := run(ctx, "networksetup", "-setmanual", service, config.IP, config.Mask, config.Gateway); err != nil {
		return fmt.Errorf("apply static IP on %q: %w", service, err)
	}

	if len(config.DNS) == 0 {
		return b.ApplyDNSAuto(ctx, service)
	}

	return b.ApplyDNSServers(ctx, service, config.DNS)
}

func (b *Backend) ExportProfile(ctx context.Context, service NetworkService) (profile.Config, error) {
	status, err := b.Status(ctx, service)
	if err != nil {
		return profile.Config{}, err
	}

	if status.IP == "" || status.Mask == "" || status.Gateway == "" {
		return profile.Config{}, fmt.Errorf("current network service does not expose a complete IPv4 configuration")
	}

	return profile.Config{
		IP:      status.IP,
		Mask:    status.Mask,
		Gateway: status.Gateway,
		DNS:     status.DNS,
	}, nil
}

func (b *Backend) Status(ctx context.Context, service NetworkService) (Status, error) {
	associated := false
	ssid := ""
	associationKnown := false
	if service.HardwarePort == "Wi-Fi" || service.HardwarePort == "AirPort" {
		var err error
		associated, ssid, err = b.associationInfo(ctx, service.Device)
		if err != nil {
			return Status{}, err
		}
		associationKnown = true
	}

	infoOut, err := run(ctx, "networksetup", "-getinfo", service.Service)
	if err != nil {
		return Status{}, fmt.Errorf("read network service info: %w", err)
	}
	info := parseNetworkSetupInfo(infoOut)

	dnsOut, err := run(ctx, "networksetup", "-getdnsservers", service.Service)
	if err != nil {
		return Status{}, fmt.Errorf("read DNS servers: %w", err)
	}

	return Status{
		Service:    service.Service,
		HardwarePort: service.HardwarePort,
		Device:     service.Device,
		SSID:       ssid,
		AssociationKnown: associationKnown,
		Associated: associated,
		Method:     info.Method,
		IP:         info.IP,
		Mask:       info.Mask,
		Gateway:    info.Gateway,
		DNS:        parseDNSServers(dnsOut),
	}, nil
}

func (b *Backend) ListServices(ctx context.Context) ([]NetworkService, error) {
	out, err := run(ctx, "networksetup", "-listnetworkserviceorder")
	if err != nil {
		return nil, fmt.Errorf("list network services: %w", err)
	}
	return parseNetworkServiceOrder(out)
}

func (b *Backend) associationInfo(ctx context.Context, device string) (bool, string, error) {
	if associated, ssid, err := b.associationFromWDUtil(ctx, device); err == nil {
		return associated, ssid, nil
	}

	return b.associationFromSystemProfiler(ctx, device)
}

func parseNetworkServiceOrder(out string) ([]NetworkService, error) {
	serviceRe := regexp.MustCompile(`^\(\d+\)\s(.+)$`)
	detailRe := regexp.MustCompile(`^\(Hardware Port: (.+), Device: (.+)\)$`)
	var services []NetworkService
	var currentService string
	for _, rawLine := range strings.Split(out, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		if matches := serviceRe.FindStringSubmatch(line); len(matches) == 2 {
			currentService = strings.TrimSpace(matches[1])
			continue
		}

		if matches := detailRe.FindStringSubmatch(line); len(matches) == 3 && currentService != "" {
			services = append(services, NetworkService{
				Service: currentService,
				HardwarePort: strings.TrimSpace(matches[1]),
				Device: strings.TrimSpace(matches[2]),
			})
			currentService = ""
		}
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("no network services found")
	}
	return services, nil
}

type networkSetupInfo struct {
	Method  string
	IP      string
	Mask    string
	Gateway string
}

func parseNetworkSetupInfo(out string) networkSetupInfo {
	info := networkSetupInfo{}

	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasSuffix(trimmed, "Configuration"):
			info.Method = strings.TrimSpace(strings.TrimSuffix(trimmed, "Configuration"))
		case strings.HasPrefix(trimmed, "IP address: "):
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "IP address: "))
			if !strings.EqualFold(value, "none") {
				info.IP = value
			}
		case strings.HasPrefix(trimmed, "Subnet mask: "):
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "Subnet mask: "))
			if !strings.EqualFold(value, "none") {
				info.Mask = value
			}
		case strings.HasPrefix(trimmed, "Router: "):
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "Router: "))
			if !strings.EqualFold(value, "none") {
				info.Gateway = value
			}
		}
	}

	return info
}

func parseNetworkSetupIP(out string) (string, error) {
	return parseNetworkSetupInfo(out).IP, nil
}

func parseDNSServers(out string) []string {
	out = strings.TrimSpace(out)
	switch {
	case out == "":
		return nil
	case strings.Contains(out, "There aren't any DNS Servers set"):
		return nil
	}

	lines := strings.Split(out, "\n")
	servers := make([]string, 0, len(lines))
	for _, line := range lines {
		value := strings.TrimSpace(line)
		if value != "" {
			servers = append(servers, value)
		}
	}

	return servers
}

func (b *Backend) associationFromWDUtil(ctx context.Context, device string) (bool, string, error) {
	out, err := run(ctx, "wdutil", "info")
	if err != nil {
		return false, "", err
	}

	return parseWDUtilAssociation(out, device)
}

func (b *Backend) associationFromSystemProfiler(ctx context.Context, device string) (bool, string, error) {
	out, err := run(ctx, "system_profiler", "SPAirPortDataType")
	if err != nil {
		return false, "", fmt.Errorf("read Wi-Fi status: %w", err)
	}

	return parseSystemProfilerAssociation(out, device)
}

func parseWDUtilAssociation(out, device string) (bool, string, error) {
	lines := strings.Split(out, "\n")
	inWiFiSection := false
	var interfaceName string
	var ssid string
	var bssid string
	var opMode string

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "WIFI") {
			inWiFiSection = true
			continue
		}

		if inWiFiSection && strings.HasPrefix(line, "BLUETOOTH") {
			break
		}

		if !inWiFiSection {
			continue
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "Interface Name":
			interfaceName = value
		case "SSID":
			ssid = value
		case "BSSID":
			bssid = value
		case "Op Mode":
			opMode = value
		}
	}

	if interfaceName == "" {
		return false, "", fmt.Errorf("wdutil output did not include Wi-Fi interface data")
	}

	if interfaceName != device {
		return false, "", fmt.Errorf("wdutil reported interface %q, expected %q", interfaceName, device)
	}

	associated := (ssid != "" && ssid != "None") || (bssid != "" && bssid != "None") || opMode == "STA"
	return associated, redactAwareSSID(ssid), nil
}

func parseSystemProfilerAssociation(out, device string) (bool, string, error) {
	lines := strings.Split(out, "\n")
	inDeviceBlock := false
	var status string
	var currentNetwork string

	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, " ")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, device+":") {
			inDeviceBlock = true
			continue
		}

		if inDeviceBlock && !strings.HasPrefix(rawLine, "          ") && !strings.HasPrefix(rawLine, "            ") {
			break
		}

		if !inDeviceBlock {
			continue
		}

		if strings.HasPrefix(trimmed, "Status: ") {
			status = strings.TrimSpace(strings.TrimPrefix(trimmed, "Status: "))
			continue
		}

		if strings.HasSuffix(trimmed, ":") && strings.Contains(rawLine, "            ") && currentNetwork == "" && trimmed != "Current Network Information:" {
			currentNetwork = strings.TrimSuffix(trimmed, ":")
		}
	}

	if !inDeviceBlock {
		return false, "", fmt.Errorf("system_profiler output did not include interface %q", device)
	}

	return strings.EqualFold(status, "Connected"), redactAwareSSID(currentNetwork), nil
}

func redactAwareSSID(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case "", "None", "<redacted>":
		return ""
	default:
		return value
	}
}

func run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(out))
		if trimmed == "" {
			return "", fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
		}

		return "", fmt.Errorf("%s %s: %s", name, strings.Join(args, " "), trimmed)
	}

	return string(out), nil
}
