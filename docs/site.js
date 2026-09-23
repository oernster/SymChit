// The site's two small jobs: the appearance toggle; and reading the newest
// release so the page can say which version it is and how big each file is.
//
// The toggle mirrors the application's own rule (FR-073). The page opens dark
// and wears light only when the reader asks for it; the choice is remembered in
// this browser. The face shown is the appearance it would move TO, so the sun
// appears while you are in the dark, exactly as the button in the window does.

const STORED = 'symdiary.site.theme'
const DARK_FACE = 'dark-mode.png'
const LIGHT_FACE = 'light-mode.png'

/** stored answers the remembered appearance, else dark. */
function stored() {
    try {
        return window.localStorage.getItem(STORED) === 'light' ? 'light' : 'dark'
    } catch {
        return 'dark'
    }
}

/** apply dresses the page and the toggle in one act, so a repaint cannot leave
 *  the button offering the appearance just departed. */
function apply(theme) {
    document.documentElement.setAttribute('data-theme', theme)
    const face = document.getElementById('theme-face')
    if (!face) {
        return
    }
    const toLight = theme === 'dark'
    face.src = toLight ? LIGHT_FACE : DARK_FACE
    face.alt = toLight ? 'Switch to light mode' : 'Switch to dark mode'
}

/** toggle moves between the two and remembers which. A browser that refuses to
 *  store it still changes appearance; it simply forgets by the next visit. */
function toggle() {
    const next = stored() === 'dark' ? 'light' : 'dark'
    try {
        window.localStorage.setItem(STORED, next)
    } catch {
        /* nothing to do: the appearance still changes for this visit */
    }
    apply(next)
}

// No version is written into this page. Every download button points at
// GitHub's releases/latest/download redirect, which always serves the newest
// release. The names carry no version, so the links cannot go stale. This
// only decorates: the version, where the release notes are and how big each
// file is. Deliberately NOT the date it was published: the site carries no
// visible dates, so a page read a year from now reads the same. One request
// feeds every part of the page that wants an answer from it; where it fails,
// what is already written stands on its own.
function decorateFromLatestRelease() {
    fetch('https://api.github.com/repos/oernster/SymDiary/releases/latest')
        .then((answer) => (answer.ok ? answer.json() : null))
        .then((release) => {
            if (!release) {
                return
            }
            showVersion(String(release.tag_name || '').replace(/^v/, ''))
            showNotes(release.html_url)
            showSizes(release.assets || [])
        })
        .catch(() => {
            /* no network, no decoration; the page is complete without it */
        })
}

/** showVersion fills the hero's pill and the download section's chip. Until
 *  there is a release the pill stays hidden rather than showing a guess, while
 *  the chip keeps the words it was written with. */
function showVersion(version) {
    if (!version) {
        return
    }
    const pill = document.getElementById('version-pill')
    if (pill) {
        pill.textContent = 'Version ' + version
        pill.style.display = 'block'
    }
    const chip = document.getElementById('dl-version')
    if (chip) {
        chip.textContent = 'Version ' + version
    }
}

/** showNotes points the release-notes link at this particular release rather
 *  than at whatever is newest when somebody follows it. */
function showNotes(url) {
    const link = document.getElementById('dl-whats-new')
    if (link && url) {
        link.href = url
    }
}

/** showSizes adds each file's size to the line naming what it is, so somebody
 *  on a slow connection knows what they are starting. */
function showSizes(assets) {
    const megabyte = 1024 * 1024
    for (const asset of assets) {
        const line = document.querySelector('[data-asset="' + asset.name + '"]')
        if (line && asset.size) {
            line.textContent = line.textContent + ' · ' + (asset.size / megabyte).toFixed(1) + ' MB'
        }
    }
}

apply(stored())
document.getElementById('theme-toggle').addEventListener('click', toggle)
decorateFromLatestRelease()
