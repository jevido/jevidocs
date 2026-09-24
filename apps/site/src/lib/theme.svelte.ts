export type ThemeMode = 'light' | 'dark' | 'system'

const KEY = 'jevidocs-theme'

function read(): ThemeMode {
  try {
    const v = localStorage.getItem(KEY)
    if (v === 'light' || v === 'dark' || v === 'system') return v
  } catch {
    // storage blocked: fall back to system
  }
  return 'system'
}

const media = typeof matchMedia === 'function' ? matchMedia('(prefers-color-scheme: dark)') : null

function apply(mode: ThemeMode) {
  const dark = mode === 'dark' || (mode === 'system' && !!media?.matches)
  document.documentElement.classList.toggle('dark', dark)
}

class Theme {
  mode = $state<ThemeMode>(read())

  constructor() {
    media?.addEventListener('change', () => {
      if (this.mode === 'system') apply('system')
    })
  }

  set(mode: ThemeMode) {
    this.mode = mode
    try {
      localStorage.setItem(KEY, mode)
    } catch {
      // ignore
    }
    apply(mode)
  }
}

export const theme = new Theme()
