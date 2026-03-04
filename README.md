# grantelder.com v0

A minimal Go tool that reads markdown files and publishes a static site.

## Usage

**Publish** (outputs to `docs/` for GitHub Pages):

```
go run . 
# or
./grantelder
```

**Serve locally** (dev):

```
go run . -serve :8080
```

**Options**

- `-content .` — directory containing markdown files (default: current dir)
- `-output docs` — output directory for HTML (default: docs)
- `-serve :8080` — run HTTP server instead of publishing

## Deploy to grantelder.com

1. Push `docs/` to GitHub
2. Enable GitHub Pages (Settings → Pages → Source: main branch, /docs folder)
3. Or use any static host (Netlify, Cloudflare Pages, etc.)
