package chrome

import "testing"

func TestParseMemoryToMB(t *testing.T) {
	cases := []struct {
		input    string
		expected float64
	}{
		{"642 MB", 642.0},
		{"1.2 GB", 1228.8},
		{"512 KB", 0.5},
		{"0 MB", 0.0},
		{"-", 0.0},
	}

	for _, tc := range cases {
		got := parseMemoryToMB(tc.input)
		if got != tc.expected {
			t.Errorf("parseMemoryToMB(%q) = %v, expected %v", tc.input, got, tc.expected)
		}
	}
}

func TestParseTaskManagerOutput(t *testing.T) {
	raw := `
Column Task, Column Memory footprint sorted in descending order, Column CPU, Column Network and Column Process ID
---DELIM---
Tab: Daily Deals: Find The Deal Of The Day at The Home Depot
---DELIM---
642 MB
---DELIM---
0.0
---DELIM---
0
---DELIM---
85529
---DELIM---
Tab: Boiler Upgrade Scheme | Daikin, 89.2 MB, 0.0, 0 and 2953
---DELIM---
`
	entries := parseTaskManagerOutput(raw)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Title != "Daily Deals: Find The Deal Of The Day at The Home Depot" || entries[0].PID != 85529 || entries[0].MemoryMB != 642.0 {
		t.Errorf("unexpected entry 0: %+v", entries[0])
	}

	if entries[1].Title != "Boiler Upgrade Scheme | Daikin" || entries[1].PID != 2953 || entries[1].MemoryMB != 89.2 {
		t.Errorf("unexpected entry 1: %+v", entries[1])
	}
}

func TestParseTopPowerOutput(t *testing.T) {
	raw := `Processes: 600 total, 2 running, 598 sleeping
2026/11/16 23:44:21
Load Avg: 1.21, 2.28, 3.49
PID    POWER
99279  0.0
99274  0.0

Processes: 600 total, 2 running, 598 sleeping
2026/11/16 23:44:22
Load Avg: 1.21, 2.28, 3.49
PID    POWER
1722   59.4
720    11.6
380    10.0
85529  18.7
2953   0.0
`
	powerMap := ParseTopPowerOutput(raw)
	if len(powerMap) != 5 {
		t.Fatalf("expected 5 power entries from 2nd iteration, got %d", len(powerMap))
	}

	if powerMap[1722] != 59.4 {
		t.Errorf("expected PID 1722 power 59.4, got %f", powerMap[1722])
	}
	if powerMap[85529] != 18.7 {
		t.Errorf("expected PID 85529 power 18.7, got %f", powerMap[85529])
	}
	if powerMap[2953] != 0.0 {
		t.Errorf("expected PID 2953 power 0.0, got %f", powerMap[2953])
	}
}

func TestCorrelateTabsWithEnergy(t *testing.T) {
	tabs := []Tab{
		{
			ID:    "1",
			Title: "Daily Deals: Find The Deal Of The Day at The Home Depot",
			URL:   "https://homedepot.com/deals",
		},
		{
			ID:    "2",
			Title: "Boiler Upgrade Scheme | Daikin",
			URL:   "https://daikin.com/bus",
		},
		{
			ID:    "3",
			Title: "Discarded / Idle Tab",
			URL:   "https://example.com",
		},
	}

	metrics := []TaskManagerEntry{
		{
			Title:     "Daily Deals: Find The Deal Of The Day at The Home Depot",
			MemoryStr: "642 MB",
			MemoryMB:  642.0,
			CPU:       12.5,
			PID:       85529,
		},
		{
			Title:     "Boiler Upgrade Scheme | Daikin",
			MemoryStr: "89.2 MB",
			MemoryMB:  89.2,
			CPU:       0.0,
			PID:       2953,
		},
	}

	energyMap := map[int]float64{
		85529: 23.4,
		2953:  0.0,
	}

	correlated := CorrelateTabs(tabs, metrics, energyMap)
	if len(correlated) != 3 {
		t.Fatalf("expected 3 correlated tabs, got %d", len(correlated))
	}

	// Tab 1: matched with TM and energy
	if correlated[0].PID != 85529 || correlated[0].Energy != 23.4 || correlated[0].EnergyStr != "23.4" {
		t.Errorf("tab 0 energy mismatch: %+v", correlated[0])
	}

	// Tab 2: matched with TM, energy 0.0
	if correlated[1].PID != 2953 || correlated[1].Energy != 0.0 || correlated[1].EnergyStr != "0.0" {
		t.Errorf("tab 1 energy mismatch: %+v", correlated[1])
	}

	// Tab 3: unmatched (idle / discarded)
	if correlated[2].PID != 0 || correlated[2].Energy != 0.0 || correlated[2].EnergyStr != "-" {
		t.Errorf("tab 2 (unmatched) energy mismatch: %+v", correlated[2])
	}
}
