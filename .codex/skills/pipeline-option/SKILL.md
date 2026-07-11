---
name: pipeline-option
description: "Add MaaFramework runtime UI options (select/checkbox/switch/input) through interface task files and Pipeline overrides. Supports pure pipeline_override behavior and Go Agent decisions with maa-framework-go/v4 Context.GetNode, typed node parameters, and Custom Action/Recognition arguments. Use when adding or reviewing user-facing options, persisted task configuration, option-driven Pipeline branches, or Go Custom logic controlled by interface options."
---

# Pipeline Option 工作流（Go Agent）

## 项目结构与联动

新增 UI 选项时同时检查以下位置：

| # | 位置 | 内容 |
| --- | --- | --- |
| 1 | `assets\tasks\<Task>.json` 的 `option` 对象 | 定义 type、cases/inputs、默认值和 `pipeline_override` |
| 2 | 同文件对应 task 的 `option` 数组 | 引用选项名，否则 UI 不显示 |
| 3 | `assets\resource\pipeline\*.json` | 预定义 override 目标节点 |
| 4 | `agent\go-service\custom\...` | 仅当 Go 必须读取配置并作运行时决策时修改 |

`assets\interface.json` 通过 `import` 加载任务文件。若任务未拆分，才直接在 interface 根文件定义 task/option。`pipeline_override` 只覆盖已加载节点，不创建节点。

## Go API 映射

项目使用 `github.com/MaaXYZ/maa-framework-go/v4 v4.0.0-beta.14`：

- 使用 `ctx.GetNode(name)` 读取 override 后的强类型 `*maa.Node`。
- 使用 `ctx.GetNodeJSON(name)` 获取原始节点 JSON。
- 使用 `ctx.OverridePipeline(value)` 或 `ctx.OverrideNext(...)` 在 Custom 运行中动态覆盖。
- 当前 Custom runner 直接从 `arg.CustomActionParam` / `arg.CustomRecognitionParam` 读取参数，不要绕行查询自己的节点。

优先让 option 直接覆盖 `next`、`enabled`、recognition/action param。只有复杂计算、组合识别或复杂动作才进入 Go Custom。

## TL;DR：3 处联动

新增一个 UI 选项需要**同时**改 3 个地方，缺一不可：

| # | 位置 | 内容 |
|---|------|------|
| 1 | `assets\tasks\<Task>.json` 的 `option` 字典 | 选项定义（type / cases / pipeline_override） |
| 2 | 同文件对应 task 的 `option: []` 数组 | 注册到具体任务（否则 UI 上看不到） |
| 3 | `assets\resource\pipeline\*.json` | **预定义**目标节点（pipeline_override 不会创建节点） |
| 4 | Go Custom（仅代码决策必需） | `ctx.GetNode()` 或 Custom 参数读取；pure override 无需 Go 改动 |

> ⚠️ **pipeline_override 只做属性合并，不会凭空创建节点。** 少了第 3 步，`ctx.GetNode()` 会报错，运行时覆盖也不会得到预期结果。

完整协议参考（嵌套 option、global_option、controller/resource 限制、占位符注入）：[references/protocol.md](references/protocol.md)

---

## 4 种 type 速查

| type | 选择 | override 字段 | 节点预定义形态 |
|------|------|---------------|---------------|
| `select` | 单选互斥 | `expected` | `recognition: "OCR"` + `expected: [...]` |
| `switch` | 二元 Yes/No | `enabled` | `{"enabled": bool}` |
| `input` | 自由文本 | `custom_action_param` | `action.param.custom_action_param` |
| `checkbox` | 多选 | `enabled` | `{"enabled": false}` |

## 选哪个模式？

| 你的需求 | 推荐模式 |
|---------|---------|
| 启用/禁用一个 Go Custom 业务分支 | **A**（优先直接覆盖执行节点；必要时 Flag + Go 读取） |
| 从多个互斥选项里选一个值 | **B**（select + OCR 节点） |
| 同时启用多个独立的功能模块 | **C**（checkbox + 多个 Flag 节点） |
| 用户输入自定义文本 | **D**（input + 占位符注入） |
| 切换行为（点哪个按钮 / 走哪条 next 链）且不需要代码判断 | **E**（pure override 现有节点字段） |

