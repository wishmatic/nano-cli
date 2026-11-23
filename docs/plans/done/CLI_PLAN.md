# CLI Binary Plan

## Goal

Add a `nano` binary that exposes the same NanoGPT features as the MCP server without requiring
the HTTP server or an MCP client. The CLI reuses `internal/nanogpt` for every upstream call and
`internal/resolve` for image inputs, and writes generated media to local files. The same binary
also runs the MCP server through `nano serve`, which replaces the separate `cmd/server` entry
point.

The command surface is:

```text
nano help
nano serve
nano login
nano logout
nano web search
nano web scrape
nano web scrape --firecrawl
nano web map
nano web crawl
nano web youtube | nano web yt
nano models list
nano models balance
nano gen image | nano generate image
nano gen video | nano generate video
```

## Scope

In scope:

- A new `cmd/nano` entry point and an `internal/cli` package.
- Moving the MCP server entry point to `nano serve` and removing `cmd/server`.
- Credential storage, `login`, and `logout`.
- All web, model, balance, image, and video commands listed above.
- Local output for generated images and videos.
- Unit and command-level tests.
- README documentation for installing and using the CLI.

Out of scope:

- Changing the MCP tool surface or server behavior.
- Publishing prebuilt binaries.
- An interactive TUI, shell completion installation, or a config file format beyond credentials.
- Storing generated media in `internal/filestore` or hosting it from the CLI.

## Decisions

### CLI library

Use `github.com/spf13/cobra`. It is more widely used than `urfave/cli` and gives nested commands,
aliases, help, and flag parsing without custom code. The root command enables `SilenceUsage` and
`SilenceErrors`, so runtime failures print one error rather than a usage block; Cobra's generated
`help` command covers `nano help` and `nano help <command>`.

### Credential precedence

The NanoGPT key is resolved in this order:

1. `NANOGPT_API_KEY`, for parity with the server and for CI use.
2. The credential saved by `nano login`.
3. Otherwise fail with a message telling the user to run `nano login` or set `NANOGPT_API_KEY`.

The key is never accepted as a command-line flag, because flags leak into shell history and process
listings.

### Credential backend

The chosen option is an OS keyring with a user-only file fallback:

- Primary: `github.com/zalando/go-keyring`, using Keychain on macOS, Secret Service on Linux, and
  Credential Manager on Windows. Service name `nano-mcp`, account `nanogpt-api-key`.
- Fallback: `$XDG_CONFIG_HOME/nano/api_key` or the platform equivalent from `os.UserConfigDir()`,
  with directory mode `0700`, file mode `0600`, and an atomic write.
- Default backend is `auto`: try the keyring, then fall back to the file when the keyring is
  unavailable. `NANO_CREDENTIAL_BACKEND=keyring|file|auto` can force one.
- `logout` deletes from both backends so a stale fallback cannot survive a backend change; an
  unreachable keyring does not fail a logout whose file fallback was already removed.
- `login` removes any file fallback after a successful keyring write.

Alternatives considered:

- File only: simplest and fully portable, but the key is plaintext to the account.
- Encrypted file with a passphrase: protects backups and copies, but requires a passphrase on every
  command or an environment variable, and adds more failure modes than the threat warrants.
- Environment only: no persistent storage, so `login` and `logout` would be meaningless.

The rest of this plan assumes the chosen keyring-plus-fallback behavior.

### Output model

- Data commands print readable summaries. Upstream payloads whose shape is not fixed by NanoGPT
  (search `data`, Firecrawl results, crawl pages) are printed as indented JSON so no fields are
  lost.
- Generated media is written to a local directory, not hosted. The default is the current
  directory; `--out` selects another directory. Files are written atomically with mode `0644` and
  named from the media kind, a UTC timestamp, and a short unique suffix.
- Costs reported by NanoGPT are printed next to successful results.

### Image inputs

The CLI builds the same resolver the server uses, including `IMAGE_URL_MAP`, so `--image`,
`--images`, and `--mask` accept http(s) URLs, data URIs, raw base64, and mapped local addresses.

## Design

### Package layout

