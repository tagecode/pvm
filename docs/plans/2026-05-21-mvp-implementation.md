# PVM MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付 Windows 平台 PHP 版本管理 CLI 的 MVP：install、list、use、setup、alias、current 及全局 flags，满足 PRD §10.1 验收清单。

**Architecture:** Cobra 命令层 + internal/app 业务编排 + pkg 领域包；Viper 管理 config.toml；Junction 切换 current 链接；RemoteIndex 解析 windows.php.net 发布页。

**Tech Stack:** Go 1.22+ · Cobra · Viper · slog · archive/zip · net/http

**关联文档:** [PRD.md](../PRD.md) · [TDD.md](../TDD.md)

**MVP 状态（2026-05-22）：** 代码与自动化验收完成；本机在线 `pvm install 8.3` 因网络中断未通过，已由 mock HTTP 集成测试覆盖同等代码路径。PRD §10.1 除 Win10/11 双环境手工矩阵外均已验证（含 `php.ini` 自动生成与缺模板拒绝安装）。

---

## File Structure (MVP)

```text
pvm/
├── cmd/
│   ├── root.go           # 全局 flags + Viper
│   ├── helpers.go        # App 初始化、错误 exit
│   ├── app.go            # App 服务聚合
│   ├── json.go           # PRD JSON schema
│   ├── version.go
│   ├── install.go
│   ├── uninstall.go
│   ├── use.go
│   ├── list.go
│   ├── list_remote.go
│   ├── current.go
│   ├── which.go
│   ├── info.go
│   ├── alias.go
│   ├── default.go
│   ├── setup.go          # setup + unsetup + refresh
│   ├── path.go
│   └── link_mode.go
├── internal/
│   ├── config/           # config.go, arch.go, linkmode.go
│   └── app/
│       ├── errors.go
│       ├── registry.go
│       ├── installer.go
│       ├── switcher.go
│       ├── alias.go
│       ├── setup.go
│       └── acceptance_windows_test.go  # PRD §10.1 集成验收
├── pkg/
│   ├── paths/paths.go
│   ├── php/              # version, spec, zipname, remote
│   ├── download/
│   ├── win/
│   └── ui/output.go
└── .github/workflows/test.yml
```

---

## Task 1: 基础设施 — paths + config + ui

**Files:**
- Create: `pkg/paths/paths.go`
- Create: `internal/config/config.go`
- Create: `pkg/ui/output.go`
- Modify: `cmd/root.go`

- [x] **Step 1:** 实现 `paths.Home()` — 读取 `PVM_HOME`，默认 `%USERPROFILE%\.pvm`
- [x] **Step 2:** 实现 `paths.VersionsDir`, `CurrentLink`, `AliasesDir`, `CacheDir`, `EnsureLayout`
- [x] **Step 3:** Viper 加载 `config.toml`，设置 defaults（link_mode=junction, verify_sha256=true；arch=auto）
- [x] **Step 4:** `ui` 包 — `PrintJSON`, `PrintError`, `JSONEnabled()` 读 persistent flag
- [x] **Step 5:** 更新 `root.go` — persistent flags、`SilenceUsage`
- [x] **Step 6:** `go build ./...` 通过
- [ ] **Deferred:** `internal/log/log.go`（slog）与 `-v` 日志输出

---

## Task 2: PHP 版本领域 — spec + version + zipname

**Files:**
- Create: `pkg/php/spec.go`, `pkg/php/version.go`, `pkg/php/zipname.go`
- Create: `pkg/php/version_test.go`

- [x] **Step 1:** 写失败测试 — `ParseSpec("8.3")` → SpecMinor; `ParseSpec("8.3.10")` → SpecExact
- [x] **Step 2:** 实现 `ParseSpec`, `DefaultVC`, `DefaultArch`
- [x] **Step 3:** 写失败测试 — `Version{8,3,10,x64,false,vs16}.ID()` 期望 `8.3.10-x64-nts-vs16`
- [x] **Step 4:** 实现 `ZipName()` 符合官方 `-nts-Win32-` 命名规则
- [x] **Step 5:** `go test ./pkg/php/...` 通过

