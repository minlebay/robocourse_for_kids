/**
 * Определения цветовых схем портала.
 * При добавлении темы или переменных обновить также инлайн-скрипт в index.html (FOUC).
 */

export const THEME_STORAGE_KEY = 'learn_kids_theme'

export type ThemeId =
  | 'default'
  | 'light'
  | 'cyberpunk'
  | 'contrast-light'
  | 'contrast-dark'
  | 'cream'
  | 'snow'
  | 'midnight'
  | 'forest'

export interface Theme {
  id: ThemeId
  name: string
  /** Четыре цвета для превью (2x2): bg, accent, text, secondary */
  previewColors: [string, string, string, string]
  cssVars: Record<string, string>
}

/** Циклические градиенты и «стеклянные» панели (прозрачнее + сильнее размытие). Карточки — второй слой матового стекла (темнее). */
const frostedDark = {
  '--bg-gradient':
    'linear-gradient(-45deg, #0f172a, #1e1b4b, #312e81, #1e1b4b, #0f172a)',
  '--glass-bg': 'rgba(30, 41, 59, 0.48)',
  '--glass-card-bg': 'rgba(30, 26, 46, 0.72)',
  '--glass-border': 'rgba(148, 163, 184, 0.25)',
  '--glass-blur': '24px',
  '--glass-shadow': '0 8px 32px rgba(0, 0, 0, 0.25)',
}
/** Карточки — второй слой матового стекла (светлее). */
const frostedLight = {
  '--bg-gradient':
    'linear-gradient(-45deg, #fce7f3, #e9d5ff, #dbeafe, #e9d5ff, #fce7f3)',
  '--glass-bg': 'rgba(255, 255, 255, 0.52)',
  '--glass-card-bg': 'rgba(255, 255, 255, 0.78)',
  '--glass-border': 'rgba(226, 232, 240, 0.65)',
  '--glass-blur': '24px',
  '--glass-shadow': '0 8px 32px rgba(0, 0, 0, 0.08)',
}

