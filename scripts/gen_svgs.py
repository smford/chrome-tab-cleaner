#!/usr/bin/env python3
import subprocess
import sys

WIDTH = 200

# ANSI color helper functions
def reset(): return "\033[0m"
def bold(): return "\033[1m"
def yellow(): return "\033[1;33m"
def red(): return "\033[1;31m"
def orange(): return "\033[38;5;214m"
def white(): return "\033[37m"
def bright_white(): return "\033[1;37m"
def gray(): return "\033[38;5;248m"
def dark_gray(): return "\033[38;5;242m"
def cadet_blue(): return "\033[38;5;73m"
def bg_blue(): return "\033[44m"
def bg_navy(): return "\033[48;5;18m"
def bg_cyan(): return "\033[48;5;30m"
def bg_dark(): return "\033[40m"

def render_header():
    # "[::b]CHROME TAB TOP CLEANER[::-] | [yellow]m[-]: Sort Mem | [yellow]c[-]: Sort CPU | [yellow]t[-]: Sort Title | [yellow]r[-]: Refresh | [yellow]Enter[-]: Focus | [yellow]x[-]: Close Tab | [yellow]q[-]: Quit"
    prefix = " CHROME TAB TOP CLEANER "
    items = [
        ("m", "Sort Mem"),
        ("c", "Sort CPU"),
        ("t", "Sort Title"),
        ("r", "Refresh"),
        ("Enter", "Focus"),
        ("x", "Close Tab"),
        ("q", "Quit"),
    ]
    raw_len = len(prefix)
    line = f"{bg_blue()}{bright_white()}{prefix}"
    for k, desc in items:
        line += f"{bright_white()}| {yellow()}{k}{bright_white()}: {desc} "
        raw_len += len(f"| {k}: {desc} ")
    padding = max(0, WIDTH - raw_len)
    line += " " * padding + reset() + "\n"
    return line

def render_filter(prompt=" Filter: ", text=""):
    raw_len = len(prompt) + len(text)
    padding = max(0, WIDTH - raw_len)
    line = f"{reset()}{bright_white()}{prompt}{yellow()}{text}{reset()}" + " " * padding + "\n"
    return line

def render_table_header():
    # Headers: "  PID", " CPU %", "  MEMORY", " WINDOW", " TAB", " TITLE", " URL"
    # Column widths:
    # PID: 7 chars
    # CPU: 7 chars
    # MEMORY: 11 chars
    # WINDOW: 8 chars
    # TAB: 5 chars
    # TITLE: 58 chars
    # URL: fills rest to WIDTH
    line = f"{yellow()}{bold()}"
    line += f"{'  PID':>7} "
    line += f"{' CPU %':>7} "
    line += f"{'  MEMORY':>11} "
    line += f"{' WINDOW':>8} "
    line += f"{' TAB':>5} "
    line += f"{' TITLE':<58} "
    line += f"{' URL':<96}"
    line += reset() + "\n"
    return line

def render_row(pid, cpu_val, cpu_str, mem_mb, mem_str, win_id, tab_idx, title, url, selected=False):
    # Truncate strings to prevent wrapping
    t_trunc = title[:56]
    u_trunc = url[:95]

    # Format fields
    pid_s = f"{pid:>7}"
    cpu_s = f"{cpu_str:>7}"
    mem_s = f"{mem_str:>11}"
    win_s = f"{win_id:>8}"
    tab_s = f"{tab_idx:>5}"
    title_s = f" {t_trunc:<57}"
    url_s = f" {u_trunc:<95}"

    if selected:
        # Dark cyan background across entire line
        line = f"{bg_cyan()}{bright_white()}{bold()}"
        line += pid_s + " "
        line += cpu_s + " "
        line += mem_s + " "
        line += win_s + " "
        line += tab_s + " "
        line += title_s + " "
        line += url_s
        # Total length so far: 7+1+7+1+11+1+8+1+5+1+58+1+96 = 198 chars
        total_len = 198
        line += " " * max(0, WIDTH - total_len) + reset() + "\n"
        return line

    # Normal row styling
    # CPU color
    if cpu_val >= 20.0:
        c_color = red()
    elif cpu_val >= 5.0:
        c_color = orange()
    else:
        c_color = white()

    # Memory color
    if mem_mb >= 800:
        m_color = red()
    elif mem_mb >= 500:
        m_color = orange()
    else:
        m_color = white()

    line = f"{gray()}{pid_s}{reset()} "
    line += f"{c_color}{cpu_s}{reset()} "
    line += f"{m_color}{mem_s}{reset()} "
    line += f"{dark_gray()}{win_s}{reset()} "
    line += f"{dark_gray()}{tab_s}{reset()} "
    line += f"{white()}{title_s}{reset()} "
    line += f"{cadet_blue()}{url_s}{reset()}"
    line += " " * max(0, WIDTH - 198) + reset() + "\n"
    return line

