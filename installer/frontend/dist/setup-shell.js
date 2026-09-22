const $ = (id) => document.getElementById(id)

// appName is the product's name, read from the setup program rather than
// written here. This page has no build step, so a name typed into it would
// survive a rename with nothing to say so.
let appName = ''

// currentState is the last reading of the machine, kept so a screen that goes
// back can return to the one that was due.
let currentState = null

function backend() {
    return window.go && window.go.main && window.go.main.App
}

/* ------------------------------------------------------------------ theme */

// applyTheme follows the Windows light or dark setting, read once by setup.
// There is no toggle, because SymChit itself has none.
function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme)
}

/* ---------------------------------------------------------------- screens */

function showScreen(name) {
    document.querySelectorAll('.screen').forEach((el) => el.classList.remove('active'))
    $('screen-' + name).classList.add('active')
}

// setFooter rebuilds the footer for the screen now showing. It is never one
// fixed row relabelled as it goes: a relabelled row has to remember what it
// used to mean, which is how a go-ahead button keeps the styling of a safe
// action while doing a destructive one.
function setFooter(buttons) {
    const footer = $('footer')
    footer.innerHTML = ''
    buttons.forEach((spec) => {
        const el = document.createElement('button')
        el.className = 'btn' + (spec.kind ? ' ' + spec.kind : '')
        el.textContent = spec.label
        el.onclick = spec.onClick
        footer.appendChild(el)
    })
    focusFooter()
}

// focusFooter puts focus on the button a screen leads with, so Enter does the
// obvious thing and the ring says where it would land.
function focusFooter() {
    const footer = $('footer')
    const first = footer.querySelector('.btn.primary') || footer.querySelector('.btn')
    if (first) first.focus()
}

/* ---------------------------------------------------------------- options */

// renderOptions fills a container with checkboxes and answers a reader for
// their values, so no screen has to know the ids of its own boxes.
function renderOptions(container, specs) {
    container.innerHTML = ''
    const boxes = {}
    specs.forEach((spec) => {
        const label = document.createElement('label')
        label.className = 'option'
        const input = document.createElement('input')
        input.type = 'checkbox'
        input.checked = !!spec.checked
        if (spec.onChange) input.onchange = () => spec.onChange(input.checked)
        const tick = document.createElement('span')
        tick.className = 'check'
        const text = document.createElement('span')
        const title = document.createElement('span')
        title.className = 'label'
        title.textContent = spec.label
        text.appendChild(title)
        if (spec.hint) {
            const hint = document.createElement('span')
            hint.className = 'hint'
            hint.textContent = spec.hint
            text.appendChild(hint)
        }
        label.append(input, tick, text)
        container.appendChild(label)
        boxes[spec.key] = input
    })
    return (key) => boxes[key].checked
}

// freshChoices are what a first install applies; a reinstall puts them back. That is the whole of the difference from a repair: a repair leaves
// every choice alone.
const freshChoices = {startMenu: true, desktop: true}

// launchOption finishes every screen that writes files. Setup's job is done
// once SymChit is running, so the same tick that starts it closes setup.
function launchOption() {
    return {
        key: 'launch',
        label: `Start ${appName} and close setup when this finishes`,
        checked: true,
    }
}

// shortcutOptions open on what is already true, never all ticked: a user who
// declined a desktop shortcut is not asked to decline it again.
function shortcutOptions(state) {
    return [
        {
            key: 'startMenu', label: 'Add a Start Menu entry',
            hint: 'Find it by typing its name in the Start Menu.',
            checked: state.installed ? state.startMenu : true,
        },
        {
            key: 'desktop', label: 'Add a Desktop shortcut',
            checked: state.installed ? state.desktop : true,
        },
    ]
}

/* ------------------------------------------------------------------- work */

function onProgress(p) {
    $('progress-fill').style.width = p.pct + '%'
    $('progress-status').textContent = p.msg
}

// run moves to the progress screen, which offers no actions at all: a screen
// with nothing safe to offer offers nothing. Every path ends in a verdict.
async function run(work, title, doneTitle, doneMsg) {
    $('progress-title').textContent = title
    $('progress-fill').style.width = '0'
    $('progress-status').textContent = 'Starting...'
    setFooter([])
    showScreen('progress')
    try {
        await work()
        $('done-title').textContent = doneTitle
        $('done-msg').textContent = doneMsg
        showScreen('done')
        setFooter([{label: 'Close', kind: 'primary', onClick: () => backend().Quit()}])
    } catch (e) {
        showError(String(e))
    }
}

// finish runs one piece of work, then honours the launch tick. A successful
// launch closes setup; a failed one leaves the error on screen, so setup never
// disappears having quietly failed.
function finish(work, wanted, title, doneTitle, doneMsg) {
    return withAppClosed(() => run(
        () => work().then(() => {
            if (wanted) return backend().LaunchApp().then(() => backend().Quit())
        }),
        title, doneTitle, doneMsg + (wanted ? ' It is starting now.' : ''),
    ))
}

function showError(message) {
    $('error-msg').textContent = message
    showScreen('error')
    setFooter([{label: 'Close', kind: 'primary', onClick: () => backend().Quit()}])
}

// withAppClosed runs the work once SymChit is not running. If it is open, the
// offer to close it comes first, before any file is touched, rather than
// failing later on a locked executable.
async function withAppClosed(proceed) {
    if (!(await backend().AppRunning())) {
        proceed()
        return
    }
    showScreen('running')
    setFooter([
        {label: 'Cancel', onClick: () => route(currentState)},
        {
            label: 'Close it and continue', kind: 'primary', onClick: async () => {
                setFooter([])
                try {
                    await backend().CloseRunningApp()
                } catch (e) {
                    showError(String(e))
                    return
                }
                proceed()
            },
        },
    ])
}

// install runs one write of the files, whatever the screen calls it.
function install(read, title, doneTitle, doneMsg) {
    const launchAfter = read('launch')
    const choices = {startMenu: read('startMenu'), desktop: read('desktop')}
    return withAppClosed(() => run(
        () => backend().Install(choices).then(() => {
            if (launchAfter) return backend().LaunchApp().then(() => backend().Quit())
        }),
        title, doneTitle,
        doneMsg + (launchAfter ? ' It is starting now.' : ''),
    ))
}
