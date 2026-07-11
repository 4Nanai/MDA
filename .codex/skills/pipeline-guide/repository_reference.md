# 仓库复用清单

修改 Pipeline 或 Go Agent Custom 时读取本文件，并以所列源文件的当前内容为最终依据。

## 通用交互

来源：`assets\resource\pipeline\Common\Interaction.json`

| 节点 | 语义 |
| --- | --- |
| `CommonConfirmReward` | OCR 识别并点击领取奖励确认 |
| `CommonConfirmAction` | 模板识别并点击操作确认 |
| `CommonSkipSettlement` | 模板识别并点击跳过结算 |
| `CommonClosePage` | 匹配多种关闭按钮并点击 |
| `CommonClickMax` | 通过模板与颜色组合识别 MAX 并点击 |
| `CommonClickBlank` | 点击左下空白区域 |
| `CommonClickMiddleBlank` | 点击中部空白区域 |
| `CommonEndTask` | 无动作的流程结束节点 |

来源：`assets\resource\pipeline\Common\Dialogues.json`

| 节点 | 语义 |
| --- | --- |
| `DialoguesSkipStory` | OCR 识别 SKIP 并点击 |
| `DialoguesClickContinue` | 模板识别并点击继续 |
| `DialoguesOptionOCR` | 通过序号 OCR 识别对话选项 |
| `DialoguesOptionTemplate` | 通过模板识别对话选项 |
| `DialoguesClickOption` | 合并两种选项识别并偏移点击 |
| `DialoguesMainVisible` | 判断剧情主界面可见 |
| `DialoguesSkipWorkflow` | 处理跳过、继续、选项和空白点击的完整流程 |
| `DialoguesSkipStoryAndEndTask` | 点击跳过并收束到结束节点 |

新增跨任务通用按钮或一般交互到 `Interaction.json`；新增跨任务通用剧情交互或剧情工作流到 `Dialogues.json`。同步添加所需模板图片；任务私有节点不要放入 Common。

## 通用状态

来源：`assets\resource\pipeline\Common\State.json`

| 节点 | 语义 |
| --- | --- |
| `CommonNoTransparentMask` | 通过颜色判断当前没有子页面透明遮罩 |
| `CommonNoAvailableTeamSet` | 判断没有可用队伍设置，尝试关闭/返回并输出提示 |

新增跨任务复用的纯状态判断到 `State.json`。如果判断需要复杂计算，优先在 `agent\go-service\custom\recognizer\common` 实现通用 Recognition，再由 `State.json` 封装为语义化节点。

## Go Custom

根目录：`agent\go-service\custom`

### Action

| 注册名 | 包目录 | 用途 |
| --- | --- | --- |
| `DailyRewardsPassClick` | `action\daily_reward` | 每日奖励通行证点击逻辑 |

### Recognition

| 注册名 | 包目录 | 用途 |
| --- | --- | --- |
| `RaidChallengeButtonUnavailableRecognition` | `recognizer\common` | OCR 与颜色组合判断挑战按钮不可用 |
| `CommonTemplateOCRRecognition` | `recognizer\common` | 模板与 OCR 组合识别 |
| `CommonWaitingPageLoadRecognition` | `recognizer\common` | OCR 与遮罩状态组合判断页面加载完成 |
| `CommonTemplateColorMatchRecognition` | `recognizer\common` | 模板与颜色组合识别 |
| `LargeEventStoryRecognition` | `recognizer\large-event` | 大型活动剧情识别 |
| `LargeEventMissionCompletedRecognition` | `recognizer\large-event` | 大型活动任务完成识别 |
| `SoloRaidRecognition` | `recognizer\soloraid` | 单人突袭入口/红点组合识别 |
| `SoloRaidSwitchTeam` | `recognizer\soloraid` | 识别当前队伍并返回下一队伍点击框 |
| `UnionRaidEntryRecognition` | `recognizer\unionraid` | 联盟突袭入口识别 |
| `UnionRaidOpenedRecognition` | `recognizer\unionraid` | 联盟突袭开启状态识别 |

新增 Custom 时：

1. 在 `custom\action\<feature>` 或 `custom\recognizer\<feature>` 创建/扩展 runner，并用编译期接口断言。
2. 在该包 `register.go` 注册稳定、唯一的 PascalCase 名称。
3. 在 `agent\go-service\register.go` 导入包并调用 `Register()`。
4. 在 Pipeline JSON 中使用完全一致的 `custom_action` 或 `custom_recognition` 名称，并传递与 Go JSON struct 对应的参数。
5. 运行 `gofmt` 和 `go test ./...`，再验证调用节点的 Pipeline 流程。
