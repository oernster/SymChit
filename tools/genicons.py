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

# DONATE_MASTER lives at the repository root, where the donate skill expects it.
DONATE_MASTER = "donate.png"

# ICO_SIZES are the sizes Windows chooses between, from the menu size to the
# large one Explorer uses in its biggest view.
ICO_SIZES = [(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]

# APPICON_SIZE is the square wails build expects build/appicon.png to be.
APPICON_SIZE = 1024

# HEADER_SIZE is the setup window's header mark, drawn at 126 pixels. Twice
# that keeps it crisp on a high-density display. The setup page has no bundler,
# so it loads the file as it finds it: shipping the master there would put a
# megabyte behind one badge.
HEADER_SIZE = 256

REPO = pathlib.Path(__file__).resolve().parent.parent
MASTERS = REPO / "assets"
OUTPUT = REPO / "frontend" / "src" / "assets" / "icons"
BUILD = REPO / "build"
ICO = BUILD / "windows" / "icon.ico"
APPICON = BUILD / "appicon.png"

# The setup program is a second Wails application, so it takes the same icon on
# its own executable, plus the header mark its page draws.
SETUP_BUILD = REPO / "installer" / "build"
SETUP_ICO = SETUP_BUILD / "windows" / "icon.ico"
SETUP_APPICON = SETUP_BUILD / "appicon.png"
SETUP_HEADER = REPO / "installer" / "frontend" / "dist" / "icon.png"


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


def mark(master: pathlib.Path, target: pathlib.Path) -> int:
    """Write the donate mark, scaled by height alone; return its byte size.

    It is a picture rather than an icon, so it does not take the squaring path
    above: a square canvas would spend the difference between its width and its
    height on nothing; the band draws it at the band's own icon height with its
    own width. SIZE is reused rather than restated, so the mark and the
    icons beside it cannot drift apart.
    """
    art = trimmed(master)
    width = max(1, round(art.width * SIZE / art.height))
    art.resize((width, SIZE), Image.LANCZOS).save(target, "PNG", optimize=True)
    return target.stat().st_size


def main() -> int:
    masters = sorted(p for p in MASTERS.glob("*.png") if p.name != APP_MASTER)
    if not masters:
        sys.exit(f"no master artwork found in {MASTERS}")
    OUTPUT.mkdir(parents=True, exist_ok=True)
    for master in masters:
        written = render(master, OUTPUT / master.name)
        print(f"{master.name:<22} {master.stat().st_size:>9,} -> {written:>7,} bytes")

    donate = REPO / DONATE_MASTER
    if not donate.exists():
        sys.exit(f"\nno donate artwork at {donate}")
    written = mark(donate, OUTPUT / DONATE_MASTER)
    print(f"{DONATE_MASTER:<22} {donate.stat().st_size:>9,} -> {written:>7,} bytes")

    app = MASTERS / APP_MASTER
    if not app.exists():
        sys.exit(f"\nno application icon at {app}")
    square = squared(trimmed(app))
    for target in (ICO, SETUP_ICO):
        target.parent.mkdir(parents=True, exist_ok=True)
        square.save(target, "ICO", sizes=ICO_SIZES)
        print(f"\n{APP_MASTER:<22} -> {target.relative_to(REPO)} ({target.stat().st_size:,} bytes)")
    large = square.resize((APPICON_SIZE, APPICON_SIZE), Image.LANCZOS)
    for target in (APPICON, SETUP_APPICON):
        large.save(target, "PNG", optimize=True)
        print(f"{'':<22} -> {target.relative_to(REPO)} ({target.stat().st_size:,} bytes)")
    square.resize((HEADER_SIZE, HEADER_SIZE), Image.LANCZOS).save(
        SETUP_HEADER, "PNG", optimize=True
    )
    print(f"{'':<22} -> {SETUP_HEADER.relative_to(REPO)} ({SETUP_HEADER.stat().st_size:,} bytes)")
    written = render(app, OUTPUT / APP_MASTER)
    print(f"{'':<22} -> About crest ({written:,} bytes)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
