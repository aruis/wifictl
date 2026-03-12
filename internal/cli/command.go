package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

type Action string

const (
	ActionStatus Action = "status"
	ActionDHCP   Action = "dhcp"
	ActionLoad   Action = "load"
	ActionExport Action = "export"
	ActionDNS    Action = "dns"
	ActionServices Action = "services"
)

var ErrHelp = errors.New("help requested")

type Command struct {
	Service     string
	Action      Action
	ProfilePath string
	DNSServers  []string
	DNSAuto     bool
}

func Usage() string {
	return strings.TrimLeft(`
Usage:
  wifictl [--service <name>] status
  wifictl [--service <name>] dhcp
  wifictl [--service <name>] load <profile>
  wifictl [--service <name>] export <profile>
  wifictl [--service <name>] dns reset
  wifictl [--service <name>] dns <server...>
  wifictl services

Examples:
  wifictl status
  wifictl --service Ethernet status
  wifictl dhcp
  wifictl load office.conf
  wifictl export office.conf
  wifictl dns reset
  wifictl dns 223.5.5.5 223.6.6.6
  wifictl services
`, "\n")
}

func Parse(args []string) (Command, error) {
	if len(args) == 0 {
		return Command{}, ErrHelp
	}

	command := Command{Service: "Wi-Fi"}
	fs := flag.NewFlagSet("wifictl", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&command.Service, "service", "Wi-Fi", "")
	fs.StringVar(&command.Service, "S", "Wi-Fi", "")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Command{}, ErrHelp
		}
		return Command{}, err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return Command{}, ErrHelp
	}

	switch rest[0] {
	case "help", "-h", "--help":
		return Command{}, ErrHelp
	case string(ActionServices):
		if len(rest) != 1 {
			return Command{}, fmt.Errorf("services does not accept arguments")
		}
		command.Action = ActionServices
		return command, nil
	case string(ActionStatus):
		if len(rest) != 1 {
			return Command{}, fmt.Errorf("status does not accept arguments")
		}
		command.Action = ActionStatus
		return command, nil
	case string(ActionDHCP):
		if len(rest) != 1 {
			return Command{}, fmt.Errorf("dhcp does not accept arguments")
		}
		command.Action = ActionDHCP
		return command, nil
	case string(ActionLoad):
		if len(rest) != 2 {
			return Command{}, fmt.Errorf("load requires <profile>")
		}
		command.Action = ActionLoad
		command.ProfilePath = rest[1]
		return command, nil
	case string(ActionExport):
		if len(rest) != 2 {
			return Command{}, fmt.Errorf("export requires <profile>")
		}
		command.Action = ActionExport
		command.ProfilePath = rest[1]
		return command, nil
	case string(ActionDNS):
		dnsCommand, err := parseDNS(rest[1:])
		if err != nil {
			return Command{}, err
		}
		command.Action = dnsCommand.Action
		command.DNSAuto = dnsCommand.DNSAuto
		command.DNSServers = dnsCommand.DNSServers
		return command, nil
	default:
		return Command{}, fmt.Errorf("unsupported command %q", rest[0])
	}
}

func parseDNS(args []string) (Command, error) {
	if len(args) == 0 {
		return Command{}, fmt.Errorf("dns requires either reset or one or more DNS server addresses")
	}

	if len(args) == 1 && args[0] == "reset" {
		return Command{Action: ActionDNS, DNSAuto: true}, nil
	}

	for _, value := range args {
		if strings.TrimSpace(value) == "" {
			return Command{}, fmt.Errorf("dns server address cannot be empty")
		}
	}

	return Command{Action: ActionDNS, DNSServers: args}, nil
}
