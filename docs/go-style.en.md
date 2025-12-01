# Go Project Development Guide (English Version)

> This document is intentionally lean. It captures the universally applicable guidance that every Go change in this repo must follow. Reach for the referenced docs when you need deeper details.

## Why this doc exists
- Provide a single source of truth for style/process rules that apply to every Go edit.
- Keep `CLAUDE.md`/`agent.md` short by pointing here for language-specific expectations.
- Remind contributors that deterministic tooling and existing patterns—not ad‑hoc prompts—define quality.

## Non‑negotiable rules
1. **English-only artifacts** – All identifiers, comments, docs, log/error messages, and tests MUST be written entirely in English. (Human-facing replies in the CLI can be Chinese; nothing in the codebase should be.)
2. **Chinese responses to users** – When chatting with the user, answer in Chinese, but still write English code artifacts.
3. **Always run `just check` before handoff** – It runs format, lint, `go vet`, and tests with the repo’s canonical configuration.
4. **Prefer deterministic tools** – Let `gofmt`, `goimports`, `golangci-lint`, and the Go compiler enforce style instead of adding subjective guidance.

## Getting oriented
- Start with `agent.md` for repo-wide context, `README.md` for SDK usage/CLI examples, and `Justfile` for the full task list.
- If you need Chinese-language guidance, see the earlier history of this repo or translate this file—there is no longer a maintained `zh` version.
- For service-specific behavior, read the package README or nearby markdown files under `docs/` or `agent_docs/`. Do **not** bloat this file with feature instructions—add/link a focused doc instead.

## Style guardrails (apply everywhere)

### Naming & packages
- Use MixedCaps/mixedCaps (no underscores). Package names are short, lowercase, singular nouns (`customer`, `transport`).
- Avoid repeating package names in exported identifiers (`customer.Service`, not `customer.CustomerService`).
- Keep receiver names short and consistent (`func (c *Client)`).

### Imports
- Group imports: stdlib ➜ external ➜ module-internal/path aliases ➜ side-effect imports.
- Never use dot imports outside very narrow test helpers. Blank imports only for side effects.

### Errors & logging
- Errors are last return values; never signal failure via sentinel values alone.
- Error strings start lowercase, no trailing punctuation, and wrap with `%w` when callers must inspect.
- Either log or return—never both. Let the highest-level caller decide how to surface errors.

### Functions & APIs
- Prefer synchronous functions; callers add concurrency if needed.
- Keep method receivers consistent (all pointer or all value) unless copying semantics require both.
- Expose concrete types from constructors; define interfaces on the consumer side for seams/mocks.
- Accept `context.Context` as the first parameter for every request/IO-facing API.

### Comments & docs
- Every exported identifier requires a doc comment that starts with the identifier name.
- Comments explain “why” or external contracts, not obvious “what”.
- Package comments live directly above the `package` clause.

### Testing & concurrency
- Table-driven tests with descriptive `name` fields are the default. Use `t.Fatal` for setup failures, `t.Error` for assertion failures.
- Prefer `cmp.Diff` (or `reflect.DeepEqual` when sufficient) over ad-hoc comparisons.
- When using goroutines, make their lifecycle explicit (`WaitGroup`, context cancellation). Specify channel direction (`chan<-`, `<-chan`) in function signatures.

## Toolchain quick reference

```bash
# All-in-one quality gate
just check

# Focused commands when needed
just fmt           # gofmt + goimports
just lint          # golangci-lint
just test          # go test -race -cover ./...
just build         # go build ./...
```

If you touch dependencies, run `go mod tidy && go mod verify`. For new code generation, prefer `just generate` or the service-specific helpers already defined in the `Justfile`.

## Language compliance checklist
- Identifiers, comments, docs, test names, log lines, and error messages are English.
- String literals shown to end users (CLI output, examples) may be localized, but code artifacts stay English.
- When you need bilingual explanations, keep Chinese text outside code blocks (e.g., in commit messages or user replies).

## Progressive disclosure
- When work requires domain specifics (API contracts, data models, rollout notes), capture them in a dedicated markdown file (e.g., `agent_docs/transfers.md`) and reference it from `agent.md` or the relevant package README.
- Keep this document focused on timeless Go style/process rules to avoid instruction bloat that models might ignore.

## Further reading
- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)

**Reminder:** Code is for humans first. Favor clarity, deterministic tooling, and existing patterns over cleverness.
