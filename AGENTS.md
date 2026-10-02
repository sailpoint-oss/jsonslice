# jsonslice

`jsonslice` is a Go library. `jsonslice.Get` returns a JSONPath slice of raw JSON bytes. The module path is `github.com/sailpoint-oss/jsonslice`.

Do not add Shipmate to this repository. Do not create `.shipmate/`, Shipmate skills, or a generated agent tree.

This repository is public. Do not put private or internal SailPoint details in files, commit messages, pull request text, or labels. That includes Jira project keys, internal ticket prefixes, and internal process labels. Dependabot uses the commit prefix `chore` and the label `dependencies` for that reason.

The license is MIT. The copyright notice lives only in `LICENSE`, and it stays `Copyright (c) 2018 bhmj`. Do not add a copyright line to a source file, workflow, or document. Do not add SailPoint as a copyright holder. Other forks in `sailpoint-oss` keep the upstream notice and leave new files unmarked.

## Commands

- `make test` runs `go test ./...`.
- `make test-short` skips fuzzy tests.
- `make lint` runs `golangci-lint` and `gocyclo`.
- `make build` writes the CLI to `./build/jsonslice`.
- The module declares Go `1.27`.
- GitHub Actions workflow `.github/workflows/ci.yml` runs `go test ./...` and `golangci-lint`.

## Change rules

- Keep `jsonslice.Get` behavior stable unless the task says otherwise.
- Keep the module path `github.com/sailpoint-oss/jsonslice`.
- Add a short comment on every new function, struct, and non-obvious block.
- New table-driven tests use a map. The map key is the test name.
- Tests that need a context use `context.TODO()`.
- Do not add a copyright line to a new file. The notice in `LICENSE` covers the repository.
- Do not commit or push unless the user asks.

## Maintenance plan

The Go version is `1.27`. CI is GitHub Actions.
Dependabot is `.github/dependabot.yml`. It checks Go modules and GitHub Actions every Monday.
It groups minor and patch updates. It allows 5 open pull requests for each ecosystem.
