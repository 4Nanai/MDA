## 1. Command Foundation

- [x] 1.1 Add `agent\go-service\cmd\pipeline-test` with a subcommand dispatcher, `node` help, and stable argument validation.
- [x] 1.2 Add typed node-command configuration with repository defaults and overrides for resource, MaaFramework library, Agent executable, expectation, output format, and artifact directory.
- [x] 1.3 Add tests for required arguments, invalid enum values, help output, and default path resolution from the repository root and install layout.

## 2. MaaFramework Runtime

- [x] 2.1 Implement MaaFramework initialization and cleanup against the configured library directory with stage-specific errors.
- [x] 2.2 Implement full Resource bundle loading, target-node existence validation, and BlankController/Tasker binding.
- [x] 2.3 Implement PNG/JPEG decoding and strict 1280 by 720 validation without image resizing.
- [x] 2.4 Implement the `auto`, explicit-path, and disabled Agent lifecycle modes with unique identifiers, bounded startup/connect timeouts, Resource binding, registration diagnostics, and child-process cleanup.
- [x] 2.5 Add runtime tests for missing libraries, invalid resources, unknown nodes, unreadable images, invalid dimensions, and unavailable Agent executables.

## 3. Recognition Harness

- [x] 3.1 Implement an in-process harness Custom Action that receives a MaaFramework Context and calls `Context.RunRecognition` for the requested node and validated image.
- [x] 3.2 Execute the harness through a private DirectHit task and transfer the recognition result or execution error back to the node command without executing target action, waits, delays, `next`, or `on_error`.
- [x] 3.3 Preserve algorithm, hit, box, detail JSON, and recursive combined-recognition details in a CLI-owned result model.
- [x] 3.4 Add integration fixtures and tests for built-in OCR/TemplateMatch, cross-node `And`/`Or`, expected miss, and a target node whose action would fail if accidentally executed.
- [x] 3.5 Add an Agent-backed integration test using an MDA Custom Recognition and verify unavailable registration is classified as setup failure rather than miss.

## 4. Assertions And Reporting

- [x] 4.1 Implement `hit` and `miss` expectations and map pass, assertion failure, and setup/execution failure to exit statuses 0, 1, and 2.
- [x] 4.2 Implement concise text output containing node, image, expectation, actual hit, algorithm, box, duration, outcome, and diagnostics.
- [x] 4.3 Implement versioned single-document JSON output and keep non-result diagnostics on stderr.
- [x] 4.4 Enable MaaFramework debug output when an artifact directory is requested, save available draw images, and return their paths without affecting assertions.
- [x] 4.5 Add golden or structural tests for text/JSON output, combined details, artifact paths, and all exit-status classes.

## 5. Pipeline Testing Skill

- [x] 5.1 Add `.codex\skills\pipeline-testing\SKILL.md` with precise triggers and a workflow based on the implemented CLI help and options.
- [x] 5.2 Document positive and negative fixture selection, Custom Recognition Agent setup, required evidence, and the 1280 by 720 screenshot rule.
- [x] 5.3 Document the boundary between schema/static checks, node recognition tests, future Pipeline/Task tests, and live controller validation, including the relationship to `.claude\skills\pipeline-debug`.
- [x] 5.4 Add troubleshooting for input, resource, Agent, OCR, template, and assertion failures without recommending unjustified threshold reductions.
- [x] 5.5 Validate every command example in the skill against the CLI and add a test or review check that detects stale flags and exit-status documentation.

## 6. Developer Integration

- [x] 6.1 Add developer documentation for building the Agent and CLI and running node tests from PowerShell 7 on Windows.
- [x] 6.2 Add a small repository fixture set with ownership and update guidance; exclude sensitive or unstable screenshots.
- [x] 6.3 Add a Windows CI job that builds required binaries and runs deterministic built-in and Custom node fixtures.
- [x] 6.4 Run Go formatting and tests, schema/static checks, CLI smoke tests, and `openspec validate add-pipeline-test-cli --strict`.
- [x] 6.5 Document the deferred `pipeline` action-aware FixtureController and PI `task` option-resolution roadmap without exposing non-functional subcommands.