---

## Task 3: 下载与校验

**Files:**
- Create: `pkg/download/client.go`, `pkg/download/checksum.go`
- Create: `pkg/download/checksum_test.go`

- [x] **Step 1:** 测试 checksum 文件解析（SHA256 + SHA1）
- [x] **Step 2:** 实现 `ParseChecksumSum`, `VerifyChecksum`, `VerifyFile`
- [x] **Step 3:** 实现 `Client.Fetch` — 写入 `.part` 后 rename；连接/TLS 超时
- [ ] **Step 3b:** HTTP proxy 接入 Client（config 已占位）
- [x] **Step 4:** `go test ./pkg/download/...` 通过

---

## Task 4: Windows 平台 — junction + env

**Files:**
- Create: `pkg/win/junction_windows.go`, `pkg/win/junction_other.go`
- Create: `pkg/win/env_windows.go`

- [x] **Step 1:** 实现 `CreateLink(link, target, mode)` — junction 用 mklink /J
- [x] **Step 2:** 实现 `RemoveLink(link)`
- [x] **Step 3:** 实现 `AppendUserPath(entry)`, `GetEnvStatus()`（读用户注册表）
- [x] **Step 4:** Windows 集成测试 — junction 切换（`app_windows_test.go`）

---

## Task 5: 远程索引 — remote.go

**Files:**
- Create: `pkg/php/remote.go`

- [x] **Step 1:** 实现 `RemoteIndex.List` — 抓取 releases 页，正则提取 zip 链接
- [x] **Step 2:** 实现 `Resolve(spec, opts)` — minor 取最新 patch
- [x] **Step 3:** 合并 archives 回退逻辑
- [x] **Step 4:** 对接 `sha256sum.txt` / `archives/sha1sum.txt` 校验哈希
- [x] **Step 5:** `list-remote --major 8`（需网络，已手工验证）

---

## Task 6: 业务层 — registry + alias

**Files:**
- Create: `internal/app/registry.go`, `internal/app/alias.go`, `internal/app/errors.go`

- [x] **Step 1:** `Registry.List()` — 扫描 versions/ 目录，解析目录名
- [x] **Step 2:** 标记 active（state/active + junction）与 default（aliases/default）
- [x] **Step 3:** `AliasStore` CRUD + `ResolveTarget()` 防循环
- [x] **Step 4:** 单元测试 alias 解析 + `ResolveID()`

---

## Task 7: 业务层 — installer

**Files:**
- Create: `internal/app/installer.go`

- [x] **Step 1:** 组装 Install 流程（见 TDD §4.2）
- [x] **Step 2:** zip 解压至 `versions/{id}.tmp.{pid}`
- [x] **Step 3:** 复制 `php.ini-development` → `php.ini`
- [x] **Step 4:** atomic rename；失败 cleanup
- [x] **Step 5:** `Uninstall` — 禁止删除 active 版本
- [x] **Bonus:** `--from-zip` 离线安装

---

## Task 8: 业务层 — switcher + setup

**Files:**
- Create: `internal/app/switcher.go`, `internal/app/setup.go`

- [x] **Step 3:** 实现 `Switcher.Use` — 解析 alias/版本 → 切换 current 链接；首次 use 自动 setup；会话 PATH 刷新
- [x] **Step 2:** `Switcher.Current`, `Which`
- [x] **Step 3:** `Setup.Run` — 追加用户 PATH + PHP_HOME
- [x] **Step 4:** `Setup.Status` — path 命令输出
- [x] **Bonus:** `Setup.Unsetup`, `pvm refresh`, MSI/ZIP 预置 `current` PATH, `defaults.auto_use_after_install = true`

---

## Task 9: CLI 命令 wiring

**Files:**
- Create: 所有 `cmd/*.go`
- Create: `cmd/helpers.go`, `cmd/app.go`

