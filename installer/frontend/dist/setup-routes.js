/* ----------------------------------------------------------------- routes */

// One reading of the machine decides the screen, its heading, its options and
// its buttons. Deciding it once is what stops those four drifting apart.

function routeInstall(state) {
    $('install-title').textContent = `Install ${appName} ${state.thisVersion}`
    $('install-path').textContent = state.installDir
    const read = renderOptions($('install-options'), shortcutOptions(state).concat([
        launchOption(),
    ]))
    showScreen('install')
    setFooter([
        {label: 'Cancel', onClick: () => backend().Quit()},
        {
            label: 'Install', kind: 'primary',
            onClick: () => install(read, `Installing ${appName}`,
                `${appName} is installed`,
                'You are on v' + state.thisVersion + '.'),
        },
    ])
}

// routeChange serves both directions of a version change: an update and a way
// back differ only in wording and in which button is the safe one. Neither
// heading names a single version, because both are about two; the flow line
// below carries them.
function routeChange(state) {
    const goingBack = state.relation === 'older'
    $('update-title').textContent = goingBack ? 'Go back a version?' : 'Update available'
    $('update-lead').textContent = goingBack
        ? 'This setup file carries an older version than the one installed. Your symptom record is untouched.'
        : 'A newer version is ready to install. Your symptom record is untouched.'
    $('update-from').textContent = 'v' + state.installedVersion
    $('update-to').textContent = 'v' + state.thisVersion
    const read = renderOptions($('update-options'), shortcutOptions(state).concat([
        launchOption(),
    ]))
    showScreen('update')
    setFooter([
        {label: 'Uninstall', kind: 'danger', onClick: () => routeUninstall(state)},
        {label: 'Not now', onClick: () => backend().Quit()},
        {
            label: goingBack ? 'Go back' : 'Update', kind: 'primary',
            onClick: () => install(read,
                goingBack ? 'Going back a version' : `Updating ${appName}`,
                goingBack ? 'Version changed' : `${appName} is updated`,
                'You are on v' + state.thisVersion + '.'),
        },
    ])
}

// routeManage is the screen for a matching version. Its boxes act immediately,
// since there is nothing to install: a box that waited for a go-ahead would
// never take effect at all.
function routeManage(state) {
    $('manage-title').textContent = `${appName} ${state.installedVersion} is installed`
    const live = () => backend().SetShortcuts(read('startMenu'), read('desktop'))
    const read = renderOptions($('manage-options'), [
        {
            key: 'startMenu', label: 'Add a Start Menu entry',
            checked: state.startMenu, onChange: () => live(),
        },
        {
            key: 'desktop', label: 'Add a Desktop shortcut',
            checked: state.desktop, onChange: () => live(),
        },
        launchOption(),
    ])
    showScreen('manage')
    setFooter([
        {label: 'Uninstall', kind: 'danger', onClick: () => routeUninstall(state)},
        {label: 'Close', onClick: () => backend().Quit()},
        {
            label: 'Reinstall', onClick: () => finish(
                () => backend().Install(freshChoices), read('launch'),
                `Reinstalling ${appName}`, `${appName} is reinstalled`,
                'The files were written again and the shortcuts put back as a new install would leave them.'),
        },
        {
            label: 'Repair', kind: 'primary',
            onClick: () => finish(
                () => backend().Repair(), read('launch'),
                `Repairing ${appName}`, 'Repair complete',
                'The files have been put back and nothing else was changed.'),
        },
    ])
}

// routeUninstall is reachable from every other screen. The record is kept
// unless the box is ticked; the box names the file it would delete.
function routeUninstall(state) {
    $('record-path').textContent = state.recordFile
    const read = renderOptions($('uninstall-options'), [
        {
            key: 'record', label: 'Also delete my symptom record',
            hint: 'Everything you have recorded goes with it. This cannot be undone,'
                + ' so export it first if you want to keep a copy.',
            checked: false,
        },
    ])
    showScreen('uninstall')
    setFooter([
        {label: 'Cancel', onClick: () => state.installed ? route(state) : backend().Quit()},
        {
            label: 'Uninstall', kind: 'danger',
            onClick: () => withAppClosed(() => run(
                () => backend().Uninstall(read('record')),
                `Removing ${appName}`, `${appName} is removed`,
                read('record')
                    ? 'The application, its shortcuts and your record are gone.'
                    : 'The application and its shortcuts are gone. Your record is still there.')),
        },
    ])
}

function route(state) {
    currentState = state
    if (state.mode === 'uninstall') {
        routeUninstall(state)
    } else if (state.mode !== 'manage') {
        routeInstall(state)
    } else if (state.relation === 'same') {
        routeManage(state)
    } else {
        routeChange(state)
    }
}

async function init() {
    applyTheme('light')
    let tries = 0
    while (!backend() && tries < 100) {
        await new Promise((resolve) => setTimeout(resolve, 50))
        tries++
    }
    if (!backend()) {
        $('error-msg').textContent = 'Could not reach the setup program.'
        showScreen('error')
        return
    }
    window.runtime.EventsOn('progress', onProgress)
    const state = await backend().DetectState()
    appName = state.appName
    document.title = `${appName} Setup`
    $('brand').textContent = `${appName} Setup`
    $('uninstall-title').textContent = `Remove ${appName}?`
    $('running-title').textContent = `${appName} is open`
    applyTheme(state.prefersDark ? 'dark' : 'light')
    route(state)
    window.focus()
    focusFooter()
}

// A missing mark leaves no broken image in the header.
const markImage = $('markimg')
markImage.onerror = () => { markImage.remove() }

window.addEventListener('DOMContentLoaded', init)
