# `nano-cli` and MCP

A CLI tool for [NanoGPT](https://nano-gpt.com)'s data endpoints.

With just this (and an API key), you or your agent can search the web, read pages, and generate
images and videos on NanoGPT's billing.

## Getting Started

### CLI

```sh
go install github.com/wishmatic/nano-mcp/cmd/nano-cli@latest
```

As with all Go tools, make sure your Go `bin/` directory is on your `PATH`. Next:

```sh
nano-cli login
```

### MCP (via Docker)

```sh
docker run -d -p 8080:8080 \
  -e API_KEY=change-me \
  -e NANOGPT_API_KEY=change-me \
  -e PUBLIC_HOST=https://nano-mcp.example.com \
  -v nano-mcp-files:/data \
  ghcr.io/wishmatic/nano-mcp:latest
```

The API key is used by callers over HTTP; it should not be the same as the NanoGPT API key.

## Usage

The CLI also exposes the MCP server. Note that this relies on environment variables. We recommend
you use the Docker image for this.

```sh
nano-cli serve
```

Examples of other commands:

```sh
nano-cli help
nano-cli web search "go generics" --provider exa
nano-cli web scrape https://example.com
nano-cli web scrape --firecrawl https://example.com --format markdown
nano-cli web map https://example.com
nano-cli web crawl https://example.com --limit 10
nano-cli web youtube https://www.youtube.com/watch?v=...
nano-cli models list image --query nano
nano-cli models balance
nano-cli gen image "a red fox in snow" --out ./images
nano-cli gen video "waves at sunset" --model veo2-video --out ./videos
```

### Served Files

Generated images and videos are served from `/v/`, under `FILES_DIR` on disk.

### Input Addresses

Some MCP tools and the CLI `gen` commands take an image as an address: an http(s) URL, a data URI,
raw base64, or a URL this deployment maps to a local file. Problem is, that local file or even
a HTTP URL might not be accessible from where it is running.

`IMAGE_URL_MAP` allows mapping addresses as they're received by the MCP tool or in `gen` CLI calls:
comma-separated `foo=bar` pairs. For example, using [Elfu](https://github.com/wishmatic/elfu-mcp):

```sh
IMAGE_URL_MAP=https://elfu.example.com=/data/elfu/files
```

A mapped directory is readable in full by anyone who can call the tool.

### Other Notes

- The CLI saves generated media to local files.
- You may use `NANOGPT_API_KEY` to store the API key instead of `nano-cli login` but please note
  that environment variables are much easier to exfiltrate even accidentally by your agent.

## Local Development

Modify `.env`; you can copy from `.env.example`. Then:

```sh
go run ./cmd/nano-cli serve
```

## License

Licensed under the Apache License, Version 2.0, derived from
[go-mcp](https://github.com/wishmatic/go-mcp). See [LICENSE](LICENSE).

This project is not affiliated with nor endorsed by NanoGPT or any other third party.
