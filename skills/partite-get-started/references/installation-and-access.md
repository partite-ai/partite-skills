# Install the CLI and connect

## Installation

First check the current installation:

```sh
command -v partite
partite --version
partite --help
```

If a working CLI is present, keep it unless an update is needed. If it is absent,
check `go version`. The CLI source used for these examples declares Go 1.26.5.
Use a compatible Go toolchain. If Go is absent, use the user's normal toolchain
installation method; do not invent a Homebrew package or a binary download URL.

Install the released CLI:

```sh
go install github.com/partite-ai/partite-cli/cmd/partite@latest
```

For repeatable environments, use the team's chosen release tag in place of
`latest`. Report the version that was installed. A version of `dev` is possible
for local builds; also check `go version -m` on the binary for module details.

Go places the executable in `GOBIN` when set, or the first `GOPATH` entry's `bin`
directory otherwise. Inspect `go env GOBIN GOPATH`. Add that directory to the
current process PATH. For example, when GOBIN is empty and GOPATH has one entry:

```sh
export PATH="$(go env GOPATH)/bin:$PATH"
partite --version
partite --help
```

Use the actual directory for multiple GOPATH entries. On Windows, use
`Get-Command partite`, the Go binary directory, and the PowerShell PATH syntax.
Explain any persistent PATH change before editing a shell profile. Installation
must finish before resource commands are attempted.

If the user explicitly wants the local CLI source and has its checkout, run
`GOWORK=off go build -o <chosen-bin-directory>/partite ./cmd/partite` from that
checkout. Do not assume a sibling repository exists in a customer's environment.
For install failures, report the actual toolchain, network, or permission error.

## Server and authentication

```sh
partite context list -o json
partite context show -o json
```

Inspect whether `PARTITE_CONTEXT`, `PARTITE_URL`, `PARTITE_API_TOKEN`, or
`PARTITE_TOKEN` is set, without printing token values. `PARTITE_URL` overrides
the context URL. `PARTITE_API_TOKEN` takes precedence over stored credentials.
Resolve unintended overrides before changing any server resources.

- Cloud: the built-in `cloud` context points at `https://api.partite.ai`.
  Use `partite --context cloud login`, or `login --no-browser` for a remote
  terminal. Let the user complete browser authorization. If needed, pass the
  user's organization with `login --org ORG`.
- Local: the built-in `local` context points at `http://localhost:8080` without
  authentication. Use it only when the user's Mesh server is running there.
  This skill does not deploy a new Mesh server.
- Another server: create a context with
  `partite context set NAME --url URL`. Use an API key through
  `PARTITE_API_TOKEN`, or feed it to
  `partite --context NAME login --with-api-key` through standard input.

Use `partite --context NAME whoami -o json` when authenticated identity is
available. Check server access with
`partite --context NAME model-provider list -o json`; local no-auth access
does not require a cloud identity. CLI credentials and model provider
credentials serve different purposes.

Cloud login stores credentials in the OS keychain. `PARTITE_NO_KEYRING=1`
selects the CLI credential file instead. Do not enable it just to bypass a login
error. Setup may need configuration read/write access, provider credential use,
mesh conversation access, and trace access. Report a denied operation; do not
automatically broaden a key's permissions.
