package cli

import (
	"errors"
	"fmt"
	"strings"
)

type Action string

const (
	ActionStatus Action = "status"
	ActionDHCP   Action = "dhcp"
	ActionLoad   Action = "load"
	ActionExport Action = "export"
	ActionDNS    Action = "dns"
)

var ErrHelp = errors.New("help requested")

type Command struct {
	Action      Action
	ProfilePath string
	DNSServers  []string
	DNSAuto     bool
}

func Usage() string {
	return strings.TrimLeft(`
Usage:
  wifictl status
  wifictl dhcp
  wifictl load <profile>
  wifictl export <profile>
  wifictl dns auto
  wifictl dns <server...>

Examples:
  wifictl status
  wifictl dhcp
  wifictl load office.conf
  wifictl export office.conf
  wifictl dns auto
  wifictl dns 114.114.114.114
  wifictl dns 223.5.5.5 119.29.29.29
`, "\n")
}

func Parse(args []string) (Command, error) {
	if len(args) == 0 {
		return Command{}, ErrHelp
	}

	switch args[0] {
	case "help", "-h", "--help":
		return Command{}, ErrHelp
	case string(ActionStatus):
		if len(args) != 1 {
			return Command{}, fmt.Errorf("status does not accept arguments")
		}
		return Command{Action: ActionStatus}, nil
	case string(ActionDHCP):
		if len(args) != 1 {
			return Command{}, fmt.Errorf("dhcp does not accept arguments")
		}
		return Command{Action: ActionDHCP}, nil
	case string(ActionLoad):
		if len(args) != 2 {
			return Command{}, fmt.Errorf("load requires <profile>")
		}
		return Command{Action: ActionLoad, ProfilePath: args[1]}, nil
	case string(ActionExport):
		if len(args) != 2 {
			return Command{}, fmt.Errorf("export requires <profile>")
		}
		return Command{Action: ActionExport, ProfilePath: args[1]}, nil
	case string(ActionDNS):
		return parseDNS(args[1:])
	default:
		return Command{}, fmt.Errorf("unsupported command %q", args[0])
	}
}

func parseDNS(args []string) (Command, error) {
	if len(args) == 0 {
		return Command{}, fmt.Errorf("dns requires either auto or one or more DNS server addresses")
	}

	if len(args) == 1 && args[0] == "auto" {
		return Command{Action: ActionDNS, DNSAuto: true}, nil
	}

	for _, value := range args {
		if strings.TrimSpace(value) == "" {
			return Command{}, fmt.Errorf("dns server address cannot be empty")
		}
	}

	return Command{Action: ActionDNS, DNSServers: args}, nil
}
