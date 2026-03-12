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
	SSID string
	IP   string
	DNS  []string
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
	ssid, err := b.CurrentSSID(ctx, wifi.Device)
	if err != nil {
		return Status{}, err
	}

	ip, err := run(ctx, "ipconfig", "getifaddr", wifi.Device)
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() != 0 {
			ip = ""
		} else {
			return Status{}, fmt.Errorf("read IP address: %w", err)
		}
	}

	dnsOut, err := run(ctx, "networksetup", "-getdnsservers", wifi.Service)
	if err != nil {
		return Status{}, fmt.Errorf("read DNS servers: %w", err)
	}

	return Status{
		SSID: ssid,
		IP:   strings.TrimSpace(ip),
		DNS:  parseDNSServers(dnsOut),
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