def render_status_bar(sort_str, showing_str, status_msg):
    text_lead = f" Sort: {sort_str} | {showing_str} | "
    raw_lead = len(text_lead)
    raw_msg = len(status_msg)
    padding = max(0, WIDTH - (raw_lead + raw_msg))
    line = f"{bg_navy()}{bright_white()}{bold()} Sort: {sort_str}{reset()}{bg_navy()}{bright_white()} | {showing_str} | {yellow()}{status_msg}{bg_navy()}" + (" " * padding) + reset() + "\n"
    return line

def generate_svg(ansi_text, output_file, rows=22):
    p = subprocess.Popen(
        ['/Users/asc/git/terminal-to-svg/termtosvg', '-cols', '200', '-rows', str(rows), '-auto-crop', '-title', 'chrome-tab-cleaner (macOS)', '-o', output_file],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=False
    )
    stdout, stderr = p.communicate(input=ansi_text.encode('utf-8'))
    if p.returncode != 0:
        print(f"Error generating {output_file}: {stderr.decode('utf-8', errors='ignore')}", file=sys.stderr)
        sys.exit(1)
    print(f"Generated {output_file}")

# 1. Overview / Memory Sort (m)
def make_screenshot_1():
    rows = [
        (48210, 4.2, "4.2%", 1420, "1,420 MB", "win-1", 3, "Figma — App Design System & Tokens (v3.4)", "https://figma.com/file/a8F92j9/Design-System-Production", True),
        (48392, 12.8, "12.8%", 1180, "1,180 MB", "win-1", 7, "Claude — Architecture Refactor & LLM Evaluation", "https://claude.ai/chat/8a7b9c-refactor-plan", False),
        (48102, 1.1, "1.1%", 985, "985 MB", "win-2", 1, "Datadog — APM Traces & Latency Dashboard (us-east-1)", "https://app.datadoghq.com/apm/traces?env=production", False),
        (48512, 18.4, "18.4%", 870, "870 MB", "win-1", 12, "Grafana — Kubernetes Cluster Metrics & Memory Burn", "https://grafana.internal.net/d/k8s-cluster/overview", False),
        (47930, 0.4, "0.4%", 740, "740 MB", "win-2", 4, "Google Sheets — Q3 Cloud Infrastructure Budget", "https://docs.google.com/spreadsheets/d/1BxiMVs0XRA/edit", False),
        (48441, 0.2, "0.2%", 620, "620 MB", "win-1", 15, "Linear — Sprint 42: Backend Reliability & PagerDuty", "https://linear.app/team/issue/ENG-4209/latency-spikes", False),
        (47819, 0.0, "0.0%", 540, "540 MB", "win-2", 8, "Notion — Engineering Incident Postmortems 2026", "https://notion.so/org/Incident-Postmortem-Q3", False),
        (48620, 2.5, "2.5%", 480, "480 MB", "win-3", 2, "YouTube — Lo-Fi Chill Beats to Code/Relax To (Live)", "https://youtube.com/watch?v=jfKfPfyJRdk", False),
        (48290, 0.1, "0.1%", 390, "390 MB", "win-1", 18, "GitHub — PR #14: Optimize tab memory and event polling", "https://github.com/asc/chrome-tab-cleaner/pull/14", False),
        (48050, 0.0, "0.0%", 310, "310 MB", "win-2", 11, "Slack Web — #engineering-core channel stream", "https://app.slack.com/client/T024/C048A9", False),
        (48180, 0.0, "0.0%", 245, "245 MB", "win-3", 5, "Hacker News — Ask HN: What are your favourite CLI tools?", "https://news.ycombinator.com/item?id=38192014", False),
        (48310, 0.0, "0.0%", 180, "180 MB", "win-1", 22, "Golang Documentation — pkg/sync/atomic memory models", "https://pkg.go.dev/sync/atomic", False),
        (48405, 0.0, "0.0%", 140, "140 MB", "win-2", 19, "Stack Overflow — Efficient IPC communication in macOS Go", "https://stackoverflow.com/questions/591823/macos-ipc", False),
        (48700, 0.0, "0.0%", 95, "95 MB", "win-3", 9, "MDN Web Docs — Memory management and Garbage Collection", "https://developer.mozilla.org/en-US/docs/Web/JavaScript/Memory", False),
        (48810, 0.0, "0.0%", 68, "68 MB", "win-1", 28, "Chrome Web Store — Extension manifest and tabs permissions", "https://chromewebstore.google.com/detail/tab-manager", False),
    ]
    txt = render_header()
    txt += render_filter(" Filter: ", "")
    txt += render_table_header()
    for r in rows:
        txt += render_row(*r)
    txt += render_status_bar("Memory", f"Showing {len(rows)} / 47 tabs", "Loaded 47 tabs (47 metrics matched) at 14:02:18")
    generate_svg(txt, "docs/images/01-tui-overview-memory.svg", rows=len(rows)+5)

