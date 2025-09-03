//go:build playwright

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/playwright-community/playwright-go"
)

type Notice struct {
	NumberTitle string
	Type        string
	Authority   string
	Region      string
	Amount      string
	PublishedOn string
	ClosingDate string
	ClosingTime string
}

func main() {
	pw, err := playwright.Run()
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

	for _, item := range items {
		// Log the raw HTML of the item before parsing
		if html, err := item.InnerHTML(); err == nil {
			log.Printf("ITEM HTML:\n%s\n", strings.TrimSpace(html))
		} else {
			log.Printf("failed to get item innerHTML: %v", err)
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
		log.Printf("%+v\n", notice)
	}
}
