# Partite skills

Developer skills for Partite Mesh and Particles. Install the collection
from GitHub or download a release bundle. No public plugin store is required.

| Skill | Purpose |
|---|---|
| `partite-get-started` | Install the CLI, connect, and run a first mesh |
| `partite-concepts` | Explain Mesh concepts without server access |
| `partite-build-mesh` | Create or extend agents, intents, links, and tools |
| `partite-integrate-app` | Connect application code to the Conversation API |
| `partite-debug-run` | Inspect conversations, requests, events, and traces |
| `partite-evaluate-mesh` | Configure and inspect Enterprise evaluations |
| `write-particle-python` | Create or extend Particle tools in Python |
| `write-particle-typescript` | Create or extend Particle tools in TypeScript or JavaScript |

The `skills/` directory is the maintained source. Packaging does not rewrite 
the instructions. References stay inside each skill, so each folder can also be 
installed on its own.

The Particle authoring skills are adapted from `particles/skills/`. Their runtime
guidance is retained, with portable descriptions, license notices, and OpenAI
display metadata. The Go embedding file `skills.go` belongs to the Particles
runtime and is not part of this skill collection.

## Install from GitHub

These commands assume this project is published as
`partite-ai/partite-skills`. The local folder alone does not create that GitHub
repository. Change the repository address if it is published elsewhere.

### Claude Code

In a Claude Code session:

```text
/plugin marketplace add partite-ai/partite-skills
/plugin install partite-ai@partite-skills
```

Start a new session if the client requires it. Try
`/partite-ai:partite-get-started`, or ask Partite to help you build a first mesh.
For a local checkout or an extracted Anthropic bundle, add its absolute directory
as the marketplace source instead of the GitHub repository name. The bundle
contains `.claude-plugin/marketplace.json` and `.claude-plugin/plugin.json`.
You can also try it for one session with `claude --plugin-dir /path/to/package`.

### Codex and ChatGPT desktop

Register the repository marketplace:

```sh
codex plugin marketplace add partite-ai/partite-skills
```

In clients that support the plugin install command:

```sh
codex plugin add partite-ai@partite-skills
```

Otherwise, open the ChatGPT desktop Plugins Directory, select the Partite Skills
marketplace, and install Partite. Restart the client if needed. The OpenAI
manifest selects `partite-get-started` as the onboarding skill.

For an extracted OpenAI bundle or a local checkout, register its absolute path
with `codex plugin marketplace add /path/to/package`. The package includes the
`.agents/plugins/marketplace.json` catalog. Pin a Git tag when needed with
`codex plugin marketplace add partite-ai/partite-skills --ref v0.1.1`.

ChatGPT workspace admins can instead import the GitHub marketplace through
Admin > Plugins. Workspace imports follow workspace access policies. This is
different from installing a local plugin in Codex or the desktop app. CLI-based
workflows need a host with terminal execution, network access, and the Partite
CLI; installing this package does not supply those capabilities to every chat.

### Pi

```sh
pi install git:github.com/partite-ai/partite-skills@v0.1.1
```

Use a published tag, or omit `@v0.1.1` to follow the repository's default branch.
The package declares only `skills/`; it adds no extensions or install hooks.
For an extracted standard bundle, use `pi install /path/to/package`.

### Other Agent Skills hosts

Download the `agent-skills` ZIP, extract it, and copy the individual skill
folders from `skills/` into the host's configured skill directory. Preserve each
folder's `SKILL.md` and `references/`. There is no universal bundle installer.
For hosts that support them, `.agents/skills/` or `~/.agents/skills/` can be used.

## Downloadable bundles

Each version produces these files under `dist/`:

| File | Contents |
|---|---|
| `partite-skills-anthropic-VERSION.zip` | Claude plugin and marketplace manifests, all portable skills |
| `partite-skills-openai-VERSION.zip` | Agent Plugins manifest, Codex compatibility manifest, marketplace, all skills and OpenAI skill metadata |
| `partite-skills-agent-skills-VERSION.zip` | All standard skill folders and Pi package metadata |
| `manifest.json` | Version, skill names, archive filenames, sizes, and SHA-256 hashes |
| `SHA256SUMS` | Checksums for all three ZIPs and the release manifest |

