package app

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/aruis/wifictl/internal/cli"
	"github.com/aruis/wifictl/internal/macos"
	"github.com/aruis/wifictl/internal/profile"
)

type App struct {
	backend *macos.Backend
	out     io.Writer
}

func New(backend *macos.Backend, out io.Writer) *App {
	return &App{
		backend: backend,
		out:     out,
	}
}

func (a *App) Run(ctx context.Context, command cli.Command, timeout time.Duration) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("wifictl currently supports macOS only")
	}

	wifi, err := a.backend.ResolveWiFi(ctx)
	if err != nil {
		return err
	}

	if err := a.backend.JoinWiFi(ctx, wifi.Device, command.SSID, command.Password); err != nil {
		return err
	}

	if err := a.waitForSSID(ctx, wifi.Device, command.SSID, timeout); err != nil {
		return err
	}

	switch command.Mode {
	case cli.ModeDHCP:
		if err := a.backend.ApplyDHCP(ctx, wifi.Service); err != nil {
			return err
		}
	case cli.ModeStatic:
		config, err := profile.Load(command.ProfilePath)
		if err != nil {
			return err
		}

		if err := a.backend.ApplyStatic(ctx, wifi.Service, config); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported mode %q", command.Mode)
	}

	status, err := a.waitForStatus(ctx, wifi, command.Mode, 10*time.Second)
	if err != nil {
		return err
	}

	a.printStatus(command.Mode, wifi, status)
	return nil
}

func (a *App) waitForSSID(ctx context.Context, device, targetSSID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ssid, err := a.backend.CurrentSSID(ctx, device)
		if err != nil {
			return err
		}

		if ssid == targetSSID {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}

	return fmt.Errorf("timed out waiting for Wi-Fi %q to become active", targetSSID)
}

func (a *App) waitForStatus(ctx context.Context, wifi macos.WiFiService, mode cli.Mode, timeout time.Duration) (macos.Status, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, err := a.backend.Status(ctx, wifi)
		if err != nil {
			return macos.Status{}, err
		}

		switch mode {
		case cli.ModeDHCP:
			if status.IP != "" {
				return status, nil
			}
		case cli.ModeStatic:
			return status, nil
		}

		select {
		case <-ctx.Done():
			return macos.Status{}, ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}

	status, err := a.backend.Status(ctx, wifi)
	if err != nil {
		return macos.Status{}, err
	}

	return status, nil
}

func (a *App) printStatus(mode cli.Mode, wifi macos.WiFiService, status macos.Status) {
	fmt.Fprintf(a.out, "Connected to %s\n", status.SSID)
	fmt.Fprintf(a.out, "Mode: %s\n", strings.ToUpper(string(mode)))
	fmt.Fprintf(a.out, "Device: %s\n", wifi.Device)
	fmt.Fprintf(a.out, "Service: %s\n", wifi.Service)
	if status.IP != "" {
		fmt.Fprintf(a.out, "IP: %s\n", status.IP)
	}
	if len(status.DNS) > 0 {
		fmt.Fprintf(a.out, "DNS: %s\n", strings.Join(status.DNS, ", "))
	}
}
