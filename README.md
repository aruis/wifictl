# wifictl

[中文文档](./README.zh-CN.md)

`wifictl` is a macOS-first CLI for joining a Wi-Fi network and immediately applying an IP strategy.

## Current Status

The current implementation is a working macOS Go CLI with:

- `connect <ssid> dhcp`
- `connect <ssid> static <profile>`
- short flag equivalents
- optional Wi-Fi password support
- static profile validation
- status output after apply

## Why wifictl

`wifictl` is intentionally focused on a direct execution workflow:

- connect to a target Wi-Fi SSID
- apply DHCP or a static IP profile immediately
- print the resulting network status

It does not depend on macOS Locations and does not try to manage long-running background rules in the first version.

## Goals

- Connect to a target Wi-Fi SSID from the command line.
- Apply `DHCP` or a `static` IP profile right after the Wi-Fi connection succeeds.
- Keep the user-facing command short and direct.
- Ship as a standalone CLI suitable for Homebrew distribution.

## Scope

The first version targets macOS only.

Linux support is intentionally out of scope for the initial release because Wi-Fi and IP configuration stacks differ too much across distributions and network managers.

## Command Design

Primary commands:

```bash
wifictl connect <ssid> dhcp
wifictl connect <ssid> static <profile>
```

Short flags:

```bash
wifictl -c <ssid> -d
wifictl -c <ssid> -s <profile>
```

Equivalent examples:

```bash
wifictl connect OfficeWiFi dhcp
wifictl connect OfficeWiFi static office.conf
wifictl connect OfficeWiFi dhcp --password secret
wifictl connect OfficeWiFi static office.conf --password secret

wifictl -c OfficeWiFi -d
wifictl -c OfficeWiFi -s office.conf
wifictl -c OfficeWiFi -d -p secret
```

## Quick Start

Build:

```bash
make build
```

Run:

```bash
sudo ./dist/wifictl connect OfficeWiFi dhcp
sudo ./dist/wifictl connect OfficeWiFi static examples/office.conf
```

Short flags:

```bash
sudo ./dist/wifictl -c OfficeWiFi -d
sudo ./dist/wifictl -c OfficeWiFi -s examples/office.conf
```

## Intended Behavior

For `wifictl connect <ssid> dhcp`:

1. Resolve the active Wi-Fi hardware device and network service on macOS.
2. Join the target SSID.
3. Wait until the current Wi-Fi network matches the target SSID.
4. Apply DHCP to the Wi-Fi service.
5. Print the resulting network status.

For `wifictl connect <ssid> static <profile>`:

1. Resolve the active Wi-Fi hardware device and network service on macOS.
2. Join the target SSID.
3. Wait until the current Wi-Fi network matches the target SSID.
4. Read the static profile.
5. Apply manual IP, subnet mask, gateway, and DNS settings.
6. Print the resulting network status.

## Profile Format

Static IP profiles use a simple key-value format.

Example `office.conf`:

```ini
ip=192.168.10.88
mask=255.255.255.0
gateway=192.168.10.1
dns=192.168.10.2,223.5.5.5
```

Required keys:

- `ip`
- `mask`
- `gateway`

Optional keys:

- `dns`

## Password Support

When the target Wi-Fi is not already saved by macOS, a password can be passed explicitly.

Examples:

```bash
wifictl connect OfficeWiFi dhcp --password secret
wifictl connect OfficeWiFi static office.conf --password secret
wifictl -c OfficeWiFi -d -p secret
wifictl -c OfficeWiFi -s office.conf -p secret
```

If no password is provided, `wifictl` relies on an open network or a credential already known to macOS.

## Example Output

```text
Connected to OfficeWiFi
Mode: DHCP
Device: en0
Service: Wi-Fi
IP: 192.168.10.25
DNS: 192.168.10.2, 223.5.5.5
```

## Permissions

Changing Wi-Fi and IP settings on macOS commonly requires administrator privileges.

The initial UX assumption is:

```bash
sudo wifictl connect OfficeWiFi dhcp
sudo wifictl connect OfficeWiFi static office.conf
```

The tool should not attempt to manage privilege escalation internally.

## Error Handling Expectations

The CLI should return explicit errors for cases such as:

- Wi-Fi hardware device not found
- Wi-Fi network service not found
- Failed to join the target SSID
- Timed out waiting for the SSID to become active
- Invalid or incomplete static profile
- Failed to apply DHCP
- Failed to apply static IP or DNS settings

## macOS Implementation Notes

The implementation relies on macOS system commands such as:

- `networksetup`
- `ipconfig`
- `ifconfig`

The implementation should prefer stable system command wrappers over private APIs in the first version.

## Project Direction

Implementation language: Go

Why Go:

- easy to ship as a single binary
- good fit for CLI argument parsing and subprocess orchestration
- low-friction Homebrew distribution

## Packaging Direction

Planned distribution path:

1. publish source and releases on GitHub
2. create and maintain a custom Homebrew tap
3. install with `brew install <tap>/wifictl`

The project is structured as a clean standalone Go CLI:

- `main.go`: entry point
- `internal/cli`: argument parsing and usage
- `internal/app`: orchestration flow
- `internal/macos`: macOS system command backend
- `internal/profile`: static profile parsing and validation

## Roadmap

Near-term improvements:

- version output and build metadata
- stricter validation for profile values
- optional scan and current status commands
- Homebrew tap and formula

Out of scope for now:

- Linux support
- background SSID rule daemon
- GUI integration

## Development

Run tests:

```bash
make test
```

Build the CLI:

```bash
make build
```
