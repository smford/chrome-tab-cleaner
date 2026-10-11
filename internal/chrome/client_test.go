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