# 2. Live Search & Duplicate Filter (/)
def make_screenshot_2():
    rows = [
        (48290, 0.1, "0.1%", 390, "390 MB", "win-1", 18, "GitHub — PR #14: Optimize tab memory and event polling", "https://github.com/asc/chrome-tab-cleaner/pull/14", True),
        (48295, 0.0, "0.0%", 320, "320 MB", "win-2", 6, "GitHub — PR #14: Optimize tab memory and event polling [Duplicate]", "https://github.com/asc/chrome-tab-cleaner/pull/14", False),
        (48301, 0.2, "0.2%", 285, "285 MB", "win-3", 14, "GitHub — asc/chrome-tab-cleaner: Fast TUI tab cleaner", "https://github.com/asc/chrome-tab-cleaner", False),
        (48305, 0.0, "0.0%", 210, "210 MB", "win-1", 24, "GitHub — Issues: Search filter regression under high load", "https://github.com/asc/chrome-tab-cleaner/issues/8", False),
        (48312, 0.0, "0.0%", 195, "195 MB", "win-2", 21, "GitHub — Discussions: Feature requests & tab grouping", "https://github.com/asc/chrome-tab-cleaner/discussions", False),
    ]
    txt = render_header()
    txt += render_filter(" Filter: ", "github")
    txt += render_table_header()
    for r in rows:
        txt += render_row(*r)
    # Fill empty lines to give standard TUI height
    for _ in range(8):
        txt += " " * WIDTH + "\n"
    txt += render_status_bar("Memory", "Showing 5 / 47 tabs", "Filter active: 'github' — 1 duplicate tab detected")
    generate_svg(txt, "docs/images/02-search-filter.svg", rows=18)

