package macos

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/aruis/wifictl/internal/profile"
)

type WiFiService struct {
	Device  string
	Service string
}

type Status struct {
	SSID       string
	Associated bool
	IP         string
	DNS        []string
}

type Backend struct{}

func New() *Backend {
	return &Backend{}
}

func (b *Backend) ResolveWiFi(ctx context.Context) (WiFiService, error) {
	device, err := b.findWiFiDevice(ctx)
	if err != nil {
		return WiFiService{}, err
	}

	service, err := b.findWiFiService(ctx, device)
	if err != nil {
		return WiFiService{}, err
	}

	return WiFiService{
		Device:  device,
		Service: service,
	}, nil
}

func (b *Backend) JoinWiFi(ctx context.Context, device, ssid, password string) error {
	args := []string{"-setairportnetwork", device, ssid}
	if password != "" {
		args = append(args, password)
	}

	if _, err := run(ctx, "networksetup", args...); err != nil {
		return fmt.Errorf("join Wi-Fi %q: %w", ssid, err)
	}

	return nil
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

func (b *Backend) IsAssociated(ctx context.Context, device string) (bool, string, error) {
	if associated, ssid, err := b.associationFromWDUtil(ctx, device); err == nil {
		return associated, ssid, nil
	}

	return b.associationFromSystemProfiler(ctx, device)
}

func (b *Backend) ApplyDHCP(ctx context.Context, service string) error {
	if _, err := run(ctx, "networksetup", "-setdhcp", service); err != nil {
		return fmt.Errorf("apply DHCP on %q: %w", service, err)
	}

	return nil
}

func (b *Backend) ApplyStatic(ctx context.Context, service string, config profile.Config) error {
	if _, err := run(ctx, "networksetup", "-setmanual", service, config.IP, config.Mask, config.Gateway); err != nil {
		return fmt.Errorf("apply static IP on %q: %w", service, err)
	}

	args := []string{"-setdnsservers", service}
	if len(config.DNS) == 0 {
		args = append(args, "empty")
	} else {
		args = append(args, config.DNS...)
	}

	if _, err := run(ctx, "networksetup", args...); err != nil {
		return fmt.Errorf("apply DNS on %q: %w", service, err)
	}

	return nil
}

func (b *Backend) Status(ctx context.Context, wifi WiFiService) (Status, error) {
	associated, ssid, err := b.IsAssociated(ctx, wifi.Device)
	if err != nil {
		return Status{}, err
	}

	ip, err := run(ctx, "ipconfig", "getifaddr", wifi.Device)
	if err != nil {
		ip = ""
	}

	dnsOut, err := run(ctx, "networksetup", "-getdnsservers", wifi.Service)
	if err != nil {
		return Status{}, fmt.Errorf("read DNS servers: %w", err)
	}

	return Status{
		SSID:       ssid,
		Associated: associated,
		IP:         strings.TrimSpace(ip),
		DNS:        parseDNSServers(dnsOut),
	}, nil
}

func (b *Backend) findWiFiDevice(ctx context.Context) (string, error) {
	out, err := run(ctx, "networksetup", "-listallhardwareports")
	if err != nil {
		return "", fmt.Errorf("list hardware ports: %w", err)
	}

	lines := strings.Split(out, "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line != "Hardware Port: Wi-Fi" && line != "Hardware Port: AirPort" {
			continue
		}

		for j := i + 1; j < len(lines) && j <= i+3; j++ {
			next := strings.TrimSpace(lines[j])
			if strings.HasPrefix(next, "Device: ") {
				return strings.TrimSpace(strings.TrimPrefix(next, "Device: ")), nil
			}
		}
	}

	return "", fmt.Errorf("Wi-Fi hardware device not found")
}

func (b *Backend) findWiFiService(ctx context.Context, device string) (string, error) {
	out, err := run(ctx, "networksetup", "-listnetworkserviceorder")
	if err != nil {
		return "", fmt.Errorf("list network services: %w", err)
	}

	serviceRe := regexp.MustCompile(`^\(\d+\)\s(.+)$`)
	deviceRe := regexp.MustCompile(`^\(Hardware Port: .+, Device: (.+)\)$`)

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

		if matches := deviceRe.FindStringSubmatch(line); len(matches) == 2 && currentService != "" {
			if strings.TrimSpace(matches[1]) == device {
				return currentService, nil
			}
		}
	}

	return "", fmt.Errorf("Wi-Fi network service not found for device %q", device)
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
