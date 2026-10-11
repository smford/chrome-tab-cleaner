# chrome-tab-cleaner

A high-performance terminal user interface (TUI) for macOS to inspect Google Chrome tabs sorted by Memory and CPU usage, and instantly bring selected tabs to the front and focus them.

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

## Building & Running

```bash
# Build binary
go build -o bin/chrome-tab-cleaner ./cmd/chrome-tab-cleaner

# Run
./bin/chrome-tab-cleaner
```