Manifests and `skills/` are at each ZIP's root. Extract into a dedicated directory;
do not unpack a new version over an existing installation. The Anthropic and
standard bundles omit `agents/openai.yaml`, which is OpenAI-specific UI metadata.
No bundle includes repository build scripts, CI configuration, or credentials.

Download from GitHub Releases after a maintainer publishes a release. Verify the
files with `sha256sum --check SHA256SUMS` on Linux, or `shasum -a 256 -c SHA256SUMS`
on macOS. The checksums detect damaged or mismatched downloads; obtain them from
the same trusted release. A release ZIP is not a universal one-click installer.

## Build and validate

Use Go 1.26.5 or later:

```sh
go run ./cmd/bundle --check
go test ./...
go run ./cmd/bundle
```

Run these commands from the repository root. If this checkout is inside a larger
Go workspace that does not include it, set `GOWORK=off` for these commands. The
GitHub Action already sets that value. Go downloads the YAML parser declared in
`go.mod`; the archive, checksum, and test code use the standard library.

Go is needed only to build and validate the bundles. Installing or loading the
skills does not require the build tooling. The getting-started skill guides
Partite CLI installation; the Mesh workflow skills require the `partite` CLI
when used. Particle authoring uses the separate `particle` CLI to build and run
tools. Installing the skill package does not install that CLI. The concepts
skill needs no CLI or credentials.

Use `--output /path/to/directory` to choose an output directory, and
`--version 0.1.1` to require a matching package version before building. The
`--root` flag selects the package root when running a compiled bundler elsewhere.

The validator checks frontmatter, package names and versions, bundled reference
paths, YAML/JSON examples, and host manifests. Tests extract every ZIP, check
resource completeness, verify checksums and repeatable builds, and reject broken
references, cross-skill dependencies, and symlinks. These checks do not prove
model behavior or live server access. Test actual skill use in supported clients
before calling a release verified for that client.

## Releases and updates

The GitHub Action validates, tests, builds, and uploads downloadable artifacts on
pushes, pull requests, and manual runs. A `vVERSION` tag also creates a **draft**
GitHub release with the three ZIPs, manifest, and checksums. A maintainer reviews
and publishes it. No action submits to a public plugin store or publishes to npm.

For a new version, update `package.json`, `plugin.json`,
`.claude-plugin/plugin.json`, and `.codex-plugin/plugin.json` together. All four
versions must match. Tag names must match that version with a `v` prefix. For
example, `v0.1.1` requires version `0.1.1`. Prerelease tags such as `v0.2.0-beta.1`
are supported. Existing release assets are never overwritten automatically.

Repository marketplace installs read repository content, not GitHub Release
ZIPs. Publishing a release does not update pinned installations. Use the host's
marketplace refresh/update flow; change a pinned Git ref explicitly when moving
to a new version. Keep skill content in this repository after the initial copy,
rather than editing separate copies for each host.

## Sources and license

- [Partite documentation](https://docs.partite.ai/)
- [Agent Skills specification](https://agentskills.io/specification)
- [Pi packages](https://pi.dev/docs/latest/packages)
- [Claude plugins](https://code.claude.com/docs/en/plugins)
- [OpenAI plugin packaging](https://developers.openai.com/plugins/build/plugins)

The Mesh skills and package tooling retain the existing
[FSL-1.1-ALv2 license](LICENSE.md). The imported Particle authoring skills retain
their source project's GPL-3.0-or-later license, included within each skill:

- [Python skill license](skills/write-particle-python/LICENSE)
- [TypeScript/JavaScript skill license](skills/write-particle-typescript/LICENSE)

The package-level license does not replace those per-skill licenses. Both are
included in every bundle. No new license grant is introduced by packaging.