> **黄金法则**：能 pure override 解决就不加 Flag + Go 改动。

---

## 模式 A：开关（switch + Flag 节点）— 最常用

**适用**：开启/关闭某个功能。

### interface.json

```jsonc
"开启5月城堡相亲": {
    "type": "switch",
    "description": "是否开启5月自动相亲",
    "default_case": "Yes",
    "cases": [
        {
            "name": "Yes",
            "pipeline_override": { "Flag_EnableMarryTask": { "enabled": true } }
        },
        {
            "name": "No",
            "pipeline_override": { "Flag_EnableMarryTask": { "enabled": false } }
        }
    ]
}
```

### 配套 pipeline 节点（必须预定义！）

```jsonc
"Flag_EnableMarryTask": { "enabled": true }
```

### 注册到 task

```jsonc
"task": [{
    "name": "推年计划",
    "entry": "Auto_YearlyTask",
    "option": ["开启5月城堡相亲", /* 其他选项 */]
}]
```

### Go 读取（仅 Custom 代码确实需要时）

```go
node, err := ctx.GetNode("Flag_EnableMarryTask")
if err != nil {
    log.Error().Err(err).Msg("failed to read option node")
    return false
}
enabled := true // 协议默认 enabled=true
if node.Enabled != nil {
    enabled = *node.Enabled
}
if !enabled {
    log.Info().Msg("feature disabled")
    return true
}
```

---

## 模式 B：单选（select + OCR 节点）

**适用**：选择城市、关卡、模式等互斥选项。

### interface.json

```jsonc
"选择刷取任务国家": {
    "type": "select",
    "description": "选择要刷取任务的目标城市",
    "default_case": "雄月城",
    "cases": [
        { "name": "王座堡", "pipeline_override": { "EnterCity": { "expected": ["王座堡"] } } },
        { "name": "雄月城", "pipeline_override": { "EnterCity": { "expected": ["雄月城"] } } }
    ]
}
```

### 配套 OCR 节点

```jsonc
"EnterCity": {
    "recognition": "OCR",          // ⚠️ 必须是 OCR，否则 expected 不生效
    "expected": ["王座堡", "圣盾堡", "雄月城", "翠庭"],
    "roi": [58, 320, 600, 682],
    "action": "Click"
}
```

### Go 读取

```go
node, err := ctx.GetNode("EnterCity")
if err != nil || node.Recognition == nil {
    return false
}
param, ok := node.Recognition.Param.(*maa.OCRParam)
if !ok || len(param.Expected) == 0 {
    return false
}
city := param.Expected[0]
```

---

## 模式 C：多选（checkbox + 多个 Flag 节点）

**适用**：多条件检测（好苗子条件）、可叠加的功能模块。

### interface.json

```jsonc
"开启好娃提醒": {
    "type": "checkbox",
    "default_case": ["科内塔之怒"],
    "cases": [
        { "name": "科内塔之怒",   "pipeline_override": { "检测_科内塔之怒":     { "enabled": true } } },
        { "name": "太阳+科内塔之怒", "pipeline_override": { "检测_太阳+科内塔之怒": { "enabled": true } } }
    ]
}
```

### 配套节点（每个 case 一个，默认全 false）

```jsonc
"检测_科内塔之怒":      { "expected": ["koneita"],            "enabled": false },
"检测_太阳+科内塔之怒": { "expected": ["sun_and_koneita"],    "enabled": false }
```

### Go 读取（遍历收集）

```go
var enabled []string
for _, name := range []string{"CheckKoneita", "CheckSunAndKoneita"} {
    node, err := ctx.GetNode(name)
    if err != nil || node.Enabled == nil || !*node.Enabled || node.Recognition == nil {
        continue
    }
    param, ok := node.Recognition.Param.(*maa.OCRParam)
    if ok && len(param.Expected) > 0 {
        enabled = append(enabled, param.Expected[0])
    }
}
```

---

