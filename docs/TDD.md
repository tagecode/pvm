# PVM 技术设计文档（TDD）

| 属性 | 内容 |
|------|------|
| **文档版本** | v1.0.0 |
| **文档状态** | Draft |
| **创建日期** | 2026-05-21 |
| **关联文档** | [PRD.md](./PRD.md) · [FEATURE.md](./FEATURE.md) |
| **实现计划** | [plans/2026-05-21-mvp-implementation.md](./plans/2026-05-21-mvp-implementation.md) |

---

## 1. 设计目标

本文档将 PRD 中的产品需求转化为可实现的模块设计、接口契约、数据流与关键算法，指导 MVP 及后续版本开发。

**设计原则：**

1. **单一职责**：`cmd` 只做参数解析与输出，`internal/app` 编排业务，`pkg/*` 提供可测试的领域逻辑。
2. **可测试性**：所有 I/O 通过 interface 注入，核心逻辑不依赖真实网络与注册表。
3. **原子操作**：安装/切换采用「临时目录 → 校验 → rename/swap」模式。
4. **Windows 优先**：平台相关代码集中在 `pkg/win`，非 Windows 编译仅用于单元测试（stub）。

---

## 2. 模块划分

### 2.1 模块依赖图

```text
                    cmd/*
                      │
                      ▼
              internal/config ──────────────┐
                      │                     │
                      ▼                     │
               internal/app ◄───────────────┘
                 /    |    \
                /     |     \
               ▼      ▼      ▼
          pkg/php  pkg/win  pkg/download
               \      |      /
                \     |     /
                 ▼    ▼    ▼
               pkg/paths  pkg/ui
```

**依赖规则：**

- `pkg/*` 不得 import `cmd` 或 `internal/app`
- `internal/app` 可 import 所有 `pkg/*`
- `cmd/*` 仅调用 `internal/app` 与 `pkg/ui`

### 2.2 模块职责

| 模块 | 路径 | 职责 |
|------|------|------|
| CLI | `cmd/` | Cobra 命令树、全局 flags、exit code |
| 配置 | `internal/config/` | Viper 加载、默认值、PVM_HOME 解析 |
| 业务编排 | `internal/app/` | Install、Switch、Alias、Setup 等用例 |
| 路径 | `pkg/paths/` | PVM_HOME 及子目录路径常量与 Ensure 函数 |
| PHP 领域 | `pkg/php/` | 版本解析、远程索引、zip 命名、VC 映射 |
| Windows | `pkg/win/` | Junction、用户/系统 PATH、注册表 |
| 下载 | `pkg/download/` | HTTP 客户端、缓存、SHA256 |
| 输出 | `pkg/ui/` | 表格、JSON、颜色、错误格式化 |

---

## 3. 核心数据结构

### 3.1 VersionSpec — 用户输入的版本规格

```go
// pkg/php/spec.go

type SpecKind int
const (
    SpecExact SpecKind = iota  // 8.3.10
    SpecMinor                  // 8.3 → 取最新 patch
    SpecLatest                 // latest
    SpecAlias                  // prod → 查 aliases/
)

type VersionSpec struct {
    Kind   SpecKind
    Major  int
    Minor  int
    Patch  int
    Alias  string
    Raw    string
}
```

### 3.2 Version — 完整 PHP 构建标识

```go
// pkg/php/version.go

type Version struct {
    Major      int
    Minor      int
    Patch      int
    Arch       string // "x64" | "x86"
    ThreadSafe bool
    VC         string // "vc15" | "vs16" | "vs17"
}

func (v Version) ID() string
func (v Version) String() string
func (v Version) ZipName() string
func (v Version) DownloadURL(base string) string
```

**ID 格式：** `{major}.{minor}.{patch}-{arch}-{ts|nts}-{vc}`

**示例：** `8.3.10-x64-nts-vs16`

### 3.3 InstalledVersion — 本地已安装记录

```go
// internal/app/registry.go

type InstalledVersion struct {
    Version   php.Version
    Path      string
    Active    bool
    Default   bool
    InstalledAt time.Time
}
```

### 3.4 RemoteRelease — 远程发布索引条目

```go
// pkg/php/remote.go

type RemoteRelease struct {
    Version    php.Version
    URL        string
    SHA256     string
    ReleaseDate time.Time
    Archived   bool
}
```

### 3.5 Alias — 版本别名

