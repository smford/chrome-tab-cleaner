package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/smford/top-tab-cleaner/internal/chrome"
)

type SortMode int

const (
	SortMemory SortMode = iota
	SortCPU
	SortTitle
)

// App manages the terminal user interface.
type App struct {
	client     *chrome.Client
	app        *tview.Application
	table      *tview.Table
	filter     *tview.InputField
	statusBar  *tview.TextView
	headerBar  *tview.TextView
	flex       *tview.Flex

	tabs       []chrome.Tab
	filtered   []chrome.Tab
	sortMode   SortMode
	filterText string
	mu         sync.RWMutex
	statusMsg  string
	isLoading  bool
}

// NewApp initializes the TUI application.
func NewApp(client *chrome.Client) *App {
	a := &App{
		client:   client,
		app:      tview.NewApplication(),
		table:    tview.NewTable(),
		filter:   tview.NewInputField(),
		statusBar: tview.NewTextView(),
		headerBar: tview.NewTextView(),
		sortMode: SortMemory,
	}

	a.setupUI()
	return a
}

func (a *App) setupUI() {
	// Header bar styling
	a.headerBar.SetDynamicColors(true).
		SetText("[::b]CHROME TAB TOP CLEANER[::-] | [yellow]m[-]: Sort Mem | [yellow]c[-]: Sort CPU | [yellow]t[-]: Sort Title | [yellow]r[-]: Refresh | [yellow]Enter[-]: Focus | [yellow]x[-]: Close Tab | [yellow]q[-]: Quit").
		SetBackgroundColor(tcell.ColorDarkBlue)

	// Filter input field
	a.filter.SetLabel(" Filter: ").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite).
		SetChangedFunc(func(text string) {
			a.mu.Lock()
			a.filterText = strings.TrimSpace(text)
			a.mu.Unlock()
			a.applyFilterAndSort()
			a.renderTable()
		})

	// Table configuration
	a.table.SetBorders(false).
		SetSelectable(true, false).
		SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorWhite).Bold(true)).
		SetSelectedFunc(func(row, column int) {
			a.focusSelectedTab()
		})

	// Status bar styling
	a.statusBar.SetDynamicColors(true).
		SetBackgroundColor(tcell.ColorNavy)

	// Flex layout
	a.flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.headerBar, 1, 0, false).
		AddItem(a.filter, 1, 0, false).
		AddItem(a.table, 0, 1, true).
		AddItem(a.statusBar, 1, 0, false)

	// Key bindings
	a.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				a.app.Stop()
				return nil
			case 'r', 'R':
				go a.RefreshData()
				return nil
			case 'm', 'M':
				a.setSort(SortMemory)
				return nil
			case 'c', 'C':
				a.setSort(SortCPU)
				return nil
			case 't', 'T':
				a.setSort(SortTitle)
				return nil
			case '/':
				a.app.SetFocus(a.filter)
				return nil
			case 'x', 'X':
				a.closeSelectedTab()
				return nil
			}
		case tcell.KeyEnter:
			a.focusSelectedTab()
			return nil
		case tcell.KeyEsc:
			if a.filterText != "" {
				a.filter.SetText("")
				return nil
			}
		}
		return event
	})

	a.filter.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter, tcell.KeyDown, tcell.KeyEsc:
			a.app.SetFocus(a.table)
			return nil
		}
		return event
	})
}

func (a *App) setSort(mode SortMode) {
	a.mu.Lock()
	a.sortMode = mode
	a.mu.Unlock()
	a.applyFilterAndSort()
	a.renderTable()
}

// RefreshData fetches tabs and performance metrics, then updates the UI.
func (a *App) RefreshData() {
	a.mu.Lock()
	if a.isLoading {
		a.mu.Unlock()
		return
	}
	a.isLoading = true
	a.statusMsg = "Refreshing tabs & reading Chrome Task Manager..."
	a.mu.Unlock()
	a.updateStatusBar()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tabs, err := a.client.FetchAllTabs(ctx)
	if err != nil {
		a.mu.Lock()
		a.isLoading = false
		a.statusMsg = fmt.Sprintf("[red]Error fetching tabs: %v[-]", err)
		a.mu.Unlock()
		a.updateStatusBar()
		return
	}

	metrics, _ := a.client.FetchTaskManagerMetrics(ctx)
	correlated := chrome.CorrelateTabs(tabs, metrics)

	a.mu.Lock()
	a.tabs = correlated
	a.isLoading = false
	a.statusMsg = fmt.Sprintf("Loaded %d tabs (%d metrics matched) at %s", len(correlated), len(metrics), time.Now().Format("15:04:05"))
	a.mu.Unlock()

	a.applyFilterAndSort()
	a.app.QueueUpdateDraw(func() {
		a.renderTable()
		a.updateStatusBar()
	})
}

func (a *App) applyFilterAndSort() {
	a.mu.Lock()
	defer a.mu.Unlock()

	var filtered []chrome.Tab
	q := strings.ToLower(a.filterText)

	for _, t := range a.tabs {
		if q == "" || strings.Contains(strings.ToLower(t.Title), q) || strings.Contains(strings.ToLower(t.URL), q) {
			filtered = append(filtered, t)
		}
	}

	switch a.sortMode {
	case SortMemory:
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].MemoryMB > filtered[j].MemoryMB
		})
	case SortCPU:
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].CPU > filtered[j].CPU
		})
	case SortTitle:
		sort.SliceStable(filtered, func(i, j int) bool {
			return strings.ToLower(filtered[i].Title) < strings.ToLower(filtered[j].Title)
		})
	}

	a.filtered = filtered
}

