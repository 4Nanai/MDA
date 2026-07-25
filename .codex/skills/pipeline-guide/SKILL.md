---
name: pipeline-guide
description: "Universal Pipeline orchestration 编写指南。基于 MaaFramework Pipeline 协议，覆盖 Pipeline JSON、Go Agent Custom Recognition/Action、节点命名、识别算法、动作类型、流程控制与可复用节点。在编写、修改或审查 Pipeline JSON，设计 Go Agent + JSON 流程，使用 TemplateMatch/OCR/Custom 识别及 Click/Swipe/Custom 动作，或从 origin 合并、重命名并接入活动图片资源时使用。"
---

# Universal Pipeline Orchestration 编写指南

## 实现选型

- 优先采用 **Go Agent + Pipeline JSON**：JSON 负责声明页面状态、识别、动作和流程连接；Go 负责纯 JSON 难以清晰表达的计算、组合识别、运行时数据处理与复杂操作。
- 简单、有限且可枚举的页面状态仍直接使用 JSON 的 `next`、`on_error`、`[JumpBack]`，不要为了包装单次识别或点击而新增 Custom。
- 不使用 Python 编排。需要扩展时使用 `maa-framework-go/v4` 的 Custom Recognition 或 Custom Action，并由 JSON 节点调用。
- 编写前先读取 [repository_reference.md](repository_reference.md)，核对仓库现有通用交互、状态和 Go Custom；不要凭记忆复制节点。
- 从 origin 合并、同步或 cherry-pick 大小活动图片时，必须读取并完整执行[活动图片合并子技能](references/event-image-merge.md)，同时更新图片命名、任务覆盖、Pipeline 默认引用、预设与本地化。

## 核心原则

1. **状态驱动**：遵循"识别 → 操作 → 识别"循环。每次操作必须基于识别结果，禁止假设操作后画面状态。
2. **高命中率**：扩充 `next` 列表，覆盖当前操作后所有可能画面，力争一次截图命中。
3. **避免硬延迟**：尽量不用 `pre_delay` / `post_delay` / `timeout`，优先通过增加中间识别节点解决；只在必须等画面稳定时才用 `pre_wait_freezes` / `post_wait_freezes`。当确实不需要延迟时，要在节点上显式将 `rate_limit` / `pre_delay` / `post_delay` 设为 0（协议默认 `rate_limit=1000ms`、`pre_delay/post_delay=200ms`，省略字段会引入隐式等待；仓库的 `tools/add_node_defaults.py` 会为 Common 节点补齐这些 0 值字段）。
4. **720p 基准**：所有坐标、ROI、图片必须基于 **1280\*720**。
5. **格式化**：JSON 遵循 `.prettierrc`（4 空格缩进，数组元素换行）。
6. **UTF-8 文本**：Pipeline JSON 与相关文本文件使用 UTF-8（可无 BOM）。读取时使用 PowerShell 7；若只能使用 Windows PowerShell 5.1，必须显式指定 `Get-Content -Encoding UTF8`，不要依赖本地代码页。出现中文乱码时先用显式 UTF-8 或严格解析器验证原始文件，不能据终端显示直接判定文件损坏。

## 版本发布

发布前先读取 `scripts\rebase-push.ps1`，确认其当前参数与副作用。该脚本按位置接收远程名和标签：`.\scripts\rebase-push.ps1 <remote> <version>`；它会 fetch 当前分支、rebase 到 `<remote>/<当前分支>`、以 `--force-with-lease` 推送当前分支，并强制更新和推送该标签。

按以下流程发布：

1. 完成实现与验证，将本次改动（含必要的技能更新）提交；确认 `git status --short` 为空。
2. 同时查询本地标签与 `git ls-remote --tags <remote>` 返回的目标远程标签；只将稳定版 `v<major>.<minor>.<patch>` 参与语义化排序，忽略 `-beta`、`-rc` 等预发布标签。以两者中的最高稳定版的 patch 加一生成新标签，并确认该标签尚不存在。
3. 运行 `.\scripts\rebase-push.ps1 <remote> <new-version>`。该命令会产生远程分支和标签写入；仅在用户明确授权发布时执行。
4. 结束后核验 `git status --short` 为空，并用 `git ls-remote --heads <remote> <branch>` 与 `git ls-remote --tags <remote> <new-version>` 确认分支和标签均已推送。