```go
// internal/app/alias.go

type Alias struct {
    Name    string
    Target  string // Version ID 或另一个 alias（禁止循环）
}
```

存储：`%PVM_HOME%\aliases\<name>` 文件内容为单行 Version ID。

---

## 4. 接口设计

### 4.1 RemoteIndex — 远程版本索引

```go
type RemoteIndex interface {
    List(ctx context.Context, filter RemoteFilter) ([]RemoteRelease, error)
    Resolve(ctx context.Context, spec VersionSpec, opts ResolveOptions) (RemoteRelease, error)
}

type RemoteFilter struct {
    Major *int
    Minor *int
}
```

**实现：** `pkg/php/remote_http.go` — 抓取 releases 页面 HTML/JSON，解析 zip 链接与 sha256sum。

**解析策略：**

1. GET `{mirror}/releases/` 或 sha256 索引
2. 正则匹配 `php-{version}-Win32-{vc}{ts|nts}-{arch}.zip`
3. 合并 archives 目录结果
4. 缓存至 `%PVM_HOME%\cache\index.json`（TTL 1h，可配置）

### 4.2 Installer — 安装器

```go
type Installer interface {
    Install(ctx context.Context, spec VersionSpec, opts InstallOptions) (*InstalledVersion, error)
    Uninstall(ctx context.Context, id string, opts UninstallOptions) error
}

type InstallOptions struct {
    Arch       string
    ThreadSafe *bool
    VC         string
    FromZip    string
    Reinstall  bool
    DryRun     bool
}
```

**Install 流程：**

```text
1. ResolveOptions → 补全 arch/ts/vc 默认值
2. RemoteIndex.Resolve → 得到 RemoteRelease
3. 若 versions/{id} 已存在且 !Reinstall → 返回 ErrAlreadyInstalled
4. download.Fetch(zip) → cache/{id}.zip
5. checksum.Verify(zip, sha256)
6. extract.ToTemp(zip) → versions/{id}.tmp.{pid}
7. postInstall: 若无 php.ini，复制 php.ini-development → php.ini（无 development 则用 php.ini-production；三者均缺则失败）
8. rename temp → versions/{id}
9. 若 defaults.auto_use_after_install → Switcher.Use(id)（含自动 setup + 会话 PATH 刷新）
```

### 4.3 Switcher — 版本切换

```go
type Switcher interface {
    Use(ctx context.Context, target string, opts UseOptions) (UseResult, error)
    Current(ctx context.Context) (*InstalledVersion, error)
    Which(ctx context.Context, target string) (string, error)
}

type UseResult struct {
    ID           string
    SetupApplied bool
}

type UseOptions struct {
    DryRun bool
}
```

**Use 流程：**

```text
1. Resolve target: alias → version id
2. 验证 versions/{id}/php.exe 存在
3. linkMode := config.defaults.link_mode
4. win.CreateLink(current, versions/{id}, linkMode)
5. SetupService.EnsureUserEnvironment() — 首次 use 自动 setup
6. win.RefreshSessionEnv(current) — 当前终端立即可用 php
7. 输出确认信息
```

**原子性：** 先创建 `current.new` 链接，成功后再删除 `current` 并重命名（或 swap）。

### 4.4 AliasStore — 别名存储

```go
type AliasStore interface {
    Set(name, versionID string) error
    Get(name string) (string, error)
    Delete(name string) error
    List() (map[string]string, error)
    Resolve(name string) (php.Version, error) // 跟踪 alias 链，检测循环
}
```

### 4.5 EnvManager — 环境变量

```go
type EnvManager interface {
    Setup(ctx context.Context, opts SetupOptions) error
    EnsureUserEnvironment(ctx context.Context, opts SetupOptions) (applied bool, err error)
    Unsetup(ctx context.Context) error
    Status() (*EnvStatus, error)
    RefreshScript() (powershell, cmd string)
}

type SetupOptions struct {
    System bool // 需管理员
}
```

**Setup 写入：**

| 变量 | 值 | 级别 |
|------|-----|------|
| PATH（前置） | `%PVM_HOME%\current` | User（默认）或 System |
| PHP_HOME | `%PVM_HOME%\current` | 同上 |
| PVM_HOME | `%USERPROFILE%\.pvm` | User（MSI / install.ps1 写入） |

**MSI / ZIP `install.ps1`：** 除 `pvm.exe` 目录外，预置 `%USERPROFILE%\.pvm\current` 到用户 PATH 及 `PHP_HOME`。