export const themes: Record<ThemeId, Theme> = {
  default: {
    id: 'default',
    name: 'Тёмная (по умолчанию)',
    previewColors: ['#0f172a', '#e07c24', '#e2e8f0', '#94a3b8'],
    cssVars: {
      ...frostedDark,
      '--bg': '#0f172a',
      '--bg-card': '#1e293b',
      '--accent': '#e07c24',
      '--accent-hover': '#f59e0b',
      '--accent-soft': '#422006',
      '--text': '#e2e8f0',
      '--text-muted': '#94a3b8',
      '--border': '#334155',
      '--success': '#10b981',
      '--success-soft': '#064e3b',
      '--header-bg': '#0f172a',
      '--header-text': '#f1f5f9',
      '--accent-on': '#ffffff',
    },
  },
  light: {
    id: 'light',
    name: 'Светлая',
    previewColors: ['#f8fafc', '#ea580c', '#1e293b', '#64748b'],
    cssVars: {
      ...frostedLight,
      '--bg': '#f1f5f9',
      '--bg-card': '#ffffff',
      '--accent': '#ea580c',
      '--accent-hover': '#c2410c',
      '--accent-soft': '#fff7ed',
      '--text': '#1e293b',
      '--text-muted': '#64748b',
      '--border': '#e2e8f0',
      '--success': '#059669',
      '--success-soft': '#d1fae5',
      '--header-bg': '#ffffff',
      '--header-text': '#1e293b',
      '--accent-on': '#ffffff',
    },
  },
  cyberpunk: {
    id: 'cyberpunk',
    name: 'Киберпанк',
    previewColors: ['#0a0118', '#00b377', '#9900b3', '#006680'],
    cssVars: {
      /* Анимированный градиент: -45deg, с переходом в чёрный */
      '--bg-gradient':
        'linear-gradient(-45deg, #0a0118, #1a0a2e, #2e1065, #000000, #1a0a2e, #0a0118)',
      '--glass-bg': 'rgba(20, 8, 36, 0.5)',
      '--glass-card-bg': 'rgba(15, 5, 28, 0.78)',
      '--glass-border': 'rgba(55, 25, 95, 0.4)',
      '--glass-blur': '24px',
      '--glass-shadow': '0 8px 32px rgba(0, 0, 0, 0.4)',
      '--bg': '#0a0118',
      '--bg-card': '#1a0a2e',
      '--accent': '#00b377',
      '--accent-hover': '#00d99e',
      '--accent-soft': '#0a2e22',
      '--text': '#c4c0d4',
      '--text-muted': '#7c6b9e',
      '--border': '#3b2d5c',
      '--success': '#00a86b',
      '--success-soft': '#0a2e22',
      '--header-bg': '#0a0118',
      '--header-text': '#00b377',
      '--accent-on': '#0a0118',
    },
  },
  // Контрастная светлая
  'contrast-light': {
    id: 'contrast-light',
    name: 'Контрастная светлая',
    previewColors: ['#ffffff', '#0066cc', '#000000', '#333333'],
    cssVars: {
      ...frostedLight,
      '--bg': '#f5f5f5',
      '--bg-card': '#ffffff',
      '--accent': '#0066cc',
      '--accent-hover': '#004499',
      '--accent-soft': '#e6f0ff',
      '--text': '#000000',
      '--text-muted': '#333333',
      '--border': '#cccccc',
      '--success': '#00884a',
      '--success-soft': '#e6f7ef',
      '--header-bg': '#ffffff',
      '--header-text': '#000000',
      '--accent-on': '#ffffff',
    },
  },
  // Контрастная тёмная
  'contrast-dark': {
    id: 'contrast-dark',
    name: 'Контрастная тёмная',
    previewColors: ['#0a0a0a', '#00aaff', '#ffffff', '#b0b0b0'],
    cssVars: {
      '--bg-gradient':
        'linear-gradient(-45deg, #0a0a0a, #0f172a, #1e3a8a, #0f172a, #0a0a0a)',
      '--glass-bg': 'rgba(26, 26, 26, 0.5)',
      '--glass-card-bg': 'rgba(18, 18, 18, 0.75)',
      '--glass-border': 'rgba(64, 64, 64, 0.5)',
      '--glass-blur': '24px',
      '--glass-shadow': '0 8px 32px rgba(0, 0, 0, 0.35)',
      '--bg': '#0a0a0a',
      '--bg-card': '#1a1a1a',
      '--accent': '#00aaff',
      '--accent-hover': '#00ccff',
      '--accent-soft': '#002233',
      '--text': '#ffffff',
      '--text-muted': '#b0b0b0',
      '--border': '#404040',
      '--success': '#00cc66',
      '--success-soft': '#002211',
      '--header-bg': '#0a0a0a',
      '--header-text': '#ffffff',
      '--accent-on': '#0a0a0a',
    },
  },
  // Светлые
  cream: {
    id: 'cream',
    name: 'Кремовая',
    previewColors: ['#fef9f3', '#c2410c', '#292524', '#78716c'],
    cssVars: {
      ...frostedLight,
      '--bg': '#fef9f3',
      '--bg-card': '#fff7ed',
      '--accent': '#c2410c',
      '--accent-hover': '#9a3412',
      '--accent-soft': '#ffedd5',
      '--text': '#292524',
      '--text-muted': '#78716c',
      '--border': '#e7e5e4',
      '--success': '#15803d',
      '--success-soft': '#dcfce7',
      '--header-bg': '#fff7ed',
      '--header-text': '#292524',
      '--accent-on': '#ffffff',
    },
  },
  snow: {
    id: 'snow',
    name: 'Снежная',
    previewColors: ['#f0f9ff', '#0369a1', '#0c4a6e', '#64748b'],
    cssVars: {
      ...frostedLight,
      '--bg': '#f0f9ff',
      '--bg-card': '#ffffff',
      '--accent': '#0369a1',
      '--accent-hover': '#0284c7',
      '--accent-soft': '#e0f2fe',
      '--text': '#0c4a6e',
      '--text-muted': '#64748b',
      '--border': '#bae6fd',
      '--success': '#0d9488',
      '--success-soft': '#ccfbf1',
      '--header-bg': '#ffffff',
      '--header-text': '#0c4a6e',
      '--accent-on': '#ffffff',
    },
  },
  // Тёмные
  midnight: {
    id: 'midnight',
    name: 'Полночь',
    previewColors: ['#0c1222', '#38bdf8', '#e0f2fe', '#7dd3fc'],
    cssVars: {
      ...frostedDark,
      '--bg': '#0c1222',
      '--bg-card': '#1e293b',
      '--accent': '#38bdf8',
      '--accent-hover': '#7dd3fc',
      '--accent-soft': '#0c4a6e',
      '--text': '#e0f2fe',
      '--text-muted': '#7dd3fc',
      '--border': '#1e3a5f',
      '--success': '#2dd4bf',
      '--success-soft': '#134e4a',
      '--header-bg': '#0c1222',
      '--header-text': '#e0f2fe',
      '--accent-on': '#0c1222',
    },
  },
  forest: {
    id: 'forest',
    name: 'Лес',
    previewColors: ['#0f1419', '#22c55e', '#dcfce7', '#86efac'],
    cssVars: {
      '--bg-gradient':
        'linear-gradient(-45deg, #0f1419, #14532d, #166534, #14532d, #0f1419)',
      '--glass-bg': 'rgba(26, 36, 33, 0.5)',
      '--glass-card-bg': 'rgba(15, 28, 24, 0.72)',
      '--glass-border': 'rgba(22, 101, 52, 0.35)',
      '--glass-blur': '24px',
      '--glass-shadow': '0 8px 32px rgba(0, 0, 0, 0.3)',
      '--bg': '#0f1419',
      '--bg-card': '#1a2421',
      '--accent': '#22c55e',
      '--accent-hover': '#4ade80',
      '--accent-soft': '#14532d',
      '--text': '#dcfce7',
      '--text-muted': '#86efac',
      '--border': '#166534',
      '--success': '#22c55e',
      '--success-soft': '#14532d',
      '--header-bg': '#0f1419',
      '--header-text': '#dcfce7',
      '--accent-on': '#0f1419',
    },
  },
}

export const themeOrder: ThemeId[] = [
  'default',
  'light',
  'cyberpunk',
  'contrast-light',
  'contrast-dark',
  'cream',
  'snow',
  'midnight',
  'forest',
]
