package scraper

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// Scrape uses an existing Playwright instance to visit the URL and extract notices.
func Scrape(ctx context.Context, pw *playwright.Playwright, url string, showHTML bool) ([]Notice, error) {
	browser, err := pw.Chromium.Launch()
	if err != nil {
		return nil, fmt.Errorf("could not launch browser: %w", err)
	}
	defer browser.Close()

	context, err := browser.NewContext()
	if err != nil {
		return nil, fmt.Errorf("new context: %w", err)
	}
	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("new page: %w", err)
	}

	if _, err = page.Goto(url); err != nil {
		return nil, fmt.Errorf("goto: %w", err)
	}
	if _, err := page.WaitForSelector(".list-group"); err != nil {
		return nil, fmt.Errorf("wait for list group: %w", err)
	}

	container, err := page.QuerySelector(".list-group")
	if err != nil || container == nil {
		return nil, fmt.Errorf("find list group: %w", err)
	}

	items, err := container.QuerySelectorAll("li.list-group-item")
	if err != nil {
		return nil, fmt.Errorf("query list items: %w", err)
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

	notices := make([]Notice, 0, len(items))
	for _, item := range items {
		if showHTML {
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
	return notices, nil
}
