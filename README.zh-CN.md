# wifictl

[English README](./README.md)

`wifictl` 是一个以 macOS 为优先目标的命令行工具，用来管理 macOS 网络服务的 IP 和 DNS 配置。

## 适用范围

`wifictl` 不负责连接指定 Wi-Fi。

这个工具只聚焦在更稳定的一层：管理目标网络服务配置，包括：

- 查看当前网络服务状态
- 把目标网络服务切回 DHCP
- 给目标网络服务加载一份静态 IP 配置
- 把目标网络服务导出成 profile 文件
- 单独调整 DNS 设置

默认目标服务是 `Wi-Fi`，也可以通过 `--service <name>` 显式指定，例如 `Ethernet`。

## 命令

```bash
wifictl [--service <name>] status
wifictl [--service <name>] dhcp
wifictl [--service <name>] load <profile>
wifictl [--service <name>] export <profile>
wifictl [--service <name>] dns reset
wifictl [--service <name>] dns <server...>
wifictl services
```

示例：

```bash
wifictl status
wifictl --service Ethernet status
sudo wifictl --service Ethernet dhcp
sudo wifictl load examples/office.conf
wifictl export office.conf
sudo wifictl dns reset
sudo wifictl dns 223.5.5.5 223.6.6.6
sudo wifictl dns 223.5.5.5 223.6.6.6
wifictl services
```

## 命令行为

- `status`
  输出目标服务、硬件端口、设备、IPv4 模式、IP、掩码、网关和 DNS。对于 Wi-Fi 服务，还会尽量输出是否已关联和 SSID。
- `dhcp`
  将 IPv4 切换为 DHCP，并把 DNS 恢复为系统默认行为。
- `load <profile>`
  加载一份静态 IP profile，并应用其中的 DNS 设置。
- `export <profile>`
  将当前目标服务配置导出为 profile 文件。
- `dns reset`
  清除手动 DNS 设置，恢复为系统默认 DNS 行为，不改 IPv4 模式。
- `dns <server...>`
  只设置一个或多个 DNS，不改 IPv4 模式。
- `services`
  列出当前 macOS 可用的 network service、硬件端口和设备名。

## Profile 格式

profile 使用简单的 `key=value` 格式。

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

构建：

```bash
make build
```

运行：

```bash
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

## 实现说明

`wifictl` 依赖 macOS 的系统命令：

- `networksetup`
- `wdutil`
- `system_profiler`

项目使用 Go 实现，结构如下：

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

项目已经配置 GitHub Actions：

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

## 后续计划

- 增加版本号和构建信息输出
- 增加更严格的 IPv4 和 DNS 校验
- 接入 Homebrew tap 和 formula
