# Pausa landing page

A dependency-free static landing page for Pausa. Plain HTML + CSS (+ a few lines of vanilla JS for the copy buttons) — no build step.

```
site/
├── index.html
├── styles.css
└── assets/        # optional: demo.mp4 / demo.gif / posters
```

## Add the demo (highest-impact improvement)

The hero of a landing page is usually a short screen recording. Drop a 20–30s
clip into `site/assets/` and swap the placeholder block in `index.html` for:

```html
<video class="demo" autoplay muted loop playsinline
       poster="assets/demo-poster.png">
  <source src="assets/demo.mp4" type="video/mp4">
</video>
```

Good clip sequence: a call starts → countdown auto-pauses → fullscreen break
appears → focus returns to the previous app. Keep it under ~4 MB.

## Preview locally

```bash
cd site
python3 -m http.server 8080
# open http://localhost:8080
```

## Deploy

The page deploys automatically to **GitHub Pages** via
`.github/workflows/pages.yml` on every push to `main` that touches `site/`.

One-time setup: **Settings → Pages → Build and deployment → Source =
"GitHub Actions"**.

The site will be available at:

```
https://<owner>.github.io/pausa/
```

To use a custom domain, add a `CNAME` file to this directory containing the
domain, and point a DNS `CNAME` record at `<owner>.github.io`.

## Optional analytics

The app itself is telemetry-free, but you may want privacy-friendly page
analytics. A commented GoatCounter snippet is included at the bottom of
`index.html` — uncomment and set your code if you want it.

## Notes

- Images are referenced from `raw.githubusercontent.com` so the repo stays
  small. If you prefer self-contained assets, copy optimized copies into
  `site/assets/` and update the `src` attributes.
- Once the site is live, add a link to it near the top of the root `README.md`.
