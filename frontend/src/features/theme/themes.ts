/**
 * Определения цветовых схем портала.
 * При добавлении темы или переменных обновить также инлайн-скрипт в index.html (FOUC).
 */

export const THEME_STORAGE_KEY = 'learn_kids_theme'

export type ThemeId = 'default' | 'light' | 'cyberpunk'

export interface Theme {
  id: ThemeId
  name: string
  /** Четыре цвета для превью (2x2): bg, accent, text, secondary */
  previewColors: [string, string, string, string]
  cssVars: Record<string, string>
}

export const themes: Record<ThemeId, Theme> = {
  default: {
    id: 'default',
    name: 'Тёмная (по умолчанию)',
    previewColors: ['#0f172a', '#e07c24', '#e2e8f0', '#94a3b8'],
    cssVars: {
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
    previewColors: ['#0d0221', '#00ff9f', '#ff00e5', '#00d4ff'],
    cssVars: {
      '--bg': '#0d0221',
      '--bg-card': '#1a0a2e',
      '--accent': '#00ff9f',
      '--accent-hover': '#00d4ff',
      '--accent-soft': '#0d3d2e',
      '--text': '#e0e0ff',
      '--text-muted': '#a78bfa',
      '--border': '#4c1d95',
      '--success': '#00ff9f',
      '--success-soft': '#0d3d2e',
      '--header-bg': '#0d0221',
      '--header-text': '#00ff9f',
      '--accent-on': '#0d0221',
    },
  },
}

export const themeOrder: ThemeId[] = ['default', 'light', 'cyberpunk']