遇到 rebase 冲突时，停止发布推送：解决冲突后执行 `git rebase --continue`，或用 `git rebase --abort` 回退，再重新确认版本号和用户意图；不要跳过冲突或改用无保护的 force push。

## 节点命名

- 使用 **PascalCase**，同一任务内节点以任务名/模块名为前缀。
- 内部实现节点以 `__` 开头（如 `__ScenePrivateXXX`），不对外暴露。
- Anchor 使用描述路由职责的独立名称，如 `LargeEventPostEntryRoute`、`BattleExitRoute`；不要与任何 Pipeline 节点同名。
- 示例：`ResellMain`、`DailyProtocolPassInMenu`、`RealTimeAutoFightEntry`。

## 文件组织

- 一个独立业务流程使用一个独立 Pipeline JSON 文件；文件名反映业务用途，并放在所属模块目录中。
- 模块入口文件只保留主入口、顶层状态编排和对独立流程入口节点的引用，不在入口文件中堆放完整子流程实现。
- 同一业务流程的入口、状态路由、回退和收尾节点放在同一文件；仅 Common 节点和明确设计为跨模块复用的节点允许跨文件实现。
- 新增流程前检查模块目录结构并沿用现有拆分方式。例如每日奖励中的独立奖励类型应放入 `assets\resource\pipeline\DailyRewards\<FlowName>.json`，再由 `DailyRewards.json` 引用入口节点。

## Pipeline v2 格式（推荐）

Universal pipeline 使用 v2 格式，recognition 和 action 放入二级字典：

```jsonc
{
    "MyNode": {
        "recognition": {
            "type": "TemplateMatch",
            "param": {
                "template": "MyTask/button.png",
                "roi": [100, 200, 300, 100],
                "threshold": 0.7,
            },
        },
        "action": {
            "type": "Click",
        },
        "next": ["NextNode"],
    },
}
```

## 常用识别算法

### TemplateMatch（找图）

```jsonc
"recognition": {
    "type": "TemplateMatch",
    "param": {
        "template": "path/to/image.png",  // 相对 image 文件夹
        "roi": [x, y, w, h],              // 720p 坐标，缩小搜索范围
        "threshold": 0.7                   // 默认 0.7，按需调整
    }
}
```

- 图片必须从无损原图裁剪并缩放到 720p。
- `green_mask: true` 可遮蔽不参与匹配的区域（用 RGB(0,255,0) 涂色）。

### OCR（文字识别）

```jsonc
"recognition": {
    "type": "OCR",
    "param": {
        "roi": [x, y, w, h],
        "expected": ["完整文本"]
    }
}
```

- `expected` 写完整文本，不要写片段。
- 无需手动维护多语言——`tools/i18n` 会自动处理。
- 需要写片段或正则时，在 `expected` 数组中加 `// @i18n-skip` 注释。

### ColorMatch（找色）

```jsonc
"recognition": {
    "type": "ColorMatch",
    "param": {
        "roi": [x, y, w, h],
        "method": 40,                     // HSV 空间（推荐）
        "lower": [h_low, s_low, v_low],
        "upper": [h_high, s_high, v_high],
        "count": 100
    }
}
```

- 优先使用 HSV（method: 40）或灰度（method: 6），避免 RGB 直接匹配（不同显卡渲染差异）。

### And / Or（组合识别）

```jsonc
// And：全部子识别都成功才算命中
"recognition": {
    "type": "And",
    "param": {
        "all_of": ["NodeA", "NodeB"],  // 可引用节点名或内联 object
        "box_index": 0
    }
}

// Or：任一子识别成功即命中
"recognition": {
    "type": "Or",
    "param": {
        "any_of": ["NodeA", "NodeB"]
    }
}
```

### Custom（自定义识别）

调用 go-service 注册的自定义识别器：

