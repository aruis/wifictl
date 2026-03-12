package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/aruis/wifictl/internal/cli"
	"github.com/aruis/wifictl/internal/macos"
	"github.com/aruis/wifictl/internal/profile"
)

type App struct {
	backend *macos.Backend
	out     io.Writer
}

func New(backend *macos.Backend, out io.Writer) *App {
	return &App{backend: backend, out: out}
}

func (a *App) Run(ctx context.Context, command cli.Command) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("wifictl currently supports macOS only")
	}

	wifi, err := a.backend.ResolveWiFi(ctx)
	if err != nil {
		return err
	}

	switch command.Action {
	case cli.ActionStatus:
		return a.runStatus(ctx, wifi)
	case cli.ActionDHCP:
		if err := requireRoot(); err != nil {
			return err
		}
		if err := a.backend.ApplyDHCP(ctx, wifi.Service); err != nil {
			return err
		}
		if err := a.backend.ApplyDNSAuto(ctx, wifi.Service); err != nil {
			return err
		}
		return a.runStatus(ctx, wifi)
	case cli.ActionLoad:
		if err := requireRoot(); err != nil {
			return err
		}
		config, err := profile.Load(command.ProfilePath)
		if err != nil {
			return err
		}
		if err := a.backend.ApplyStatic(ctx, wifi.Service, config); err != nil {
			return err
		}
		return a.runStatus(ctx, wifi)
	case cli.ActionExport:
		return a.runExport(ctx, wifi, command.ProfilePath)
	case cli.ActionDNS:
		if err := requireRoot(); err != nil {
			return err
		}
		if command.DNSAuto {
			if err := a.backend.ApplyDNSAuto(ctx, wifi.Service); err != nil {
				return err
			}
		} else {
			if err := a.backend.ApplyDNSServers(ctx, wifi.Service, command.DNSServers); err != nil {
				return err
			}
		}
		return a.runStatus(ctx, wifi)
	default:
		return fmt.Errorf("unsupported action %q", command.Action)
	}
}

func requireRoot() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("administrator privileges required; run with sudo")
	}
	return nil
}

func (a *App) runStatus(ctx context.Context, wifi macos.WiFiService) error {
	status, err := a.backend.Status(ctx, wifi)
	if err != nil {
		return err
	}
	a.printStatus(status)
	return nil
}

func (a *App) runExport(ctx context.Context, wifi macos.WiFiService, path string) error {
	exported, err := a.backend.ExportProfile(ctx, wifi)
	if err != nil {
		return err
	}
	if err := profile.Save(path, exported); err != nil {
		return err
	}
	fmt.Fprintf(a.out, "Exported current Wi-Fi profile to %s\n", path)
	return nil
}

func (a *App) printStatus(status macos.Status) {
	fmt.Fprintf(a.out, "Service: %s\n", status.Service)
	fmt.Fprintf(a.out, "Device: %s\n", status.Device)
	if status.SSID != "" {
		fmt.Fprintf(a.out, "SSID: %s\n", status.SSID)
	}
	if status.Associated {
		fmt.Fprintf(a.out, "Associated: yes\n")
	} else {
		fmt.Fprintf(a.out, "Associated: no\n")
	}
	if status.Method != "" {
		fmt.Fprintf(a.out, "IPv4: %s\n", status.Method)
	}
	if status.IP != "" {
		fmt.Fprintf(a.out, "IP: %s\n", status.IP)
	}
	if status.Mask != "" {
		fmt.Fprintf(a.out, "Mask: %s\n", status.Mask)
	}
	if status.Gateway != "" {
		fmt.Fprintf(a.out, "Gateway: %s\n", status.Gateway)
	}
	if len(status.DNS) > 0 {
		fmt.Fprintf(a.out, "DNS: %s\n", strings.Join(status.DNS, ", "))
	} else {
		fmt.Fprintf(a.out, "DNS: reset (no DNS servers currently reported by macOS)\n")
	}
}
