package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "office.conf")
	content := "ip=192.168.10.88\nmask=255.255.255.0\ngateway=192.168.10.1\ndns=192.168.10.2, 223.5.5.5\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	config, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if config.IP != "192.168.10.88" {
		t.Fatalf("unexpected IP: %q", config.IP)
	}
	if len(config.DNS) != 2 {
		t.Fatalf("unexpected DNS count: %d", len(config.DNS))
	}
}

func TestLoadMissingGateway(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.conf")
	content := "ip=192.168.10.88\nmask=255.255.255.0\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected validation error")
	}
}
