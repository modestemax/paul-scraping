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

	// Helpers for label normalization (accent stripping, punctuation, spaces)
	replacer := strings.NewReplacer(
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"à", "a", "â", "a", "ä", "a",
		"î", "i", "ï", "i",
		"ô", "o", "ö", "o",
		"ù", "u", "û", "u", "ü", "u",
		"ç", "c",
		"’", "'",
		":", "",
	)
	normalize := func(s string) string {
		s = strings.ToLower(strings.TrimSpace(s))
		s = replacer.Replace(s)
		// collapse whitespace
		fields := strings.Fields(s)
		s = strings.Join(fields, " ")
		return s
	}

	notices := make([]Notice, 0, len(items))
	for _, item := range items {
		if showHTML {
			if html, err := item.InnerHTML(); err == nil {
				log.Printf("ITEM HTML:\n%s\n", strings.TrimSpace(html))
			} else {
				log.Printf("failed to get item innerHTML: %v", err)
			}
		}
		notice := Notice{NumberTitle: textContent(item, "strong")}

		// Extract all label/value cells and map via normalized label synonyms
		cells, err := item.QuerySelectorAll("div.d-table-cell")
		if err == nil && len(cells) >= 2 {
			for i := 0; i+1 < len(cells); i += 2 {
				rawLabel, _ := cells[i].TextContent()
				rawValue, _ := cells[i+1].TextContent()
				label := normalize(rawLabel)
				value := strings.TrimSpace(rawValue)

				switch label {
				case "mo/ac", "po/ca":
					if notice.Authority == "" {
						notice.Authority = value
					}
				case "type":
					if notice.Type == "" {
						notice.Type = value
					}
				case "region":
					if notice.Region == "" {
						notice.Region = value
					}
				case "montant", "amount":
					if notice.Amount == "" {
						notice.Amount = value
					}
				case "publie le", "published on", "published":
					if notice.PublishedOn == "" {
						notice.PublishedOn = value
					}
				case "date de cloture", "closing date":
					if notice.ClosingDate == "" {
						notice.ClosingDate = value
					}
				case "heure de cloture", "closing time":
					if notice.ClosingTime == "" {
						notice.ClosingTime = value
					}
				}
			}
		}
		notices = append(notices, notice)
	}
	return notices, nil
}
