# wifictl

[English README](./README.md)

`wifictl` 是一个以 macOS 为优先目标的命令行工具，用来管理当前 Wi-Fi 服务的 IP 和 DNS 配置。

## 适用范围

`wifictl` 不再负责连接指定 Wi-Fi。

现在这个工具只聚焦在一件更稳定的事情上：管理当前已经连接好的 Wi-Fi 服务配置，包括：

- 查看当前 Wi-Fi 服务状态
- 把当前 Wi-Fi 服务切回 DHCP
- 给当前 Wi-Fi 服务加载一份静态 IP 配置
- 把当前 Wi-Fi 服务导出成 profile 文件
- 单独调整 DNS 设置

## 命令

```bash
wifictl status
wifictl dhcp
wifictl load <profile>
wifictl export <profile>
wifictl dns auto
wifictl dns <server...>
```

示例：

```bash
wifictl status
sudo wifictl dhcp
sudo wifictl load examples/office.conf
wifictl export office.conf
sudo wifictl dns auto
sudo wifictl dns 114.114.114.114
sudo wifictl dns 223.5.5.5 119.29.29.29
```

## 命令行为

- `status`
  输出当前 Wi-Fi 服务、设备、是否已关联、IPv4 模式、IP、掩码、网关和 DNS。
- `dhcp`
  将 IPv4 切换为 DHCP，并把 DNS 恢复为自动。
- `load <profile>`
  加载一份静态 IP profile，并应用其中的 DNS 设置。
- `export <profile>`
  将当前 Wi-Fi 服务配置导出为 profile 文件。
- `dns auto`
  只恢复 DNS 为自动，不改 IPv4 模式。
- `dns <server...>`
  只设置一个或多个 DNS，不改 IPv4 模式。

## Profile 格式

profile 使用简单的 `key=value` 格式。

示例 [`examples/office.conf`](/Users/aruis/develop/workspace-ai/test/wifictl/examples/office.conf)：

```ini
ip=192.168.10.88
mask=255.255.255.0
gateway=192.168.10.1
dns=192.168.10.2,223.5.5.5
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
sudo wifictl dns auto
sudo wifictl dns 114.114.114.114
```

`status` 和 `export` 一般不需要 `sudo`。

## 快速开始

构建：

```bash
make build
```

运行：

```bash
./dist/wifictl status
sudo ./dist/wifictl dhcp
sudo ./dist/wifictl load examples/office.conf
./dist/wifictl export examples/office.current.conf
sudo ./dist/wifictl dns 114.114.114.114
```

## 输出示例

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

## 后续计划

- 增加版本号和构建信息输出
- 增加更严格的 IPv4 和 DNS 校验
- 接入 Homebrew tap 和 formula
