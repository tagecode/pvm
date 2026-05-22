# PVM 产品需求文档（PRD）

| 属性 | 内容 |
|------|------|
| **产品名称** | PVM — PHP Version Manager for Windows |
| **文档版本** | v1.0.0 |
| **文档状态** | Draft |
| **创建日期** | 2026-05-21 |
| **最后更新** | 2026-05-21（php.ini 安装初始化） |
| **关联文档** | [FEATURE.md](./FEATURE.md) |
| **技术栈** | Go 1.22+ · Cobra · Viper · 单一可执行文件 |

---

## 1. 文档目的与范围

### 1.1 目的

本文档定义 PVM 的产品目标、功能需求、非功能需求、技术架构与验收标准，作为设计、开发、测试与发布的唯一需求基线。所有实现决策须可回溯至本文档条目。

### 1.2 范围

| 在范围内 | 不在范围内 |
|----------|------------|
| Windows 10/11 上的 PHP 多版本安装、切换与管理 | Linux / macOS 支持（后续版本另行规划） |
| 命令行工具 `pvm.exe` 及 Shell 集成 | GUI 图形界面 |
| php.ini 与 Windows PECL 扩展管理 | 自行编译 PHP 源码 |
| 项目级 `.pvmrc` 版本锁定 | Web 管理控制台 |
| Composer 绑定、镜像、诊断与运维工具 | PHP 运行时本身的功能缺陷修复 |

### 1.3 术语表

| 术语 | 定义 |
|------|------|
| **PVM_HOME** | PVM 数据根目录，默认 `%USERPROFILE%\.pvm` |
| **激活版本** | 当前通过 `current` 链接指向、处于 PATH 中的 PHP 版本 |
| **默认版本** | 新开终端时自动使用的版本，存储于 `aliases/default` |
| **NTS / TS** | Non-Thread Safe / Thread Safe，Windows PHP 构建变体 |
| **VC / VS** | Visual C++ 运行时版本（vc15 / vs16 / vs17） |
| **Junction** | Windows 目录联接，无需管理员权限即可实现目录"软链接" |
| **.pvmrc** | 项目级 PHP 版本声明文件，类似 `.nvmrc` |
| **pvm 架构** | `pvm.exe` 可执行文件本身的 CPU 架构（amd64 → x64，386 → x86） |
| **PHP 架构** | 待安装/切换的 PHP zip 构建架构（`x64` / `x86`），写入版本 ID 与目录名 |
| **系统 PHP** | PVM 安装前已存在于系统/用户 PATH 中的原生 PHP，不由 `%PVM_HOME%\versions` 管理 |

---

## 2. 背景与问题陈述

### 2.1 背景

Windows 平台缺乏成熟、统一的 PHP 版本管理工具。开发者通常需要：