## 模式 D：自由输入（input + 占位符注入）

**适用**：用户输入自定义关卡号、自定义黑名单任务等。

### interface.json

```jsonc
"自定义任务黑名单": {
    "type": "input",
    "inputs": [
        {
            "name": "任务名称",
            "pipeline_type": "string",
            "default": "",
            "verify": "^[^,，]*$",
            "pattern_msg": "不能包含逗号"
        }
    ],
    "pipeline_override": {
        "CustomTaskBlacklist": {
            "expected": ["{任务名称}"]   // {名称} 占位符被实际输入替换
        }
    }
}
```

### Go 读取

```go
node, err := ctx.GetNode("CustomTaskBlacklist")
if err != nil || node.Recognition == nil {
    return false
}
param, ok := node.Recognition.Param.(*maa.OCRParam)
if !ok || len(param.Expected) == 0 {
    return false
}
value := param.Expected[0]
```

若输入供 Custom 使用，优先把占位符注入 `action.param.custom_action_param` 或 `recognition.param.custom_recognition_param`。Go runner 使用 `encoding/json` 将参数字符串反序列化为明确 struct：

```go
type OptionParam struct {
    Value string `json:"value"`
}

var param OptionParam
if err := json.Unmarshal([]byte(arg.CustomActionParam), &param); err != nil {
    log.Error().Err(err).Msg("invalid custom option")
    return false
}
```

---

## 模式 E：行为覆盖（pure override 现有节点字段）— 最简

**适用**：行为切换映射到现有 pipeline 节点的字段（`next` / `action` / `recognition` / 任何可覆盖字段），且不需要 Go 判断。

**核心思路**：用户切换 UI 选项 → 改变 pipeline 节点字段 → 框架执行。**Go 代码完全不动。**

### 典型场景：开关决定点哪个按钮

```jsonc
"开启自动接受佣兵": {
    "type": "switch",
    "default_case": "No",
    "cases": [
        {
            "name": "Yes",
            "description": "直接点确认",
            "pipeline_override": {
                "Event_MercenaryJoin": {
                    "next": ["Event_MercenaryJoinConfirm"]
                }
            }
        },
        {
            "name": "No",
            "description": "直接点取消",
            "pipeline_override": {
                "Event_MercenaryJoin": {
                    "next": ["Event_MercenaryJoinCancel"]
                }
            }
        }
    ]
}
```

目标节点必须在 `assets\resource\pipeline` 中有完整定义；`pipeline_override` 只覆盖 `next`，其他字段保持原值。

### `next` 数组的单元素 vs 多元素语义

| 写法 | 语义 | 何时用 |
|------|------|-------|
| `["A"]` | **强约束**：只走 A | 行为已确定，单路径足够（**模式 E 的典型形态**） |
| `["A", "B"]` | **回退链**：优先 A，A 失败走 B | 兜底机制（"优先点确认，找不到才点取消"） |
| `["A", "B", "[JumpBack]C"]` | 失败后跳回 C 节点重试 | 复杂回退 |

### 可被 override 的字段

| 字段 | override 效果 | 典型用途 |
|------|--------------|---------|
| `next` | 改变后续节点列表 | 切换行为路径（模式 E 主力） |
| `action` | 改变点击/滑动/输入动作 | 切换操作类型 |
| `recognition` | 改变识别算法 | 切换识别方式（OCR ↔ Template） |
| `expected` | 改变识别期望值 | 配合 select 选值 |
| `roi` | 改变识别区域 | 适配不同界面尺寸 |
| `timeout` | 改变超时时间 | 适配不同网络/性能 |

> **关键认识**：上面这些字段都是普通 JSON 值，pipeline_override 一视同仁做深合并。**模式 A 用的 `enabled` 字段只是最常见的入口，不是唯一可 override 的字段。**

### 模式 A vs 模式 E 对比

