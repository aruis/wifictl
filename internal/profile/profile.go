package profile

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	IP      string
	Mask    string
	Gateway string
	DNS     []string
}

func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open profile %q: %w", path, err)
	}
	defer file.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return Config{}, fmt.Errorf("invalid profile line %d: expected key=value", lineNumber)
		}

		key = strings.TrimSpace(strings.ToLower(key))
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			return Config{}, fmt.Errorf("invalid profile line %d: empty key or value", lineNumber)
		}

		values[key] = value
	}

	if err := scanner.Err(); err != nil {
		return Config{}, fmt.Errorf("read profile %q: %w", path, err)
	}

	config := Config{
		IP:      values["ip"],
		Mask:    values["mask"],
		Gateway: values["gateway"],
	}

	if config.IP == "" {
		return Config{}, fmt.Errorf("profile %q is missing ip", path)
	}
	if config.Mask == "" {
		return Config{}, fmt.Errorf("profile %q is missing mask", path)
	}
	if config.Gateway == "" {
		return Config{}, fmt.Errorf("profile %q is missing gateway", path)
	}
	if err := validateIPv4("ip", config.IP); err != nil {
		return Config{}, fmt.Errorf("profile %q: %w", path, err)
	}
	if err := validateIPv4Mask(config.Mask); err != nil {
		return Config{}, fmt.Errorf("profile %q: %w", path, err)
	}
	if err := validateIPv4("gateway", config.Gateway); err != nil {
		return Config{}, fmt.Errorf("profile %q: %w", path, err)
	}

	if dns := values["dns"]; dns != "" {
		config.DNS = splitCSV(dns)
		for _, server := range config.DNS {
			if err := validateIPv4("dns", server); err != nil {
				return Config{}, fmt.Errorf("profile %q: %w", path, err)
			}
		}
	}

	return config, nil
}

func Save(path string, config Config) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("profile path cannot be empty")
	}

	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create profile directory %q: %w", dir, err)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create profile %q: %w", path, err)
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "ip=%s\nmask=%s\ngateway=%s\n", config.IP, config.Mask, config.Gateway); err != nil {
		return fmt.Errorf("write profile %q: %w", path, err)
	}
	if len(config.DNS) > 0 {
		if _, err := fmt.Fprintf(file, "dns=%s\n", strings.Join(config.DNS, ",")); err != nil {
			return fmt.Errorf("write profile %q: %w", path, err)
		}
	}

	return nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}

	return out
}

func validateIPv4(field, value string) error {
	ip := net.ParseIP(value)
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("%s must be a valid IPv4 address", field)
	}

	return nil
}

func validateIPv4Mask(value string) error {
	ip := net.ParseIP(value)
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("mask must be a valid IPv4 subnet mask")
	}

	mask := net.IPMask(ip.To4())
	ones, bits := mask.Size()
	if bits != 32 || ones < 0 {
		return fmt.Errorf("mask must be a valid IPv4 subnet mask")
	}

	return nil
}
