# wifictl

[中文文档](./README.zh-CN.md)

`wifictl` is a macOS-first CLI for managing the IP and DNS configuration of a macOS network service.

## Scope

`wifictl` does not try to join a Wi-Fi network.

The tool is focused on a narrower and more reliable workflow on macOS:

- inspect the current network service status
- switch the target network service to DHCP
- load a static IP profile onto the target network service
- export the target network service config into a profile
- change only DNS settings when needed

By default, `wifictl` targets the `Wi-Fi` service. You can override that with `--service <name>`.

## Commands

```bash
wifictl version
wifictl [--service <name>] status
wifictl [--service <name>] dhcp
wifictl [--service <name>] load <profile>
wifictl [--service <name>] export <profile>
wifictl [--service <name>] dns reset
wifictl [--service <name>] dns <server...>
wifictl services
```

Examples:

```bash
wifictl version
wifictl status
wifictl --service Ethernet status
sudo wifictl --service Ethernet dhcp
sudo wifictl load examples/office.conf
wifictl export office.conf
sudo wifictl dns reset
sudo wifictl dns 223.5.5.5 223.6.6.6
wifictl services
```

## Command Behavior

- `status`
  Prints the target service, hardware port, device, IPv4 mode, IP, mask, gateway, and DNS servers. For Wi-Fi services it also prints association state and SSID when available.
- `version`
  Prints the application version, commit, and build time embedded in the binary.
- `dhcp`
  Sets IPv4 to DHCP and resets DNS to system default behavior.
- `load <profile>`
  Loads a static IP profile and applies its DNS settings.
- `export <profile>`
  Exports the current target service configuration into a profile file.
- `dns reset`
  Clears manual DNS settings and returns DNS resolution to the system default behavior without changing the IPv4 mode.
- `dns <server...>`
  Applies one or more DNS servers without changing the IPv4 mode.
- `services`
  Lists available macOS network services with hardware port and device.

## Profile Format

Profiles use a simple `key=value` format.

Example [`examples/office.conf`](/Users/aruis/develop/workspace-ai/test/wifictl/examples/office.conf):

```ini
ip=192.168.10.88
mask=255.255.255.0
gateway=192.168.10.1
dns=223.5.5.5,223.6.6.6
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
sudo wifictl dns reset
sudo wifictl dns 223.5.5.5 223.6.6.6
```

`status`, `export`, and `services` do not require `sudo` in normal cases.

## Quick Start

Build:

```bash
make build
```

Run:

```bash
./dist/wifictl version
./dist/wifictl status
./dist/wifictl services
./dist/wifictl --service Ethernet status
sudo ./dist/wifictl dhcp
sudo ./dist/wifictl load examples/office.conf
./dist/wifictl export examples/office.current.conf
sudo ./dist/wifictl dns reset
sudo ./dist/wifictl dns 223.5.5.5 223.6.6.6
```

## Example Output

Version output:

```text
version: v1.26.2
commit: ae062f4
built: 2026-03-12T07:06:05Z
```

```text
Service: Wi-Fi
Hardware Port: Wi-Fi
Device: en1
Associated: yes
IPv4: Manual
IP: 10.60.1.94
Mask: 255.255.0.0
Gateway: 10.60.1.254
DNS: 223.5.5.5, 223.6.6.6
```

Services output example:

```text
Ethernet	Ethernet	en0
Wi-Fi	Wi-Fi	en1
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

## Release

GitHub Actions is configured to:

- run CI on pushes to `main` and on pull requests
- create a GitHub Release when a tag matching `v*` is pushed
- build macOS archives for `darwin/arm64` and `darwin/amd64`

Release flow:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Release assets are uploaded as:

- `wifictl_<version>_Darwin_arm64.tar.gz`
- `wifictl_<version>_Darwin_x86_64.tar.gz`

## Roadmap

- add version output and release metadata
- add stricter IPv4 and DNS validation
- add Homebrew tap and formula