| 场景 | 模式 A（Flag + Go） | 模式 E（pure override） |
|------|------------------------|----------------------|
| Go Custom 必须执行条件逻辑 | ✅ 适用 | ❌ 无法替代 |
| 行为由 pipeline 字段决定 | ❌ 多此一举 | ✅ 最简 |
| 需要运行时根据 flag 走不同代码分支 | ✅ 唯一选择 | ❌ 不行 |
| 改动 Go 代码 | ✅ 需要 | ❌ 不需要 |
| 需要新加 Flag 节点 | ✅ 需要 | ❌ 不需要 |

### 实战决策流程

```
要加新选项
│
├─ 行为切换对应到一个 pipeline 节点的某个字段？
│   └─ ✅ 用模式 E（pure override）
│       示例：佣兵加入时点"确认"还是"取消"
│
└─ ❌ Go Custom 必须按配置执行复杂分支
    └─ 优先把值注入 Custom param；需要读取共享状态时用 Flag + ctx.GetNode()
```

---

## 补充：能用状态机就别写 Go orchestration

**MaaFramework 的 `next` + `[JumpBack]` 是跨页面状态推进原语**。若流程可列举为有限页面状态，优先用 JSON 状态机；不要在 Go 中循环调用 `ctx.RunTask()` 串联页面。

详见相邻 `pipeline-guide` skill 的跨页面流程模式。

### 状态机 vs Go Custom 对比

| 场景 | 状态机（推荐） | Go Custom |
|------|--------------|--------------------------|
| 有限页面状态推进（如活动流程） | ✅ 链 `next` + `[JumpBack]` | ❌ 自己写 `for/while` 调度 |
| 按配置执行复杂计算或动作 | ❌ 不适合 | ✅ 参数化 Custom |
| 组合识别与运行时数据处理 | ❌ 难表达 | ✅ Go 实现 |

### 跨文件节点引用的测试陷阱

MaaFramework 加载 `assets\resource\pipeline\*.json` 后合并节点命名空间，跨文件引用可解析。单文件节点工具无法代表完整集成加载。

**应对**：
- 集成测试必须用 MaaFramework GUI / CLI 触发，不能依赖 `run_pipeline`
- 单元测试每个节点用 `run_pipeline` 是 OK 的（无跨文件依赖）
- 若某个流程有跨文件引用，本地调试时考虑用 `MaaCli` 跑全 bundle

---

## 补充：状态机驱动的「流程型选项」

如果一个 UI 选项代表的是**进入某个跨页面流程**（如"开启成长试炼"→ 大地图 → 难度选择 → 队伍 → 战斗），把选项的 `pipeline_override` 用于：
1. 切换"是否启用流程"的 Flag 节点
2. 注入该流程入口节点所需参数（如难度 `expected`）

不要用 Go orchestration 串联流程中的每个节点。完整流程示例：

```jsonc
// 选项定义
"开启3月成长试炼": {
    "type": "switch",
    "default_case": "No",
    "cases": [
        {
            "name": "Yes",
            "pipeline_override": {
                "Flag_GrowthTrialMode": { "enabled": true },
                "GrowthTrial_Difficulty_Select": { "expected": ["噩梦"] }
            }
        },
        {
            "name": "No",
            "pipeline_override": {
                "Flag_GrowthTrialMode": { "enabled": false }
            }
        }
    ]
}
```

```jsonc
// 入口节点（路由）
"GrowthTrial_Start": {
    "next": [
        "GrowthTrial_TeamReady",                  // 已在队伍配置页
        "[JumpBack]GrowthTrial_Difficulty_Select", // 在难度选择页
        "[JumpBack]GrowthTrial_Enter"             // 在大地图
    ]
}
```

战斗入口自动接力：

```jsonc
"GrowthTrial_EnterBattle": {
    "action": "Click",
    "next": [
        "GrowthTrial_FightStart",                   // 战斗开始
        "[JumpBack]GrowthTrial_TravelSelect_Boat",  // 弹出旅行框
        "[JumpBack]GrowthTrial_TravelSelect_Walk"   // fallback
    ]
}
```

---

## 命名与默认值

### 命名约定