- 手动从 [windows.php.net](https://windows.php.net/) 下载 zip 包并解压；
- 手工维护 PATH、PHPRC、扩展 DLL 与 php.ini；
- 在不同项目间频繁切换 PHP 版本时容易出错；
- 缺少类似 Unix 生态中 `nvm`、`pyenv` 的一致体验。

### 2.2 核心痛点

1. **安装复杂**：需理解 x64/x86、TS/NTS、VC 版本等多维组合。
2. **切换成本高**：修改环境变量后常需重启终端甚至系统。
3. **配置分散**：php.ini、扩展 DLL、Composer 绑定各自独立，难以追溯。
4. **离线/内网困难**：缺少镜像、缓存与校验机制。
5. **项目协作不一致**：团队成员 PHP 版本无法通过版本控制文件统一。

### 2.3 产品定位

PVM 是面向 Windows 开发者的 **PHP 版本管理 CLI 工具**，提供「下载 → 安装 → 切换 → 配置 → 诊断」全链路能力，以单一二进制、零运行时依赖的方式交付，目标成为 Windows PHP 开发的事实标准工具。

---

## 3. 产品目标

### 3.1 业务目标

| 目标 | 衡量指标 |
|------|----------|
| 降低 PHP 环境搭建时间 | 从零到可用 PHP CLI ≤ 3 分钟（含下载） |
| 提升版本切换效率 | `pvm use <version>` 生效无需重启终端 |
| 提高配置可维护性 | 支持项目级 `.pvmrc` 与 ini  diff/模板 |
| 保障安装安全性 | 100% 下载包 SHA256 校验 |

### 3.2 用户目标

- 一条命令安装任意 PHP 版本；
- 一条命令切换全局或项目级 PHP 版本；
- 可视化的健康检查与自动修复；
- 脚本友好的 JSON 输出与稳定退出码。

### 3.3 非目标（Out of Scope）

- 不提供 PHP 源码编译器或自定义 build 流水线；
- 不替代 IIS / Apache / Nginx 的 Web Server 配置管理；
- 不内置 PHP IDE 或调试器（可与 Xdebug 扩展配合）；
- v1.x 不支持 WSL 内部安装（WSL 用户建议使用 Linux 版工具）。

---

## 4. 目标用户与典型场景

### 4.1 用户画像

| 角色 | 描述 | 核心诉求 |
|------|------|----------|
| **独立开发者** | 本地维护多个 PHP 项目 | 快速切换版本、管理扩展 |
| **团队工程师** | 企业内网或 CI 环境 | `.pvmrc` 锁定、镜像、JSON 输出 |
| **DevOps / 运维** | 批量部署与诊断 | doctor、export/import、日志 |
| **PHP 初学者** | 不熟悉 Windows PHP 构建变体 | 智能默认、友好错误提示 |

### 4.2 典型用户故事

| ID | 角色 | 故事 | 验收条件 |
|----|------|------|----------|
| US-01 | 开发者 | 作为开发者，我希望一条命令安装 PHP 8.3 最新补丁版，以便快速开始项目 | `pvm install 8.3` 自动解析最新补丁号并完成安装 |
| US-02 | 开发者 | 作为开发者，我希望切换 PHP 版本后当前终端立即生效 | `pvm use 8.2.20` 后 `php -v` 输出正确版本 |
| US-03 | 团队工程师 | 作为工程师，我希望项目根目录有 `.pvmrc`，进入目录自动切换版本 | `pvm auto on` 后 cd 到项目目录自动 use |
| US-04 | 运维 | 作为运维，我希望诊断工具能检测并修复 PATH 问题 | `pvm doctor fix` 修复可识别的问题 |
| US-05 | CI 脚本 | 作为 CI 脚本，我希望 list/info 支持 JSON 输出 | `pvm list --json` 输出合法 JSON |

---

## 5. 功能需求

功能按优先级分为 **P0（MVP）**、**P1（V1.0）**、**P2（V1.1）**、**P3（V2.0）**。编号格式：`FR-<模块>-<序号>`。

---

### 5.1 核心版本管理（P0 — MVP）

#### 5.1.1 列表与查询

| 编号 | 命令 | 需求描述 | 优先级 |
|------|------|----------|--------|
| FR-CORE-001 | `pvm list-remote` / `pvm ls-remote` | 从 `downloads.php.net/~windows/releases/` 及 `archives/` 拉取可下载版本列表；支持 `--major`、`--minor` 过滤 | P0 |
| FR-CORE-002 | `pvm list` / `pvm ls` | 列出本地已安装版本，标注：当前激活(*)、默认(default)、架构、TS/NTS、VC 版本 | P0 |
| FR-CORE-003 | `pvm current` | 显示当前激活 PHP 版本及完整路径 | P0 |
| FR-CORE-004 | `pvm which <version>` | 输出指定版本安装目录，供脚本集成 | P0 |
| FR-CORE-005 | `pvm info <version>` | 显示版本详情：发布日期、下载 URL、SHA256、内置扩展、php.ini 路径 | P0 |

**行为细则：**

- `list-remote` 须合并正式区与 archives 区结果并去重；
- 网络不可达时给出明确错误与重试建议；
- 所有 list/info/current 类命令须支持 `--json`（见 §5.9）。

#### 5.1.2 安装与卸载

| 编号 | 命令 | 需求描述 | 优先级 |
|------|------|----------|--------|
| FR-CORE-010 | `pvm install <version>` | 下载、解压至 `%PVM_HOME%\versions\<version>\`，并按 §5.1.2 初始化 `php.ini` | P0 |
| FR-CORE-011 | `pvm install --arch x64\|x86` | 指定 PHP 架构；未指定时按 §5.1.2「架构解析」规则推断（默认与 `pvm.exe` 架构一致） | P0 |
| FR-CORE-012 | `pvm install --ts\|--nts` | 指定 TS/NTS，**默认 NTS** | P0 |
| FR-CORE-013 | `pvm install --vc vs16\|vs17` | 指定 VC 版本；未指定时按 PHP 主版本自动匹配（见 §7.6） | P0 |
| FR-CORE-014 | `pvm install latest` | 安装最新正式版 | P0 |
| FR-CORE-015 | `pvm install 8.3` | 主.次版本号自动取该系列最新补丁号 | P0 |
| FR-CORE-016 | `pvm install --from-zip <path>` | 从本地 zip 离线安装 | P1 |
| FR-CORE-017 | `pvm install --reinstall` | 强制重新下载安装 | P1 |
| FR-CORE-018 | `pvm uninstall <version>` | 卸载指定版本；须确认提示；**禁止删除当前激活版本** | P0 |
| FR-CORE-019 | `pvm uninstall --all-but <version>` | 保留指定版本，清理其余 | P1 |

**安装目录命名规范：**

```text
{major}.{minor}.{patch}-{arch}-{ts|nts}-{vc}
示例：8.3.10-x64-nts-vs16
```

**架构分发与解析（arch）：**

1. **分发约定（PVM 自身）**  
   - Windows **x64** 用户须安装 **amd64** 构建的 `pvm.exe`（发布页标注 `Windows x64`）。  
   - Windows **x86** 用户须安装 **386** 构建的 `pvm.exe`（发布页标注 `Windows x86`）。  
   - 禁止在 x64 系统上默认推荐 x86 版 `pvm.exe`；安装文档须明确二者不可混用。

2. **PHP 包默认架构**  
   - 用户未指定 `--arch` 且 `defaults.arch` 为 `auto`（或未配置）时，默认 PHP 架构 **与当前运行的 `pvm.exe` 架构一致**（实现：`runtime.GOARCH`；在用户装对 pvm 的前提下等价于系统架构）。  
   - 在 x64 系统上若需 32 位 PHP（兼容/测试），须显式：`pvm install <version> --arch x86`。

3. **解析优先级（高 → 低）**  
   1. 命令行 `--arch x64|x86`  
   2. `config.toml` → `defaults.arch`（`auto` \| `x64` \| `x86`）  
   3. 当前 `pvm.exe` 架构（`auto` 的最终回退）

4. **校验与诊断（P1）**  
   - `pvm doctor` 检测：`pvm.exe` 为 x86 但宿主系统为 x64 时输出**警告**（可能误装 32 位 pvm），并提示安装 x64 版或确认是否需要 x86 PHP。

**行为细则：**

- 安装前检查磁盘空间（至少 zip 大小的 2 倍）；
- 安装后自动执行 SHA256 校验（P1 可配置跳过，默认开启）；
- **php.ini 初始化（解压后、写入 `versions/` 前）：**
  - 若版本目录下**已存在** `php.ini`（如 `--reinstall` 保留的旧配置），**不覆盖**；
  - 若不存在 `php.ini`，优先将 `php.ini-development` **复制为** `php.ini`；
  - 若不存在 `php.ini-development` 但存在 `php.ini-production`，则将 `php.ini-production` **复制为** `php.ini`；
  - 若 `php.ini`、`php.ini-development`、`php.ini-production` **均不存在**，**安装失败**并报错；
- 安装成功后默认自动 `use`（配置项 `defaults.auto_use_after_install`，默认 **true**）；设为 `false` 则仅安装不切换；
- 若目标目录已存在且非 `--reinstall`，报错并提示已有版本路径。

#### 5.1.3 版本切换

| 编号 | 命令 | 需求描述 | 优先级 |
|------|------|----------|--------|
| FR-CORE-020 | `pvm use <version>` | 切换激活版本，更新 `%PVM_HOME%\current` 链接，**无需重启终端** | P0 |
| FR-CORE-021 | `pvm use default` | 切回默认版本 | P0 |
| FR-CORE-022 | `pvm use system` | 当前终端临时使用系统 PATH 中的原生 PHP；**不修改** `current` junction（见 §5.1.3 共存） | P1 |
| FR-CORE-023 | `pvm default <version>` | 设置默认版本 | P0 |
| FR-CORE-024 | `pvm unuse` | 当前终端从 PATH **最前** 临时移除 `%PVM_HOME%\current`；**不修改** `current`（见 §5.1.3 共存） | P1 |

**行为细则：**

- `<version>` 支持别名解析（见 §5.1.4）；
- 切换失败时保持原 `current` 不变（原子操作）；
- 切换成功后输出简短确认：`Now using PHP x.y.z (x64-nts-vs16)`；若本次为首次环境配置，附加 `(environment configured)`；
- **首次 `use` 自动 setup：** 若用户 PATH 中尚无 `%PVM_HOME%\current`（或未置顶），自动写入用户 PATH 与 `PHP_HOME`（等同 `pvm setup`）；
- **`use` 后会话刷新：** 当前终端立即更新 `Path` / `PHP_HOME`，无需再执行 `pvm refresh` 即可 `php -v`；

**与系统 PHP 共存：**

机器在安装 PVM 前若已通过安装包、XAMPP、Scoop、Chocolatey 等方式配置过 PHP，PVM **不得删除或覆盖** 既有 PHP 的 PATH 条目；仅管理 `%PVM_HOME%\current` 及 PVM 写入的环境变量。

| 场景 | 命令 / 机制 | 是否改 `current` | 作用范围 | 之后如何回到 PVM |
|------|-------------|------------------|----------|------------------|
| PVM 版本互切 | `pvm use 8.3` → `pvm use 8.2` → `pvm use 8.3` | ✅ 每次更新 junction | 持久；**首次 `use` 自动 setup** | 任意 `pvm use <version>` |
| 切回默认 PVM 版 | `pvm use default` | ✅ | 持久 | — |
| 临时用系统 PHP | `pvm use system` | ❌ 保持原 junction | **仅当前终端** | 同终端 `pvm refresh` 或新开终端；或 `pvm use <version>` |
| 临时不用 PVM | `pvm unuse` | ❌ | **仅当前终端** | 同终端 `pvm refresh`；或 `pvm use <version>` |
| 长期停用 PVM | `pvm unsetup` | ❌（junction 仍在，但 PATH 无 PVM） | 持久（用户环境变量） | 再次 `pvm setup` + `pvm use <version>` |

**`pvm use system` / `pvm unuse` 细则（P1）：**

- 二者均为**会话级**操作，不写入注册表，不改变 `%PVM_HOME%\current` 指向；
- `use system`：在当前进程 PATH 中，将 `%PVM_HOME%\current` **后置**（或临时移除），使 PATH 中下一顺位可见的系统 `php.exe` 生效；若 PATH 中无其它 PHP，报错并提示 `pvm path`；
- `unuse`：从当前进程 PATH **最前** 移除 `%PVM_HOME%\current` 一项（与 `refresh` 前置策略相反）；
- 执行 `pvm use <version>` 后，除更新 `current` 外，应在本终端将 `%PVM_HOME%\current` 重新置于 PATH 最前（与 `pvm refresh` 一致）；
- `PHP_HOME` / `PHPRC`：`use system` / `unuse` 时在本终端清空或恢复为切换前快照；`pvm use <version>` 时将 `PHP_HOME` 设为 `%PVM_HOME%\current`（与 `setup` 一致）。

**MVP 说明：** `pvm use <version>`（P0）已覆盖 PVM 版本互切，并含首次自动 setup 与会话 PATH 刷新；`use system` / `unuse` 为 P1。MSI / ZIP 安装包预置 `%PVM_HOME%\current` 与 `PHP_HOME`；默认 `auto_use_after_install = true`，通常 `pvm install 8.3` 后即可在本终端 `php -v`。

#### 5.1.4 别名管理

| 编号 | 命令 | 需求描述 | 优先级 |
|------|------|----------|--------|
| FR-CORE-030 | `pvm alias <name> <version>` | 创建版本别名 | P0 |
| FR-CORE-031 | `pvm alias` | 列出所有别名 | P0 |
| FR-CORE-032 | `pvm unalias <name>` | 删除别名（`default` 为保留名，不可删除只可覆盖） | P0 |
| FR-CORE-033 | `pvm alias default <version>` | 设置默认版本（等同 `pvm default`） | P0 |

---

### 5.2 Windows 平台专属（P0 — MVP）

#### 5.2.1 环境变量与 PATH

| 编号 | 命令 | 需求描述 | 优先级 |
|------|------|----------|--------|
| FR-WIN-001 | `pvm setup` | 向用户级 PATH **前置** `%PVM_HOME%\current`；设置 `PHP_HOME` 指向 current；**不删除** 既有 PHP 路径；**首次 `use` 时自动执行**（可手动重复） | P0 |
| FR-WIN-002 | `pvm setup --system` | 写入系统级环境变量；需管理员权限，自动检测并提示 UAC | P1 |
| FR-WIN-003 | `pvm unsetup` | 仅移除 PVM 写入的 PATH 条目与 `PHP_HOME`；**不删除** 系统中原生 PHP 路径 | P1 |
| FR-WIN-004 | `pvm path` | 显示 PATH、PHP_HOME、PHPRC 当前状态；标注 PVM 条目位置及是否可能被系统 PHP 抢先 | P0 |
| FR-WIN-005 | `pvm refresh` | 输出可在**当前会话**将 `%PVM_HOME%\current` 置于 PATH 最前的命令（PowerShell / cmd） | P1 |

**PATH 与环境变量细则：**

1. **前置而非追加：** `setup` 须将 `%PVM_HOME%\current` 放在**用户级 PATH 的最前面**（新建终端生效），确保在已存在系统 PHP 时默认 `php` 指向 PVM 激活版本。  
2. **不破坏既有配置：** 禁止在 `setup` / `unsetup` 中删除、修改非 PVM 写入的 PATH 段；系统级 / 用户级其它 PHP、Composer 等路径保持原样。  
3. **查找顺序：** Windows 合并系统 PATH + 用户 PATH；`php` 解析以**先匹配者为准**。`pvm path` / `pvm doctor` 须提示：若系统 PATH 中另有 PHP 且排在 PVM 之前，可能出现「已 use 但 `php -v` 仍为系统版」——建议 `pvm setup` 或 `pvm refresh`。  
4. **会话刷新：** `use` 成功后当前进程自动前置 `current` 并设置 `PHP_HOME`；`refresh` 用于其它已打开终端复制命令。  
5. **安装包：** MSI / ZIP 的 `install.ps1` 预置 `%USERPROFILE%\.pvm\current` 到用户 PATH 及 `PHP_HOME`。  
6. **与 §5.1.3 的关系：** 持久切换靠 `current` junction + 用户 PATH 前置；临时回退系统 PHP 靠 `use system` / `unuse`（P1）或 `unsetup`（长期）。

#### 5.2.2 链接策略

> **MVP 实现范围：** 链接模式**仅需实现 `junction`**。`symlink`、`copy` 暂不实现，待整体稳定后再补；下表中已划掉项为延后项。

| 编号 | 命令 | 需求描述 | 优先级 |
|------|------|----------|--------|
| FR-WIN-010 | `pvm link-mode junction` | 配置版本切换方式（MVP 仅 `junction`；~~`symlink`~~、~~`copy`~~ 延后） | P0 |
| FR-WIN-011 | `pvm doctor link` | 检测 current 链接健康状态 | P1 |

**链接模式说明：**

| 模式 | 权限要求 | 性能 | 默认 | MVP |
|------|----------|------|------|-----|
| junction | 普通用户 | 快 | ✓ | **实现** |
| ~~symlink~~ | ~~开发者模式或管理员~~ | ~~快~~ | | ~~延后~~ |
| ~~copy~~ | ~~无~~ | ~~慢，占磁盘~~ | | ~~延后~~ |

---

### 5.3 配置文件管理（P1 — V1.0）

| 编号 | 命令 | 需求描述 | 优先级 |
|------|------|----------|--------|
| FR-CFG-001 | `pvm config edit [version]` | 用 `$EDITOR` 或系统默认编辑器打开 php.ini | P1 |
| FR-CFG-002 | `pvm config get <key> [version]` | 读取 ini 配置项 | P1 |
| FR-CFG-003 | `pvm config set <key> <value> [version]` | 设置配置项，**自动备份原文件** | P1 |
| FR-CFG-004 | `pvm config enable <extension>` | 取消扩展行注释以启用 | P1 |
| FR-CFG-005 | `pvm config disable <extension>` | 注释扩展行以禁用 | P1 |
| FR-CFG-006 | `pvm config diff <v1> <v2>` | 对比两版本 ini 差异 | P1 |
| FR-CFG-007 | `pvm config copy <from> <to>` | 复制 ini 配置 | P1 |
| FR-CFG-008 | `pvm config reset [version]` | 用 php.ini-development 或 php.ini-production 重置 | P1 |
| FR-CFG-009 | `pvm config template dev\|prod` | 一键应用开发/生产模板 | P1 |

**行为细则：**

- 每次 `set` / `enable` / `disable` 前备份至 `%PVM_HOME%\versions\<version>\ini-backups\`；
- 未指定 `[version]` 时操作当前激活版本；
- ini 解析须处理 Windows 换行符（CRLF）及 `;` 注释。

---

### 5.4 扩展管理（P1 — V1.0）

| 编号 | 命令 | 需求描述 | 优先级 |
|------|------|----------|--------|
| FR-EXT-001 | `pvm ext list` | 列出当前版本已加载扩展（解析 `php -m`） | P1 |
| FR-EXT-002 | `pvm ext list-available` | 列出与当前 PHP 版本兼容的可安装扩展 | P1 |
| FR-EXT-003 | `pvm ext search <keyword>` | 搜索扩展 | P1 |
| FR-EXT-004 | `pvm ext install <name>` | 下载匹配 arch/TS-NTS/VC 的 DLL 至 `ext/`，写入 php.ini | P1 |
| FR-EXT-005 | `pvm ext install <name>@<version>` | 安装指定版本扩展 | P2 |
| FR-EXT-006 | `pvm ext uninstall <name>` | 删除 DLL 并从 ini 移除 | P1 |
| FR-EXT-007 | `pvm ext enable\|disable <name>` | 启用/禁用扩展 | P1 |
| FR-EXT-008 | `pvm ext info <name>` | 显示扩展详情 | P1 |
| FR-EXT-009 | `pvm ext sync <from> <to>` | 跨版本同步扩展配置（自动匹配兼容 DLL） | P2 |

**数据源：**

- PECL：`pecl.php.net` / `windows.php.net/downloads/pecl/`

**内置扩展白名单匹配规则：**

`xdebug`, `redis`, `mongodb`, `imagick`, `apcu`, `swoole`, `yaml`, `ldap`, `sqlsrv`, `mysqli`, `opcache`, `intl` 等。

---

### 5.5 项目级版本（P1 — V1.0）

| 编号 | 命令 | 需求描述 | 优先级 |
|------|------|----------|--------|
| FR-PROJ-001 | `pvm local <version>` | 在当前目录生成 `.pvmrc` | P1 |
| FR-PROJ-002 | `pvm auto on\|off` | 开启/关闭目录切换自动识别 | P1 |
| FR-PROJ-003 | `pvm use`（无参数） | 向上递归查找 `.pvmrc` 并切换 | P1 |
| FR-PROJ-004 | `pvm exec <version> -- <command>` | 临时用指定版本执行命令，不改变全局 | P1 |
| FR-PROJ-005 | `pvm run <command>` | 在 `.pvmrc` 指定版本下执行命令 | P1 |

**.pvmrc 格式规范：**

**格式 A — 纯文本（推荐简单场景）：**

```text
8.2.20
```

**格式 B — JSON（推荐复杂场景）：**

```json
{
  "version": "8.2.20",
  "arch": "x64",
  "ts": false,
  "extensions": ["redis", "xdebug"]
}
```

**行为细则：**

- 向上递归查找直至文件系统根或 `$PVM_HOME`；
- `auto on` 注入 PowerShell `Set-Location` Hook 或 cmd prompt Hook；
- `.pvmrc` 中版本未安装时，提示 `pvm install <version>` 而非静默失败。

---

### 5.6 生态集成（P2 — V1.1）

#### 5.6.1 Composer

| 编号 | 命令 | 需求描述 | 优先级 |
|------|------|----------|--------|
| FR-ECO-001 | `pvm composer install` | 为当前 PHP 版本安装独立 composer.phar | P2 |
| FR-ECO-002 | `pvm composer install --global` | 安装全局 Composer 并绑定当前版本 | P2 |
| FR-ECO-003 | `pvm composer use <version>` | 切换 Composer 绑定的 PHP 版本 | P2 |
| FR-ECO-004 | `pvm composer self-update` | 升级 Composer | P2 |

#### 5.6.2 镜像与代理

| 编号 | 命令 | 需求描述 | 优先级 |
|------|------|----------|--------|
| FR-ECO-010 | `pvm mirror set <url>` | 设置下载镜像 | P2 |
| FR-ECO-011 | `pvm mirror list` | 列出预置镜像：official / huawei / aliyun / custom | P2 |
| FR-ECO-012 | `pvm mirror test` | 测试镜像连通性与速度，推荐最快 | P2 |
| FR-ECO-013 | `pvm proxy <url>` | 设置 HTTP/HTTPS 代理 | P2 |
| FR-ECO-014 | `pvm proxy off` | 关闭代理 | P2 |

#### 5.6.3 校验与安全

| 编号 | 需求描述 | 优先级 |
|------|----------|--------|
| FR-ECO-020 | 每次下载后自动 SHA256 校验 | P0（install）/ P2（ext） |
| FR-ECO-021 | 可选 GPG 签名校验（调用本地 `gpg.exe`） | P3 |
| FR-ECO-022 | `pvm verify <version>` 重新校验已安装版本完整性 | P2 |

---

### 5.7 运维诊断与维护（P2 — V1.1）

| 编号 | 命令 | 需求描述 | 优先级 |
|------|------|----------|--------|
| FR-OPS-001 | `pvm doctor` | 全面健康检查（PATH、可执行性、ext、ini、链接） | P2 |
| FR-OPS-002 | `pvm doctor fix` | 自动修复可识别问题 | P2 |
| FR-OPS-003 | `pvm cache list` | 列出下载缓存 | P2 |
| FR-OPS-004 | `pvm cache clean` | 清理下载缓存 | P2 |
| FR-OPS-005 | `pvm cache dir` | 显示缓存目录 | P2 |
| FR-OPS-006 | `pvm prune` | 清理无效版本目录、损坏 junction | P2 |
| FR-OPS-007 | `pvm logs` | 查看 `%PVM_HOME%\logs\pvm.log` | P2 |
| FR-OPS-008 | `pvm env` | 输出所有 PVM 相关环境变量与配置路径 | P2 |
| FR-OPS-009 | `pvm self-update` | 升级 PVM 自身 | P2 |
| FR-OPS-010 | `pvm self-uninstall` | 完全卸载 PVM（含环境变量回滚） | P2 |

**doctor 检查项清单：**

1. `%PVM_HOME%` 目录存在且可写
2. `current` 链接有效且指向已安装版本
3. PATH 中包含 `%PVM_HOME%\current` 且顺序合理
4. `php.exe` 可执行且 `--version` 正常
5. `PHP_HOME` / `PHPRC` 一致性
6. 已安装版本目录完整性（php.exe、ext/ 存在）
7. **pvm.exe 与系统位数**（P1）：x86 pvm 运行在 x64 系统时警告
8. **PATH 与系统 PHP 冲突**（P1）：用户 PATH 中 `%PVM_HOME%\current` 未置顶，或系统 PATH 中存在其它 `php.exe` 且可能抢先时警告

---

### 5.8 高级功能（P3 — V2.0）

| 编号 | 命令 | 需求描述 | 优先级 |
|------|------|----------|--------|
| FR-ADV-001 | `pvm export <version> --to <path>` | 打包导出版本（含 ini + 扩展） | P3 |
| FR-ADV-002 | `pvm import <package.zip>` | 导入打包版本 | P3 |
| FR-ADV-003 | `pvm snapshot create\|restore\|list\|delete` | 版本快照管理 | P3 |
| FR-ADV-004 | `pvm matrix run <cmd>` | 多版本并行执行命令（兼容测试矩阵） | P3 |
| FR-ADV-005 | `pvm bench <v1> <v2>` | 跨版本 micro-benchmark 对比 | P3 |
| FR-ADV-006 | `pvm completion powershell\|bash\|cmd` | Shell 自动补全脚本 | P3 |
| FR-ADV-007 | `pvm shell pwsh\|cmd` | 启动隔离子 Shell，仅在其中激活指定版本 | P3 |

---

### 5.9 全局选项与人性化设计（全阶段）

| 编号 | 选项 | 需求描述 | 优先级 |
|------|------|----------|--------|
| FR-UX-001 | `--json` | list/info/current 等命令 JSON 输出 | P0 |
| FR-UX-002 | `--quiet` / `-q` | 静默模式，仅输出结果 | P0 |
| FR-UX-003 | `--verbose` / `-v` | 详细日志（可 `-vv` 更详细） | P0 |
| FR-UX-004 | `--no-color` | 关闭 ANSI 彩色输出 | P0 |
| FR-UX-005 | `--yes` / `-y` | 跳过所有确认提示 | P1 |
| FR-UX-006 | `--dry-run` | 模拟执行，打印计划操作 | P1 |
| FR-UX-007 | `pvm <cmd> --help` | 每个子命令独立帮助 | P0 |
| FR-UX-008 | `pvm help` | 顶层帮助 + 常见用例 | P0 |

**输出规范：**

- 成功：stdout 输出结果，exit code `0`
- 用户错误（如版本不存在）：stderr 友好提示 + 建议命令，exit code `1`
- 系统错误（如网络/权限）：stderr 详细错误 + 日志路径，exit code `2`
- JSON 模式：stdout 仅输出 JSON，错误信息放入 JSON `error` 字段

---

## 6. 非功能需求

### 6.1 性能

| 编号 | 需求 | 指标 |
|------|------|------|
| NFR-PERF-001 | CLI 冷启动 | `pvm --version` ≤ 100ms（SSD，i5+） |
| NFR-PERF-002 | 本地 list/current | ≤ 50ms |
| NFR-PERF-003 | 版本切换 | `pvm use` ≤ 200ms（junction 模式） |
| NFR-PERF-004 | 并发下载 | 支持 configurable 并发数，默认 4 线程 |

### 6.2 可靠性

| 编号 | 需求 |
|------|------|
| NFR-REL-001 | 安装/切换操作具备原子性，失败时不留半成品目录 |
| NFR-REL-002 | 所有写操作先写临时目录再 rename |
| NFR-REL-003 | 网络中断支持断点续传（P2） |
| NFR-REL-004 | 配置文件损坏时自动回退至默认配置并告警 |

### 6.3 兼容性

| 编号 | 需求 |
|------|------|
| NFR-COMPAT-001 | 支持 Windows 10 1903+ 及 Windows 11 |
| NFR-COMPAT-002 | 支持 PHP 7.4.x — 最新正式版（随官方 releases 扩展） |
| NFR-COMPAT-003 | 支持 x64 与 x86 PHP 架构；**分别提供** amd64 / 386 两个 `pvm.exe` 发布工件，与目标 Windows 位数对应 |
| NFR-COMPAT-004 | 支持 PowerShell 5.1+、PowerShell 7+、cmd.exe |

### 6.4 安全

| 编号 | 需求 |
|------|------|
| NFR-SEC-001 | 所有远程下载必须 SHA256 校验 |
| NFR-SEC-002 | 不在日志中记录代理凭证 |
| NFR-SEC-003 | 系统级 setup 须显式 `--system` 并 UAC 确认 |
| NFR-SEC-004 | 不将 `%PVM_HOME%` 写入需管理员权限的系统目录（默认用户目录） |

### 6.5 可维护性

| 编号 | 需求 |
|------|------|
| NFR-MAINT-001 | 结构化日志（text + 可选 JSON 行格式） |
| NFR-MAINT-002 | 语义化版本号（SemVer） |
| NFR-MAINT-003 | 单元测试覆盖率 ≥ 70%（核心 pkg） |
| NFR-MAINT-004 | 集成测试覆盖 install/use/uninstall 主流程 |

### 6.6 可访问性 / 国际化

| 编号 | 需求 |
|------|------|
| NFR-I18N-001 | v1.x 仅英文 CLI 输出 |
| NFR-I18N-002 | 错误消息包含可操作的下一步建议 |
| NFR-I18N-003 | 中文本地化列入 v2.x 路线图（非 v1 阻塞项） |

---

## 7. 技术架构

### 7.1 技术栈选型

| 层次 | 选型 | 理由 |
|------|------|------|
| 语言 | **Go 1.22+** | 单一二进制、交叉编译、标准库丰富、Windows 支持成熟 |
| CLI 框架 | **Cobra** | 子命令树、自动 help、Shell 补全生态 |
| 配置管理 | **Viper** | 多源配置（文件 + 环境变量 + flags）、TOML/YAML |
| HTTP | `net/http` + 可选 `github.com/hashicorp/go-retryablehttp` | 下载、镜像、PECL 元数据 |
| 压缩 | `archive/zip` | 解压 PHP zip，无外部 7z 依赖 |
| Windows API | `golang.org/x/sys/windows` + `os/exec`（mklink） | Junction/Symlink、注册表、环境变量 |
| 日志 | `log/slog` 或 `github.com/charmbracelet/log` | 结构化、级别控制 |
| 测试 | `testing` + `github.com/stretchr/testify` | 单元/集成测试 |

**明确禁止：**

- 不使用 PHP / Node.js 实现 PVM 本体（避免循环依赖）；
- 不依赖外部 7-Zip、Cygwin 等工具作为运行时前提。

### 7.2 系统架构

```text
┌─────────────────────────────────────────────────────────────┐
│                        用户 / Shell                          │
│              (PowerShell / cmd / CI Script)                  │
└─────────────────────────┬───────────────────────────────────┘
                          │ pvm.exe
┌─────────────────────────▼───────────────────────────────────┐
│                     CLI Layer (cmd/)                         │
│         Cobra Commands · Global Flags · Help/Completion      │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                   Application Layer (internal/app/)          │
│    Install · Use · Config · Ext · Doctor · Project · ...    │
└─────────────────────────┬───────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        ▼                 ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  pkg/php     │  │  pkg/win     │  │  pkg/http    │
│  版本解析     │  │  Junction    │  │  下载/镜像    │
│  ini 管理    │  │  PATH/Reg    │  │  SHA256      │
│  扩展匹配    │  │  UAC 检测    │  │  代理         │
└──────────────┘  └──────────────┘  └──────────────┘
        │                 │                 │
        └─────────────────┼─────────────────┘
                          ▼
              ┌───────────────────────┐
              │   %PVM_HOME% 文件系统   │
              │  versions/ current/   │
              │  cache/ config.toml   │
              └───────────────────────┘
                          │
                          ▼
              ┌───────────────────────┐
              │   外部数据源            │
              │  windows.php.net      │
              │  pecl.php.net         │
              └───────────────────────┘
```

### 7.3 项目目录结构（代码仓库）

```text
pvm/
├── main.go
├── go.mod
├── cmd/                          # Cobra 命令定义
│   ├── root.go
│   ├── install.go
│   ├── use.go
│   ├── list.go
│   └── ...
├── internal/
│   ├── app/                      # 业务编排
│   │   ├── installer.go
│   │   ├── switcher.go
│   │   └── ...
│   └── config/                   # Viper 配置加载
│       └── config.go
├── pkg/
│   ├── php/                      # PHP 版本/ini/扩展
│   │   ├── version.go
│   │   ├── remote.go
│   │   └── ini.go
│   ├── win/                      # Windows 专属
│   │   ├── junction.go
│   │   ├── env.go
│   │   └── registry.go
│   ├── download/                 # HTTP 下载与校验
│   │   ├── client.go
│   │   └── checksum.go
│   └── ui/                       # 输出格式化
│       ├── table.go
│       └── json.go
├── docs/
│   ├── FEATURE.md
│   └── PRD.md
└── test/
    └── integration/
```

### 7.4 数据目录结构（运行时）

```text
%USERPROFILE%\.pvm\                 # PVM_HOME
├── pvm.exe                         # 主程序（amd64 或 386，须与 Windows 位数匹配；见 §7.5.1）
├── config.toml                     # 全局配置
├── current\                        # → versions/8.3.10-x64-nts-vs16 (junction)
├── versions\
│   ├── 8.3.10-x64-nts-vs16\
│   │   ├── php.exe
│   │   ├── ext\
│   │   ├── php.ini
│   │   └── ini-backups\
│   └── 8.2.20-x64-nts-vs16\
├── aliases\
│   ├── default                     # 内容：版本 ID 字符串
│   └── prod
├── cache\                          # 下载 zip 缓存
├── snapshots\
├── logs\
│   └── pvm.log
└── composer\
    └── 8.3.10-x64-nts-vs16\
        └── composer.phar
```

### 7.5 配置管理（Viper）

#### 7.5.1 架构（arch）分发与解析

**原则：** 先按**系统架构**分发正确的 `pvm.exe`，再由 PVM 默认下载**同架构**的 PHP zip；避免依赖「猜测操作系统」的 Win32 API 作为 MVP 默认路径。

| 阶段 | 规则 |
|------|------|
| 获取 PVM | x64 Windows → `pvm-windows-amd64.exe`（或统一名 `pvm.exe`）；x86 Windows → `pvm-windows-386.exe` |
| 安装 PHP（默认） | `pvm install 8.3` → 选择与 `pvm.exe` 同架构的 zip（如 amd64 pvm → `…-x64-…zip`） |
| 覆盖默认 | `pvm install 8.3 --arch x86` 或 `defaults.arch = "x86"` |

**实现要点（`pkg/php`）：**

- `DefaultArch()`：`runtime.GOARCH == "386"` → `x86`，否则 `x64`。  
- `Resolve()`：`opts.Arch` 为空或 `auto` 时，先读 `defaults.arch`（Viper），仍为 `auto` 则调用 `DefaultArch()`。  
- 发布流水线须产出 **两个** Windows 工件，安装说明中写清对应关系。

**配置文件路径：** `%PVM_HOME%\config.toml`

**配置优先级（高 → 低）：**

1. 命令行 flags
2. 环境变量（前缀 `PVM_`）
3. `%PVM_HOME%\config.toml`
4. 内置默认值

**config.toml 示例：**

```toml
[defaults]
arch = "auto"          # auto：跟 pvm.exe；或显式 x64 | x86（见 §7.5.1 解析优先级）
thread_safe = false    # false = NTS
link_mode = "junction" # MVP 仅 junction；延后：~~symlink~~ | ~~copy~~
auto_use_after_install = true

[mirror]
url = "https://windows.php.net/downloads/releases/"
# preset = "official"  # official | huawei | aliyun

[proxy]
enabled = false
http = ""
https = ""

[download]
cache_dir = ""         # 空则使用 %PVM_HOME%\cache
concurrency = 4
retry = 3

[security]
verify_sha256 = true
verify_gpg = false

[logging]
level = "info"         # debug | info | warn | error
file = ""              # 空则 %PVM_HOME%\logs\pvm.log

[project]
auto_switch = false    # 等同 pvm auto on
```

**环境变量映射：**

| 环境变量 | 配置项 |
|----------|--------|
| `PVM_HOME` | 数据根目录（默认 `%USERPROFILE%\.pvm`） |
| `PVM_MIRROR_URL` | `mirror.url` |
| `PVM_PROXY_HTTP` | `proxy.http` |
| `PVM_LOG_LEVEL` | `logging.level` |
| `PVM_DEFAULTS_ARCH` | `defaults.arch` |

### 7.6 PHP 版本解析决策树

安装时若用户未指定完整参数，按以下规则自动推断：

```text
输入 version 字符串
    │
    ├─ "latest" → 取 list-remote 最新正式版
    ├─ "8.3"    → 取 8.3.x 最新补丁
    ├─ "8.3.10" → 精确版本
    └─ alias    → 查 aliases/ 映射
         │
         ▼
    架构 arch（见 §7.5.1）：
         ├─ 已指定 --arch → 使用该值
         ├─ 否则读 defaults.arch
         │    ├─ x64 / x86 → 使用该值
         │    └─ auto 或未配置 → 跟 pvm.exe（GOARCH：386→x86，其余→x64）
         │
         ▼
    TS/NTS：未指定 → NTS（默认）
         │
         ▼
    VC 版本：未指定 → 按 PHP 主版本映射
         ├─ 7.4.x  → vc15
         ├─ 8.0–8.3 → vs16
         └─ 8.4+   → vs17
         │
         ▼
    下载源：优先 releases/ → 失败则 archives/
         │
         ▼
    校验 SHA256 → 解压
         │
         ▼
    php.ini 初始化：无 php.ini 时复制 php.ini-development（无则 php.ini-production）→ php.ini
         │（三者均不存在 → 安装失败）
         ▼
    写入 versions/
```

### 7.7 Cobra 命令树

```text
pvm
├── install
├── uninstall | remove
├── use
├── default
├── unuse
├── list | ls
├── list-remote | ls-remote
├── current
├── which
├── info
├── alias
│   └── unalias
├── setup
├── unsetup
├── path
├── refresh
├── link-mode
├── config
│   ├── edit
│   ├── get
│   ├── set
│   ├── enable
│   ├── disable
│   ├── diff
│   ├── copy
│   ├── reset
│   └── template
├── ext
│   ├── list
│   ├── list-available
│   ├── search
│   ├── install
│   ├── uninstall
│   ├── enable
│   ├── disable
│   ├── info
│   └── sync
├── local
├── auto
├── exec
├── run
├── composer
│   ├── install
│   ├── use
│   └── self-update
├── mirror
│   ├── set
│   ├── list
│   └── test
├── proxy
├── verify
├── doctor
│   ├── link
│   └── fix
├── cache
│   ├── list
│   ├── clean
│   └── dir
├── prune
├── logs
├── env
├── export
├── import
├── snapshot
├── matrix
├── bench
├── completion
├── shell
├── self-update
├── self-uninstall
├── help
└── version
```

---

## 8. 接口与集成规范

### 8.1 外部数据源

| 数据源 | URL | 用途 |
|--------|-----|------|
| PHP Windows 正式版 | `https://windows.php.net/downloads/releases/` | 版本列表、zip 下载 |
| PHP Windows 归档 | `https://windows.php.net/downloads/releases/archives/` | 历史版本 |
| SHA256 校验文件 | 同目录下 `sha256sum.txt` 或等效 | 完整性校验 |
| PECL Windows | `https://windows.php.net/downloads/pecl/` | 预编译扩展 DLL |

### 8.2 JSON 输出 schema（示例）

**`pvm list --json`：**

```json
{
  "versions": [
    {
      "id": "8.3.10-x64-nts-vs16",
      "version": "8.3.10",
      "arch": "x64",
      "thread_safe": false,
      "vc": "vs16",
      "path": "C:\\Users\\dev\\.pvm\\versions\\8.3.10-x64-nts-vs16",
      "active": true,
      "default": false
    }
  ]
}
```

**`pvm current --json`：**

```json
{
  "id": "8.3.10-x64-nts-vs16",
  "version": "8.3.10",
  "path": "C:\\Users\\dev\\.pvm\\current",
  "php_exe": "C:\\Users\\dev\\.pvm\\current\\php.exe"
}
```

### 8.3 退出码规范

| Code | 含义 |
|------|------|
| 0 | 成功 |
| 1 | 用户输入错误 / 业务逻辑拒绝 |
| 2 | 系统错误（IO、网络、权限） |
| 3 | 部分成功（如 batch 操作） |
| 130 | 用户中断（Ctrl+C） |

---

## 9. 发布路线图

| 里程碑 | 周期 | 功能范围 | 交付标准 |
|--------|------|----------|----------|
| **MVP** | 2–3 周 | §5.1 + §5.2.1 + §5.1.4 + 全局选项 | 能 install/list/use/setup；发布 **amd64 + 386** 两个 `pvm.exe`（§7.5.1）；覆盖 90% 基础场景 |
| **V1.0** | +3–4 周 | §5.3 + §5.4 + §5.5 | 生产可用：ini、扩展、.pvmrc |
| **V1.1** | +2–3 周 | §5.6 + §5.7 | 镜像、Composer、doctor、self-update |
| **V2.0** | +4 周 | §5.8 + completion + shell | 快照、矩阵测试、差异化功能 |

---

## 10. 验收标准

### 10.1 MVP 验收清单

- [x] `pvm install 8.3` 成功安装并在 `pvm list` 中可见（mock HTTP 集成测试；本机在线下载受网络影响）
- [x] `pvm use 8.3.x` 切换后，`php -v` 版本正确（离线 zip + acceptance 测试）
- [x] `pvm setup` 正确写入用户 PATH 与 PHP_HOME
- [x] `pvm list-remote --major 8` 返回过滤结果
- [x] `pvm alias prod 8.2.20` + `pvm use prod` 生效
- [x] `pvm current --json` 输出合法 JSON
- [x] 安装包 SHA256 校验失败时拒绝安装并报错
- [x] 尝试 uninstall 当前激活版本时被拒绝
- [x] `--help` 对所有 MVP 命令可用
- [ ] Windows 10/11 x64 双环境测试通过（使用 amd64 版 `pvm.exe`）
- [x] `defaults.arch=auto` 时安装的 PHP 与 `pvm.exe` 架构一致
- [x] `pvm install 8.3 --arch x86` 在 x64 系统可安装 x86 PHP（显式覆盖）
- [x] 已存在系统 PHP 时：`pvm use A` → `pvm use B` → `pvm use A` 均可切回；`php -v` 与激活版本一致
- [x] `pvm setup` 前置 `current` 且不删除既有 PHP PATH 条目；`pvm unsetup` 后系统 PHP 仍可通过原 PATH 使用
- [x] 安装后版本目录存在 `php.ini`（由 `php.ini-development` 或 `php.ini-production` 生成）
- [x] zip 内无 `php.ini` / `php.ini-development` / `php.ini-production` 时安装失败

### 10.2 V1.0 验收清单

- [ ] `pvm config set memory_limit 256M` 生效且产生备份
- [ ] `pvm ext install xdebug` 下载 DLL 并写入 ini
- [ ] `pvm local 8.2.20` 生成 `.pvmrc`，`pvm use` 自动识别
- [ ] `pvm exec 7.4 -- php -v` 不改变全局版本

### 10.3 质量门禁

- 所有 P0 功能有单元测试
- `go test ./...` 通过
- `golangci-lint run` 无 error 级问题
- 集成测试在 CI（GitHub Actions windows-latest）通过

---

## 11. 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 官方 PHP Windows 下载页结构变更 | 高 | 中 | 抽象 remote 解析层；集成测试监控；快速 hotfix |
| PECL DLL 版本矩阵复杂 | 中 | 高 | 内置白名单 + 明确错误；v1 仅支持常用扩展 |
| Windows PATH 刷新需重启终端 | 低 | 低 | `use` / `install` 自动刷新当前会话；`refresh` 供其它终端 |
| Junction 权限/杀毒误报 | 中 | 低 | doctor 检测；~~copy 降级模式~~（整体稳定后再实现） |
| 国内网络下载慢 | 中 | 高 | 镜像预设 + cache + 断点续传 |
| Go os.Symlink 在 Windows 行为差异 | 低 | 中 | 优先 mklink /J；充分测试 |
| 用户误装 x86 版 pvm 于 x64 系统 | 中 | 中 | 发布页区分 amd64/386 工件；`pvm doctor` 警告；文档说明 `--arch` 覆盖 |
| 系统 PHP 与 PVM 争抢 PATH | 中 | 高 | `setup` 前置 `current`；`pvm path` / doctor 提示；`use system` / `unuse`（P1） |

---

## 12. 测试策略

### 12.1 单元测试

- `pkg/php/version`：版本字符串解析、VC 映射
- `pkg/php/ini`：读写、enable/disable、备份
- `pkg/win/junction`：创建/删除/检测（Windows CI）
- `pkg/download/checksum`：SHA256 校验逻辑

### 12.2 集成测试

- 使用 mock HTTP server 模拟 PHP 下载
- 完整流程：install → use → php -v → uninstall
- config.toml 加载与 override

### 12.3 手工测试矩阵

| 环境 | 测试项 |
|------|--------|
| Win10 x64 + PowerShell 5.1 | MVP 全流程 |
| Win11 x64 + PowerShell 7 | MVP + auto hook |
| Win10 x86 | 386 版 pvm + x86 PHP 全流程 |
| Win11 x64 + 误装 x86 pvm | doctor 警告；`--arch x64` 仍可装 x64 PHP |
| Win10 x64 + 预装系统 PHP（PATH） | setup 后默认走 PVM；unsetup 后恢复系统 `php -v` |
| Win10 x64 + 预装系统 PHP | `use 8.3` → `use 8.2` → `use 8.3` 互切正确（§5.1.3） |

---

## 13. 附录

### 13.1 竞品参考

| 工具 | 平台 | 借鉴点 |
|------|------|--------|
| nvm-windows | Node.js | 安装目录结构、switch 机制 |
| pyenv-win | Python | 版本命名、shim 思路 |
| phpenv | Unix | 版本切换 UX |
| Scoop/Chocolatey | Windows | 包管理体验（PVM 更专注 PHP 多版本） |

### 13.2 参考链接

- [PHP for Windows Downloads](https://windows.php.net/download/)
- [PECL Windows Downloads](https://windows.php.net/downloads/pecl/)
- [Cobra Documentation](https://github.com/spf13/cobra)
- [Viper Documentation](https://github.com/spf13/viper)

### 13.3 文档修订记录

| 版本 | 日期 | 作者 | 变更说明 |
|------|------|------|----------|
| v1.0.0 | 2026-05-21 | — | 初始版本，基于 FEATURE.md 与技术栈 Go+Cobra+Viper |

---

*本文档为 PVM 项目的需求基线。功能变更须更新本文档版本号并同步 [FEATURE.md](./FEATURE.md)。*
