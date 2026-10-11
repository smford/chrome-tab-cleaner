package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/smford/chrome-tab-cleaner/internal/chrome"
	"github.com/smford/chrome-tab-cleaner/internal/tui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	showVer := flag.Bool("version", false, "Show version information")
	serveHealthz := flag.String("serve-healthz", "", "Run lightweight HTTP healthcheck daemon on given address (e.g. :8080) for container DAST / health checking")
	flag.Parse()

	if *showVer {
		fmt.Printf("chrome-tab-cleaner version %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	if *serveHealthz != "" {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status":  "ok",
				"service": "chrome-tab-cleaner",
				"version": version,
			})
		})
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status": "ok",
			})
		})
		server := &http.Server{
			Addr:              *serveHealthz,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       30 * time.Second,
		}
		fmt.Fprintf(os.Stderr, "chrome-tab-cleaner daemon listening on %s (health endpoint: /healthz)\n", *serveHealthz)
		// nosemgrep: go.lang.security.audit.net.use-tls.use-tls
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	client := chrome.NewClient()
	app := tui.NewApp(client)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