| 角色 | 风格 | 示例 |
|------|------|------|
| option 名（用户可见） | 中文动词起头 | `开启5月城堡相亲`、`选择刷取任务国家` |
| 节点名（pipeline） | 英文 | `Flag_EnableMarryTask`、`EnterCity`、`检测_科内塔之怒` |
| switch case 名 | **严格 `Yes` / `No`** | 不要用 `true/false` 或 `是/否`（Client 解析跨平台不一致） |

### 默认值策略

> **保持现有行为是底线。** 老用户不该因新选项而行为改变。

| 场景 | 推荐 default |
|------|-------------|
| 新开关让功能默认关闭 | `No`（明确告知用户"关了"） |
| 新开关让功能默认开启 | `Yes`（保留旧行为） |
| 旧代码无条件开启 | `Yes`（兼容） |
| 旧代码无条件关闭 | `No`（兼容） |

---

## Go 读取位置

| 决策类型 | 放哪读 | 理由 |
|---------|-------|------|
| 是否执行 Custom 逻辑 | 对应 runner 的 `Run` 入口 | 逻辑与配置消费位置保持一致 |
| 多处复用的选项值 | `Run` 入口读取后传给辅助函数 | 一次读取、多次复用 |

> 不要把各功能开关集中堆进通用 dispatch；让实际消费该配置的 runner 自治。

---

## ✅ 推荐做法

1. **先复用现有模式**：参考同项目里现成的同类选项（开关 → `开启5月城堡相亲`；选择 → `选择刷取任务国家`）
2. **3 处同步改完再跑**：不要中途停下来"先编译试试"
3. **JSON 改完运行 `python tools\validate_schema.py --task-dirs assets\tasks`**：检查 interface、task 和 Pipeline schema
4. **默认值遵循现状**：选项是"开"还是"关"取决于旧代码行为，不是你的偏好
5. **在 task 的 `doc` 数组里加一行说明**：用户能看懂每个选项的作用
6. **能用 pure override 解决就不要加 Flag + Go 改动**。Flag 仅在 Go Custom 必须读取共享开关时使用

---

## ❌ 不要做

### 1. 不要只通过 pipeline_override 定义节点

```jsonc
// ❌ 错：节点没在 Pipeline JSON 中预定义，override 无目标

// ✅ 对：在 pipeline JSON 里预定义
"Flag_EnableSailingFestivalPurchase": { "enabled": true }
```

**验证方法**：运行 schema 校验，并在 Go 中处理 `ctx.GetNode()` 错误。

### 2. 不要忘了注册到 task 的 option 数组

```jsonc
// ❌ 错：option 定义了但 task 不引用 → UI 上看不到
"option": []

// ✅ 对：同步注册
"option": ["开启3月启航节购买"]
```

### 3. 不要把判断塞到通用 dispatch

```go
// 在实际消费配置的 Custom runner 入口读取，不要让通用分发器累积选项分支。
node, err := ctx.GetNode("FlagEnableFeature")
if err != nil {
    return false
}
```

### 4. 不要混淆字段路径

| 用途 | 字段路径 | 备注 |
|------|---------|------|
| `select` | `node.Recognition.Param.(*maa.OCRParam).Expected` | 先检查 recognition 与类型断言 |
| `input` | `json.Unmarshal(arg.CustomActionParam, &param)` | 当前 Custom 直接读取参数 |
| `switch` / `checkbox` | `node.Enabled` | 处理 nil 所代表的默认 true |
| 模式 E 不读 | （不读，直接执行 override 后节点） | pure override 不需要 flag |

### 5. 不要用非 `Yes`/`No` 的 switch case 名

```jsonc
// ❌ 错：Client 解析可能不一致
{ "name": "true" } / { "name": "是" } / { "name": "ON" }

// ✅ 对：跨 Client 一致
{ "name": "Yes" } / { "name": "No" }
```

### 6. 不要在 input 里塞 OCR expected 路径

`input` 用 `custom_action_param` 注入自定义文本，**与 `select` 的 `expected` 是两套独立机制**。混用会导致节点配置混乱、后续维护者读不懂。

### 7. 不要用中文做 pipeline 节点名

```jsonc
// ❌ 错：中文节点名 + 英文字段访问
"开启5月": { "enabled": true }

// ✅ 对：英文 Flag_ 命名
"Flag_EnableMarryTask": { "enabled": true }
```

