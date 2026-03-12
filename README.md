# wifictl

[中文文档](./README.zh-CN.md)

`wifictl` is a macOS-first CLI for managing the IP and DNS configuration of the current Wi-Fi service.

## Scope

`wifictl` does not try to join a Wi-Fi network anymore.

The tool is intentionally focused on a narrower and more reliable workflow on macOS:

- inspect the current Wi-Fi service status
- switch the current Wi-Fi service to DHCP
- load a static IP profile onto the current Wi-Fi service
- export the current Wi-Fi service config into a profile
- change only DNS settings when needed

## Commands

```bash
wifictl status
wifictl dhcp
wifictl load <profile>
wifictl export <profile>
wifictl dns auto
wifictl dns <server...>
```

Examples:

```bash
wifictl status
sudo wifictl dhcp
sudo wifictl load examples/office.conf
wifictl export office.conf
sudo wifictl dns auto
sudo wifictl dns 114.114.114.114
sudo wifictl dns 223.5.5.5 119.29.29.29
```

## Command Behavior

- `status`
  Prints the current Wi-Fi service, device, association state, IPv4 mode, IP, mask, gateway, and DNS servers.
- `dhcp`
  Sets IPv4 to DHCP and resets DNS to automatic.
- `load <profile>`
  Loads a static IP profile and applies its DNS settings.
- `export <profile>`
  Exports the current Wi-Fi service configuration into a profile file.
- `dns auto`
  Resets DNS servers to automatic without changing the IPv4 mode.
- `dns <server...>`
  Applies one or more DNS servers without changing the IPv4 mode.

## Profile Format

Profiles use a simple `key=value` format.

Example [`examples/office.conf`](/Users/aruis/develop/workspace-ai/test/wifictl/examples/office.conf):

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

## Permissions

Commands that modify network settings require administrator privileges:

```bash
sudo wifictl dhcp
sudo wifictl load office.conf
sudo wifictl dns auto
sudo wifictl dns 114.114.114.114
```

`status` and `export` do not require `sudo` in normal cases.

## Quick Start

Build:

```bash
make build
```

Run:

```bash
./dist/wifictl status
sudo ./dist/wifictl dhcp
sudo ./dist/wifictl load examples/office.conf
./dist/wifictl export examples/office.current.conf
sudo ./dist/wifictl dns 114.114.114.114
```

## Example Output

```text
Service: Wi-Fi
Device: en1
Associated: yes
IPv4: Manual
IP: 10.60.1.94
Mask: 255.255.0.0
Gateway: 10.60.1.254
DNS: 114.114.114.114
```

## Implementation

`wifictl` relies on stable macOS system commands:

- `networksetup`
- `wdutil`
- `system_profiler`

The project is implemented in Go and structured as:

- `main.go`: entry point
- `internal/cli`: command parsing
- `internal/app`: command orchestration
- `internal/macos`: macOS backend
- `internal/profile`: profile parsing and export

## Development

Run tests:

```bash
make test
```

Build:

```bash
make build
```

## Roadmap

- add version output and release metadata
- add stricter IPv4 and DNS validation
- add Homebrew tap and formula
