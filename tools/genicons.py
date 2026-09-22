"""Generate the icons from the master artwork.

Ported from ED Voyage Companion's tools/genicons.py.

The nav-band icons: the masters in assets/ are 1254 pixels square, which is
right for artwork and wrong for a band that draws them at about seventy. Each is
trimmed to its own content, centred on a square canvas so every icon carries the
same optical weight, then written small enough to embed. donate.png sits at the repository root, where the donate skill expects it; it is
treated the same way.

The application icon: assets/application-icon.png becomes the multi-size Windows
.ico that wails build puts on the executable, which is where the taskbar button
and the shortcuts take theirs. It is also written as build/appicon.png, since
wails build writes its own logo there when the file is absent.

Run it when a master changes:

    python tools/genicons.py

It is not part of the build. The output is committed, so a clone needs neither
Python nor Pillow to build the application.
"""

from __future__ import annotations

import pathlib
import sys

try:
    from PIL import Image
except ImportError:  # pragma: no cover - a plain message beats a traceback
    sys.exit("Pillow is required: python -m pip install pillow")

# SIZE is three times the largest size the band draws an icon at, so the artwork
# stays crisp on a high-density display without carrying detail nothing shows.
SIZE = 208

# PAD keeps the trimmed artwork off the edge of its square.
PAD = 2

# APP_MASTER is the application's identity rather than a band icon.
APP_MASTER = "application-icon.png"

# DONATE_MASTER lives at the repository root.
DONATE_MASTER = "donate.png"

# ICO_SIZES are the sizes Windows chooses between, from the menu size to the
# large one Explorer uses in its biggest view.
ICO_SIZES = [(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]

# APPICON_SIZE is the square wails build expects build/appicon.png to be.
APPICON_SIZE = 1024

REPO = pathlib.Path(__file__).resolve().parent.parent
MASTERS = REPO / "assets"
OUTPUT = REPO / "frontend" / "src" / "assets" / "icons"
BUILD = REPO / "build"
ICO = BUILD / "windows" / "icon.ico"
APPICON = BUILD / "appicon.png"


def trimmed(master: pathlib.Path) -> Image.Image:
    """Open a master and crop away the transparent margin around its artwork."""
    image = Image.open(master).convert("RGBA")
    box = image.getbbox()
    return image.crop(box) if box is not None else image


def squared(image: Image.Image) -> Image.Image:
    """Centre artwork on a transparent square the size of its longer side."""
    side = max(image.width, image.height)
    square = Image.new("RGBA", (side, side), (0, 0, 0, 0))
    square.paste(image, ((side - image.width) // 2, (side - image.height) // 2), image)
    return square


def render(master: pathlib.Path, target: pathlib.Path) -> int:
    """Write one downscaled band icon; return its byte size."""
    art = trimmed(master)
    inner = SIZE - 2 * PAD
    scale = min(inner / art.width, inner / art.height)
    scaled = art.resize(
        (max(1, round(art.width * scale)), max(1, round(art.height * scale))),
        Image.LANCZOS,
    )
    canvas = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
    canvas.paste(scaled, ((SIZE - scaled.width) // 2, (SIZE - scaled.height) // 2), scaled)
    canvas.save(target, "PNG", optimize=True)
    return target.stat().st_size


def main() -> int:
    masters = sorted(p for p in MASTERS.glob("*.png") if p.name != APP_MASTER)
    donate = REPO / DONATE_MASTER
    if donate.exists():
        masters.append(donate)
    if not masters:
        sys.exit(f"no master artwork found in {MASTERS}")
    OUTPUT.mkdir(parents=True, exist_ok=True)
    for master in masters:
        written = render(master, OUTPUT / master.name)
        print(f"{master.name:<22} {master.stat().st_size:>9,} -> {written:>7,} bytes")

    app = MASTERS / APP_MASTER
    if not app.exists():
        sys.exit(f"\nno application icon at {app}")
    ICO.parent.mkdir(parents=True, exist_ok=True)
    square = squared(trimmed(app))
    square.save(ICO, "ICO", sizes=ICO_SIZES)
    print(f"\n{APP_MASTER:<22} -> {ICO.relative_to(REPO)} ({ICO.stat().st_size:,} bytes)")
    square.resize((APPICON_SIZE, APPICON_SIZE), Image.LANCZOS).save(APPICON, "PNG", optimize=True)
    print(f"{'':<22} -> {APPICON.relative_to(REPO)} ({APPICON.stat().st_size:,} bytes)")
    written = render(app, OUTPUT / APP_MASTER)
    print(f"{'':<22} -> About crest ({written:,} bytes)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