```jsonc
"recognition": {
    "type": "Custom",
    "param": {
        "custom_recognition": "ExpressionRecognition",
        "custom_recognition_param": {
            "expression": "{CreditOCR}<300"
        }
    }
}
```

## 常用动作类型

| 动作                   | 用途            | 关键字段                               |
| ---------------------- | --------------- | -------------------------------------- |
| `Click`                | 点击            | `target`, `target_offset`              |
| `LongPress`            | 长按            | `target`, `duration`                   |
| `Swipe`                | 滑动            | `begin`, `end`, `duration`             |
| `Scroll`               | 滚轮（仅Win32） | `target`, `dx`, `dy`                   |
| `ClickKey`             | 按键            | `key`（虚拟键码）                      |
| `InputText`            | 输入文本        | `input_text`                           |
| `StartApp` / `StopApp` | 启停应用        | `package`                              |
| `StopTask`             | 停止当前任务链  | 无                                     |
| `Custom`               | 自定义动作      | `custom_action`, `custom_action_param` |
| `DoNothing`            | 不执行（默认）  | 无                                     |

`target` 支持：`true`（当前识别结果）、节点名字符串、`[x, y]`、`[x, y, w, h]`。

## 流程控制

### next 列表

按序识别，首个命中的节点执行其 action 后成为当前节点。`next` 为空或全部超时则任务结束。

### on_error

识别超时或动作失败时执行的节点列表。

### Node Attributes（节点属性）

**`[JumpBack]`**：命中后执行完该节点链，自动返回父节点继续识别 next。适用于处理弹窗、加载等中断场景。

```jsonc
"next": [
    "BusinessNode",
    "[JumpBack]HandlePopup",
    "[JumpBack]WaitLoading"
]
```

**`[Anchor]`**：动态引用锚点，运行时解析为最后设置该锚点的节点。锚点是可重定向的路由插槽，不是节点别名；名称应描述连接点或后续职责，并与所有节点名保持不同。为正常流程设置明确的默认映射，再由复用流程按需覆盖同一锚点。

### 等待画面稳定

只在必须时使用 `pre_wait_freezes` / `post_wait_freezes` 等待画面静止，不要为了执行稳定而使用延迟：

```jsonc
"post_wait_freezes": {
    "time": 200,
    "target": [0, 0, 0, 0]  // 全屏
}
```

避免对同一按钮重复点击——第二次点击可能作用于下一界面的其他元素。

### max_hit

限制节点最大命中次数，超过后自动跳过：

```jsonc
"max_hit": 3
```

## 可复用节点

编写前先检查是否已有可复用节点，避免重复造轮子。

### 通用交互

