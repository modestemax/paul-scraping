package scraper

// Notice represents a single tender/notice item scraped from the page.
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
