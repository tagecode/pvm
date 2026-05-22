# PVM (PHP Version Manager for Windows) 命令行工具功能清单

下面是为 Windows 平台设计的 PHP 版本管理工具 `pvm` 的完整功能清单。我按照"核心功能 → 进阶功能 → 生态集成 → 运维诊断"的顺序组织，便于你按优先级排期开发。

---

## 一、核心版本管理（MVP 必备）

### 1.1 列表与查询

| 命令 | 功能说明 |
|------|---------|
| `pvm list-remote` / `pvm ls-remote` | 从 `downloads.php.net/~windows/releases/` 或者 `https://downloads.php.net/~windows/releases/archives/` 拉取所有可下载的 PHP 版本，支持过滤：`--major 8`、`--minor 8.3` |
| `pvm list` / `pvm ls` | 列出本地已安装的所有 PHP 版本，并标注：当前激活版本(*)、默认版本(default)、架构(x64/x86)、TS/NTS、VC 版本 |
| `pvm current` | 显示当前激活的 PHP 版本及其完整路径 |
| `pvm which <version>` | 输出指定版本的安装目录路径，便于脚本集成 |
| `pvm info <version>` | 显示某个版本的详细信息：发布日期、下载 URL、SHA256、内置扩展列表、php.ini 路径 |

### 1.2 安装与卸载

