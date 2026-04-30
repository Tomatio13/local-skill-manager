<h1 align="center">Local Skill Manager</h1>

<p align="center">A terminal UI for managing local Agent Skills with symbolic links.</p>

<p align="center">
  <a href="README_JP.md"><img src="https://img.shields.io/badge/README-日本語-white.svg" alt="Japanese README"></a>
  <a href="README.md"><img src="https://img.shields.io/badge/README-English-white.svg" alt="English README"></a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white" alt="Go 1.22+">
  <img src="https://img.shields.io/badge/TUI-Bubble%20Tea-1f6feb" alt="Bubble Tea">
  <img src="https://img.shields.io/badge/Deploy-Symlink-4c956c" alt="Symlink deployment">
  <img src="https://img.shields.io/badge/Platforms-Linux%20%7C%20macOS%20%7C%20Windows-555" alt="Linux macOS Windows">
</p>

<p align="center">
  <img src="./docs/screenshot-main.png" alt="Local Skill Manager screenshot" width="1100">
</p>

## ✨ Overview

Local Skill Manager reads skills from `store_path/skills` and enables or disables them for each configured agent root by creating or removing symbolic links in the resolved skill directory.

## 🚀 Features

- Browse local skills in a two-pane terminal UI
- Read `description:` from each `SKILL.md`
- Search by skill name or description
- Toggle each skill per target
- Use symbolic links instead of copying files
- Manage multiple agent roots from one config file

## 📋 Requirements

- Go `1.22+`
- A terminal with arrow key support and standard ANSI rendering
- A local skill store containing `skills/<skill-name>/SKILL.md`
- A filesystem that allows symbolic links

## 📁 Directory Layout

```text
<store_path>/
  skills/
    skill-a/
      SKILL.md
    skill-b/
      SKILL.md
```

When a skill is enabled, the app creates:

```text
<resolved skill directory>/<skill-name> -> <store_path>/skills/<skill-name>
```

## ⚙️ Configuration

Create `config.json` in the project root. If `--config` is omitted, the app reads `./config.json`.

```json
{
  "store_path": "~/.open_skills",
  "link_targets": [
    "~/.claude",
    "~/.agents",
    "~/.cursor",
    "~/.gemini",
    "~/.gemini/antigravity",
    "~/.copilot",
    "~/.config/opencode",
    "~/.codeium/windsurf"
  ]
}
```

### Fields

- `store_path`
  Root directory of the local skill store. The app scans `store_path/skills`.
- `link_targets`
  Agent root directories. The app resolves the actual skill directory for each target.

### Resolved Skill Directories

- `~/.claude` -> `~/.claude/skills`
- `~/.agents` -> `~/.agents/skills`
- `~/.config/opencode` -> `~/.config/opencode/skills`
- `~/.cursor` -> `~/.cursor/skills-cursor` if present, otherwise `~/.cursor/skills`
- any other target -> `<target>/skills`

## ▶️ Run

Build the binary:

```bash
go build -buildvcs=false ./cmd/local-skill-manager
```

Run with the default `config.json` in the current directory:

```bash
./local-skill-manager
```

Run with another config file:

```bash
./local-skill-manager --config ./config.example.json
```

If the binary is in your `PATH`, you can run:

```bash
local-skill-manager
```

Use `go run` only for development:

```bash
go run ./cmd/local-skill-manager
```

## 📦 Release Builds

Build all supported binaries locally:

```bash
make build-cross
```

Artifacts are generated in `dist/` as both raw binaries and zip archives:

- `local-skill-manager_linux_amd64`
- `local-skill-manager_linux_arm64`
- `local-skill-manager_darwin_amd64`
- `local-skill-manager_darwin_arm64`
- `local-skill-manager_windows_amd64.exe`
- `local-skill-manager_windows_arm64.exe`
- `local-skill-manager_linux_amd64.zip`
- `local-skill-manager_linux_arm64.zip`
- `local-skill-manager_darwin_amd64.zip`
- `local-skill-manager_darwin_arm64.zip`
- `local-skill-manager_windows_amd64.zip`
- `local-skill-manager_windows_arm64.zip`

You can upload the generated zip files in `dist/` to GitHub Releases manually.

## 🎮 Usage

### Basic Flow

1. Launch the app.
2. Pick a skill in the left pane.
3. Move to the right pane.
4. Toggle targets on or off.
5. Apply the pending changes.

### Key Bindings

- `↑ / ↓`
  Move selection
- `Tab` / `←` / `→`
  Switch pane
- `Enter`
  Move from the skill pane to the target pane, or toggle the selected target
- `Space`
  Toggle the selected target in the right pane
- `a`
  Apply pending changes
- `r`
  Reload skills from the store
- `q`
  Quit

### Search

- `/`
  Start search in the skill pane
- text input
  Filter by skill name and description
- `Enter`
  Confirm the current search
- `Esc`
  Exit search mode
- `Ctrl+u`
  Clear the current search input while typing
- `c`
  Clear the active filter

## 🔗 Behavior

- A skill is valid when its directory contains `SKILL.md`
- The app reads `description:` from `SKILL.md` frontmatter when present
- `ON` means the symbolic link should exist in the resolved skill directory
- `OFF` means the symbolic link should be removed

## 📝 Notes

- Source skill files are never copied
- Editing a source skill affects every target linked to it
- Disabling a target does not remove the source skill
- Enabling replaces an existing target entry with the same skill name
- On Windows, creating symbolic links may require Developer Mode or elevated privileges
- On Windows, launch the app from the extracted `local-skill-manager_windows_<arch>.zip` contents

## 🛠️ Troubleshooting

### No skills appear

- Check that `config.json` exists in the current directory, or that `--config` points to a valid file
- Check that `store_path/skills` exists
- Check that each skill directory contains `SKILL.md`

### A skill does not enable

- Check that the target directory is writable
- Check that the filesystem allows symbolic links
- Check that an existing file or directory at the destination is not protected
- On Windows, verify that Developer Mode is enabled or run the terminal with administrator privileges

### Search shows no results

- Check that the query matches the skill name or description
- Press `c` to clear the current filter

## 🧪 Development

Format:

```bash
gofmt -w ./cmd ./internal
```

Test:

```bash
GOCACHE=$PWD/.cache/go-build GOMODCACHE=$PWD/.cache/go-mod go test ./...
```

Build:

```bash
GOCACHE=$PWD/.cache/go-build GOMODCACHE=$PWD/.cache/go-mod go build -buildvcs=false ./cmd/local-skill-manager
```

Cross-build all release binaries:

```bash
make build-cross
```
