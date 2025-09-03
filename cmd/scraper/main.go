package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	ps "github.com/mitchellh/go-ps"
	"gopkg.in/yaml.v3"

	"paul-scraping/internal/scraper"
)

const defaultURL = "https://www.armp.cm/recherche_avancee?recherche_avancee_do=1&reference_avis=&maitre_ouvrage=0&region=0&departement=0&type_publication%5B%5D=AO"

func main() {
	// Flags
	formatFlag := flag.String("format", "", "output format: json or yaml (default yaml)")
	outputFlag := flag.String("output", "", "output file path (optional); stdout if empty")
	showHTMLFlag := flag.Bool("show-html", false, "also log each item's inner HTML before parsing")
	autoInstallFlag := flag.Bool("auto-install", true, "auto-install Playwright driver/browsers if missing")
	pauseFlag := flag.Bool("pause", false, "on Windows, wait for Enter before exiting")
	urlFlag := flag.String("url", defaultURL, "target page URL to scrape")
	flag.Parse()

	// Resolve format (flags override env)
	format := "yaml"
	if f := strings.ToLower(strings.TrimSpace(*formatFlag)); f != "" {
		format = f
	} else if f := strings.ToLower(strings.TrimSpace(os.Getenv("FORMAT"))); f != "" {
		format = f
	}
	if format != "yaml" && format != "json" {
		log.Printf("unknown format %q; defaulting to yaml", format)
		format = "yaml"
	}

	// Resolve output file (flag overrides env)
	outputPath := *outputFlag
	if outputPath == "" {
		outputPath = strings.TrimSpace(os.Getenv("OUTPUT_FILE"))
	}

	// Toggle HTML dump via env or flag
	showHTML := func() bool {
		v := strings.ToLower(strings.TrimSpace(os.Getenv("SHOW_HTML")))
		envOn := v == "1" || v == "true" || v == "yes" || v == "y" || v == "on"
		return envOn || *showHTMLFlag
	}()

	// Run scraper
	notices, err := scraper.Scrape(context.Background(), scraper.Options{
		URL:         *urlFlag,
		ShowHTML:    showHTML,
		AutoInstall: *autoInstallFlag,
	})
	if err != nil {
		log.Fatalf("scrape: %v", err)
	}

	// Marshal output
	var out []byte
	switch format {
	case "json":
		if b, err := json.MarshalIndent(notices, "", "  "); err == nil {
			out = b
		} else {
			log.Fatalf("json marshal: %v", err)
		}
	default: // yaml
		if b, err := yaml.Marshal(notices); err == nil {
			out = b
		} else {
			log.Fatalf("yaml marshal: %v", err)
		}
	}

	// Write to file or stdout
	if outputPath != "" {
		if err := os.WriteFile(outputPath, out, 0o644); err != nil {
			log.Fatalf("write output file: %v", err)
		}
		fmt.Fprintf(os.Stderr, "wrote %d bytes to %s\n", len(out), outputPath)
	} else {
		os.Stdout.Write(out)
	}

	// Pause on Windows (flag/env or double-click via explorer)
	pauseEnv := strings.ToLower(strings.TrimSpace(os.Getenv("PAUSE_ON_EXIT")))
	pauseOnExit := *pauseFlag || pauseEnv == "1" || pauseEnv == "true" || pauseEnv == "yes" || pauseEnv == "on" || pauseEnv == "y"
	if runtime.GOOS == "windows" && !pauseOnExit {
		pid := os.Getppid()
		for i := 0; i < 3 && pid > 0; i++ {
			if proc, err := ps.FindProcess(pid); err == nil && proc != nil {
				if strings.ToLower(proc.Executable()) == "explorer.exe" {
					pauseOnExit = true
					break
				}
				pid = proc.PPid()
			} else {
				break
			}
		}
	}
	if runtime.GOOS == "windows" && pauseOnExit {
		fmt.Fprint(os.Stderr, "\nPress Enter to exit...")
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
	}
}