| 命令 | 功能说明 |
|------|---------|
| `pvm install <version>` | 下载并解压指定 PHP 版本到 `%PVM_HOME%\versions\<version>\` |
| `pvm install <version> --arch x64\|x86` | 指定架构，默认跟随当前系统 |
| `pvm install <version> --ts\|--nts` | 指定 Thread Safe 或 Non-Thread Safe，默认 NTS（用于 CLI/FPM 场景） |
| `pvm install <version> --vc vs16\|vs17` | 指定 VC 编译版本（PHP 8.4+ 用 vs17，PHP 8.0–8.3 用 vs16，PHP 7.4 用 vc15） |
| `pvm install latest` | 安装最新版 |
| `pvm install 8.3` | 仅指定主.次版本号时，自动取该系列最新补丁号 |
| `pvm install --from-zip <path>` | 从已下载的本地 zip 文件安装（离线场景） |
| `pvm install --reinstall <version>` | 强制重新下载安装 |
| （安装行为） | 解压后若无 `php.ini`，从 `php.ini-development`（优先）或 `php.ini-production` 生成；缺模板则安装失败；默认 `auto_use_after_install = true` 装完自动 `use` |
| `pvm uninstall <version>` / `pvm remove <version>` | 卸载某个版本（含确认提示，禁止删除当前激活版本） |
| `pvm uninstall --all-but <version>` | 清理除指定版本外的所有版本 |

### 1.3 版本切换

| 命令 | 功能说明 |
|------|---------|
| `pvm use <version>` | 切换到指定版本（修改 `%PVM_HOME%\current` 软链接或 junction，无需重启终端） |
| `pvm use default` | 切回默认版本 |
| `pvm use system` | 临时回退到系统原生 PHP（如果存在） |
| `pvm default <version>` / `pvm alias default <version>` | 设置开机默认版本 |
| `pvm unuse` | 取消当前会话的 PHP 关联（从 PATH 中临时移除） |

### 1.4 别名管理

| 命令 | 功能说明 |
|------|---------|
| `pvm alias <name> <version>` | 创建版本别名，如 `pvm alias prod 8.2.20` |
| `pvm alias` | 列出所有别名 |
| `pvm unalias <name>` | 删除别名 |
| `pvm alias default <version>` | 设置默认版本别名（特殊保留别名） |

---

## 二、Windows 平台专属功能

### 2.1 环境变量与 PATH 管理

| 命令 | 功能说明 |
|------|---------|
| `pvm setup` | 手动向用户级 PATH 注入 `%PVM_HOME%\current` 并设置 `PHP_HOME`（**首次 `use` 会自动执行**，通常无需单独运行） |
| `pvm setup --system` | 写入系统级环境变量（需管理员权限，自动检测并提示 UAC） |
| `pvm unsetup` | 清理 pvm 写入的 PATH 条目与 `PHP_HOME` |
| `pvm path` | 显示当前 PHP 相关环境变量状态（PATH、PHP_HOME、PHPRC） |
| `pvm refresh` | 输出可在其它已打开终端刷新 PATH 的命令（当前终端由 `use` / `install` 自动刷新） |

### 2.2 软链接 / Junction 策略

| 命令 | 功能说明 |
|------|---------|
| `pvm link-mode junction\|symlink\|copy` | 选择版本切换方式：junction（默认，无需管理员）、symlink（需开发者模式/管理员）、copy（兼容性最好但慢） |
| `pvm doctor link` | 检测当前链接方式是否健康 |

---

## 三、配置文件（php.ini）管理

| 命令 | 功能说明 |
|------|---------|
| `pvm config edit [version]` | 用默认编辑器打开 php.ini |
| `pvm config get <key> [version]` | 读取某个配置项的值，如 `pvm config get memory_limit` |
| `pvm config set <key> <value> [version]` | 设置配置项（自动备份原文件） |
| `pvm config enable <extension>` | 启用扩展（取消对应行的分号注释） |
| `pvm config disable <extension>` | 禁用扩展 |
| `pvm config diff <v1> <v2>` | 对比两个版本的 ini 配置差异 |
| `pvm config copy <from> <to>` | 把某个版本的 ini 复制到另一个版本 |
| `pvm config reset [version]` | 用 `php.ini-development` 或 `php.ini-production` 重置 |
| `pvm config template dev\|prod` | 一键应用开发或生产环境模板 |

---

## 四、扩展管理（PECL / 原生）

> Windows 下 PECL 扩展通常需要预编译 DLL，需对接 `pecl.php.net` 或 `windows.php.net/downloads/pecl/`。

| 命令 | 功能说明 |
|------|---------|
| `pvm ext list` | 列出当前版本已加载的扩展（解析 `php -m`） |
| `pvm ext list-available` | 列出当前 PHP 版本兼容的可安装扩展 |
| `pvm ext search <keyword>` | 搜索扩展 |
| `pvm ext install <name>` | 自动下载匹配 PHP 版本/架构/TS-NTS/VC 的 DLL 到 `ext/` 目录，并写入 `php.ini` |
| `pvm ext install <name>@<version>` | 安装指定版本扩展 |
| `pvm ext uninstall <name>` | 删除扩展 DLL 并从 ini 移除 |
| `pvm ext enable\|disable <name>` | 启用 / 禁用扩展 |
| `pvm ext info <name>` | 显示扩展信息（版本、依赖、配置项） |
| `pvm ext sync <from-version> <to-version>` | 把一个版本的扩展配置复制到另一版本（自动匹配兼容 DLL） |

常见扩展白名单内置匹配规则：`xdebug`、`redis`、`mongodb`、`imagick`、`apcu`、`swoole`、`yaml`、`ldap`、`sqlsrv`、`mysqli`、`opcache`、`intl` 等。

---

## 五、项目级版本（类似 `.nvmrc`）

| 命令 | 功能说明 |
|------|---------|
| `pvm local <version>` | 在当前目录生成 `.pvmrc` 文件，记录该项目需要的 PHP 版本 |
| `pvm auto on\|off` | 开启 / 关闭目录切换自动识别（注入 PowerShell `Set-Location` Hook 或 cmd `prompt` Hook） |
| `pvm use auto` / `pvm use` | 不带参数时自动读取 `.pvmrc`（向上递归查找） |
| `pvm exec <version> -- <command>` | 用指定版本临时执行命令（不切换全局），如 `pvm exec 7.4 -- composer install` |
| `pvm run <command>` | 在 `.pvmrc` 指定的版本下执行命令 |

`.pvmrc` 文件格式示例：
```text
8.2.20
# 或者 JSON
{
  "version": "8.2.20",
  "arch": "x64",
  "ts": false,
  "extensions": ["redis", "xdebug"]
}
```

---

## 六、生态集成

### 6.1 Composer 集成

| 命令 | 功能说明 |
|------|---------|
| `pvm composer install` | 为当前版本下载并安装 Composer（每个 PHP 版本独立一个 composer.phar，避免全局污染） |
| `pvm composer install --global` | 安装全局 Composer，并绑定到当前版本 |
| `pvm composer use <version>` | 切换默认 Composer 绑定的 PHP 版本 |
| `pvm composer self-update` | 升级 Composer |

### 6.2 镜像与代理

| 命令 | 功能说明 |
|------|---------|
| `pvm mirror set <url>` | 设置下载镜像（国内可设为 `https://www.php.net/distributions/` 或自建镜像） |
| `pvm mirror list` | 列出预置镜像：official / huawei / aliyun / 自定义 |
| `pvm mirror test` | 测试各镜像的连通性与速度，自动选择最快的 |
| `pvm proxy http://...` | 设置 HTTP/HTTPS 代理 |
| `pvm proxy off` | 关闭代理 |

### 6.3 校验与安全

| 命令 | 功能说明 |
|------|---------|
| 自动校验 SHA256 | 每次下载后自动比对官方 sha256 文件 |
| 自动校验 GPG 签名（可选） | 调用本地 `gpg.exe` 验证 `.asc` 签名 |
| `pvm verify <version>` | 重新校验某个已安装版本的完整性 |

---

## 七、运维诊断与维护

