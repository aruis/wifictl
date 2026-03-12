package cli

import "testing"

func TestParseStatus(t *testing.T) {
	command, err := Parse([]string{"status"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if command.Action != ActionStatus {
		t.Fatalf("unexpected action: %q", command.Action)
	}
	if command.Service != "Wi-Fi" {
		t.Fatalf("unexpected default service: %q", command.Service)
	}
}

func TestParseLoad(t *testing.T) {
	command, err := Parse([]string{"--service", "Ethernet", "load", "office.conf"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if command.Action != ActionLoad {
		t.Fatalf("unexpected action: %q", command.Action)
	}
	if command.Service != "Ethernet" {
		t.Fatalf("unexpected service: %q", command.Service)
	}
	if command.ProfilePath != "office.conf" {
		t.Fatalf("unexpected profile path: %q", command.ProfilePath)
	}
}

func TestParseDNSReset(t *testing.T) {
	command, err := Parse([]string{"dns", "reset"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if command.Action != ActionDNS || !command.DNSAuto {
		t.Fatalf("unexpected dns command: %+v", command)
	}
}

func TestParseDNSServers(t *testing.T) {
	command, err := Parse([]string{"dns", "223.5.5.5", "119.29.29.29"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(command.DNSServers) != 2 {
		t.Fatalf("unexpected dns server count: %d", len(command.DNSServers))
	}
}

func TestParseInvalidCommand(t *testing.T) {
	if _, err := Parse([]string{"connect", "OfficeWiFi"}); err == nil {
		t.Fatal("expected error for unsupported command")
	}
}

func TestParseServices(t *testing.T) {
	command, err := Parse([]string{"services"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if command.Action != ActionServices {
		t.Fatalf("unexpected action: %q", command.Action)
	}
}
