# 活动图片合并子技能

用于从 `origin` 合并或 cherry-pick 大活动、小活动的新主题图片，并适配本仓库命名与引用。不要只复制图片或只修改 task override；必须完成引用闭环。

## 目录

- [1. 审计来源与本地状态](#1-审计来源与本地状态)
- [2. 按本仓库规范重命名图片](#2-按本仓库规范重命名图片)
- [3. 建立完整引用矩阵](#3-建立完整引用矩阵)
- [4. 防漏验证](#4-防漏验证)

## 1. 审计来源与本地状态

1. 运行 `git status --short`，区分用户已有改动与本次改动，禁止覆盖无关修改。
2. 获取 origin 后，用以下命令检查来源提交的完整范围：

    ```powershell
    git show --stat --summary --find-renames <commit>
    git show --name-status --find-renames <commit>
    git diff <commit>^ <commit> -- assets
    ```

3. 列出来源提交新增或替换的所有图片、task、Pipeline、preset 和 locale 文件。不要根据提交标题推断范围。
4. 用户要求 cherry-pick 时保留来源提交的有效逻辑变化；解决冲突后再按本仓库规范重命名和适配，禁止直接照搬来源仓库的节点名、文件名或 task 结构。

## 2. 按本仓库规范重命名图片

1. 先查看相邻活动主题目录，沿用当前目录与语义文件名约定。
2. 使用主题目录表达活动名，文件名只表达用途，避免重复主题前缀。例如：

    | origin 名称                             | 本仓库名称                                          |
    | --------------------------------------- | --------------------------------------------------- |
    | `ProjectMatisLogo.png`                  | `SmallEvent\ProjectMatis\Logo.png`                  |
    | `ProjectMatisStageNormal.png`           | `SmallEvent\ProjectMatis\StageNormal.png`           |
    | `ProjectMatisStageNormalRepeatable.png` | `SmallEvent\ProjectMatis\StageNormalRepeatable.png` |

3. 对已跟踪文件使用 `git mv`，使 Git 保留重命名关系。
4. 保留来源图片原始二进制内容；除非用户要求，不要重新压缩、缩放或转码。

## 3. 建立完整引用矩阵

先用新旧主题名、目录名和每个旧文件名全仓搜索：

```powershell
rg -n "OldTheme|OldLogoName|OldStageName|NewTheme" assets
```

逐项检查并修改实际存在的消费者：

- `assets\resource\pipeline\Event\<EventType>\`：入口、剧情、挑战、商店等 Pipeline。
- `assets\tasks\<EventType>.json`：主题 case、`default_case` 与所有 `pipeline_override`。
- `assets\tasks\preset\*.json`：显式选择活动主题的预设。
- `assets\locales\interface\*.json`：当前主题或新增 case 的标签。
- 其他由 `rg` 找到的测试、配置或文档引用。

### SmallEvent 强制检查

合并小活动主题时至少核对以下位置：

1. `SmallEvent.json` 中 `SmallEventEntry` 的默认 `Logo.png` 引用。
2. `SmallEventStory.json` 中普通关、可扫荡普通关，以及来源提交实际提供的困难关图片引用。
3. `assets\tasks\SmallEvent.json` 中当前主题、默认 case 和对应节点 override。
4. 使用 SmallEvent 主题选项的所有 preset 与中英文 locale。

Pipeline 中标注“占位，应该去 task 页面修改”的默认模板也必须更新为当前默认主题。它们用于未注入 override 的加载、调试与测试，不能保留旧活动图片。本次 PROJECT MATIS 的入口必须同时满足：

```jsonc
// assets\resource\pipeline\Event\SmallEvent\SmallEvent.json
"template": ["SmallEvent/ProjectMatis/Logo.png"]
```

```jsonc
// assets\tasks\SmallEvent.json
"template": ["SmallEvent/ProjectMatis/Logo.png"]
```

同理，`StageNormal.png` 和 `StageNormalRepeatable.png` 必须同时覆盖基础 Pipeline 与 task override。

## 4. 防漏验证

1. 搜索来源文件名、旧主题路径和被替换主题名，确认不存在意外残留：

    ```powershell
    rg -n "ProjectMatisLogo|ProjectMatisStage|SmallEvent/OldTheme" assets
    ```

2. 搜索新主题的全部引用，逐个确认目标图片存在：

    ```powershell
    rg -n "SmallEvent/ProjectMatis|LargeEvent/<Theme>" assets
    Get-ChildItem -Recurse "assets\resource\image\SmallEvent\ProjectMatis"
    ```

3. 检查最终差异与重命名：

    ```powershell
    git diff --check
    git diff --stat
    git diff --name-status --find-renames
    ```

4. 对修改的 JSON/JSONC 执行格式检查，再运行仓库校验：

    ```powershell
    npx.cmd prettier --check <modified-json-files>
    npx.cmd @nekosu/maa-tools check
    python tools\validate_schema.py --resource-dirs assets\resource --exclude-dirs assets\resource\announcement --interface-files assets\interface.json
    ```

5. 提交前根据来源提交的文件清单逐项回查：每张图片已导入并按规范命名；每个新名称已被正确引用；入口 Logo、关卡图、task override、基础 Pipeline、preset 和 locale 均未遗漏。
