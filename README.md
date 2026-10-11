# chrome-tab-cleaner

[![CI](https://github.com/smford/chrome-tab-cleaner/actions/workflows/ci.yml/badge.svg)](https://github.com/smford/chrome-tab-cleaner/actions/workflows/ci.yml)
[![SAST](https://github.com/smford/chrome-tab-cleaner/actions/workflows/sast.yml/badge.svg)](https://github.com/smford/chrome-tab-cleaner/actions/workflows/sast.yml)
[![DAST](https://github.com/smford/chrome-tab-cleaner/actions/workflows/dast.yml/badge.svg)](https://github.com/smford/chrome-tab-cleaner/actions/workflows/dast.yml)
[![Release](https://img.shields.io/github/v/release/smford/chrome-tab-cleaner?color=blue&logo=github)](https://github.com/smford/chrome-tab-cleaner/releases/latest)
[![Homebrew](https://img.shields.io/badge/homebrew-smford%2Ftap-orange.svg?logo=homebrew)](https://github.com/smford/homebrew-tap)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance terminal user interface (TUI) for macOS to inspect Google Chrome tabs sorted by Memory and CPU usage, and instantly bring selected tabs to the front and focus them.

Visit the [GitHub Pages Site](https://smford.github.io/chrome-tab-cleaner/) for live documentation and interactive installation instructions.

---

## Visual Showcase

### 1. Memory Sort & Resource Overview
Inspect Chrome tabs across all profiles and windows, correlated with live Task Manager metrics:

![Memory Sort Overview](docs/images/01-tui-overview-memory.svg)

### 2. Live Search & Title / URL Filtering (`/`)
Instant interactive search filtering across tab titles and URLs as you type:

![Live Search and Filtering](docs/images/02-search-filter.svg)

### 3. CPU Intensive Tabs & Runaway Renderer Alerts (`c`)
Sort by CPU usage (`c`) to immediately spotlight background tabs draining battery or processor resources:

![CPU Sort and Alerts](docs/images/03-cpu-sort-alert.svg)

### 4. Direct Focus (`Enter`) & Instant Tab Close (`x`)
Switch directly to a tab and raise its window in macOS, or terminate memory-heavy tabs directly from the terminal:

![Instant Close Action](docs/images/04-close-action.svg)

---

## Installation

### Homebrew (Recommended)

```bash
brew install smford/tap/chrome-tab-cleaner
```

### Pre-built Binaries

Download pre-compiled hermetic binaries for macOS (Apple Silicon & Intel) and Linux from the [GitHub Releases](https://github.com/smford/chrome-tab-cleaner/releases/latest) page.

```bash
# macOS Apple Silicon
curl -sSL "https://github.com/smford/chrome-tab-cleaner/releases/download/v0.1.0/chrome-tab-cleaner_0.1.0_darwin_arm64.tar.gz" | tar -xz
sudo mv chrome-tab-cleaner /usr/local/bin/

# macOS Intel
curl -sSL "https://github.com/smford/chrome-tab-cleaner/releases/download/v0.1.0/chrome-tab-cleaner_0.1.0_darwin_amd64.tar.gz" | tar -xz
sudo mv chrome-tab-cleaner /usr/local/bin/
```

### Go Toolchain

```bash
go install github.com/smford/chrome-tab-cleaner/cmd/chrome-tab-cleaner@latest
```

## Features

- **Resource Monitoring**: Correlates Chrome tab metadata with live Chrome Task Manager metrics (Memory footprint, CPU %, Process ID).
- **Interactive TUI**:
  - `↑` / `↓` (or `k` / `j`): Navigate tabs.
  - `Enter`: Bring selected tab and its parent window to the front and focus it.
  - `m`: Sort by Memory (highest first).
  - `c`: Sort by CPU (highest first).
  - `t`: Sort alphabetically by Title.
  - `/`: Search / live filter by title or URL (press `Enter` or `Esc` to return to table).
  - `x`: Close the selected tab directly from the TUI.
  - `r`: Refresh tabs and resource metrics.
  - `q`: Quit.

## Architecture

1. **Tab Enumeration**: Discovers all open windows and tabs across profiles via macOS AppleScript in a single batch query (~0.8s for 400+ tabs).
2. **Metrics Ingestion**: Reads process metrics from Chrome's Task Manager via macOS Accessibility UI APIs.
3. **Tab Focus**: Switches directly to the target tab and raises its window using Chrome's native window/tab indexing API.

## Requirements

- macOS
- Google Chrome
- Terminal permissions: Ensure your Terminal / iTerm has Accessibility permissions in **System Settings > Privacy & Security > Accessibility** so it can interact with Chrome via Apple Events.

## Building & Running Locally

```bash
# Build binary
make build

# Run application
./bin/chrome-tab-cleaner

# Run tests and linter
make test
make lint
```
