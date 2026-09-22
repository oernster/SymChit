#!/usr/bin/env bash
# Builds the SymChit macOS DMG for Apple Silicon. Run on an arm64 Mac from the repo root:
#
#   bash builddmg.sh
#
# Flow: wails build for darwin/arm64, stamp the bundle version from VERSION, codesign the
# .app (hardened runtime), notarize and staple the .app, stage it with ditto, create-dmg,
# stamp the DMG file icon, sign the DMG, then notarize and staple the DMG.
#
# Notarization is mandatory. A Developer ID signature alone is not enough: since macOS
# 10.15 Gatekeeper rejects signed-but-unnotarized apps with "Apple could not verify ... is
# free of malware". APPLE_ID and APPLE_APP_PASSWORD must both be set, else a keychain
# profile must exist, else the build stops before anything is built.
#
# Ported from PigeonPost's builddmg.sh. One difference: there is no icon generation step.
# SymChit's icons are written by tools/genicons.py and committed, so build/appicon.png is
# already there and a clone needs neither Python nor Pillow. Wails writes build/darwin
# itself on the first run for this platform, so no Info.plist is committed either; the
# version is stamped into the produced bundle below.
#
# Environment overrides:
#   DEVELOPER_ID_APPLICATION   signing identity (defaults to Oliver's Developer ID)
#   APPLE_ID, APPLE_APP_PASSWORD, APPLE_TEAM_ID   notarization credentials
#   APPLE_KEYCHAIN_PROFILE     the stored credential to use instead
#   ALLOW_UNNOTARIZED=1        build without notarizing; local testing only, never released
#
# Output: SymChit.dmg in the repo root
set -euo pipefail

APP_NAME="SymChit"
PLATFORM="darwin/arm64"
APP_BUNDLE="build/bin/${APP_NAME}.app"
DIST_DIR="dist-dmg"
STAGE_DIR="${DIST_DIR}/stage"
VERSION="$(tr -d '[:space:]' < VERSION)"
# DIST_DIR and STAGE_DIR are scratch space; the final DMG lands in the repo root.
DMG_PATH="${APP_NAME}.dmg"

DEVELOPER_ID="${DEVELOPER_ID_APPLICATION:-Developer ID Application: Oliver Ernster (W7K465GKFJ)}"
APPLE_ID="${APPLE_ID:-}"
APPLE_APP_PASSWORD="${APPLE_APP_PASSWORD:-}"
APPLE_TEAM_ID="${APPLE_TEAM_ID:-W7K465GKFJ}"
# Escape hatch for local test builds. Distribution builds must never set this: an
# unnotarized DMG is rejected by Gatekeeper on every machine but the one that signed it,
# and the failure is invisible at build time.
ALLOW_UNNOTARIZED="${ALLOW_UNNOTARIZED:-}"
# The notarization credential for this app, created once with
#   xcrun notarytool store-credentials SymChit \
#     --apple-id <id> --team-id <team> --password <app-specific>
# One profile per app means a leaked credential can be revoked for a single app. Stated
# explicitly rather than derived from APP_NAME: the profile is a fact registered with
# Apple; deriving it would silently change which credential the build looks for if
# that name were ever edited.
NOTARY_PROFILE="${APPLE_KEYCHAIN_PROFILE:-SymChit}"
# The notary service accepts only an app-specific password from appleid.apple.com and
# rejects the Apple account password with HTTP 401. The shape is distinctive, so it is
# checked before the build rather than discovered after it.
APP_SPECIFIC_PASSWORD_RE='^[a-z]{4}-[a-z]{4}-[a-z]{4}-[a-z]{4}$'
# Notarization is the default and the keychain profile always resolves, so the only way to
# skip it is to ask for that explicitly.
NOTARIZING=1
[ "${ALLOW_UNNOTARIZED}" = "1" ] && NOTARIZING=0

section() { printf '\n\033[1m== %s ==\033[0m\n' "$1"; }

# notarytool_submit uploads a file to Apple and waits for the verdict. notarytool exits
# non-zero on an Invalid verdict and set -e then fails the build, rather than leaving an
# artifact that looks distributable. Stapling is a separate step because the file that is
# submitted and the file that carries the ticket differ for a bundle: a zip goes up, the
# .app gets stapled. The password never reaches the echoed command: with a keychain
# profile it is not on the command line at all; the other branch prints a masked form.
notarytool_submit() {
    if [ -n "${APPLE_ID}" ] && [ -n "${APPLE_APP_PASSWORD}" ]; then
        echo "\$ xcrun notarytool submit $1 --apple-id ${APPLE_ID} --password ******** --team-id ${APPLE_TEAM_ID} --wait"
        xcrun notarytool submit "$1" \
            --apple-id "${APPLE_ID}" \
            --password "${APPLE_APP_PASSWORD}" \
            --team-id "${APPLE_TEAM_ID}" \
            --wait
        return
    fi
    echo "\$ xcrun notarytool submit $1 --keychain-profile ${NOTARY_PROFILE} --wait"
    xcrun notarytool submit "$1" \
        --keychain-profile "${NOTARY_PROFILE}" \
        --wait
}

