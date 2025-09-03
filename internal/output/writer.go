package output

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"paul-scraping/internal/scraper"
)

// Write marshals notices to the chosen format and writes to file or stdout.
func Write(notices []scraper.Notice, format, outputPath string) error {
	var out []byte
	switch format {
	case "json":
		b, err := json.MarshalIndent(notices, "", "  ")
		if err != nil {
			return fmt.Errorf("json marshal: %w", err)
		}
		out = b
	case "yaml", "yml", "":
		b, err := yaml.Marshal(notices)
		if err != nil {
			return fmt.Errorf("yaml marshal: %w", err)
		}
		out = b
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	if outputPath != "" {
		if err := os.WriteFile(outputPath, out, 0o644); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "wrote %d bytes to %s\n", len(out), outputPath)
		return nil
	}
	_, _ = os.Stdout.Write(out)
	return nil
}
