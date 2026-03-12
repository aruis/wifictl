package macos

import "testing"

func TestIPFromNetworkSetupOutput(t *testing.T) {
	out := `
Manual Configuration
IP address: 10.60.1.94
Subnet mask: 255.255.0.0
Router: 10.60.1.254
`

	ip, err := parseNetworkSetupIP(out)
	if err != nil {
		t.Fatalf("parseNetworkSetupIP returned error: %v", err)
	}
	if ip != "10.60.1.94" {
		t.Fatalf("unexpected ip: %q", ip)
	}
}

func TestIPFromNetworkSetupOutputEmpty(t *testing.T) {
	out := `
DHCP Configuration
IP address: none
`

	ip, err := parseNetworkSetupIP(out)
	if err != nil {
		t.Fatalf("parseNetworkSetupIP returned error: %v", err)
	}
	if ip != "" {
		t.Fatalf("expected empty ip, got %q", ip)
	}
}
