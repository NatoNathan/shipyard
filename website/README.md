# Shipyard Website

The project website and docs, built with [Astro](https://astro.build) and Tailwind CSS 4.

The docs section is generated from the repo's top-level [`docs/`](../docs) folder via an
Astro content collection — edit the markdown there and the site picks it up. Relative
`.md` links between docs are rewritten to page URLs at build time, so keep authoring
links the GitHub-friendly way.

## Development

```bash
pnpm install
pnpm dev          # http://localhost:4321/shipyard/
pnpm build        # outputs to dist/
pnpm preview      # serve the production build
```

## Deployment

Deployed to GitHub Pages by `.github/workflows/deploy-website.yml` on pushes to `main`
that touch `website/` or `docs/`. The site is served at `/shipyard/` (configured via
`base` in `astro.config.mjs`).
