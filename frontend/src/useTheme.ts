// The DOM half of the theme: what the page is wearing, plus remembering it.
//
// The tokens live in theme.css, which dresses the document dark unless the root
// element says otherwise, so this hook only ever sets one attribute.

import { useCallback, useEffect, useState } from 'react'
import { defaultTheme, nextTheme, storedTheme, themeKey, type Theme } from './theme'

/**
 * The theme remembered in this window, else the default.
 *
 * Storage can refuse both reading and writing: a window opened with site data
 * blocked answers by throwing rather than by answering nothing. A theme is not
 * worth a dead page, so a refusal leaves the default showing and the choice
 * lasting only as long as the window.
 */
export function rememberedTheme(): Theme {
  try {
    return storedTheme(window.localStorage.getItem(themeKey))
  } catch {
    return defaultTheme
  }
}

/** remember keeps the choice for the next opening, where it can. */
function remember(theme: Theme): void {
  try {
    window.localStorage.setItem(themeKey, theme)
  } catch {
    // Nothing to do and nothing to say: the window is already wearing it.
  }
}

/**
 * Dress the document, then answer the theme with the way to change it.
 *
 * The attribute is set on the root element rather than a class on the body,
 * because the tokens hang off `:root` and a dialog drawn outside the body's
 * tree would miss a class set there.
 */
export function useTheme(): [Theme, () => void] {
  const [theme, setTheme] = useState<Theme>(rememberedTheme)

  useEffect(() => {
    document.documentElement.dataset.theme = theme
  }, [theme])

  const toggle = useCallback(() => {
    setTheme((current) => {
      const wanted = nextTheme(current)
      remember(wanted)
      return wanted
    })
  }, [])

  return [theme, toggle]
}
