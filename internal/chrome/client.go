package chrome

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Tab represents an open tab in Google Chrome.
type Tab struct {
	ID        string  // Chrome tab ID
	WindowID  string  // Chrome window ID
	TabIndex  int     // 1-based index within window
	Title     string  // Title reported by Chrome
	URL       string  // URL
	PID       int     // Renderer Process ID (if correlated)
	CPU       float64 // CPU percentage from Task Manager / ps
	MemoryMB  float64 // Memory footprint in MB
	MemoryStr string  // Human-readable memory string (e.g. "636 MB")
	Energy    float64 // macOS Energy Impact (POWER metric from top / Activity Monitor)
	EnergyStr string  // Human-readable energy string (e.g. "12.4" or "-")
	IsActive  bool    // Whether it was reported active in its window
}

// Client manages interaction with Google Chrome via AppleScript and macOS APIs.
type Client struct {
	mu sync.Mutex
}

// NewClient initializes a Chrome management client.
func NewClient() *Client {
	return &Client{}
}

// FetchAllTabs returns all open tabs across all Chrome windows.
func (c *Client) FetchAllTabs(ctx context.Context) ([]Tab, error) {
	script := `
set out to ""
tell application "Google Chrome"
	if not running then
		return ""
	end if
	repeat with w in windows
		set wId to id of w
		set tIdx to 0
		repeat with t in tabs of w
			set tIdx to tIdx + 1
			set out to out & (wId as text) & "\t" & (tIdx as text) & "\t" & (id of t as text) & "\t" & (title of t as text) & "\t" & (URL of t as text) & linefeed
		end repeat
	end repeat
end tell
return out
`
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tabs via AppleScript: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var tabs []Tab
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 5 {
			continue
		}
		wID := parts[0]
		tIdx, _ := strconv.Atoi(parts[1])
		tID := parts[2]
		title := parts[3]
		url := parts[4]

		tabs = append(tabs, Tab{
			ID:        tID,
			WindowID:  wID,
			TabIndex:  tIdx,
			Title:     title,
			URL:       url,
			MemoryStr: "-",
			EnergyStr: "-",
		})
	}

	return tabs, nil
}

// TaskManagerEntry holds resource metrics parsed from Chrome's Task Manager window.
type TaskManagerEntry struct {
	Title     string
	MemoryStr string
	MemoryMB  float64
	CPU       float64
	PID       int
}

// FetchTaskManagerMetrics inspects Chrome's Task Manager accessibility table.
// If the Task Manager window is not open, it attempts to open it via Shift+Esc.
func (c *Client) FetchTaskManagerMetrics(ctx context.Context) ([]TaskManagerEntry, error) {
	script := `
tell application "System Events"
	tell process "Google Chrome"
		set hasTM to false
		repeat with w in windows
			if name of w is "Task manager" then
				set hasTM to true
				exit repeat
			end if
		end repeat

		if not hasTM then
			set frontmost to true
			keystroke (ASCII character 27) using {shift down}
			delay 0.5
		end if

		repeat with w in windows
			if name of w is "Task manager" then
				set theTable to table 1 of group 1 of scroll area 1 of group 1 of group 1 of group 1 of group 1 of group 1 of w
				set out to ""
				repeat with el in (UI elements of theTable)
					set d to (description of el as text)
					set out to out & d & linefeed & "---DELIM---" & linefeed
				end repeat
				return out
			end if
		end repeat
		return ""
	end tell
end tell
`
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch task manager metrics: %w", err)
	}

	return parseTaskManagerOutput(string(output)), nil
}

// FetchProcessEnergy samples macOS process power metrics using top.
// It returns a mapping of PID -> Energy Impact score.
func (c *Client) FetchProcessEnergy(ctx context.Context) (map[int]float64, error) {
	cmd := exec.CommandContext(ctx, "top", "-l", "2", "-s", "1", "-stats", "pid,power")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch energy metrics via top: %w", err)
	}

	return ParseTopPowerOutput(string(output)), nil
}

