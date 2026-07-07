---
name: pp-substack-reader
description: "Read any Substack publication as a local, full-text-searchable corpus — keyless for free posts, your own session for what you subscribe to. Trigger phrases: `archive this Substack`, `read this Substack post`, `search my Substack corpus`, `what's new in my newsletters`, `use substack-reader`, `run substack`."
author: "Maxime Delavergne"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - substack-reader-pp-cli
    install:
      - kind: go
        bins: [substack-reader-pp-cli]
        module: github.com/mvanhorn/printing-press-library/library/media-and-entertainment/substack-reader/cmd/substack-reader-pp-cli
---

# Substack — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `substack-reader-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer. It defaults binaries to `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows:
   ```bash
   npx -y @mvanhorn/printing-press-library install substack-reader --cli-only
   ```
2. Verify: `substack-reader-pp-cli --version`
3. Ensure the reported install directory is on `$PATH` for the agent/runtime that will invoke this skill.

If the `npx` install fails (no Node, offline, etc.), fall back to a direct Go install (requires Go 1.26.4 or newer). This installs into `$GOPATH/bin` (default `$HOME/go/bin`), so add that directory to `$PATH` instead:

```bash
go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/substack-reader/cmd/substack-reader-pp-cli@latest
```

If `--version` reports "command not found" after install, the runtime cannot see the binary directory on `$PATH`. Do not proceed with skill commands until verification succeeds.

Substack Reader archives whole publications into a local SQLite mirror you can search, SQL-query, and read offline. Free posts need no login; paid posts you're entitled to unlock with your own session cookie — never redistributed, always opt-in. Unlike every other Substack tool it builds a corpus that compounds instead of fetching live per call.

## When to Use This CLI

Use Substack Reader when you want a durable, searchable local copy of one or more Substack publications for reading, agent workflows, or analysis — especially reading a specific post's full text or searching across newsletters offline. It is the right tool when you value a corpus that compounds over live per-call fetching.

## Anti-triggers

Do not use this CLI for:
- Do not use it to publish, schedule, or manage a Substack you own (this is read-only) — use the Substack web app or a publishing tool.
- Do not use it to bulk-scrape or redistribute paid content you are not entitled to — it reads only your own entitled content, on demand.
- Do not use it to manage subscribers, payments, or analytics for your own publication.

## Unique Capabilities

These capabilities aren't available in any other tool for this API.

### Local corpus that compounds
- **`archive`** — Archive a whole Substack publication into a local SQLite mirror you can read, search, and query offline — no other Substack tool builds a persistent corpus.

  _Reach for this to turn a live newsletter into a durable, queryable knowledge base instead of re-fetching every time._

  ```bash
  substack-reader-pp-cli archive astralcodexten --limit 200
  ```

### Entitlement-aware reading
- **`read`** — Read a post's full text; free posts keyless, and paid posts you subscribe to via your own session cookie — with an honest 'preview only, you're not entitled' signal.

  _Use to pull a specific post's full text into an agent workflow, respecting exactly what the user is entitled to._

  ```bash
  substack-reader-pp-cli read astralcodexten/open-thread-441
  ```

### Topic & comparative intelligence
- **`digest`** — A time-windowed digest across every publication in your local corpus — what's new since you last synced, ranked, in one view.

  _Use as a personal 'what did I miss across my newsletters' briefing._

  ```bash
  substack-reader-pp-cli digest --since 7d
  ```
- **`author-compare`** — Compare two publications' cadence, topics, and free/paid mix from the local corpus.

  _Use to size up a newsletter before subscribing, or to study what a successful author publishes._

  ```bash
  substack-reader-pp-cli author-compare astralcodexten blog.bytebytego.com
  ```

## Command Reference

**categories** — Browse Substack's publication categories

- `substack-reader-pp-cli categories browse` — List publications in a category
- `substack-reader-pp-cli categories list` — List all Substack categories

**publications** — Discover Substack publications

- `substack-reader-pp-cli publications <query>` — Search Substack publications by name (best-effort; may return few results anonymously)


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
substack-reader-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

## Recipes

### Build a searchable corpus

```bash
substack-reader-pp-cli archive astralcodexten --limit 200 && substack-reader-pp-cli search "prediction markets"
```

Mirror a publication then search it offline with FTS ranking.

