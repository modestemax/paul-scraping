package main

import (
	"context"
	"log"

	"paul-scraping/internal/config"
	"paul-scraping/internal/install"
	"paul-scraping/internal/output"
	"paul-scraping/internal/scraper"
	"paul-scraping/internal/winutil"
)

func main() {
	cfg := config.Parse()

	pw, err := install.StartPlaywright(cfg.AutoInstall)
	if err != nil {
		log.Fatalf("playwright: %v", err)
	}
	defer pw.Stop()

	notices, err := scraper.Scrape(context.Background(), pw, cfg.URL, cfg.ShowHTML)
	if err != nil {
		log.Fatalf("scrape: %v", err)
	}

	if err := output.Write(notices, cfg.Format, cfg.OutputPath); err != nil {
		log.Fatalf("output: %v", err)
	}

	if winutil.ShouldPause(cfg.Pause) {
		winutil.Pause()
	}
}