```text
cmd/nano/main.go             process entry point and root command
cmd/nano/main_test.go        root command and help tests
cmd/nano/login.go            login command
cmd/nano/logout.go           logout command
cmd/nano/serve.go            MCP server command, logger, and signal handling
cmd/nano/web.go              web parent command and subcommand registration
cmd/nano/web_search.go       web search command
cmd/nano/web_scrape.go       web scrape and firecrawl scrape command
cmd/nano/web_map.go          web map command
cmd/nano/web_crawl.go        web crawl command
cmd/nano/web_youtube.go      web youtube and yt command
cmd/nano/models.go           models parent command and subcommand registration
cmd/nano/models_list.go      models list command
cmd/nano/models_balance.go   models balance command
cmd/nano/gen.go              gen/generate parent command and aliases
cmd/nano/gen_image.go        gen/generate image command
cmd/nano/gen_video.go        gen/generate video command
internal/cli/deps.go         shared command dependencies, credential lookup, and client construction
internal/cli/asset.go        local media writing and downloads
internal/cli/media.go        image resolution and data URL helpers
internal/cli/output.go       shared formatting helpers
internal/cli/flags.go        shared flag helpers
internal/credential/store.go credential interface, selection, file backend
internal/credential/keyring.go OS keyring backend
```

Every command has its own file directly under `cmd/nano/`, next to `main.go`, so the command
surface is visible without walking into an implementation package. There is no separate root
file: `main.go` builds the root command and `main_test.go` tests it. `internal/cli` keeps only
behavior shared across commands. The `cmd/nano` package depends on `internal/nanogpt`,
`internal/credential`, `internal/resolve`, `internal/sourcemap`, and `internal/server`. Only the
`serve` command starts the HTTP server, and the CLI never imports `internal/mcp` directly.

### Command tree and flags

`nano serve`

- Loads the server environment (`.env`, `API_KEY`, `NANOGPT_API_KEY`, `PUBLIC_HOST`, and the rest)
  with `internal/config` and runs `internal/server`.
- Handles `SIGINT` and `SIGTERM` with the same graceful shutdown window as the old entry point.
- This is the only way to run the MCP server; `cmd/server` is removed.

`nano web search <query...>`

- `--provider`
- `--depth`
- `--output-type`

`nano web scrape <url...>`

- `--provider`
- `--stealth`
- `--firecrawl`, which switches to Firecrawl scrape and accepts exactly one URL
- `--format` (repeatable, Firecrawl only)
- `--proxy` (Firecrawl only)
- `--only-main-content[=bool]` (Firecrawl only; pass a pointer only when the flag was changed)
- `--only-clean-content[=bool]` (Firecrawl only; same pointer rule)

`nano web map <url>`

- `--sitemap`
- `--limit`

`nano web crawl <url>`

- `--format` (repeatable)
- `--proxy`
- `--only-main-content[=bool]`
- `--only-clean-content[=bool]`
- `--sitemap`
- `--limit`
- `--crawl-entire-domain`
- `--allow-subdomains`
- `--max-discovery-depth`
- `--wait-for-finish`
- `--max-returned-pages`

`nano web youtube <url...>` and `nano web yt <url...>`

- No flags; the client enforces the documented 1 to 10 URL limit.

`nano models list [image|video]`

- If the kind is omitted, list both kinds.
- `--query`, which filters on id, name, description, owner, and tags.
- `--limit`, default 25 and max 100 per kind.

`nano models balance`

- No flags.

`nano gen image <prompt...>` and `nano generate image <prompt...>`

- `--model`
- `--n`
- `--size`
- `--image`
- `--images` (repeatable)
- `--mask`
- `--seed`
- `--strength`
- `--guidance-scale`
- `--num-inference-steps`
- `--out`

`nano gen video <prompt...>` and `nano generate video <prompt...>`

- `--model` (required; NanoGPT needs one)
- `--image`
- `--negative-prompt`
- `--duration`
- `--aspect-ratio`
- `--resolution`
- `--mode`
- `--seed`
- `--generate-audio`
- `--video-url`
- `--out`

Prompt arguments are joined with spaces, so quoted and unquoted prompts both work. An empty prompt
is rejected before any request is sent. A prompt beginning with `-` uses `--` as usual.

### Credential package

`internal/credential` exposes:

```go
var ErrNotFound = errors.New("no stored NanoGPT API key")

type Store interface {
    Load() (string, error)
    Save(key string) error
    Delete() error
}
```