# require ensures a tool is on PATH, running the given install command where it is not. The
# install command is a full shell command rather than just a brew formula, so go-installed
# tools bootstrap the same way brew-installed ones do.
require() {
    local tool="$1" install="$2"
    if ! command -v "$tool" > /dev/null 2>&1; then
        section "Installing missing tool: $tool"
        eval "$install"
    fi
    command -v "$tool" > /dev/null 2>&1 || { echo "error: $tool is required (install: $install)" >&2; exit 1; }
}

# stamp_dmg_icon gives the .dmg file itself a custom Finder icon. create-dmg only sets the
# mounted volume icon, so without this the .dmg shows the generic disk-image icon in
# Finder. Cosmetic: warn and skip where the icns or the classic resource tools are
# unavailable, rather than failing the build.
stamp_dmg_icon() {
    local dmg="$1" icns="$2" tool
    [ -f "$icns" ] || { echo "warning: ${icns} missing; DMG file icon not set" >&2; return 0; }
    for tool in sips DeRez Rez SetFile; do
        command -v "$tool" > /dev/null 2>&1 || { echo "warning: ${tool} missing; DMG file icon not set" >&2; return 0; }
    done
    local work
    work="$(mktemp -d)"
    cp "$icns" "${work}/icon.icns"
    sips -i "${work}/icon.icns" > /dev/null
    DeRez -only icns "${work}/icon.icns" > "${work}/icon.rsrc"
    Rez -append "${work}/icon.rsrc" -o "$dmg"
    SetFile -a C "$dmg"
    rm -rf "$work"
}

section "Notarization credentials"
# Checked before any build work, so a missing password costs a second rather than a full
# wails build.
if [ "${NOTARIZING}" -eq 0 ]; then
    echo "warning: ALLOW_UNNOTARIZED=1; local test build only, do not release the result" >&2
elif [ -n "${APPLE_ID}" ] && [ -n "${APPLE_APP_PASSWORD}" ]; then
    if ! [[ "${APPLE_APP_PASSWORD}" =~ ${APP_SPECIFIC_PASSWORD_RE} ]]; then
        echo "error: APPLE_APP_PASSWORD is not an app-specific password." >&2
        echo "Expected four lowercase groups of four, like abcd-efgh-ijkl-mnop." >&2
        echo "An Apple account password is rejected by the notary service with" >&2
        echo "'HTTP status code: 401. Invalid credentials'." >&2
        echo "Generate one at https://appleid.apple.com (Sign-In and Security," >&2
        echo "App-Specific Passwords). Leave both variables unset instead and store" >&2
        echo "credential in the keychain as profile ${NOTARY_PROFILE}." >&2
        exit 1
    fi
    echo "Notarizing as ${APPLE_ID} (team ${APPLE_TEAM_ID})"
else
    # The profile is checked here rather than assumed. notarytool stores it in the
    # data-protection keychain, where `security find-generic-password` cannot see it, so
    # asking notarytool itself is the only reliable probe; a history query is a couple of
    # seconds against a full wails build; it also proves the credential authenticates rather
    # than merely existing. Exit 69 is the missing-profile case.
    if ! xcrun notarytool history --keychain-profile "${NOTARY_PROFILE}" > /dev/null 2>&1; then
        echo "error: keychain profile ${NOTARY_PROFILE} is missing or its credential is rejected." >&2
        echo "Create it once with:" >&2
        echo "  xcrun notarytool store-credentials ${NOTARY_PROFILE} \\" >&2
        echo "    --apple-id <apple-id> --team-id ${APPLE_TEAM_ID} --password <app-specific>" >&2
        echo "The password is an app-specific one from https://appleid.apple.com" >&2
        echo "(Sign-In and Security, App-Specific Passwords), shaped abcd-efgh-ijkl-mnop." >&2
        echo "Set APPLE_ID and APPLE_APP_PASSWORD instead to pass it on the command line." >&2
        exit 1
    fi
    echo "Notarizing with keychain profile ${NOTARY_PROFILE}"
fi

section "Platform guard"
[ "$(uname -s)" = "Darwin" ] || { echo "error: this script must run on macOS" >&2; exit 1; }
[ "$(uname -m)" = "arm64" ] || { echo "error: this script targets Apple Silicon (arm64)" >&2; exit 1; }
command -v go > /dev/null 2>&1 || { echo "error: go is required (install from https://go.dev/dl/)" >&2; exit 1; }
# wails installs into the Go bin dir; put it on PATH so a fresh install resolves.
GOBIN_DIR="$(go env GOPATH)/bin"
case ":${PATH}:" in *":${GOBIN_DIR}:"*) ;; *) PATH="${GOBIN_DIR}:${PATH}" ;; esac
export PATH
require wails "go install github.com/wailsapp/wails/v2/cmd/wails@latest"
require npm "brew install node"
require create-dmg "brew install create-dmg"

