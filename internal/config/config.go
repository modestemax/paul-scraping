package config

import (
	"flag"
	"os"
	"strings"
)

const DefaultURL = "https://www.armp.cm/recherche_avancee?recherche_avancee_do=1&reference_avis=&maitre_ouvrage=0&region=0&departement=0&type_publication%5B%5D=AO"

type Config struct {
	Format      string
	OutputPath  string
	ShowHTML    bool
	AutoInstall bool
	Pause       bool
	URL         string
}

// Parse reads flags and env vars to produce the runtime config.
func Parse() *Config {
	formatFlag := flag.String("format", "", "output format: json or yaml (default yaml)")
	outputFlag := flag.String("output", "", "output file path (optional); stdout if empty")
	showHTMLFlag := flag.Bool("show-html", false, "also log each item's inner HTML before parsing")
	autoInstallFlag := flag.Bool("auto-install", true, "auto-install Playwright driver/browsers if missing")
	pauseFlag := flag.Bool("pause", false, "on Windows, wait for Enter before exiting")
	urlFlag := flag.String("url", DefaultURL, "target page URL to scrape")
	flag.Parse()

	// Resolve format
	format := "yaml"
	if f := strings.ToLower(strings.TrimSpace(*formatFlag)); f != "" {
		format = f
	} else if f := strings.ToLower(strings.TrimSpace(os.Getenv("FORMAT"))); f != "" {
		format = f
	}
	if format != "yaml" && format != "json" {
		format = "yaml"
	}

	// Resolve output file
	outputPath := *outputFlag
	if outputPath == "" {
		outputPath = strings.TrimSpace(os.Getenv("OUTPUT_FILE"))
	}

	// Resolve show HTML
	showHTML := func() bool {
		v := strings.ToLower(strings.TrimSpace(os.Getenv("SHOW_HTML")))
		envOn := v == "1" || v == "true" || v == "yes" || v == "y" || v == "on"
		return envOn || *showHTMLFlag
	}()

	return &Config{
		Format:      format,
		OutputPath:  outputPath,
		ShowHTML:    showHTML,
		AutoInstall: *autoInstallFlag,
		Pause:       *pauseFlag,
		URL:         *urlFlag,
	}
}
