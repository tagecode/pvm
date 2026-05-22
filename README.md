# PVM — PHP Version Manager for Windows

PVM 是 Windows 平台的 PHP 版本管理 CLI，类似 nvm，用于安装、切换和管理多个 PHP 版本。

## 系统要求

- Windows 10 / 11
- 64 位系统请使用 **x64** 构建；32 位系统请使用 **x86** 构建
- 无需管理员权限（默认 per-user 安装）

## 安装

从 [GitHub Releases](https://github.com/tagecode/pvm/releases) 下载对应架构的产物：

| 文件 | 说明 |
|------|------|
| `pvm-v0.1.0-windows-x64.msi` | 64 位 Windows 安装包（推荐） |
| `pvm-v0.1.0-windows-x86.msi` | 32 位 Windows 安装包 |
| `pvm-v0.1.0-windows-x64.zip` | 64 位便携包（含 `install.ps1`） |
| `pvm-v0.1.0-windows-x64.exe` | 64 位单文件，可直接运行 |

**MSI 安装：**

```powershell
msiexec /i pvm-v0.1.0-windows-x64.msi
```

**ZIP 安装：**

```powershell
Expand-Archive pvm-v0.1.0-windows-x64.zip -DestinationPath .
cd pvm-v0.1.0-windows-x64
.\install.ps1
```

安装后**重新打开终端**，验证：

```powershell
pvm version
```

## 快速开始

```powershell
# 安装 PHP 8.3（默认自动 use + 配置环境）
pvm install 8.3

# 当前终端即可验证（无需单独 setup）
php -v
pvm current
pvm list

# 切换其它版本
pvm use 8.2
php -v
```

> MSI / ZIP 安装包会预置 `PVM_HOME` 与 `current` 的 PATH；首次 `use` 也会自动 setup。  
> 若关闭了 `defaults.auto_use_after_install`，安装后需手动 `pvm use <version>`。  
> 其它已打开的终端可运行 `pvm refresh` 获取 PATH 刷新命令。

## 常用命令

| 命令 | 说明 |
|------|------|
| `pvm install <version>` | 安装 PHP（如 `8.3`、`8.3.31`） |
| `pvm install --from-zip <path>` | 从本地 zip 离线安装 |
| `pvm use <version>` | 切换当前 PHP 版本 |
| `pvm list` / `pvm ls` | 列出已安装版本 |
| `pvm list-remote` / `pvm ls-remote` | 列出可下载版本 |
| `pvm setup` | 手动配置用户 PATH（首次 `use` 会自动执行） |
| `pvm unsetup` | 移除 PVM 环境变量 |
| `pvm alias prod 8.3.31` | 设置别名 |
| `pvm refresh` | 输出当前会话 PATH 刷新命令 |

全局选项：`--json`、`-q`、`--dry-run`、`-y`

## 配置

配置文件位于 `%PVM_HOME%\config.toml`（默认 `%USERPROFILE%\.pvm\config.toml`）。

常用默认值：

```toml
[defaults]
arch = "auto"          # 跟随 pvm.exe 架构
thread_safe = false    # 默认 NTS
link_mode = "junction"
auto_use_after_install = true   # 安装后自动 use

[mirror]
url = "https://downloads.php.net/~windows/releases/"
```

显式安装 x86 PHP（在 64 位系统上）：

```powershell
pvm install 8.3 --arch x86
```

## 目录结构

```
%PVM_HOME%\
├── versions\          # 已安装的 PHP 版本
├── current\           # 当前激活版本（junction）
├── cache\             # 下载缓存
├── aliases\           # 版本别名
└── config.toml
```

## 从源码构建

需要 Go 1.25+：

```powershell
go build -o pvm.exe .
go test ./...
```

## 文档

- [产品需求（PRD）](docs/PRD.md)
- [技术设计（TDD）](docs/TDD.md)
- [功能清单（FEATURE）](docs/FEATURE.md)
- [Release Notes 编写说明](docs/releases/README.md)

## 许可证

[MIT](LICENSE)