| 命令 | 功能说明 |
|------|---------|
| `pvm doctor` | 全面健康检查：环境变量是否正确、当前版本可执行性、ext 目录是否完整、php.ini 路径是否一致、PATH 顺序是否合理 |
| `pvm doctor fix` | 自动修复常见问题 |
| `pvm cache list` | 列出下载缓存（zip 包） |
| `pvm cache clean` | 清理下载缓存 |
| `pvm cache dir` | 显示缓存目录 |
| `pvm prune` | 清理无效的版本目录、损坏的 junction |
| `pvm logs` | 查看 pvm 自身的操作日志（位于 `%PVM_HOME%\logs\pvm.log`） |
| `pvm env` | 输出当前所有 pvm 相关环境变量、配置文件路径，便于排错 |
| `pvm self-update` | 升级 pvm 自身 |
| `pvm self-uninstall` | 完全卸载 pvm（含环境变量回滚） |

---

## 八、高级功能（差异化亮点）

| 命令 | 功能说明 |
|------|---------|
| `pvm export <version> --to <path>` | 把某个版本打包导出（含 ini + 扩展），便于分发到其他机器 |
| `pvm import <package.zip>` | 导入打包好的版本 |
| `pvm snapshot create <name>` | 为当前版本（含 ini 和扩展）创建快照 |
| `pvm snapshot restore <name>` | 还原快照 |
| `pvm snapshot list\|delete` | 管理快照 |
| `pvm matrix run <cmd>` | 在多个版本下并行执行命令，输出矩阵结果（用于跨版本兼容测试） |
| `pvm bench <version1> <version2>` | 在多个版本下跑基准测试（micro_time 对比） |
| `pvm completion powershell\|bash\|cmd` | 输出 shell 自动补全脚本 |
| `pvm shell pwsh\|cmd` | 启动一个内嵌的子 shell，仅在该子 shell 内激活指定 PHP 版本（隔离影响） |

---

## 九、全局选项与人性化设计

| 选项 | 说明 |
|------|------|
| `--json` | 所有 list/info/current 类命令支持 JSON 输出，便于脚本集成 |
| `--quiet` / `-q` | 静默模式 |
| `--verbose` / `-v` | 详细日志 |
| `--no-color` | 关闭彩色输出 |
| `--yes` / `-y` | 跳过所有确认 |
| `--dry-run` | 模拟执行，仅打印将要发生的操作 |
| `pvm <cmd> --help` | 每个子命令独立帮助 |
| `pvm help` | 顶层帮助 + 常见用例 |

---

## 十、目录结构建议

```text
%USERPROFILE%\.pvm\           （即 %PVM_HOME%）
├── pvm.exe                   # 主程序
├── config.toml               # 全局配置（镜像、代理、默认架构等）
├── current\                  # 指向当前版本的 junction（要被加入 PATH）
├── versions\
│   ├── 8.3.10-x64-nts\
│   ├── 8.2.20-x64-nts\
│   └── 7.4.33-x64-nts\
├── aliases\
│   ├── default
│   └── prod
├── cache\                    # 下载的 zip 缓存
├── snapshots\
├── logs\
│   └── pvm.log
└── composer\                 # 各版本独立的 composer.phar
```

---

## 十一、版本号与系列优先级（决策树）

设计安装时，对 PHP for Windows 的多重组合需要明确默认值：

| 维度 | 选项 | 推荐默认 |
|------|------|---------|
| 架构 | x64 / x86 | 跟随系统（64 位机器默认 x64） |
| 线程安全 | TS / NTS | **NTS**（PHP-FPM、CLI 场景的标配；TS 仅 Apache mod_php 需要） |
| VC 编译器 | vc15 / vs16 / vs17 | 根据 PHP 主版本自动匹配 |
| 来源 | 正式发布 / archives 历史 | 优先正式区，找不到再回退到 `releases/archives/` |

---

## 十二、开发优先级建议（路线图）

| 阶段 | 包含功能 | 目标 |
|------|---------|------|
| **MVP（2–3 周）** | 一、二.1 一.4 | 能装、能切、能列、能用，覆盖 90% nvm 用户的基本需求 |
| **V1.0** | 三、四、五 | 配置 + 扩展 + .pvmrc，达到生产可用 |
| **V1.1** | 六、七 | 镜像、Composer、诊断工具，提升体验和稳定性 |
| **V2.0** | 八、十 | 快照、矩阵测试、自动补全，差异化竞争 |

---

## 十三、技术栈建议（仅供参考）

- **Rust** 或 **Go**：单一二进制、跨架构编译、启动快、无运行时依赖（推荐 Rust，因为 Windows API 操作 junction、注册表更熟练；Go 的开发效率也很高）
- **不要用 PHP/Node 写 pvm 本身**：会引入循环依赖
- **CLI 框架**：Rust 用 `clap`，Go 用 `cobra` + `viper`
- **HTTP 客户端**：Rust `reqwest`，Go `net/http`
- **Junction 创建**：Rust `mklink` 调用 + `winapi::CreateSymbolicLinkW`；Go `os.Symlink` 或 `mklink /J`
- **压缩**：直接用 `zip` crate / `archive/zip`，避免依赖外部 `7z`