func (a *App) renderTable() {
	a.table.Clear()

	// Column headers
	headers := []string{"  PID", " CPU %", "  MEMORY", " WINDOW", " TAB", " TITLE", " URL"}
	for col, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetAttributes(tcell.AttrBold)
		a.table.SetCell(0, col, cell)
	}

	a.mu.RLock()
	tabs := a.filtered
	a.mu.RUnlock()

	for rowIdx, t := range tabs {
		row := rowIdx + 1

		pidStr := "-"
		if t.PID > 0 {
			pidStr = fmt.Sprintf("%6d", t.PID)
		}

		cpuStr := "-"
		if t.CPU > 0.0 || t.MemoryMB > 0 {
			cpuStr = fmt.Sprintf("%5.1f%%", t.CPU)
		}

		memStr := t.MemoryStr
		if memStr == "" || memStr == "-" {
			memStr = "    -    "
		} else {
			memStr = fmt.Sprintf("%9s", memStr)
		}

		cpuColor := tcell.ColorWhite
		if t.CPU >= 20.0 {
			cpuColor = tcell.ColorRed
		} else if t.CPU >= 5.0 {
			cpuColor = tcell.ColorOrange
		}

		memColor := tcell.ColorWhite
		if t.MemoryMB >= 500 {
			memColor = tcell.ColorOrange
		} else if t.MemoryMB >= 800 {
			memColor = tcell.ColorRed
		}

		a.table.SetCell(row, 0, tview.NewTableCell(pidStr).SetTextColor(tcell.ColorLightGray))
		a.table.SetCell(row, 1, tview.NewTableCell(cpuStr).SetTextColor(cpuColor))
		a.table.SetCell(row, 2, tview.NewTableCell(memStr).SetTextColor(memColor))
		a.table.SetCell(row, 3, tview.NewTableCell(fmt.Sprintf("%7s", t.WindowID)).SetTextColor(tcell.ColorDarkGray))
		a.table.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("%4d", t.TabIndex)).SetTextColor(tcell.ColorDarkGray))
		a.table.SetCell(row, 5, tview.NewTableCell(" "+t.Title).SetTextColor(tcell.ColorWhite).SetExpansion(2))
		a.table.SetCell(row, 6, tview.NewTableCell(" "+t.URL).SetTextColor(tcell.ColorCadetBlue).SetExpansion(3))
	}

	if len(tabs) > 0 {
		currentRow, _ := a.table.GetSelection()
		if currentRow <= 0 || currentRow > len(tabs) {
			a.table.Select(1, 0)
		} else {
			a.table.Select(currentRow, 0)
		}
	}
}

func (a *App) updateStatusBar() {
	a.mu.RLock()
	sortStr := "Memory"
	switch a.sortMode {
	case SortCPU:
		sortStr = "CPU"
	case SortTitle:
		sortStr = "Title"
	}

	countStr := fmt.Sprintf("Showing %d / %d tabs", len(a.filtered), len(a.tabs))
	msg := a.statusMsg
	a.mu.RUnlock()

	text := fmt.Sprintf(" [::b]Sort: %s[::-] | %s | %s", sortStr, countStr, msg)
	a.statusBar.SetText(text)
}

func (a *App) getSelectedTab() *chrome.Tab {
	row, _ := a.table.GetSelection()
	if row <= 0 {
		return nil
	}
	tabIdx := row - 1
	a.mu.RLock()
	defer a.mu.RUnlock()
	if tabIdx < len(a.filtered) {
		t := a.filtered[tabIdx]
		return &t
	}
	return nil
}

func (a *App) focusSelectedTab() {
	tab := a.getSelectedTab()
	if tab == nil {
		return
	}

	a.mu.Lock()
	a.statusMsg = fmt.Sprintf("Bringing '%s' to front...", tab.Title)
	a.mu.Unlock()
	a.updateStatusBar()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := a.client.FocusTab(ctx, *tab)
		a.mu.Lock()
		if err != nil {
			a.statusMsg = fmt.Sprintf("[red]Focus failed: %v[-]", err)
		} else {
			a.statusMsg = fmt.Sprintf("[green]Focused tab: %s[-]", tab.Title)
		}
		a.mu.Unlock()
		a.app.QueueUpdateDraw(func() {
			a.updateStatusBar()
		})
	}()
}

func (a *App) closeSelectedTab() {
	tab := a.getSelectedTab()
	if tab == nil {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := a.client.CloseTab(ctx, *tab)
		a.mu.Lock()
		if err != nil {
			a.statusMsg = fmt.Sprintf("[red]Close failed: %v[-]", err)
		} else {
			a.statusMsg = fmt.Sprintf("[yellow]Closed tab: %s[-]", tab.Title)
		}
		a.mu.Unlock()

		// Trigger background refresh
		a.RefreshData()
	}()
}

// Run starts the TUI application loop.
func (a *App) Run() error {
	a.app.SetRoot(a.flex, true)
	go a.RefreshData()
	return a.app.Run()
}