// ParseTopPowerOutput parses the output from `top -stats pid,power`.
// It targets the last iteration block to capture delta power consumption.
func ParseTopPowerOutput(raw string) map[int]float64 {
	results := make(map[int]float64)
	lines := strings.Split(raw, "\n")

	// Find the start index of the last "PID" / "POWER" header block
	lastHeaderIdx := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "PID") && strings.Contains(trimmed, "POWER") {
			lastHeaderIdx = i
		}
	}

	if lastHeaderIdx == -1 {
		return results
	}

	for _, line := range lines[lastHeaderIdx+1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "Processes:") || (strings.HasPrefix(trimmed, "PID") && strings.Contains(trimmed, "POWER")) {
			break
		}

		parts := strings.Fields(trimmed)
		if len(parts) < 2 {
			continue
		}

		pid, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		power, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}

		results[pid] = power
	}

	return results
}

var (
	memRegex      = regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)\s*(MB|GB|KB|B)$`)
	cpuRegex      = regexp.MustCompile(`^\d+\.\d+$`)
	combinedRegex = regexp.MustCompile(`(?i)^(?:Tab|App):\s*(.*?),\s*(\d+(?:\.\d+)?\s*(?:MB|GB|KB|B)),\s*(\d+\.\d+),\s*\d+\s+and\s+(\d+)`)
)

func parseTaskManagerOutput(raw string) []TaskManagerEntry {
	items := strings.Split(raw, "---DELIM---")
	var cleanItems []string
	for _, it := range items {
		it = strings.TrimSpace(it)
		if it != "" {
			cleanItems = append(cleanItems, it)
		}
	}

	var results []TaskManagerEntry

	for idx, item := range cleanItems {
		m := combinedRegex.FindStringSubmatch(item)
		if len(m) == 5 {
			cleanTitle := strings.TrimSpace(m[1])
			memStr := strings.TrimSpace(m[2])
			cpuVal, _ := strconv.ParseFloat(m[3], 64)
			pidVal, _ := strconv.Atoi(m[4])

			results = append(results, TaskManagerEntry{
				Title:     cleanTitle,
				MemoryStr: memStr,
				MemoryMB:  parseMemoryToMB(memStr),
				CPU:       cpuVal,
				PID:       pidVal,
			})
			continue
		}

		if strings.HasPrefix(item, "Tab: ") || strings.HasPrefix(item, "App: ") {
			cleanTitle := strings.TrimSpace(item[5:])
			memStr := "-"
			var cpuVal float64
			var pidVal int

			maxLook := idx + 7
			if maxLook > len(cleanItems) {
				maxLook = len(cleanItems)
			}

			for _, fwd := range cleanItems[idx+1 : maxLook] {
				if strings.HasPrefix(fwd, "Tab: ") || strings.HasPrefix(fwd, "App: ") ||
					strings.HasPrefix(fwd, "Subframe: ") || strings.HasPrefix(fwd, "Extension: ") ||
					strings.HasPrefix(fwd, "Tool: ") || strings.HasPrefix(fwd, "Browser") ||
					strings.HasPrefix(fwd, "GPU Process") {
					break
				}
				if memRegex.MatchString(fwd) && memStr == "-" {
					memStr = fwd
				} else if cpuRegex.MatchString(fwd) && cpuVal == 0.0 {
					cpuVal, _ = strconv.ParseFloat(fwd, 64)
				} else if isDigits(fwd) {
					p, _ := strconv.Atoi(fwd)
					if p > 100 && pidVal == 0 {
						pidVal = p
					}
				}
			}

			results = append(results, TaskManagerEntry{
				Title:     cleanTitle,
				MemoryStr: memStr,
				MemoryMB:  parseMemoryToMB(memStr),
				CPU:       cpuVal,
				PID:       pidVal,
			})
		}
	}

	return results
}

func parseMemoryToMB(memStr string) float64 {
	m := memRegex.FindStringSubmatch(strings.TrimSpace(memStr))
	if len(m) < 3 {
		return 0
	}
	val, _ := strconv.ParseFloat(m[1], 64)
	unit := strings.ToUpper(m[2])
	switch unit {
	case "GB":
		return val * 1024
	case "MB":
		return val
	case "KB":
		return val / 1024
	case "B":
		return val / (1024 * 1024)
	default:
		return val
	}
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// CorrelateTabs joins tab metadata from Chrome with Task Manager metrics and OS energy metrics.
func CorrelateTabs(tabs []Tab, metrics []TaskManagerEntry, energyMap map[int]float64) []Tab {
	matchedMetric := make(map[int]bool)

	// First pass: exact title match
	for i := range tabs {
		t := &tabs[i]
		for j, m := range metrics {
			if matchedMetric[j] {
				continue
			}
			if strings.TrimSpace(t.Title) == strings.TrimSpace(m.Title) {
				t.PID = m.PID
				t.CPU = m.CPU
				t.MemoryMB = m.MemoryMB
				t.MemoryStr = m.MemoryStr
				matchedMetric[j] = true
				break
			}
		}
	}

	// Second pass: prefix or substring match for truncated titles
	for i := range tabs {
		t := &tabs[i]
		if t.MemoryMB > 0 || t.PID > 0 {
			continue
		}
		for j, m := range metrics {
			if matchedMetric[j] {
				continue
			}
			tTitle := strings.ToLower(strings.TrimSpace(t.Title))
			mTitle := strings.ToLower(strings.TrimSpace(m.Title))
			if len(mTitle) > 10 && (strings.HasPrefix(tTitle, mTitle) || strings.HasPrefix(mTitle, tTitle)) {
				t.PID = m.PID
				t.CPU = m.CPU
				t.MemoryMB = m.MemoryMB
				t.MemoryStr = m.MemoryStr
				matchedMetric[j] = true
				break
			}
		}
	}

	// Third pass: correlate energy impact metrics by PID
	for i := range tabs {
		t := &tabs[i]
		if t.PID > 0 && energyMap != nil {
			if p, ok := energyMap[t.PID]; ok {
				t.Energy = p
				t.EnergyStr = fmt.Sprintf("%.1f", p)
			} else {
				t.Energy = 0.0
				t.EnergyStr = "0.0"
			}
		} else {
			t.Energy = 0.0
			t.EnergyStr = "-"
		}
	}

	return tabs
}

// FocusTab activates Google Chrome, raises the specific window, and sets the active tab index.
func (c *Client) FocusTab(ctx context.Context, tab Tab) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	script := fmt.Sprintf(`
set found to false
tell application "Google Chrome"
	repeat with w in windows
		set tIdx to 0
		repeat with t in tabs of w
			set tIdx to tIdx + 1
			if (id of t as text) is "%s" then
				tell w
					set active tab index to tIdx
					set index to 1
				end tell
				activate
				set found to true
				exit repeat
			end if
		end repeat
		if found then exit repeat
	end repeat
end tell

if found then
	tell application "System Events"
		tell process "Google Chrome"
			set frontmost to true
			try
				perform action "AXRaise" of window 1
			end try
		end tell
	end tell
	return "success"
else
	return "not_found"
end if
`, tab.ID)

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to focus tab: %w", err)
	}
	res := strings.TrimSpace(string(out))
	if res == "not_found" {
		return fmt.Errorf("tab id %s not found in Chrome", tab.ID)
	}
	return nil
}

// CloseTab closes the specified tab.
func (c *Client) CloseTab(ctx context.Context, tab Tab) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	script := fmt.Sprintf(`
tell application "Google Chrome"
	repeat with w in windows
		repeat with t in tabs of w
			if (id of t as text) is "%s" then
				close t
				return "closed"
			end if
		end repeat
	end repeat
	return "not_found"
end tell
`, tab.ID)

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to close tab: %w", err)
	}
	res := strings.TrimSpace(string(out))
	if res == "not_found" {
		return fmt.Errorf("tab id %s not found in Chrome", tab.ID)
	}
	return nil
}
