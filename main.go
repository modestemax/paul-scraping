package main

import (
        "fmt"
        "html"
        "io"
        "log"
        "net/http"
        "regexp"
        "strings"
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

func parseNotice(item string) Notice {
	unescape := func(s string) string {
		return strings.TrimSpace(html.UnescapeString(s))
	}
	get := func(pattern string) string {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(item)
		if len(match) >= 2 {
			return unescape(match[1])
		}
		return ""
	}
        getAny := func(patterns ...string) string {
                for _, p := range patterns {
                        if v := get(p); v != "" {
                                return v
                        }
                }
                return ""
        }
        return Notice{
                NumberTitle: get(`<strong[^>]*>(?s)(.*?)</strong>`),
                Authority:   get(`PO/CA:\s*</div>\s*<div class="d-table-cell">\s*(.*?)\s*</div>`),
                Type:        get(`Type:\s*</div>\s*<div class="d-table-cell">\s*(.*?)\s*</div>`),
                Region:      get(`Region\s*:\s*</div>\s*<div class="d-table-cell">\s*(.*?)\s*</div>`),
                Country:     getAny(`Pays\s*:\s*</div>\s*<div class="d-table-cell">\s*(.*?)\s*</div>`, `Country\s*:\s*</div>\s*<div class="d-table-cell">\s*(.*?)\s*</div>`),
                Amount:      get(`Amount\s*:\s*</div>\s*<div class="d-table-cell[^>]*>\s*(.*?)\s*</div>`),
                Funding:     getAny(`Type de financement\s*:\s*</div>\s*<div class="d-table-cell">\s*(.*?)\s*</div>`, `Financing Type\s*:\s*</div>\s*<div class="d-table-cell">\s*(.*?)\s*</div>`),
                ClosingDate: get(`Closing date\s*:\s*</div>\s*<div class="d-table-cell">\s*(.*?)\s*</div>`),
                ClosingTime: get(`Closing time\s*:\s*</div>\s*<div class="d-table-cell">\s*(.*?)\s*</div>`),
        }
}

func main() {
        url := "https://www.armp.cm/recherche_avancee?recherche_avancee_do=1&reference_avis=&maitre_ouvrage=0&region=0&departement=0&type_publication%5B%5D=AO"
        resp, err := http.Get(url)
        if err != nil {
                log.Fatalf("request: %v", err)
        }
        defer resp.Body.Close()
        body, err := io.ReadAll(resp.Body)
        if err != nil {
                log.Fatalf("read body: %v", err)
        }
        liRe := regexp.MustCompile(`(?s)<li[^>]*class="list-group-item[^"]*".*?</li>`)
        items := liRe.FindAllString(string(body), -1)
        for _, item := range items {
                n := parseNotice(item)
                fmt.Printf("%+v\n\n", n)
        }
}
