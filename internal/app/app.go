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

	if command.Action == cli.ActionServices {
		return a.runServices(ctx)
	}

	service, err := a.backend.ResolveService(ctx, command.Service)
	if err != nil {
		return err
	}

	switch command.Action {
	case cli.ActionStatus:
		return a.runStatus(ctx, service)
	case cli.ActionDHCP:
		if err := requireRoot(); err != nil {
			return err
		}
		if err := a.backend.ApplyDHCP(ctx, service.Service); err != nil {
			return err
		}
		if err := a.backend.ApplyDNSAuto(ctx, service.Service); err != nil {
			return err
		}
		return a.runStatus(ctx, service)
	case cli.ActionLoad:
		if err := requireRoot(); err != nil {
			return err
		}
		config, err := profile.Load(command.ProfilePath)
		if err != nil {
			return err
		}
		if err := a.backend.ApplyStatic(ctx, service.Service, config); err != nil {
			return err
		}
		return a.runStatus(ctx, service)
	case cli.ActionExport:
		return a.runExport(ctx, service, command.ProfilePath)
	case cli.ActionDNS:
		if err := requireRoot(); err != nil {
			return err
		}
		if command.DNSAuto {
			if err := a.backend.ApplyDNSAuto(ctx, service.Service); err != nil {
				return err
			}
		} else {
			if err := a.backend.ApplyDNSServers(ctx, service.Service, command.DNSServers); err != nil {
				return err
			}
		}
		return a.runStatus(ctx, service)
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

func (a *App) runStatus(ctx context.Context, service macos.NetworkService) error {
	status, err := a.backend.Status(ctx, service)
	if err != nil {
		return err
	}
	a.printStatus(status)
	return nil
}

func (a *App) runExport(ctx context.Context, service macos.NetworkService, path string) error {
	exported, err := a.backend.ExportProfile(ctx, service)
	if err != nil {
		return err
	}
	if err := profile.Save(path, exported); err != nil {
		return err
	}
	fmt.Fprintf(a.out, "Exported current %s profile to %s\n", service.Service, path)
	return nil
}

func (a *App) runServices(ctx context.Context) error {
	services, err := a.backend.ListServices(ctx)
	if err != nil {
		return err
	}
	for _, service := range services {
		fmt.Fprintf(a.out, "%s\t%s\t%s\n", service.Service, service.HardwarePort, service.Device)
	}
	return nil
}

func (a *App) printStatus(status macos.Status) {
	fmt.Fprintf(a.out, "Service: %s\n", status.Service)
	if status.HardwarePort != "" {
		fmt.Fprintf(a.out, "Hardware Port: %s\n", status.HardwarePort)
	}
	fmt.Fprintf(a.out, "Device: %s\n", status.Device)
	if status.SSID != "" {
		fmt.Fprintf(a.out, "SSID: %s\n", status.SSID)
	}
	if status.AssociationKnown {
		if status.Associated {
			fmt.Fprintf(a.out, "Associated: yes\n")
		} else {
			fmt.Fprintf(a.out, "Associated: no\n")
		}
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