使用 `pkg/win/registry` 操作用户/系统环境变量（`HKCU\Environment` / `HKLM\...\Environment`）。

### 4.6 Downloader — 下载与校验

```go
type Downloader interface {
    Fetch(ctx context.Context, url, dest string) error
    FetchCached(ctx context.Context, url, cacheKey string) (localPath string, err error)
}

type ChecksumVerifier interface {
    SHA256File(path string) (string, error)
    Verify(path, expected string) error
}
```

---

## 5. 关键算法

### 5.1 VC 版本自动映射

```go
func DefaultVC(major, minor int) string {
    switch {
    case major == 7 && minor == 4:
        return "vc15"
    case major == 8 && minor <= 3:
        return "vs16"
    case major == 8 && minor >= 4:
        return "vs17"
    case major >= 9:
        return "vs17"
    default:
        return "vs16"
    }
}
```

### 5.2 架构默认值

```go
func DefaultArch() string {
    if runtime.GOARCH == "386" {
        return "x86"
    }
    return "x64"
}
```

### 5.3 Zip 文件名构造

Windows PHP zip 命名规则（官方）：

```text
php-{major}.{minor}.{patch}-Win32-{vc}{-nts|空}{-x64|空}.zip

示例：
  php-8.3.10-Win32-vs16-x64-nts.zip   (NTS x64)
  php-8.3.10-Win32-vs16-x64.zip       (TS x64, 无 -nts 后缀)
  php-7.4.33-Win32-vc15-x86-nts.zip
```

**实现注意：** TS 构建文件名不含 `-nts`；NTS 含 `-nts`。解析远程列表时双向匹配。

### 5.4 SHA256 校验

1. 下载 `{mirror}/releases/sha256sum.txt`（或 archives 对应文件）
2. 解析 `{hash}  {filename}` 行
3. 本地 `sha256.Sum256` 比对
4. 不匹配 → `ErrChecksumMismatch`，删除已下载文件

### 5.5 Junction 创建（Windows）

```go
// pkg/win/junction_windows.go

func CreateJunction(link, target string) error {
    // 1. os.RemoveAll(link) if exists
    // 2. exec.Command("cmd", "/c", "mklink", "/J", link, target)
    // 3. 验证 link 指向 target
}
```

**降级策略：**

| 模式 | 实现 |
|------|------|
| junction | `mklink /J` |
| symlink | `mklink /D`（需 SeCreateSymbolicLinkPrivilege） |
| copy | `robocopy /MIR` 或 `io.Copy`（慢，仅兼容模式） |

---

## 6. 配置系统（Viper）

### 6.1 初始化流程

```text
cmd/root.go init()
    │
    ├─ paths.ResolveHome() → PVM_HOME
    ├─ config.Load(home)
    │     ├─ SetDefault(...)
    │     ├─ ReadInConfig() // 可选，文件不存在不报错
    │     └─ AutomaticEnv(PVM_)
    └─ slog setup from config.logging.level
```

### 6.2 配置与 Flag 绑定

| Flag | Viper Key | 说明 |
|------|-----------|------|
| `--json` | — | 仅运行时，不入配置文件 |
| `--verbose` | — | 临时提升 log level |
| `--config` | — | 覆盖 config 文件路径 |

Persistent flags 在 root 定义，子命令通过 `internal/config.Get()` 读取。

---

## 7. CLI 层设计

### 7.1 命令与 App 服务映射

| 命令 | App 服务 | 主要 pkg |
|------|----------|----------|
| `install` | Installer.Install | php, download, win |
| `uninstall` | Installer.Uninstall | paths |
| `use` | Switcher.Use | win, alias |
| `list` | Registry.List | php |
| `list-remote` | RemoteIndex.List | php, download |
| `current` | Switcher.Current | paths |
| `which` | Switcher.Which | alias |
| `info` | Registry.Info | php |
| `alias` | AliasStore | paths |
| `default` | AliasStore.Set("default") | — |
| `setup` | EnvManager.Setup | win |
| `path` | EnvManager.Status | win |

### 7.2 全局 Flags（root.go）

```go
rootCmd.PersistentFlags().Bool("json", false, "JSON output")
rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Quiet mode")
rootCmd.PersistentFlags().CountP("verbose", "v", "Verbose logging")
rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output")
rootCmd.PersistentFlags().BoolP("yes", "y", false, "Skip confirmations")
rootCmd.PersistentFlags().Bool("dry-run", false, "Simulate operations")
```

