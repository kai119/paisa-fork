# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Paisa

Paisa is a personal finance manager built on top of [Ledger](https://www.ledger-cli.org/) (and optionally HLedger or Beancount). It provides a web UI and CLI for tracking investments, expenses, income, net worth, capital gains, and budgets — with India-specific features like mutual fund NAV scraping, NPS, and CII data.

## Build & Development Commands

```bash
# Development (Go server + Svelte dev server in parallel)
make develop

# Production build (JS + Go)
make build

# Install (build + go install)
make install

# Lint (prettier + svelte-check + gofmt)
npm run lint

# Type check only
npm run check

# Format code
npm run format
```

### Testing

```bash
# Full test suite
make test

# JS/TS unit tests only (Bun)
bun test --preload ./src/happydom.ts src

# Single JS test file
bun test src/lib/bulk_edit.test.ts

# Go tests only
go test ./...

# Specific Go package
go test ./internal/ledger

# Specific Go test
go test ./internal/ledger -run TestName

# Regression/integration tests (requires compiled ./paisa binary)
TZ=UTC bun test tests

# Regenerate regression test fixtures
make regen
```

### Parser Development

```bash
# Rebuild Lezer language parsers (sheet + query DSLs)
npm run parser-build

# Build parsers with debug names
npm run parser-build-debug
```

## Architecture

### Tech Stack

- **Backend**: Go 1.24 with Gin (HTTP), GORM (SQLite ORM), Cobra (CLI)
- **Frontend**: SvelteKit + TypeScript, Vite build, D3.js visualizations, CodeMirror editor
- **Testing**: Bun test runner (JS), standard Go testing, regression tests with fixture snapshots
- **Desktop**: Wails (optional, in `desktop/`)

### How It Works

1. A YAML config (`paisa.yaml`) specifies the user's ledger file path and options
2. Go backend invokes ledger/hledger/beancount CLI to parse transactions, then caches results in SQLite
3. Gin REST API serves computed data (allocations, portfolio, gains, income, etc.) as JSON
4. SvelteKit frontend fetches from the API and renders D3-based charts and Tabulator data tables
5. A CodeMirror-based editor with a custom Lezer grammar allows editing ledger files in-browser

### Key Directory Layout

```
internal/
  accounting/     # Double-entry accounting logic, P&L calculations
  ledger/         # Ledger file parsing (invokes ledger/hledger/beancount)
  model/          # Data models: transaction, posting, commodity, price, portfolio
  server/         # Gin REST API handlers
  scraper/        # NAV/price data scrapers (mutual funds, metals, India CII)
  config/         # YAML config parsing
  query/          # Query engine for filtering postings
  prediction/     # TF-IDF transaction tagging predictions
  xirr/           # XIRR (Extended IRR) calculations
  taxation/       # Tax calculations
  generator/      # Transaction generation from templates

src/
  routes/         # SvelteKit pages
  lib/            # Frontend business logic (allocation, portfolio, gain, etc.)
  lib/components/ # Reusable Svelte components
  lib/sheet/      # Lezer grammar for spreadsheet-like expressions
  lib/search/     # Lezer grammar for transaction search queries

tests/
  fixture/        # Integration test datasets (main, eur-hledger, bulk_edit, etc.)
  regression.test.ts  # Spins up paisa server, hits API, diffs against snapshots

notes/            # Notes and research created during development sessions
```

### Custom DSLs

Two Lezer-based parsers are maintained:
- **Sheet language** (`src/lib/sheet/`): spreadsheet-like expression language for custom reports
- **Search/query language** (`src/lib/search/`): transaction filtering queries

After editing `.grammar` files, run `npm run parser-build` to regenerate the parsers.

### Regression Tests

Regression tests in `tests/` spawn the compiled `./paisa` binary, call real API endpoints, and compare responses to JSON snapshots in `tests/fixture/*/`. The server runs on port 5700 with `TZ=UTC` and a fixed date (`2022-02-07`). Run `make regen` to regenerate snapshots when API responses intentionally change.

### Frontend Build Output

`npm run build` emits to `web/static/`. The Go binary embeds this directory via `//go:embed`, so you must build the frontend before building the Go binary for production.
