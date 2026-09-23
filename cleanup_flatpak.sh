#!/usr/bin/env bash
# Uninstalls and purges the SymDiary Flatpak. Run from the repo root:
#
#   bash cleanup_flatpak.sh
#
# Scoped to flatpak artefacts only. It deliberately does NOT touch what the other build
# paths produce (build/, dist-installer, dist-dmg), so the three stay independent. It
# never touches the user's record: removing a symptom record is something the user asks
# for, not something a cleanup script decides.
set -euo pipefail

APP_ID="uk.codecrafter.SymDiary"
BIN_NAME="symdiary"

bold=$(tput bold 2>/dev/null || true)
reset=$(tput sgr0 2>/dev/null || true)
section() { echo; echo "${bold}=== $* ===${reset}"; }

section "Uninstalling ${APP_ID}"
if flatpak list --user | grep -q "${APP_ID}"; then
    flatpak uninstall --user -y "${APP_ID}"
    echo "  Uninstalled."
else
    echo "  Not installed, skipping."
fi

section "Removing flatpak build artefacts"
rm -f "${BIN_NAME}.flatpak"
rm -rf .flatpak-build .flatpak-repo .flatpak-builder
rm -f "${APP_ID}.yml"
rm -rf packaging/
echo "  Done."

echo
echo "${bold}Purge complete.${reset}"
echo "Your record was left alone: ~/.var/app/${APP_ID}/config/SymDiary/symdiary.db"