### 7.3 错误处理与 Exit Code

```go
// cmd/helpers.go

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        var exitErr *app.ExitError
        if errors.As(err, &exitErr) {
            os.Exit(exitErr.Code)
        }
        ui.PrintError(err)
        os.Exit(2)
    }
}
```

```go
// internal/app/errors.go

type ExitError struct {
    Code    int
    Message string
    Hint    string
}

func (e *ExitError) Error() string
```

---

## 8. 文件系统约定

### 8.1 目录 Ensure 策略

启动任何写操作前调用 `paths.EnsureLayout(home)`：

```go
func EnsureLayout(home string) error {
    dirs := []string{
        home,
        VersionsDir(home),
        CacheDir(home),
        AliasesDir(home),
        LogsDir(home),
    }
    // MkdirAll 0755
}
```

### 8.2 临时目录命名

| 用途 | 模式 |
|------|------|
| 安装解压 | `versions/{id}.tmp.{pid}` |
| 下载中 | `cache/{id}.zip.part` |
| 链接切换 | `current.new` |

失败时 defer 清理 `.tmp.*` 与 `.part`。

---

## 9. 日志

使用 Go 1.21+ `log/slog`：

```go
// internal/log/log.go

func Init(level string, logFile string) (*slog.Logger, error)
```

- CLI 默认 text handler → stderr（verbose 时）
- 文件 handler → `%PVM_HOME%\logs\pvm.log`（JSON lines）
- 不在日志中记录 proxy 密码

---

## 10. 测试策略

### 10.1 单元测试（无 Windows 依赖）

| 包 | 测试重点 |
|----|----------|
| `pkg/php` | Spec 解析、ID 生成、ZipName、VC 映射 |
| `pkg/download` | checksum 解析与 verify |
| `internal/app` | mock RemoteIndex/Downloader 测 Install 状态机 |

### 10.2 Windows 集成测试

```go
//go:build windows

// test/integration/install_test.go
```

使用本地 fixture zip（`test/fixtures/php-8.3.0-test.zip` 最小结构）测 install → use → current。

### 10.3 Mock 约定

```go
// internal/app/mocks/remote_index.go

type MockRemoteIndex struct {
    Releases []php.RemoteRelease
}
```

---

## 11. 安全设计

| 威胁 | 对策 |
|------|------|
| 中间人篡改 zip | 强制 SHA256；HTTPS only |
| PATH 劫持 | setup 仅追加 `%PVM_HOME%\current`，doctor 检测顺序 |
| 符号链接攻击 | 安装前 canonicalize 路径；拒绝 `..` 跳转 |
| 凭证泄露 | proxy URL 脱敏日志 |
| 误删激活版本 | uninstall 前检查 current 指向 |

---

## 12. MVP 实现范围（Phase 1）

### 12.1 本阶段交付

| 组件 | 完成度 |
|------|--------|
| `pkg/paths` | 100% |
| `pkg/php/version, spec` | 100% |
| `pkg/php/remote` | 80%（HTML 解析） |
| `pkg/download` | 80% |
| `pkg/win/junction, env` | 70%（用户 PATH） |
| `internal/config` | 100% |
| `internal/app/installer, switcher, alias, registry` | 80% |
| `cmd/*` MVP 命令 | 骨架 + install/use/list 主流程 |

### 12.2 本阶段不实现

- ext 管理
- config ini 编辑
- .pvmrc / auto
- doctor fix
- self-update
- 断点续传

---

## 13. 后续扩展点

| 版本 | 扩展方式 |
|------|----------|
| V1.0 config | 新增 `pkg/php/ini.go` + `cmd/config/*` |
| V1.0 ext | 新增 `pkg/php/extension.go` + PECL 索引 |
| V1.1 mirror | `RemoteIndex` 实现 mirror 切换，Viper 配置 |
| V2.0 snapshot | 新增 `internal/app/snapshot.go`，zip 打包 versions/{id} |

---

## 14. 文档修订记录

| 版本 | 日期 | 变更说明 |
|------|------|----------|
| v1.0.0 | 2026-05-21 | 初始版本，MVP 模块设计与接口契约 |

---

*实现细节以代码为准；接口变更须同步更新本文档与 [PRD.md](./PRD.md)。*
