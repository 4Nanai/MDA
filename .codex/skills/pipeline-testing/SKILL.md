---
name: pipeline-testing
description: Test and diagnose MaaFramework Pipeline node recognition against local screenshots with the MDA pipeline-test CLI. Use when changing recognition parameters, templates, OCR, And/Or references, or Go Agent Custom Recognition; verifying a node hit or miss; or collecting reproducible evidence for a recognition bug.
---

# Pipeline Testing

Use `pipeline-test node` for deterministic recognition-only checks. It loads the complete Resource bundle but never runs the target action, delays, waits, `next`, or `on_error`.

## Workflow

1. Read the target and referenced recognition nodes. Identify built-in or Agent-backed `Custom` recognition.
2. Select a representative 1280 by 720 PNG or JPEG. Never resize in the command.
3. Run a positive fixture with `--expect hit`.
4. Run at least one representative negative fixture with `--expect miss` when false positives matter.
5. Add `--artifacts <directory>` when draw images are useful.
6. Report command, fixture, expectation, actual hit, algorithm, box, outcome, and artifact paths.

```powershell
& ".\install\pipeline-test.exe" node --node "ArenaClickCheerUp" --image ".\fixtures\arena-cheer-hit.png" --expect hit --output text
& ".\install\pipeline-test.exe" node --node "ArenaClickCheerUp" --image ".\fixtures\arena-cheer-miss.png" --expect miss --output json --artifacts ".\debug\pipeline-test"
```

Defaults are discovered from the repository or install layout. Override them when needed:

```powershell
& ".\install\pipeline-test.exe" node --node "TargetNode" --image ".\fixtures\target.png" --resource ".\assets\resource" --library-dir ".\install\maafw" --agent ".\install\agent\go-service.exe" --agent-timeout 10s --expect hit --output json
```

Use `--no-agent` only for built-in nodes or to verify that Custom setup fails. Auto mode starts the default Agent only when it exists. An unavailable required Custom registration is exit `2`, never a miss.

## Results

- Exit `0`: expectation passed.
- Exit `1`: recognition ran, but actual hit differed from `--expect`.
- Exit `2`: setup or execution failed.
- JSON stdout is one `pipeline-test/v1` document. Agent logs and non-result diagnostics remain on stderr.
- Inspect recursive `combined` entries for `And` and `Or` failures.

Do not treat exit `1` as infrastructure failure or exit `2` as a valid miss.

## Boundaries

- Use schema checks and `.claude\skills\pipeline-debug` for static fields, references, graphs, and naming. Static checks do not prove recognition.
- Use this skill for one node's recognition against fixed pixels.
- Action, timing, controller input, transitions, routing, retries, and `on_error` require a future action-aware Pipeline fixture runner or live test.
- Project Interface option merging and task entry resolution require the future PI `task` command.
- The CLI exposes only `node`; do not invent `pipeline` or `task` commands.

## Troubleshooting

- `image`: verify path, PNG/JPEG decoding, and exact 1280 by 720 dimensions.
- `framework`: verify the MaaFramework DLL set matches the Go binding.
- `resource`: verify the complete bundle and required models/templates.
- `node`: verify the exact case-sensitive name after overlays load.
- `agent`: build the production Agent, verify its `maafw` working-directory dependency, and confirm the required registration. Increase `--agent-timeout` only for measured startup delay.
- OCR miss: verify model language, ROI, expected text, and screenshot scale.
- Template miss: verify the Resource-relative template path, ROI, masking, and pixel scale.
- Assertion failure: inspect actual hit, box, combined details, and draw artifacts.

Do not lower thresholds merely to make a fixture pass. First establish fixture and ROI validity, then justify a threshold change with positive and negative evidence.