`FileStore` uses `os.UserConfigDir()/nano/api_key`. `Save` trims the key, rejects empty input,
creates the directory with mode `0700`, writes a temporary file with mode `0600`, and renames it
over the target. `Load` rejects an empty or whitespace-only file as `ErrNotFound`. `Delete`
treats a missing file as success.

`KeyringStore` wraps `go-keyring`. Its `ErrNotFound` maps to `ErrNotFound`. `Open` implements the
`auto` selection and returns a store that prefers the keyring and falls back to the file.

`login` reads the key from stdin. When stdin is a terminal it prints a prompt to stderr and uses
`golang.org/x/term.ReadPassword` so the key is not echoed. When stdin is a pipe or file it reads
stdin up to a small bound and trims surrounding whitespace. The key is never printed in output or
in an error. `logout` deletes from both backends, is idempotent, and succeeds when the file
fallback is gone even if the keyring cannot be reached.

### Client construction

`internal/cli/client.go` resolves the key with the precedence above and calls
`nanogpt.New(nanogpt.Config{APIKey: key})`. The base URL remains NanoGPT's default. Tests inject a
`*nanogpt.Client` pointed at an `httptest` server rather than changing production behavior.

### Local asset writing

`internal/cli/asset.go` provides:

- `writeAsset(dir, prefix string, data []byte, mediaType string) (string, error)`, which chooses an
  extension, creates the directory, writes through a temporary file, and renames it into place.
- `downloadAsset(ctx, source string) ([]byte, string, error)`, which fetches a provider URL with a
  timeout and size cap, reads the content type, and rejects non-2xx responses.

Images normally arrive inline. If NanoGPT returns a URL instead, the CLI downloads it. Videos
always arrive as a URL and are downloaded the same way. Detection falls back to sniffing the bytes
when the response content type is generic or missing, and finally to `.mp4` for an unlabelled video.

### Error handling and exit codes

- Successful commands exit `0`.
- Usage errors and upstream failures exit `1` and print one line to stderr.
- Partial successes still exit `0` when NanoGPT answered successfully but individual entries
  failed, such as a YouTube video without captions or a scrape result with `success: false`.
- `nano logout` with nothing stored exits `0` and reports that the user is logged out.

## Implementation Units

### Unit 1: Credential storage, CLI scaffold, and serve

Add the Cobra dependency, `cmd/nano/main.go`, `cmd/nano/main_test.go`, `cmd/nano/login.go`,
`cmd/nano/logout.go`, `cmd/nano/serve.go`, `internal/cli/deps.go`, `internal/credential`, and the
`login`, `logout`, and `serve` commands. Move the old
`cmd/server/main.go` logging, signal handling, and graceful shutdown into the `serve` command,
then delete `cmd/server`. Update the Dockerfile to build `cmd/nano` and run `nano serve`. Wire
dependency injection so tests can supply stdin, stdout, stderr, a credential store, and a prebuilt
NanoGPT client.

Acceptance criteria:

- [x] `go build ./cmd/nano` produces a binary with the command tree above and `nano help` exits 0.
- [x] `nano serve` loads the server config, starts `internal/server`, and shuts down gracefully on
      `SIGINT` or `SIGTERM`.
- [x] `cmd/server` no longer exists, and no build or user-facing documentation path refers to it.
- [x] The Dockerfile builds `cmd/nano` and uses `nano serve` as its entry point.
- [x] `nano login` from a pipe stores a trimmed non-empty key; empty input is rejected.
- [x] `nano login` on a terminal does not echo the key.
- [x] The file backend creates the config directory with mode `0700` and the key file with mode
      `0600`.
- [x] `nano logout` removes the key from every reachable backend and succeeds when no key exists
      or the file fallback was used without a reachable keyring.
- [x] Command errors never include the API key.
- [x] Tests cover file save/load/delete, empty input, permissions, `ErrNotFound`, and keyring
      mapping with a mock.
- [ ] Human check: `nano login` and `nano logout` work against the real OS keyring on a desktop
      session.

### Unit 2: Web commands

Implement `web search`, `web scrape` including `--firecrawl`, `web map`, `web crawl`, and
`web youtube` with the `yt` alias. Map every flag to the existing `nanogpt` request types. Use the
pointer rule for Firecrawl booleans so an unset flag stays distinguishable from an explicit false.