# 3. CPU Spike Alert (c)
def make_screenshot_3():
    rows = [
        (48991, 74.6, "74.6%", 610, "610 MB", "win-1", 9, "Infinite Loop Benchmark & Web Worker Stress Test", "https://stress-test.internal.corp/cpu-benchmark", True),
        (48512, 18.4, "18.4%", 870, "870 MB", "win-1", 12, "Grafana — Kubernetes Cluster Metrics & Memory Burn", "https://grafana.internal.net/d/k8s-cluster/overview", False),
        (48392, 12.8, "12.8%", 1180, "1,180 MB", "win-1", 7, "Claude — Architecture Refactor & LLM Evaluation", "https://claude.ai/chat/8a7b9c-refactor-plan", False),
        (48750, 8.5, "8.5%", 420, "420 MB", "win-3", 17, "Three.js WebGL Interactive 3D Simulation Canvas", "https://threejs.org/examples/#webgl_animation_skinning", False),
        (48210, 4.2, "4.2%", 1420, "1,420 MB", "win-1", 3, "Figma — App Design System & Tokens (v3.4)", "https://figma.com/file/a8F92j9/Design-System-Production", False),
        (48620, 2.5, "2.5%", 480, "480 MB", "win-3", 2, "YouTube — Lo-Fi Chill Beats to Code/Relax To (Live)", "https://youtube.com/watch?v=jfKfPfyJRdk", False),
        (48102, 1.1, "1.1%", 985, "985 MB", "win-2", 1, "Datadog — APM Traces & Latency Dashboard (us-east-1)", "https://app.datadoghq.com/apm/traces?env=production", False),
        (47930, 0.4, "0.4%", 740, "740 MB", "win-2", 4, "Google Sheets — Q3 Cloud Infrastructure Budget", "https://docs.google.com/spreadsheets/d/1BxiMVs0XRA/edit", False),
        (48301, 0.2, "0.2%", 285, "285 MB", "win-3", 14, "GitHub — asc/chrome-tab-cleaner: Fast TUI tab cleaner", "https://github.com/asc/chrome-tab-cleaner", False),
        (48441, 0.2, "0.2%", 620, "620 MB", "win-1", 15, "Linear — Sprint 42: Backend Reliability & PagerDuty", "https://linear.app/team/issue/ENG-4209/latency-spikes", False),
        (48290, 0.1, "0.1%", 390, "390 MB", "win-1", 18, "GitHub — PR #14: Optimize tab memory and event polling", "https://github.com/asc/chrome-tab-cleaner/pull/14", False),
    ]
    txt = render_header()
    txt += render_filter(" Filter: ", "")
    txt += render_table_header()
    for r in rows:
        txt += render_row(*r)
    for _ in range(3):
        txt += " " * WIDTH + "\n"
    txt += render_status_bar("CPU", f"Showing {len(rows)} / 47 tabs", "HIGH CPU: 1 process exceeding 70% CPU threshold")
    generate_svg(txt, "docs/images/03-cpu-sort-alert.svg", rows=19)

# 4. Instant Tab Close (x)
def make_screenshot_4():
    rows = [
        (48991, 74.6, "74.6%", 610, "610 MB", "win-1", 9, "Infinite Loop Benchmark & Web Worker Stress Test", "https://stress-test.internal.corp/cpu-benchmark", True),
        (48210, 4.2, "4.2%", 1420, "1,420 MB", "win-1", 3, "Figma — App Design System & Tokens (v3.4)", "https://figma.com/file/a8F92j9/Design-System-Production", False),
        (48392, 12.8, "12.8%", 1180, "1,180 MB", "win-1", 7, "Claude — Architecture Refactor & LLM Evaluation", "https://claude.ai/chat/8a7b9c-refactor-plan", False),
        (48102, 1.1, "1.1%", 985, "985 MB", "win-2", 1, "Datadog — APM Traces & Latency Dashboard (us-east-1)", "https://app.datadoghq.com/apm/traces?env=production", False),
        (48512, 18.4, "18.4%", 870, "870 MB", "win-1", 12, "Grafana — Kubernetes Cluster Metrics & Memory Burn", "https://grafana.internal.net/d/k8s-cluster/overview", False),
        (47930, 0.4, "0.4%", 740, "740 MB", "win-2", 4, "Google Sheets — Q3 Cloud Infrastructure Budget", "https://docs.google.com/spreadsheets/d/1BxiMVs0XRA/edit", False),
        (48441, 0.2, "0.2%", 620, "620 MB", "win-1", 15, "Linear — Sprint 42: Backend Reliability & PagerDuty", "https://linear.app/team/issue/ENG-4209/latency-spikes", False),
    ]
    txt = render_header()
    txt += render_filter(" Filter: ", "")
    txt += render_table_header()
    for r in rows:
        txt += render_row(*r)
    for _ in range(6):
        txt += " " * WIDTH + "\n"
    txt += render_status_bar("CPU", "Showing 7 / 46 tabs", "Closed tab 'Infinite Loop Benchmark' (PID 48991 reclaimed)")
    generate_svg(txt, "docs/images/04-close-action.svg", rows=18)

if __name__ == '__main__':
    make_screenshot_1()
    make_screenshot_2()
    make_screenshot_3()
    make_screenshot_4()
    print("All screenshots generated successfully.")
