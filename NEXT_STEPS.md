Context
- Current scraper uses a structured DOM approach: for each `li.list-group-item`, it reads `div.d-table-cell` in pairs (label, value), normalizes the label (lowercase, trim, remove `:`, strip accents), and maps to fields.
- Extracted fields: number_title (from `strong`), authority (MO/AC|PO/CA), type, region, amount, published_on, closing_date, closing_time.
- CLI: `-format`, `-output`, `-show-html`, `-auto-install`, `-url`, `-pause` (Windows) + corresponding env vars. Auto-install of Playwright driver/browsers is enabled by default.

Proposed Next Steps
1) Label Discovery Mode (list view)
   - Add a flag `-discover-labels` that logs, per item, the list of normalized labels encountered and a global frequency summary at the end.
   - Purpose: quickly audit what labels exist on the site without scanning HTML manually.

2) Detail Page Scraping (enrichment)
   - Follow the detail link for each item (e.g., main anchor within the list item).
   - Extract additional fields likely present on the detail page (to be confirmed on the site):
     - department (département/department), procedure, nature, funding (type de financement/funding type), execution_period, town/city, opening date/time, attachments (files), publication type, etc.
   - Add options: `-max-workers` (concurrency), `-timeout` per page, and polite delay/rate-limiting.
   - Always include `detail_url` in results when followed.

3) Configurable Synonyms
   - Move label normalization/mapping to a configurable YAML file (e.g., `labels.yml`), loaded at startup.
   - Allows adding/removing synonyms without recompilation.

4) Amount Normalization
   - Parse `amount` to extract numeric value (remove spaces, NBSP) and detect currency (e.g., FCFA → XAF) and store both raw and normalized forms.

5) Testing & Samples
   - Unit tests: `internal/config` (flag/env precedence), `internal/output` (JSON/YAML serialization), label normalization helpers.
   - Integration-like test using saved HTML snapshots of list items to validate parsing without hitting the network.

6) Resilience & Controls
   - Timeouts for navigation and selectors; retry transient failures.
   - Headless/headful toggle; optional screenshot on parse error for debugging.
   - Structured logging and log levels (e.g., `LOG_LEVEL=info|debug`).

7) Windows UX
   - Keep current auto-pause behavior when launched via Explorer.
   - Optional `.cmd` wrapper template users can double-click.

How We’ll Continue Next Session
- Start with Step 1 (label discovery mode) to confirm the exact labels on the live site.
- If needed, implement Step 2 (detail page scraping) to fetch additional fields.
- Then implement Step 3 (configurable synonyms) for maintainability.
- Finally, add amount normalization and tests.