- [x] **Step 1:** `NewApp()`, `handleErr()` exit code 映射
- [x] **Step 2:** `install`, `use`, `list`, `current` — 完整 wiring
- [x] **Step 3:** `list-remote`, `which`, `info`, `alias`, `default`, `setup`, `path`, `uninstall`
- [x] **Step 4:** list/current/info/which/install 等支持 `--json`
- [x] **Step 5:** `pvm version` 输出版本号
- [x] **Bonus:** `unsetup`, `refresh`, `link-mode`
- [ ] **Deferred:** `-v` 详细日志、`-no-color` 实际生效

---

## Task 10: MVP 验收

- [x] **Step 1:** `go test ./...` 全部通过
- [x] **Step 2:** `go build -o pvm.exe .` 成功
- [x] **Step 3:** PRD §10.1 验收 — 见 `internal/app/acceptance_windows_test.go` + CLI 离线脚本
- [x] **Step 4:** 修复编译阻塞（env_windows.go）、补全 link-mode

### PRD §10.1 快速对照

| 验收项 | 状态 |
|--------|------|
| `pvm install 8.3` 在线安装 | ✅ mock HTTP（`TestAcceptanceOnlineInstallMock`）；⚠️ 本机网络实测失败 |
| `pvm use` + 新终端 `php -v` | ✅ 首次 use 自动 setup + use 后会话刷新 |
| `pvm setup` PATH/PHP_HOME | ✅ 手动 setup + 首次 use 自动 setup |
| `pvm list-remote --major 8` | ✅ 本机网络实测 + 官方 HTML 1822 zip 解析 |
| alias + use prod | ✅ `TestAcceptanceAliasUseProd` + CLI |
| `current --json` | ✅ CLI 实测 |
| SHA256 校验失败拒绝 | ✅ `TestChecksumVerificationFails` |
| uninstall 激活版本拒绝 | ✅ `TestUninstallActiveVersion` |
| 所有命令 `--help` | ✅ |
| Win10/11 x64 双环境 | ⚠️ 当前仅 Win10 x64 单环境验证 |
| `defaults.arch=auto` | ✅ `arch_test.go` |
| `install --arch x86` | ✅ `TestAcceptanceInstallArchX86FromZip` + docs/zip |
| A→B→A 版本互切 | ✅ `TestAcceptanceOfflineInstallUseSwitch` |
| setup 保留系统 PATH / unsetup 恢复 | ✅ `TestAcceptanceSetupUnsetupPreservesOtherPath` |
| 安装后生成 `php.ini` | ✅ `TestInstallReleaseFromZip` + `TestAcceptanceInstallEnsuresPhpIni` |
| 缺 ini 模板拒绝安装 | ✅ `TestInstallReleaseFailsWithoutIniTemplate` |

---

## 建议 commit 顺序

1. `feat: add paths, config, and ui foundation`
2. `feat: add php version parsing and zip naming`
3. `feat: add download client and checksum verification`
4. `feat: add windows junction and env helpers`
5. `feat: add remote index and app services`
6. `feat: wire MVP cobra commands`
7. `fix: registry access types and add link-mode command`

---

## 预估工时

| Task | 工时 | 实际 |
|------|------|------|
| Task 1–2 | 1 天 | ✅ |
| Task 3–4 | 1 天 | ✅ |
| Task 5–6 | 1 天 | ✅ |
| Task 7–8 | 2 天 | ✅ |
| Task 9–10 | 2 天 | ✅ |
| **合计** | **约 7 个工作日** | **~7 天当量（MVP 可交付）** |

---

## 下一步（V1.0 或 MVP 收尾）

1. 网络稳定时复测 `pvm install 8.3` 真实下载（可选，mock 已覆盖逻辑）
2. Win11 / x86 pvm 环境手工矩阵（PRD §12.3）
3. 可选：接入 download proxy、`internal/log` + `-v`、`golangci-lint`
4. 进入 V1.0：config ini / ext / `.pvmrc`
