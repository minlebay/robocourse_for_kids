/**
 * Разрешены только http и https. Запрещены javascript:, data: и др. для защиты от XSS.
 */

const SAFE_PROTOCOLS = ['https:', 'http:']

export function isSafeHttpUrl(url: string): boolean {
  try {
    const u = new URL(url, window.location.origin)
    return SAFE_PROTOCOLS.includes(u.protocol)
  } catch {
    return false
  }
}

/** Возвращает URL, если он безопасен (только http/https), иначе null. */
export function getSafeUrl(url: string): string | null {
  if (!url || typeof url !== 'string') return null
  const trimmed = url.trim()
  if (!trimmed) return null
  return isSafeHttpUrl(trimmed) ? trimmed : null
}
