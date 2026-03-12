package cli

import "testing"

func TestParseConnectDHCP(t *testing.T) {
	command, err := Parse([]string{"connect", "OfficeWiFi", "dhcp"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if command.SSID != "OfficeWiFi" {
		t.Fatalf("unexpected SSID: %q", command.SSID)
	}
	if command.Mode != ModeDHCP {
		t.Fatalf("unexpected mode: %q", command.Mode)
	}
}

func TestParseConnectStatic(t *testing.T) {
	command, err := Parse([]string{"connect", "OfficeWiFi", "static", "office.conf"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if command.Mode != ModeStatic {
		t.Fatalf("unexpected mode: %q", command.Mode)
	}
	if command.ProfilePath != "office.conf" {
		t.Fatalf("unexpected profile path: %q", command.ProfilePath)
	}
}

func TestParseShortFlags(t *testing.T) {
	command, err := Parse([]string{"-c", "OfficeWiFi", "-d"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if command.Mode != ModeDHCP {
		t.Fatalf("unexpected mode: %q", command.Mode)
	}
}

func TestParseMutuallyExclusiveModes(t *testing.T) {
	if _, err := Parse([]string{"-c", "OfficeWiFi", "-d", "-s", "office.conf"}); err == nil {
		t.Fatal("expected error for conflicting modes")
	}
}