Acceptance criteria:

- [x] Each command sends the expected path, method, credential header, and JSON body to a stub
      server, including provider, depth, output type, stealth, formats, proxy, sitemap, limits, and
      crawl settings.
- [x] `web scrape --firecrawl` rejects zero or multiple URLs before calling upstream and accepts
      exactly one.
- [x] `web scrape` sends 1 to 5 URLs and surfaces per-URL failures without hiding successful
      entries.
- [x] `web youtube` and `web yt` both send the transcript request and print failed entries.
- [x] Search, Firecrawl, map, and crawl output preserve unfixed upstream fields.
- [x] Tests cover success, upstream HTTP errors, malformed JSON, and flag validation.
- [ ] Human check: run one command against the real NanoGPT key and confirm the charge is shown.

### Unit 3: Model and balance commands

Implement `models list` and `models balance`. Reuse the filtering and limit semantics from
`internal/mcp/models.go`: query matches id, name, description, owner, or tags case-insensitively;
limit defaults to 25 and is capped at 100.

Acceptance criteria:

- [x] `nano models list` without a kind lists both image and video models.
- [x] `nano models list image` and `nano models list video` request the correct paths.
- [x] `--query` filters all five fields and `--limit` applies per kind without exceeding 100.
- [x] Invalid kind, negative limit, and zero limit follow the same rules as the MCP tool.
- [x] `nano models balance` prints USD, Nano, and the deposit address when present.
- [x] Tests cover filtering, limits, both kinds, and an upstream error.
- [ ] Human check: `nano models balance` shows the same balance as the MCP `check_balance` tool.

### Unit 4: Generation commands

Implement `gen image`, `generate image`, `gen video`, and `generate video`. Resolve image inputs
through `internal/resolve` with `IMAGE_URL_MAP`, build the existing `nanogpt.ImageRequest` and
`nanogpt.VideoRequest` types, write results with `internal/cli/asset.go`, and print each saved path
and the reported cost.

Acceptance criteria:

- [x] `nano gen image "prompt"` writes one file for an inline image response and prints its path.
- [x] `nano generate image "prompt" --n 2 --out <dir>` writes two files in `<dir>`.
- [x] `nano gen image` downloads an image when NanoGPT returns a URL instead of base64.
- [x] Image, multiple image, and mask inputs become `imageDataUrl`, `imageDataUrls`, and
      `maskDataUrl` with the correct media types.
- [x] `nano gen video` requires `--model`, waits for completion, downloads the video, and prints
      its path and cost.
- [x] An image passed to `gen video` sets `imageDataUrl` and defaults `mode` to `image-to-video`.
- [x] Download failures, oversized assets, non-2xx responses, empty prompts, and missing model
      produce clear errors without writing partial files.
- [x] Tests cover inline images, URL images, multiple images, masks, video download, output
      directories, and the error cases above.
- [ ] Human check: generate one image and one short video with the real key and open the files.

### Unit 5: Documentation and final verification

Update `README.md` with installation, command examples, credential precedence, backend selection,
and the security note about `NANOGPT_API_KEY` versus the stored key. Keep the README high level and
point to `nano help` for flags. Update the server's local-development and Docker examples to use
`nano serve`, and confirm the MCP behavior is unchanged.

Acceptance criteria:

- [x] `README.md` has a short CLI section covering `go install`, `nano login`, and one example per
      command group.
- [x] `go build ./...`, `go vet ./...`, and `go test ./... -race -count=1` pass.
- [x] `go build ./cmd/nano` succeeds and `go build ./cmd/server` is no longer expected to exist.
- [x] `nano help` and every `nano help <command>` path exit 0 and show the documented flags.
- [x] No credential material appears in test logs, command output, or README examples.
- [x] The Docker image still builds, its entry point is `nano serve`, and the server's `/healthz`
      and `/mcp` behavior is unchanged.

## Verification Commands

```sh
go build ./...
go vet ./...
go test ./... -race -count=1
go build -o /tmp/nano ./cmd/nano
/tmp/nano help
/tmp/nano serve --help
/tmp/nano web search --help
```

Human-only checks are marked above and require a real NanoGPT key and a desktop keyring.
