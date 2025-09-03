//go:build playwright

package main

import (
    "log"
    "strings"

    "github.com/playwright-community/playwright-go"
)

type Notice struct {
    NumberTitle string
    Type        string
    Authority   string
    Region      string
    Country     string
    Amount      string
    Funding     string
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
    if err := page.WaitForSelector("li.list-group-item"); err != nil {
        log.Fatalf("wait for list items: %v", err)
    }

    items, err := page.QuerySelectorAll("li.list-group-item")
    if err != nil {
        log.Fatalf("query list items: %v", err)
    }

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

    for _, item := range items {
        notice := Notice{
            NumberTitle: textContent(item, "strong"),
            Authority:   textContent(item, "div:has-text(\"PO/CA:\") + div"),
            Type:        textContent(item, "div:has-text(\"Type:\") + div"),
            Region:      textContent(item, "div:has-text(\"Region\") + div"),
            Country:     textContent(item, "div:has-text(\"Pays\") + div"),
            Amount:      textContent(item, "div:has-text(\"Amount\") + div"),
            Funding:     textContent(item, "div:has-text(\"Type de financement\") + div"),
            ClosingDate: textContent(item, "div:has-text(\"Closing date\") + div"),
            ClosingTime: textContent(item, "div:has-text(\"Closing time\") + div"),
        }
        log.Printf("%+v\n", notice)
    }
}

