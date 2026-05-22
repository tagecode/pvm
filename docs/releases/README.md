# Release Notes

每个版本在发版前新增对应 Markdown 文件，GitHub Release 会自动读取。

## 命名规则

```
docs/releases/v{版本号}.md
```

示例：

| Tag | 文件 |
|-----|------|
| `v0.1.0` | `docs/releases/v0.1.0.md` |
| `v0.2.0` | `docs/releases/v0.2.0.md` |

## 发版流程

1. 编写 `docs/releases/vX.Y.Z.md`
2. 更新根目录 `CHANGELOG.md`
3. 提交并 push
4. 打 tag 并推送：

   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```

5. `release` workflow 会构建产物，并将本文件内容作为 GitHub Release 说明

若缺少对应文件，Release 步骤会失败——发版前务必创建。

## 模板

复制 `v0.1.0.md` 结构，替换版本号、功能列表与已知限制即可。
