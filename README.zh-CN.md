# wifictl

[English README](./README.md)

`wifictl` 是一个以 macOS 为优先目标的命令行工具，用来连接指定 Wi-Fi，并在连接成功后立即应用对应的 IP 策略。

## 项目目标

- 通过命令行连接指定的 Wi-Fi SSID
- 在连接成功后，立即切换为 `DHCP` 或 `静态 IP`
- 保持用户命令简短直接
- 以单独的 CLI 工具形式发布，后续适配 Homebrew

## 当前状态

当前版本已经实现：

- `connect <ssid> dhcp`
- `connect <ssid> static <profile>`
- 对应的短参数形式
- 可选的 Wi-Fi 密码参数
- 静态 profile 文件校验
- 应用完成后的网络状态输出

## 为什么做这个工具

`wifictl` 解决的是一个非常具体的执行型场景：

- 连接到某个 Wi-Fi
- 立即应用 DHCP 或静态 IP 配置
- 直接输出结果

第一版不依赖 macOS `Location`，也不做后台守护进程或复杂规则管理。

## 适用范围

第一版只支持 macOS。

暂不支持 Linux，原因是不同发行版在 Wi-Fi 管理和 IP 配置上的基础设施差异太大，比如：

- `NetworkManager`
- `systemd-networkd`
- `wpa_supplicant`
- `iwd`

如果一开始就同时兼容 Linux 和 macOS，会明显增加实现和维护复杂度。

## 命令设计

标准写法：

```bash
wifictl connect <ssid> dhcp
wifictl connect <ssid> static <profile>
```

简写：

```bash
wifictl -c <ssid> -d
wifictl -c <ssid> -s <profile>
```

使用示例：

```bash
wifictl connect OfficeWiFi dhcp
wifictl connect OfficeWiFi static office.conf
wifictl connect OfficeWiFi dhcp --password secret

wifictl -c OfficeWiFi -d
wifictl -c OfficeWiFi -s office.conf
wifictl -c OfficeWiFi -d -p secret
```

## 快速开始

构建：

```bash
make build
```

运行：

```bash
sudo ./dist/wifictl connect OfficeWiFi dhcp
sudo ./dist/wifictl connect OfficeWiFi static examples/office.conf
```

简写形式：

```bash
sudo ./dist/wifictl -c OfficeWiFi -d
sudo ./dist/wifictl -c OfficeWiFi -s examples/office.conf
```

## 预期行为

执行 `wifictl connect <ssid> dhcp` 时：

1. 解析当前 macOS 上的 Wi-Fi 设备名和网络服务名
2. 连接目标 SSID
3. 等待直到当前 Wi-Fi 确认切换为目标 SSID
4. 对该 Wi-Fi 服务应用 DHCP
5. 输出最终网络状态

执行 `wifictl connect <ssid> static <profile>` 时：

1. 解析当前 macOS 上的 Wi-Fi 设备名和网络服务名
2. 连接目标 SSID
3. 等待直到当前 Wi-Fi 确认切换为目标 SSID
4. 读取静态 profile 文件
5. 设置手动 IP、子网掩码、网关和 DNS
6. 输出最终网络状态

## Profile 格式

静态 IP 配置文件使用简单的 `key=value` 格式。

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

## 密码支持

如果目标 Wi-Fi 没有提前保存在 macOS 中，可以显式传入密码：

```bash
wifictl connect OfficeWiFi dhcp --password secret
wifictl connect OfficeWiFi static office.conf --password secret
wifictl -c OfficeWiFi -d -p secret
wifictl -c OfficeWiFi -s office.conf -p secret
```

如果不传密码，`wifictl` 会依赖以下条件之一：

- 目标网络是开放网络
- 目标网络密码已被 macOS 记住

## 示例输出

```text
Connected to OfficeWiFi
Mode: DHCP
Device: en0
Service: Wi-Fi
IP: 192.168.10.25
DNS: 192.168.10.2, 223.5.5.5
```

## 权限说明

macOS 下修改 Wi-Fi 和 IP 配置通常需要管理员权限。

当前建议的使用方式是：

```bash
sudo wifictl connect OfficeWiFi dhcp
sudo wifictl connect OfficeWiFi static office.conf
```

工具本身不负责内部提权。

## 错误处理目标

CLI 应该明确区分以下失败场景：

- 找不到 Wi-Fi 硬件设备
- 找不到对应的 Wi-Fi 网络服务
- 连接目标 SSID 失败
- 等待目标 SSID 激活超时
- 静态 profile 缺字段或格式错误
- DHCP 应用失败
- 静态 IP 或 DNS 应用失败

## macOS 实现说明

当前实现依赖 macOS 自带系统命令，例如：

- `networksetup`
- `ipconfig`
- `ifconfig`

第一版优先使用稳定的系统命令封装，而不是私有 API。

## 项目结构

- `main.go`：程序入口
- `internal/cli`：命令行参数解析和帮助文案
- `internal/app`：主执行流程编排
- `internal/macos`：macOS 平台实现
- `internal/profile`：静态 profile 读取和校验

## 开发

运行测试：

```bash
make test
```

构建 CLI：

```bash
make build
```

## 后续计划

近期可迭代方向：

- 版本号和构建信息输出
- 更严格的 profile 值校验
- 增加查看当前状态、扫描 Wi-Fi 等辅助命令
- 接入 Homebrew tap 和 formula

当前不打算做：

- Linux 支持
- 后台 SSID 规则守护进程
- GUI 集成