中文做 option 名（用户可见），英文做 pipeline 节点名（代码访问）。混了会让代码和配置都对不上。

### 8. 不要在多文件 pipeline 里重复定义同名节点

`parse_and_override_once` 合并所有 pipeline JSON 时**严格拒绝**重复顶层 key。检查方法：

```powershell
rg -n '^\s*"YourNodeName"\s*:' 'assets\resource\pipeline'
```

两个文件定义同一顶层节点会导致资源加载失败；同时执行 schema 校验和完整 Client/CLI 加载验证。

### 9. 不要为了"配置统一"硬塞 Flag 节点

```jsonc
// ❌ 错：行为切换只动 pipeline 字段，却额外添加 Flag 节点 + Go 分支
"Flag_AcceptMercenary": { "enabled": true },   // ← 不必要
// ✅ 对：直接 override `next`，零 Go 改动
"开启自动接受佣兵": {
    "type": "switch",
    "pipeline_override": {
        "Event_MercenaryJoin": { "next": ["Event_MercenaryJoinConfirm"] }
    }
}
```

**判断口诀**：如果 Go 分支只调用 `ctx.RunTask()` 把控制权交回 Pipeline，优先改成 `pipeline_override`。Flag + Go 只用于真正的运行时条件逻辑。

### 10. 不要用 Go orchestration 替代状态机

若跨页面流程可列举为有限状态，使用 `next` + `[JumpBack]`；不要写 Go 循环 + `ctx.RunTask()` 调度。

```jsonc
// ✅ 对：纯 JSON 状态机（推荐）
"GrowthTrial_Start": {
    "next": [
        "GrowthTrial_TeamReady",                  // 已在队伍配置页
        "[JumpBack]GrowthTrial_Difficulty_Select", // 在难度选择页
        "[JumpBack]GrowthTrial_Enter"             // 在大地图
    ]
}

"GrowthTrial_Enter": {
    "next": [
        "GrowthTrial_Enter_Click",                  // 找到图标
        "[JumpBack]BigMap_Activity_Resident",       // 切"常驻"tab
        "[JumpBack]BigMap_Activity"                 // 打开活动页
    ]
}

"GrowthTrial_EnterBattle": {
    "action": "Click",
    "next": [
        "GrowthTrial_FightStart",                    // 战斗开始
        "[JumpBack]GrowthTrial_TravelSelect_Boat",   // 弹出旅行框
        "[JumpBack]GrowthTrial_TravelSelect_Walk"    // fallback
    ]
}
```

**自检问题**：
- 我的 Go 代码是否只在调用 `ctx.RunTask()` 把控制权交给 Pipeline？是 → 改用 `next` 链
- 我的"流程推进"是否依赖**显式的状态变量**（如 `found`）？是 → 改用 `[JumpBack]` 让框架自动回退
- 我的"流程"是否**可以画成状态机图**？是 → 用 JSON `next` 链

**注意**：MaaFramework 全局加载时跨文件节点引用会解析（`main_ui.json` 里的 `BigMap_Activity*` 能在 `growth_trial.json` 引用），但**`run_pipeline` 测试工具只加载单文件**——集成测试必须用 MaaFramework GUI/CLI 触发。

---

## 验证流程

改完一次完整流程，**按顺序**做这 4 步：

1. **JSON 语法检查**

   ```powershell
   Get-Content -Raw -Encoding utf8 -LiteralPath 'assets\interface.json' | ConvertFrom-Json | Out-Null
   ```

2. **资源加载检查**

   ```powershell
   python tools\validate_schema.py --task-dirs assets\tasks
   ```

3. **Go 测试**：修改 Go 时在 `agent\go-service` 运行 `gofmt` 与 `go test ./...`。

4. **端到端验证**：用 Pipeline Testing Skill 跑一次实际流程

---

## 完整协议

更多 type 字段、嵌套 option、global_option、controller/resource 限制、`{占位符}` 注入机制等高级特性见 [references/protocol.md](references/protocol.md)。
