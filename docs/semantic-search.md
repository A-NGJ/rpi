# Semantic Search

`rpi serve` exposes an `rpi_search` MCP tool that returns ranked, semantically relevant `.rpi/` artifacts for a natural-language query. This page covers installation, warmup, runtime behavior, the status contract, and what happens when the backend is unavailable.

## One-time setup

```bash
npm install -g @tobilu/qmd      # or: bun install -g @tobilu/qmd
rpi search --warmup             # spawns qmd's HTTP MCP daemon and downloads ~2 GB of GGUF models
```

The warmup is a one-time cost. qmd caches models in `~/.cache/qmd/models/`, so subsequent sessions reuse the same files.

## Warm vs cold behavior

Once warmed, the daemon keeps models loaded in VRAM and subsequent searches are fast. An internal debounce skips redundant index refreshes for back-to-back skill calls. Expect bimodal latency: most queries return in milliseconds; the call right after a batch of writes pays a re-index cost (and embed inference if files changed).

## Automatic warmup from skills

When qmd is installed but cold (daemon down or models not yet downloaded), the seven calling skills — `rpi-research`, `rpi-propose`, `rpi-plan`, `rpi-verify`, `rpi-explain`, `rpi-diagnose`, `rpi-spec-sync` — run `rpi search --warmup` on the user's behalf and retry the query. The user sees the install hint only when qmd is genuinely missing, not on transient first-run states.

## Status contract

`rpi_search` returns one of four states, each with an actionable hint so the failure mode is unambiguous:

- **`ok`** — hits returned.
- **`empty`** — query ran but no matches scored above the relevance floor.
- **`backend_error`** — qmd is installed but failing (daemon down, models missing, parse error). The hint names the stage that failed.
- **`backend_unavailable`** — qmd is not on PATH. The hint points to the install command.

## Without qmd

When qmd is not installed, `rpi_search` returns `backend_unavailable` with an install hint, and the calling skills fall back to `rpi_scan` plus keyword grep automatically. RPI ships fully functional without qmd; semantic search is a strict upgrade, not a requirement.
