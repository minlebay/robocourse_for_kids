/**
 * Извлекает YouTube video ID из URL.
 * Поддерживает: youtube.com/watch?v=ID, youtu.be/ID, youtube.com/embed/ID
 */
export function getYouTubeVideoId(url: string): string | null {
  try {
    const u = new URL(url)
    if (u.hostname === 'www.youtube.com' || u.hostname === 'youtube.com') {
      return u.searchParams.get('v') || (u.pathname.startsWith('/embed/') ? u.pathname.split('/')[2] : null)
    }
    if (u.hostname === 'youtu.be') {
      return u.pathname.slice(1).split('?')[0] || null
    }
  } catch {
    // invalid URL
  }
  return null
}

export function isYouTubeUrl(href: string): boolean {
  try {
    const u = new URL(href)
    const host = u.hostname.toLowerCase()
    return host === 'www.youtube.com' || host === 'youtube.com' || host === 'youtu.be'
  } catch {
    return false
  }
}

/** Отклоняет небезопасные схемы (javascript:, data:, vbscript: и т.п.) */
export function isSafeUrl(href: string | undefined): boolean {
  if (!href) return false
  try {
    const u = new URL(href)
    return u.protocol === 'https:' || u.protocol === 'http:'
  } catch {
    // Относительные пути разрешены (начинаются с / или #)
    return href.startsWith('/') || href.startsWith('#')
  }
}
