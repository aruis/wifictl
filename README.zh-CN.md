# wifictl

[English README](./README.md)

`wifictl` 是一个用于查看和修改 macOS 网络服务 IP 与 DNS 配置的命令行工具。

## 它能做什么

`wifictl` 面向已经存在的 macOS network service，只负责配置管理，不处理连接 Wi-Fi 或管理 Wi-Fi 凭据这类入网动作。

你可以用它来：

- 查看当前网络服务状态
- 把目标网络服务切回 DHCP
- 给目标网络服务加载一份静态 IP 配置
- 把目标网络服务导出成 profile 文件
- 单独调整 DNS 设置

默认目标服务是 `Wi-Fi`，也可以通过 `--service <name>` 指定其他服务，例如 `Ethernet`。

## 命令

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

## 安装

使用 Homebrew：

```bash
brew install aruis/tap/wifictl
```

从源码构建：

```bash
make build
```

示例：

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

## 命令行为

- `status`
  输出目标服务、硬件端口、设备、IPv4 模式、IP、掩码、网关和 DNS。对于 Wi-Fi 服务，还会在可用时输出关联状态和 SSID。
- `version`
  输出当前二进制里嵌入的版本号、commit 和构建时间。
- `dhcp`
  将 IPv4 切换为 DHCP，并把 DNS 恢复为系统默认行为。
- `load <profile>`
  加载一份静态 IP profile，并应用其中的 DNS 设置。
- `export <profile>`
  将当前目标服务配置导出到 profile 文件。
- `dns reset`
  清除手动 DNS 设置，恢复为系统默认 DNS 行为，不改 IPv4 模式。
- `dns <server...>`
  只设置一个或多个 DNS，不改 IPv4 模式。
- `services`
  列出当前 macOS 可用的 network service、硬件端口和设备名。

## Profile 格式

profile 使用简单的 `key=value` 格式：

示例 [`examples/office.conf`](/Users/aruis/develop/workspace-ai/test/wifictl/examples/office.conf)：

```ini
ip=192.168.10.88
mask=255.255.255.0
gateway=192.168.10.1
dns=223.5.5.5,223.6.6.6
```

必填字段：

- `ip`
- `mask`
- `gateway`

可选字段：

- `dns`

## 权限

会修改网络配置的命令需要管理员权限：

```bash
sudo wifictl dhcp
sudo wifictl load office.conf
sudo wifictl dns reset
sudo wifictl dns 223.5.5.5 223.6.6.6
```

`status`、`export` 和 `services` 一般不需要 `sudo`。

## 快速开始

运行构建出的二进制：

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

## 输出示例

版本输出示例：

```text
version: v1.26.7
commit: 49dd605
built: 2026-03-12T08:12:59Z
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

可用服务输出示例：

```text
Ethernet	Ethernet	en0
Wi-Fi	Wi-Fi	en1
```

## 工作方式

`wifictl` 使用 Go 编写，并调用 macOS 自带的系统命令：

- `networksetup`
- `wdutil`
- `system_profiler`

项目结构：

- `main.go`：程序入口
- `internal/cli`：命令解析
- `internal/app`：执行编排
- `internal/macos`：macOS 平台实现
- `internal/profile`：profile 解析与导出

## 开发

运行测试：

```bash
make test
```

构建：

```bash
make build
```

## 发布

当前已经配置 GitHub Actions：

- 在 `main` 分支 push 和 PR 时自动跑 CI
- 在推送符合 `v*` 的 tag 时自动创建 GitHub Release
- 自动构建 `darwin/arm64` 和 `darwin/amd64` 两个 macOS 压缩包

发布流程：

```bash
git tag v0.1.0
git push origin v0.1.0
```

Release 产物命名为：

- `wifictl_<version>_Darwin_arm64.tar.gz`
- `wifictl_<version>_Darwin_x86_64.tar.gz`

## 说明

- `load <profile>` 在应用配置前会校验 IPv4 地址、子网掩码、网关和 DNS。
- 当前 release 会发布 macOS `arm64` 和 `amd64` 两个构建产物。
