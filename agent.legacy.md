# 1Money Go SDK – AGENT.md

This file is injected into every session: keep it lean, universally applicable, and focused on onboarding.

## Principles
- LLM sessions are stateless—agents forget everything between runs unless we restate it here.
- Less is more: only include instructions that apply to almost every task, and point to other docs for the rest.
- Prefer references to authoritative files instead of duplicating snippets; stale context is worse than no context.
- Deterministic tooling beats heuristics: run formatters/linters/tests instead of encoding style advice.

## Why this repository exists
- Public Go SDK for the 1Money API. It ships a typed client (`pkg/onemoney`) plus feature-specific service packages (`pkg/service/...`) that wrap HTTP transport, signing, and shared response handling.
- A small CLI in `cmd/` consumes the SDK the same way users do, so changes must keep both library and CLI surfaces consistent.
- Reliability matters because the SDK is how customers automate money movement; treat auth, transport, and error handling changes with extra care.

## What to know about the codebase
- **Runtime**: Go 1.25.2 (`go.mod`). Formatting/linting should match gofmt/goimports + `golangci-lint`.
- **Core flow**: `pkg/onemoney/client.go` wires credentials → auth signer → `internal/transport` HTTP client → `pkg/service/service.go` base helpers → concrete services (e.g., `pkg/service/customer`, `pkg/service/echo`, `pkg/service/conversions`, ...).
- **Internal packages**: `internal/auth` (HMAC signer), `internal/credentials` (env/file/profile chain), `internal/transport` (request/response types, API errors), `internal/utils` (shared helpers). Do not export from `internal`.
- **Service layout**: Each service folder defines a `Service` interface, an implementation embedding `service.BaseService`, request/response DTOs, and tests. New services should follow this pattern (see comments in `pkg/service/service.go`).
- **Existing docs**: `README.md` for quickstart + CLI usage, `docs/go-style.en.md` for detailed Go style expectations, `Justfile` for reproducible tasks. Reference these instead of duplicating content here.

## How to work here
1. **Start with context**: identify the relevant service directory (or `internal/*` module), skim existing request/response types and tests, and only open adjacent docs when needed.
2. **Follow Go conventions**: keep packages idiomatic, exported APIs documented, and run formatters/linters instead of encoding style rules here. If deeper guidance is needed, read `docs/go-style.en.md`.
3. **Prefer deterministic tooling**:
   - Format + imports: `just fmt`
   - Lint/static checks: `just lint` or `just check` (fmt-check + lint + `go vet`)
   - Unit tests: `just test`
   - Integration tests: set `ONEMONEY_ACCESS_KEY`, `ONEMONEY_SECRET_KEY`, `ONEMONEY_BASE_URL`, then run `INTEGRATION_TEST=true just test-integration`
   - Build CLI: `just build-cli` (injects version metadata), or `just build` for the whole module.
   - Discover more commands: `just --list`
4. **Credentials & configuration**: the SDK loads credentials in this order—explicit config ➜ environment variables ➜ `~/.onemoney/credentials` profiles. When tests need live calls, use the same chain and never hard-code secrets.
5. **Adding or changing services**:
   - Embed `service.BaseService` and use the generic helpers (`GetJSON`, `PostJSON`, etc.) so transport/auth logic stays centralized.
   - Register new service instances in `pkg/onemoney/client.go` and expose them on the `Client` struct.
   - Add table-driven tests that rely on fake transport/mocks instead of real HTTP.
6. **Progressive disclosure**: if a task needs domain-specific instructions (API contracts, schema notes, etc.), place that info in a dedicated markdown file under `docs/` or `agent_docs/` and reference it here only by name. Ask for approval before loading large auxiliary docs.
7. **Verification before handoff**: rerun the smallest `just` target that covers your changes (`just test` for service logic, `just check` when touching multiple packages, `just build-cli` if CLI code changed) and summarize what ran in your final message.

Remember: this file should stay short. If you find yourself wanting to paste long instructions, add a purpose-built doc and point to it from the “Existing docs” or “Progressive disclosure” notes above.
