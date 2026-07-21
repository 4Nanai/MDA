## Context

MDA currently runs schema and static checks but has no repository command for executing Pipeline recognition against captured screens. MaaFramework already supplies the required runtime primitives: Resource loading, Tasker execution, Context recognition by Pipeline entry, detailed recognition results, Custom Controllers, and Agent-backed Custom Recognition. The design should expose those primitives without reimplementing OCR, template matching, combined recognition, or Pipeline parsing.

The complete product direction includes node, Pipeline, and Project Interface task tests. The current change deliberately implements only the node-test vertical slice while defining reusable boundaries for later commands.

## Goals / Non-Goals

**Goals:**

- Execute the exact recognition configuration of a named, fully loaded Pipeline node against a local 1280 by 720 screenshot.
- Support built-in, composed, cross-file, and MDA Agent-backed Custom Recognition.
- Produce deterministic assertions, useful diagnostics, JSON output, and CI-compatible exit codes.
- Prevent node tests from executing actions or traversing the Pipeline.
- Create reusable runtime and result components for later Pipeline and task commands.
- Publish a repository skill that turns the CLI into a consistent contributor workflow.

**Non-Goals:**

- Simulating screen transitions or implementing the future `pipeline` command in this change.
- Parsing Project Interface tasks, presets, or option merge precedence in this change.
- Replacing MaaDebugger or providing a graphical debugger.
- Testing whether a click or swipe has the intended effect on a live application.
- Automatically resizing screenshots or silently modifying test inputs.

## Decisions

### Use a subcommand-based Go CLI in the existing Go module

The executable will live under `agent\go-service\cmd\pipeline-test`, with reusable implementation packages outside `cmd`. A subcommand interface makes `pipeline-test node` stable now while allowing later `pipeline` and `task` commands.

Alternative considered: a Python script. This was rejected because the project standardizes complex Pipeline extensions on Go, the existing Agent and binding are Go, and a second MaaFramework binding would increase packaging and behavior differences.

### Run recognition by node entry through `Context.RunRecognition`

The CLI will register a small in-process harness Custom Action on the loaded Resource. It will post a temporary DirectHit task whose action receives a MaaFramework Context, loads the already validated screenshot, and invokes `Context.RunRecognition(targetNode, image)`. The harness transfers the resulting `RecognitionDetail` back to the command and does not call the target node's action or task flow.

Alternative considered: `Tasker.PostRecognition`. This is useful for direct algorithms but does not itself express “load this named Pipeline node,” making node references and exact production node configuration harder to preserve.

Alternative considered: post the target node as a task and override its action and next list. This still subjects the test to Pipeline retry, timeout, delay, and wait behavior, and makes a recognition miss less direct to diagnose.

### Use the production resource and Agent protocol

The command will load the complete resource directory using `Resource.PostBundle`. In the default `auto` Agent mode, it will start the production Go Agent with a unique identifier when the default or explicitly configured executable exists, bind the Agent client to the same Resource, and confirm registration before executing the harness. Built-in recognition can still run when no Agent executable is available; `--no-agent` provides an explicit built-in-only mode. A node that requires an unavailable Custom Recognition fails as setup, not as a miss. Defaults will target the repository development layout, with explicit flags available for built/install layouts.

The Agent lifecycle wrapper will own child-process startup, connection timeout, disconnect, and cleanup. Setup failures will never be converted into recognition misses.

Alternative considered: duplicate all Custom Recognition registration inside the CLI process. This would create a second registration list and could diverge from production Agent behavior.

### Model results independently from console rendering

Execution will return a typed result containing schema version, node, image, expectation, actual hit, algorithm, box, detail, combined children, duration, assertion outcome, and artifact paths. Text and JSON renderers consume this model. Expected misses are successful assertions rather than runtime errors.

Diagnostics go to stderr when JSON mode is active, preserving stdout as one parseable JSON document.

### Validate screenshot dimensions rather than normalize them

MDA coordinates and templates use a 1280 by 720 baseline. The MVP rejects other dimensions and reports both sizes. Silent resize could make a fixture pass while hiding an invalid capture workflow.

### Reserve exit statuses by failure class

- `0`: recognition completed and matched the expected hit/miss assertion.
- `1`: recognition completed but the assertion failed.
- `2`: arguments, setup, resource, Agent, image, or MaaFramework execution prevented a valid assertion.

This split lets CI distinguish a product recognition regression from broken test infrastructure.

### Keep fixtures as ordinary image files in the MVP

The node command accepts command-line assertions and image paths. A scenario manifest is deferred until Pipeline playback needs ordered screens and actions. Repository test code can still table-drive multiple command configurations without defining a premature public fixture schema.

### Add one canonical project skill

The dynamic testing skill will be added at `.codex\skills\pipeline-testing` and will reference the CLI's actual help and result contract. It will include positive and negative fixture practices, Custom Recognition setup, evidence requirements, and escalation to integration/live tests. The existing `.claude\skills\pipeline-debug` remains a static structure-analysis skill; the new skill will document this boundary rather than duplicating its schema and graph checks. A future Claude-compatible mirror can be added through a separate synchronization decision.

## Future Extension Boundary

- `pipeline`: add an action-aware fixture controller that holds a screen stable across retries and advances only after an expected action; assert node/action traces using TaskDetail.
- `task`: add a Project Interface resolver for imports, selected task entry, options, presets, and ordered pipeline overrides, then delegate execution to the Pipeline test engine.
- Both commands reuse MaaFramework initialization, Resource/Agent lifecycle, result rendering, diagnostics, and exit status classification from the node MVP.

## Risks / Trade-offs

- [The harness requires a Context even though only recognition is tested] -> Keep the temporary task private, DirectHit, action-only, and covered by tests that assert the target action and next list never execute.
- [Agent process startup can be slow or fail in CI] -> Use bounded startup timeouts, stage-specific errors, explicit executable flags, and a reusable session for table-driven tests where practical.
- [Go binding and installed MaaFramework DLL versions can drift] -> Report both versions in verbose diagnostics and test against the repository's installed development runtime.
- [Debug draw images are not always available] -> Enable MaaFramework debug output when artifacts are requested, treat returned draw images as optional, and never require them for assertion correctness.
- [OCR or image recognition can vary across runtime/model updates] -> Record runtime/resource context in JSON output and use representative positive and negative fixtures.
- [A node hit does not prove workflow correctness] -> Make the limitation explicit in command help and the testing skill.

## Migration Plan

1. Add the reusable runtime, Agent lifecycle, harness, result, and renderer packages.
2. Add the `pipeline-test node` command and unit/integration tests.
3. Add a small curated fixture set, including built-in and Custom Recognition coverage.
4. Add and validate the Pipeline testing skill.
5. Add a Windows CI job after runtime dependencies and fixtures are stable.

The change is additive. Rollback consists of removing the command, fixtures, skill, and CI job; production resources and tasks remain unchanged.
