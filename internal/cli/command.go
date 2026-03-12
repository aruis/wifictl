package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

type Mode string

const (
	ModeDHCP   Mode = "dhcp"
	ModeStatic Mode = "static"
)

var ErrHelp = errors.New("help requested")

type Command struct {
	SSID        string
	Mode        Mode
	ProfilePath string
	Password    string
}

func Usage() string {
	return strings.TrimLeft(`
Usage:
  wifictl connect <ssid> dhcp [--password <password>]
  wifictl connect <ssid> static <profile> [--password <password>]
  wifictl -c <ssid> -d [-p <password>]
  wifictl -c <ssid> -s <profile> [-p <password>]

Examples:
  wifictl connect OfficeWiFi dhcp
  wifictl connect OfficeWiFi static office.conf
  wifictl connect OfficeWiFi dhcp --password secret
  wifictl -c OfficeWiFi -d
  wifictl -c OfficeWiFi -s office.conf

Options:
  -c, --connect <ssid>     connect to the target SSID
  -d, --dhcp               use DHCP on the Wi-Fi service
  -s, --static <profile>   apply a static profile file
  -p, --password <value>   Wi-Fi password for the target SSID
  -h, --help               show this help
`, "\n")
}

func Parse(args []string) (Command, error) {
	if len(args) == 0 {
		return Command{}, ErrHelp
	}

	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		return Command{}, ErrHelp
	}

	if args[0] == "connect" {
		return parseConnect(args[1:])
	}

	return parseFlags(args)
}

func parseConnect(args []string) (Command, error) {
	var command Command

	fs := flag.NewFlagSet("connect", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&command.Password, "password", "", "")
	fs.StringVar(&command.Password, "p", "", "")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Command{}, ErrHelp
		}

		return Command{}, err
	}

	rest := fs.Args()
	if len(rest) < 2 {
		return Command{}, fmt.Errorf("connect requires <ssid> and a network mode")
	}

	command.SSID = rest[0]
	switch rest[1] {
	case string(ModeDHCP):
		if len(rest) != 2 {
			return Command{}, fmt.Errorf("dhcp mode does not accept a profile")
		}
		command.Mode = ModeDHCP
	case string(ModeStatic):
		if len(rest) != 3 {
			return Command{}, fmt.Errorf("static mode requires a profile path")
		}
		command.Mode = ModeStatic
		command.ProfilePath = rest[2]
	default:
		return Command{}, fmt.Errorf("unsupported network mode %q", rest[1])
	}

	return command, nil
}

func parseFlags(args []string) (Command, error) {
	var command Command
	var dhcp bool

	fs := flag.NewFlagSet("wifictl", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&command.SSID, "connect", "", "")
	fs.StringVar(&command.SSID, "c", "", "")
	fs.BoolVar(&dhcp, "dhcp", false, "")
	fs.BoolVar(&dhcp, "d", false, "")
	fs.StringVar(&command.ProfilePath, "static", "", "")
	fs.StringVar(&command.ProfilePath, "s", "", "")
	fs.StringVar(&command.Password, "password", "", "")
	fs.StringVar(&command.Password, "p", "", "")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Command{}, ErrHelp
		}

		return Command{}, err
	}

	if len(fs.Args()) > 0 {
		return Command{}, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}

	if command.SSID == "" {
		return Command{}, fmt.Errorf("missing target SSID")
	}

	switch {
	case dhcp && command.ProfilePath != "":
		return Command{}, fmt.Errorf("use either DHCP or a static profile, not both")
	case !dhcp && command.ProfilePath == "":
		return Command{}, fmt.Errorf("missing network mode, use -d or -s <profile>")
	case dhcp:
		command.Mode = ModeDHCP
	default:
		command.Mode = ModeStatic
	}

	return command, nil
}