通用按钮和交互以 `assets\resource\pipeline\Common\` 为唯一事实来源。优先复用：

| 节点                                              | 用途                   |
| ------------------------------------------------- | ---------------------- |
| `CommonConfirmReward`                             | 领取奖励确认           |
| `CommonConfirmAction`                             | 通用操作确认           |
| `CommonSkipSettlement`                            | 跳过结算               |
| `CommonClosePage`                                 | 匹配多种右上角关闭按钮 |
| `CommonClickMax`                                  | 点击 MAX               |
| `CommonClickBlank` / `CommonClickMiddleBlank`     | 点击空白区域           |
| `CommonEndTask`                                   | 显式结束流程           |
| `DialoguesSkipWorkflow`                           | 完整跳过剧情流程       |
| `DialoguesClickContinue` / `DialoguesClickOption` | 对话继续或选择选项     |

完整辅助识别节点与语义见 [repository_reference.md](repository_reference.md)。新增可跨任务复用的按钮、对话交互或工作流时，添加到上述对应文件；任务私有交互留在任务自己的 Pipeline 文件中。

### 通用状态

通用状态判断以 `assets\resource\pipeline\Common\State.json` 为唯一事实来源。当前可复用 `CommonNoTransparentMask`（无子页面透明遮罩）和 `CommonNoAvailableTeamSet`（无可用队伍设置）。新增跨任务复用的页面状态判断时添加到该文件；任务私有状态留在任务文件中。

### Go Custom

项目自有 Custom 放在 `agent\go-service\custom`：Action 位于 `custom\action\<feature>`，Recognition 位于 `custom\recognizer\<feature>`。每个包用 `register.go` 调用 `maa.AgentServerRegisterCustomAction` 或 `maa.AgentServerRegisterCustomRecognition`，并在 `agent\go-service\register.go` 导入和调用包的 `Register()`。

先复用已注册 Custom，再新增实现；名称、类别和用途见 [repository_reference.md](repository_reference.md)。新增 Custom 必须同时完成 Go runner、包内注册、总注册入口与 Pipeline JSON 调用，参数使用明确的 JSON struct 校验。

### Custom Recognition 结果与滚动搜索

Custom Recognition 成功时应返回**实际命中的最小可操作框**；如需安全边距，只对该命中框做几何变换。普通点击仍由 Pipeline 节点的 `Click` action 执行，不要在 Recognition 内重复派发点击；只有为继续识别而必需的动作（如滚动到未见内容）才在 Custom 内执行。

实现可滚动搜索时，每轮遵循“完整识别 → 保存滚动前画面探针 → 滚动 → 获取新截图 → 判断是否仍有内容变化 → 从完整识别重新开始”。为搜索设置明确的最大滚动次数；用稳定、与目标无关的视觉区域判断列表是否停止变化，避免因目标本身消失而误判到达末尾。

### 排查 Go Custom 动作失败

先读取同一运行时间窗口内的两个日志：`agent\go-service\debug\go-service.log` 用于定位 Go runner 的报错位置，`agent\go-service\debug\maafw.log` 用于还原框架事件与动作生命周期。用任务 ID、节点名、`context_id` 和时间戳关联记录；不要只搜索错误文本后直接归因。

对 `RunActionDirect` / `RunRecognitionDirect`，按以下顺序判断失败阶段：

1. 找到对应 `MaaContextRun*Direct ... | enter`，确认调用参数与动作类型。
2. 在其后检查 `Node.ActionNode.Starting`、`Node.Action.Starting`、`Node.Action.Succeeded` 或 `Node.ActionNode.Failed`。已出现 `Node.Action.Starting` 后失败，说明框架已经派发动作，Go 中的 `failed to run action direct` 只是对底层失败的泛化结果；不要据此断言 Custom 回调或参数构造未被支持。
3. 若动作已 `Succeeded`，而后续 `screencap`/`GetImage` 报错，应将其归为截图阶段，而不是该动作失败。
4. Custom Recognition 可能在同一 Pipeline 节点内反复执行。以单次 `handle_recognition_request` 到 `_CustomRecognitionResponse` 的窗口判断“开始/中间/结束”；不要把整个节点的首次启动或最终失败误当作当前动作的阶段。
5. Go 与 MaaFramework 日志未提供控制器失败详情时，结论止于“控制器执行失败”；需要继续定位时再查客户端/控制器日志。并发截图导致的 `unexpected ...Screencap...Response` 是回包乱序证据，须与动作成功/失败事件分开分析。

## 典型模式

### 带弹窗处理的任务入口

```jsonc
{
    "MyTaskEntry": {
        "next": [
            "MyTaskMainStep",
            "[JumpBack]SceneDialogConfirm",
            "[JumpBack]SceneWaitLoadingExit",
            "[JumpBack]SceneAnyEnterWorld",
        ],
    },
}
```

### 跨页面活动流程

当一个任务涉及**多个页面跳转**（如：大地图 → 活动入口 → 难度选择 → 队伍配置 → 战斗），用 MaaFramework 的 `next` + `[JumpBack]` 串接可枚举页面；把复杂运行时判断或数据处理放入 Go Custom，再由 JSON 继续编排。不要写 Python orchestration。

```jsonc
{
    "MyActivity_Start": {
        "next": [
            "MyActivity_TeamReady",                       // 已在队伍配置页
            "[JumpBack]MyActivity_Difficulty_Select",     // 在难度选择页
            "[JumpBack]MyActivity_Enter"                  // 在大地图
        ],
        "timeout": 10000
    },

    "MyActivity_Enter": {
        "next": [
            "MyActivity_Enter_Click",                    // 找到图标
            "[JumpBack]BigMap_Activity_Resident",         // 切"常驻"tab
            "[JumpBack]BigMap_Activity"                  // 打开活动页
        ],
        "timeout": 10000
    },

    "MyActivity_EnterBattle": {
        "recognition": { "type": "OCR", "param": { "expected": ["进入战斗"], "roi": [...] } },
        "action": { "type": "Click" },
        "next": [
            "MyActivity_FightStart",                       // 战斗开始
            "[JumpBack]MyActivity_TravelSelect_Boat",      // 乘船
            "[JumpBack]MyActivity_TravelSelect_Walk"       // 步行
        ]
    }
}
```

**关键设计要点**：

- **`[JumpBack]` 是状态回退原语**：命中后执行完节点链，自动返回父节点的 `next` 继续识别。
- **窄 ROI 区分同名字段**：用 y 范围 [490, 740, 100, 80] vs [490, 590, 100, 80] 区分两个"确定"按钮行。
- **`target_offset` 偏移点击**：识别难度文字后用 `target_offset: [270, 0, 0, 0]` 右移到"确定"按钮位置。
- **跨文件节点引用**：MaaFramework 全局加载会合并所有 `pipeline/*.json`，跨文件引用 OK。但 `run_pipeline` 测试工具只加载单文件，集成测试需用 GUI/CLI。

**实战决策流程**：

```
要实现一个跨页面流程
│
├─ 流程可枚举为有限页面状态（A→B→C→D）？
│   └─ ✅ 用 JSON 状态机（next + [JumpBack]）
│       示例：成长试炼、相亲、英雄副本
│
└─ 流程涉及复杂计算、组合识别或运行时数据处理？
    └─ 实现 Go Custom Recognition / Action，由 JSON 节点调用并连接后续状态
```

详细反模式参见 [.claude/skills/pipeline-option/SKILL.md](../pipeline-option/SKILL.md) 的「不要做 #10」。

### 确认后验证画面变化

```jsonc
{
    "ClickConfirm": {
        "recognition": { "type": "TemplateMatch", "param": { "template": "confirm.png", "roi": [...] } },
        "action": { "type": "Click" },
        "post_wait_freezes": { "time": 200, "target": [0, 0, 0, 0] },
        "next": ["VerifyNextScreen", "[JumpBack]ClickConfirm"]
    }
}
```

### And 组合识别（背景 + 图标）

```jsonc
{
    "MyButton": {
        "recognition": {
            "type": "And",
            "param": {
                "all_of": ["ButtonBackground", "ButtonIcon"],
                "box_index": 0,
            },
        },
        "action": {"type": "Click"},
    },
}
```

## 审查清单

- [ ] 字段名拼写正确、类型合法（核对 Pipeline 协议）
- [ ] 无不必要的 `pre_delay` / `post_delay` / `timeout`
- [ ] `next` 列表覆盖所有可能画面，含弹窗/加载/异常
- [ ] 每次点击后有识别验证，不假设操作后状态
- [ ] ROI / target 坐标基于 1280×720
- [ ] JSON 格式化符合 `.prettierrc`
- [ ] `locales/` 已添加新增任务的多语言文本
- [ ] OCR `expected` 写完整文本
- [ ] 优先通过中间节点避免重复点击，只在必须时用 `post_wait_freezes`
- [ ] 未引用 `__ScenePrivate*` 内部节点
- [ ] Anchor 名称表达路由职责，且不与任何 Pipeline 节点同名
- [ ] 已优先复用 Common 交互、状态或现有 Go Custom
- [ ] 新增通用交互/状态/Custom 已放入约定文件或目录并完成注册
- [ ] 独立业务流程已放入独立 JSON，模块入口文件仅保留顶层编排

## 参考

- Pipeline 协议完整规范：[PipelineProtocol](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/3.1-PipelineProtocol.md)
- 仓库复用清单与扩展位置：[repository_reference.md](repository_reference.md)
- 开发手册：`docs/zh_cn/developers/README.md`
- 节点测试：`docs/zh_cn/developers/node-testing.md`
