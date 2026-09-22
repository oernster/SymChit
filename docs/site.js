// The site's two small jobs: the appearance toggle and the version pill.
//
// The toggle mirrors the application's own rule (FR-073). The page opens dark
// and wears light only when the reader asks for it; the choice is remembered in
// this browser. The face shown is the appearance it would move TO, so the sun
// appears while you are in the dark, exactly as the button in the window does.

const STORED = 'symchit.site.theme'
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

// The version is never written into this page. It comes from the newest
// release on GitHub, so the site cannot claim a version that was never cut.
// Until there is one, the pill stays hidden rather than showing a guess.
function showVersion() {
    const pill = document.getElementById('version-pill')
    if (!pill) {
        return
    }
    fetch('https://api.github.com/repos/oernster/SymChit/releases/latest')
        .then((answer) => (answer.ok ? answer.json() : null))
        .then((release) => {
            const tag = release && release.tag_name
            if (!tag) {
                return
            }
            pill.textContent = 'Version ' + String(tag).replace(/^v/, '')
            pill.style.display = 'block'
        })
        .catch(() => {
            /* no network, no pill; the page is complete without it */
        })
}

apply(stored())
document.getElementById('theme-toggle').addEventListener('click', toggle)
showVersion()