section "Building ${APP_NAME} ${VERSION} (${PLATFORM})"
# The version reaches the binary the same way build.ps1 sends it, through -ldflags -X
# against a var: a const would be silently ignored.
wails build -clean -platform "${PLATFORM}" -ldflags "-X main.appVersion=${VERSION}"
[ -d "${APP_BUNDLE}" ] || { echo "error: ${APP_BUNDLE} was not produced" >&2; exit 1; }

section "Stamping bundle version from VERSION"
PLIST="${APP_BUNDLE}/Contents/Info.plist"
/usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString ${VERSION}" "${PLIST}" \
    || /usr/libexec/PlistBuddy -c "Add :CFBundleShortVersionString string ${VERSION}" "${PLIST}"
/usr/libexec/PlistBuddy -c "Set :CFBundleVersion ${VERSION}" "${PLIST}" \
    || /usr/libexec/PlistBuddy -c "Add :CFBundleVersion string ${VERSION}" "${PLIST}"

section "Codesigning the app bundle"
codesign --force --deep --options runtime --sign "${DEVELOPER_ID}" "${APP_BUNDLE}"
codesign --verify --deep --strict "${APP_BUNDLE}"

# Notarize the .app before it goes into the DMG. Stapling only the DMG leaves the
# copied-out .app with no local ticket, so Gatekeeper falls back to an online check and the
# app fails to launch for anyone offline or behind a restrictive network. notarytool takes
# archives only, so ditto zips the bundle first; the ticket goes on the bundle, since a zip
# cannot carry one.
if [ "${NOTARIZING}" -eq 1 ]; then
    section "Notarizing the app bundle (this waits on Apple)"
    APP_ZIP="$(mktemp -d)/${APP_NAME}.zip"
    ditto -c -k --keepParent "${APP_BUNDLE}" "${APP_ZIP}"
    notarytool_submit "${APP_ZIP}"
    xcrun stapler staple "${APP_BUNDLE}"
    # Proving the ticket is attached, not merely that stapling was attempted.
    # A notarized build with no ticket on it still launches for whoever built
    # it, then asks the network on someone else's machine and fails there. The
    # DMG gets the same check further down; this one covers the bundle a user
    # copies out of it, which is the copy they actually run.
    xcrun stapler validate "${APP_BUNDLE}"
    rm -rf "$(dirname "${APP_ZIP}")"
fi

section "Creating the DMG"
rm -rf "${DIST_DIR}"
rm -f "${DMG_PATH}"
mkdir -p "${STAGE_DIR}"
# ditto preserves the symlinks and metadata the embedded signature depends on.
ditto "${APP_BUNDLE}" "${STAGE_DIR}/${APP_NAME}.app"
VOLICON="${APP_BUNDLE}/Contents/Resources/iconfile.icns"
CREATE_DMG_ARGS=(
    --volname "${APP_NAME}"
    --window-size 540 380
    --icon-size 128
    --icon "${APP_NAME}.app" 140 190
    --app-drop-link 400 190
)
[ -f "${VOLICON}" ] && CREATE_DMG_ARGS+=(--volicon "${VOLICON}")
set +e
create-dmg "${CREATE_DMG_ARGS[@]}" "${DMG_PATH}" "${STAGE_DIR}"
STATUS=$?
set -e
# create-dmg exits 2 when it cannot set a custom window background (headless); that is
# still a good DMG.
if [ "${STATUS}" -ne 0 ] && [ "${STATUS}" -ne 2 ]; then
    echo "error: create-dmg failed with exit ${STATUS}" >&2
    exit "${STATUS}"
fi
rm -rf "${DIST_DIR}"

# The icon stamp writes a resource fork into the DMG, so it runs before signing and
# notarization. Doing it afterwards would modify a file Gatekeeper has already been told
# the hash of.
section "Stamping the DMG file icon"
stamp_dmg_icon "${DMG_PATH}" "${VOLICON}"

section "Signing the DMG"
codesign --force --sign "${DEVELOPER_ID}" "${DMG_PATH}"
codesign --verify "${DMG_PATH}"

if [ "${NOTARIZING}" -eq 1 ]; then
    section "Notarizing the DMG (this waits on Apple)"
    notarytool_submit "${DMG_PATH}"
    xcrun stapler staple "${DMG_PATH}"
    # stapler validate proves a ticket is attached; spctl replays the check Gatekeeper runs
    # on the end user's machine. Together they catch the silent case where signing
    # succeeded but notarization never happened.
    xcrun stapler validate "${DMG_PATH}"
    spctl --assess --type install -vv "${DMG_PATH}"
else
    section "Notarization skipped: unnotarized DMG, do not publish this build"
fi

section "Done"
echo "${DMG_PATH}"
echo "The record lives at ~/Library/Application Support/${APP_NAME}/symchit.db"
