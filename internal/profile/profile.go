package profile

import (
	"bufio"
	"fmt"
	"os"
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

	if dns := values["dns"]; dns != "" {
		config.DNS = splitCSV(dns)
	}

	return config, nil
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
