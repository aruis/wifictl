package macos

import "testing"

func TestParseWDUtilAssociation(t *testing.T) {
	out := `
————————————————————————————————————————————————————————————————————
WIFI
————————————————————————————————————————————————————————————————————
    Interface Name       : en1
    Op Mode              : STA
    SSID                 : <redacted>
    BSSID                : <redacted>
————————————————————————————————————————————————————————————————————
BLUETOOTH
`

	associated, ssid, err := parseWDUtilAssociation(out, "en1")
	if err != nil {
		t.Fatalf("parseWDUtilAssociation returned error: %v", err)
	}

	if !associated {
		t.Fatal("expected associated=true")
	}

	if ssid != "" {
		t.Fatalf("expected empty ssid for redacted value, got %q", ssid)
	}
}

func TestParseSystemProfilerAssociation(t *testing.T) {
	out := `
Wi-Fi:
      Interfaces:
        en1:
          Status: Connected
          Current Network Information:
            <redacted>:
              PHY Mode: 802.11ax
`

	associated, ssid, err := parseSystemProfilerAssociation(out, "en1")
	if err != nil {
		t.Fatalf("parseSystemProfilerAssociation returned error: %v", err)
	}

	if !associated {
		t.Fatal("expected associated=true")
	}

	if ssid != "" {
		t.Fatalf("expected empty ssid for redacted value, got %q", ssid)
	}
}
