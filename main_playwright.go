//go:build playwright

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/playwright-community/playwright-go"
	"gopkg.in/yaml.v3"
)

type Notice struct {
	NumberTitle string `yaml:"number_title" json:"number_title"`
	Type        string `yaml:"type" json:"type"`
	Authority   string `yaml:"authority" json:"authority"`
	Region      string `yaml:"region" json:"region"`
	Amount      string `yaml:"amount" json:"amount"`
	PublishedOn string `yaml:"published_on" json:"published_on"`
	ClosingDate string `yaml:"closing_date" json:"closing_date"`
	ClosingTime string `yaml:"closing_time" json:"closing_time"`
}

func main() {
	// Flags: format (json|yaml), output file, and show-html toggle
	formatFlag := flag.String("format", "", "output format: json or yaml (default yaml)")
	outputFlag := flag.String("output", "", "output file path (optional); stdout if empty")
	showHTMLFlag := flag.Bool("show-html", false, "also log each item's inner HTML before parsing")
	autoInstallFlag := flag.Bool("auto-install", true, "auto-install Playwright driver/browsers if missing")
	flag.Parse()

	pw, err := playwright.Run()
	if err != nil && *autoInstallFlag && strings.Contains(err.Error(), "please install the driver") {
		log.Printf("Playwright driver missing; attempting auto-install…")
		if err2 := playwright.Install(); err2 != nil {
			log.Fatalf("auto-install failed: %v", err2)
		}
		pw, err = playwright.Run()
	}
	if err != nil {
		log.Fatalf("could not start Playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	context, err := browser.NewContext()
	if err != nil {
		log.Fatalf("new context: %v", err)
	}
	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("new page: %v", err)
	}

	url := "https://www.armp.cm/recherche_avancee?recherche_avancee_do=1&reference_avis=&maitre_ouvrage=0&region=0&departement=0&type_publication%5B%5D=AO"
	if _, err = page.Goto(url); err != nil {
		log.Fatalf("goto: %v", err)
	}
	if _, err := page.WaitForSelector(".list-group"); err != nil {
		log.Fatalf("wait for list group: %v", err)
	}

	container, err := page.QuerySelector(".list-group")
	if err != nil || container == nil {
		log.Fatalf("find list group: %v", err)
	}

	items, err := container.QuerySelectorAll("li.list-group-item")
	if err != nil {
		log.Fatalf("query list items: %v", err)
	}
	log.Printf("items found: %d", len(items))

	textContent := func(el playwright.ElementHandle, sel string) string {
		h, err := el.QuerySelector(sel)
		if err != nil || h == nil {
			return ""
		}
		txt, err := h.TextContent()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(txt)
	}

	// Helper: find the value next to a label, trying multiple label variants.
	labelValue := func(el playwright.ElementHandle, labels ...string) string {
		for _, lbl := range labels {
			sel := fmt.Sprintf("div:has-text(\"%s\") + div", lbl)
			if val := textContent(el, sel); val != "" {
				return val
			}
		}
		return ""
	}

	// Detect page language to prioritize label variants
	pageLang := ""
	if htmlEl, err := page.QuerySelector("html"); err == nil && htmlEl != nil {
		if lang, err := htmlEl.GetAttribute("lang"); err == nil && lang != "" {
			pageLang = strings.ToLower(strings.TrimSpace(lang))
		}
	}
	log.Printf("page lang: %s", pageLang)

	// Resolve output format: flags override env FORMAT; default yaml
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

	// Resolve output file: flag overrides env OUTPUT_FILE
	outputPath := *outputFlag
	if outputPath == "" {
		outputPath = strings.TrimSpace(os.Getenv("OUTPUT_FILE"))
	}

	// Toggle HTML dump via env SHOW_HTML (true/1/yes) or --show-html
	showHTML := func() bool {
		v := strings.ToLower(strings.TrimSpace(os.Getenv("SHOW_HTML")))
		envOn := v == "1" || v == "true" || v == "yes" || v == "y" || v == "on"
		return envOn || *showHTMLFlag
	}()

	notices := make([]Notice, 0, len(items))

	for _, item := range items {
		if showHTML {
			// Log the raw HTML of the item before parsing
			if html, err := item.InnerHTML(); err == nil {
				log.Printf("ITEM HTML:\n%s\n", strings.TrimSpace(html))
			} else {
				log.Printf("failed to get item innerHTML: %v", err)
			}
		}
		// Build ordered variants depending on detected lang
		concat := func(a, b []string) []string { return append(append([]string{}, a...), b...) }

		var authVariants, typeVariants, regionVariants, amountVariants, publishedVariants, dateVariants, timeVariants []string
		if strings.HasPrefix(pageLang, "fr") {
			authVariants = concat([]string{"MO/AC:"}, []string{"PO/CA:"})
			typeVariants = concat([]string{"Type", "Type:"}, []string{"Type", "Type:"})
			regionVariants = concat([]string{"Région", "Région:", "Region", "Region:"}, []string{"Region", "Region:"})
			amountVariants = concat([]string{"Montant", "Montant:"}, []string{"Amount", "Amount:"})
			publishedVariants = concat([]string{"Publié le", "Publié le :", "Publié le:"}, []string{"Published on", "Published on:", "Published:"})
			dateVariants = concat([]string{"Date de clôture", "Date de cloture"}, []string{"Closing date"})
			timeVariants = concat([]string{"Heure de clôture", "Heure de cloture"}, []string{"Closing time"})
		} else {
			authVariants = concat([]string{"PO/CA:"}, []string{"MO/AC:"})
			typeVariants = concat([]string{"Type", "Type:"}, []string{"Type", "Type:"})
			regionVariants = concat([]string{"Region", "Region:"}, []string{"Région", "Région:"})
			amountVariants = concat([]string{"Amount", "Amount:"}, []string{"Montant", "Montant:"})
			publishedVariants = concat([]string{"Published on", "Published on:", "Published:"}, []string{"Publié le", "Publié le :", "Publié le:"})
			dateVariants = concat([]string{"Closing date"}, []string{"Date de clôture", "Date de cloture"})
			timeVariants = concat([]string{"Closing time"}, []string{"Heure de clôture", "Heure de cloture"})
		}

		notice := Notice{
			NumberTitle: textContent(item, "strong"),
			Authority:   labelValue(item, authVariants...),
			Type:        labelValue(item, typeVariants...),
			Region:      labelValue(item, regionVariants...),
			Amount:      labelValue(item, amountVariants...),
			PublishedOn: labelValue(item, publishedVariants...),
			ClosingDate: labelValue(item, dateVariants...),
			ClosingTime: labelValue(item, timeVariants...),
		}
		notices = append(notices, notice)
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
}
