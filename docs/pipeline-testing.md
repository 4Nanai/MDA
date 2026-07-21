# Pipeline Recognition Testing

`pipeline-test` is a Windows PowerShell 7 CLI for testing one MaaFramework Pipeline node's recognition against a local screenshot. It loads the complete Resource bundle and uses the production Go Agent when Custom Recognition is required. It never executes the target node's action or flow.

## Build

Prepare `install\maafw` with MaaFramework DLLs compatible with `agent\go-service\go.mod`, then run:

```powershell
Push-Location ".\agent\go-service"
go build -o "..\..\install\agent\go-service.exe" .
go build -o "..\..\install\pipeline-test.exe" ".\cmd\pipeline-test"
Pop-Location
```

`python .\tools\build_and_install.py` also builds both executables during a normal developer install.

## Run A Node Test

The screenshot must be an unmodified 1280 by 720 PNG or JPEG.

```powershell
& ".\install\pipeline-test.exe" node --node "ArenaClickCheerUp" --image ".\fixtures\arena-cheer-hit.png" --expect hit
& ".\install\pipeline-test.exe" node --node "ArenaClickCheerUp" --image ".\fixtures\arena-cheer-miss.png" --expect miss --output json --artifacts ".\debug\pipeline-test"
```

Use `--resource`, `--library-dir`, and `--agent` to override discovered paths. Use `--no-agent` for built-in recognition only. Exit statuses are `0` for a passing expectation, `1` for an assertion failure, and `2` for setup or execution failure.

## Test Fixtures

Synthetic fixtures owned by this CLI live in `agent\go-service\internal\pipelinetest\testdata`. Regenerate them from `agent\go-service` with:

```powershell
go run ".\internal\pipelinetest\testdata\generate"
```

Do not add account data, user screenshots, time-sensitive UI, or unstable captures. Keep the pipeline and generator synchronized.

## Verification

```powershell
Push-Location ".\agent\go-service"
go test ./...
$env:MDA_PIPELINE_TEST_INTEGRATION = "1"
go test ".\internal\pipelinetest" -run "^TestMaa" -v
Remove-Item Env:\MDA_PIPELINE_TEST_INTEGRATION
Pop-Location
```

## Scope And Roadmap

The current CLI exposes only `node`. Static checks remain separate, and live controller validation is required for device input and real transitions.

A future `pipeline` command can add an action-aware FixtureController that advances screenshot sequences after expected actions and verifies routing, timing, retries, and `on_error`. A future `task` command can resolve Project Interface entries and option-driven `pipeline_override` data before delegating to that runner. Neither command is exposed until those semantics are implemented.
