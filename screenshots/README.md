# Screenshots

Images used by the README and the landing page (`site/`). All are referenced
by filename, so replacing a file here updates both.

| File | Shows |
|---|---|
| `dashboard.png` | Dashboard at rest — next-break countdown and the stats row |
| `preferences-breaks.png` | Preferences → **Breaks** |
| `preferences-detection.png` | Preferences → **Detection** |
| `preferences-experience.png` | Preferences → **Experience** |
| `preferences-general.png` | Preferences → **General** |
| `break-overlay.png` | Fullscreen break overlay |
| `og-image.png` | 1200×630 social preview (`og:image`) |

## How these were captured

Consistency matters more than any single shot, so keep to these specs:

- **App state:** default configuration (accent color **sky `#0ea5e9`** — the
  built-in default). A non-default accent reads as a different app next to the
  others.
- **Window:** the main window is 640×720 (frameless, no title bar). Capture the
  window, not the whole screen — no desktop, no menu bar, no Dock.
- **Display scale:** capture on a Retina display at the default scaled
  resolution so the PNGs come out at 2× (~1280×1440).
- **Method:** `⌘⇧4` then **Space**, click the window (or `screencapture -o -w`).
  A window capture includes a drop shadow; that is fine — it gets cropped.
- **Preferences:** capture each of the four sidebar sections in turn.
- **Overlay:** trigger a break and capture the fullscreen overlay. For the hero
  image a **long** break is preferred (it includes the breathing guide).

## Post-processing

Raw captures are cropped to the panel, downscaled to ≤1600 px on the long edge,
and saved as optimized PNGs. `og-image.png` is a 1200×630 center-crop of the
overlay. Any tool that does this is fine (e.g. Pillow, ImageMagick, `sips`).

After changing the app icon source (`build/icons/appicon.svg`), run
`make icons` to regenerate `build/appicon.png` — the README hero and the site
favicon both read that file.