### Narrow a large post to fields

```bash
substack-reader-pp-cli read astralcodexten/open-thread-441 --agent --select title,post_date,audience,body_html
```

Pull only the fields an agent needs from a verbose post object.

### Audience mix analytics

```bash
substack-reader-pp-cli sql "SELECT audience, count(*) FROM posts GROUP BY audience"
```

Read-only SQL over the local corpus for arbitrary analytics.

## Auth Setup

Free/public posts are keyless — zero setup. To read paid posts you already subscribe to, provide your own Substack session cookie (substack.sid); this reads only what you are already entitled to and is never required for free content.

Run `substack-reader-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  substack-reader-pp-cli categories list --agent --select id,name,status
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag
- **Read-only** — do not use this CLI for create, update, delete, publish, comment, upvote, invite, order, send, or other mutating requests

### Response envelope

Commands that read from the local store or the API wrap output in a provenance envelope:

```json
{
  "meta": {"source": "live" | "local", "synced_at": "...", "reason": "..."},
  "results": <data>
}
```

Parse `.results` for data and `.meta.source` to know whether it's live or local. A human-readable `N results (live)` summary is printed to stderr only when stdout is a terminal AND no machine-format flag (`--json`, `--csv`, `--compact`, `--quiet`, `--plain`, `--select`) is set — piped/agent consumers and explicit-format runs get pure JSON on stdout.

## Paths and state

Agents should treat the CLI's path resolver as part of the runtime contract:

- Use `--home <dir>` for one invocation, or set `SUBSTACK_HOME=<dir>` to relocate all four path kinds under one root.
- Use per-kind env vars only when a specific kind must diverge: `SUBSTACK_CONFIG_DIR`, `SUBSTACK_DATA_DIR`, `SUBSTACK_STATE_DIR`, `SUBSTACK_CACHE_DIR`.
- Resolution order is per-kind env var, `--home`, `SUBSTACK_HOME`, XDG (`XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`), then platform defaults.
- `config` contains settings like `config.toml` and profiles. `data` contains `credentials.toml`, `data.db`, cookies, and auth sidecars. `state` contains persisted queries, jobs, and `teach.log`. `cache` contains regenerable HTTP/cache files.
- Stored secrets live in `credentials.toml` under the data dir. Existing legacy `config.toml` secrets are read for compatibility and leave `config.toml` on the first auth write.
- Run `substack-reader-pp-cli doctor --fail-on warn` to surface path and credential-location warnings. `agent-context` exposes a schema v4 `paths` block for agents that need the resolved dirs.
- For MCP, pass relocation through the MCP host config. The MCP binary does not inherit CLI flags:

  ```json
  {
    "mcpServers": {
      "substack": {
        "command": "substack-reader-pp-mcp",
        "env": {
          "SUBSTACK_HOME": "/srv/substack"
        }
      }
    }
  }
  ```

Fleet precedence: an inherited per-kind env var overrides an explicit `--home` for that kind. Use `SUBSTACK_HOME` or per-kind vars as durable fleet levers, and use `--home` only for a single invocation. Relocation is not reversible by unsetting env vars; move files manually before clearing `SUBSTACK_HOME`, or `doctor` will not find credentials left under the former root.

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
substack-reader-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
substack-reader-pp-cli feedback --stdin < notes.txt
substack-reader-pp-cli feedback list --json --limit 10
```

Entries are stored locally as `feedback.jsonl` under the resolved data dir. They are never POSTed unless `SUBSTACK_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `SUBSTACK_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

Write what *surprised* you, not a bug report. Short, specific, one line: that is the part that compounds.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename) |
| `webhook:<url>` | POST the output body to the URL (`application/json` or `application/x-ndjson` when `--compact`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

A profile is a saved set of flag values, reused across invocations. Use it when a scheduled agent calls the same command every run with the same configuration - HeyGen's "Beacon" pattern.

```
substack-reader-pp-cli profile save briefing --json
substack-reader-pp-cli --profile briefing categories list
substack-reader-pp-cli profile list --json
substack-reader-pp-cli profile show briefing
substack-reader-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `substack-reader-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/substack-reader/cmd/substack-reader-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add substack-reader-pp-mcp -- substack-reader-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which substack-reader-pp-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   substack-reader-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `substack-reader-pp-cli <command> --help`.
