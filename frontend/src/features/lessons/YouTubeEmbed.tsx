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

function isYouTubeUrl(href: string): boolean {
  try {
    const u = new URL(href)
    const host = u.hostname.toLowerCase()
    return host === 'www.youtube.com' || host === 'youtube.com' || host === 'youtu.be'
  } catch {
    return false
  }
}

/** Отклоняет небезопасные схемы (javascript:, data:, vbscript: и т.п.) */
function isSafeUrl(href: string | undefined): boolean {
  if (!href) return false
  try {
    const u = new URL(href)
    return u.protocol === 'https:' || u.protocol === 'http:'
  } catch {
    // Относительные пути разрешены (начинаются с / или #)
    return href.startsWith('/') || href.startsWith('#')
  }
}

export function YouTubeEmbed({ videoId }: { videoId: string }) {
  return (
    <figure className="youtube-embed">
      <iframe
        src={`https://www.youtube.com/embed/${encodeURIComponent(videoId)}`}
        title="YouTube video player"
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
        sandbox="allow-scripts allow-same-origin allow-presentation allow-popups"
        allowFullScreen
      />
    </figure>
  )
}

export const markdownComponents = {
  a: ({ href, children, ...props }: React.AnchorHTMLAttributes<HTMLAnchorElement>) => {
    if (href && isYouTubeUrl(href)) {
      const videoId = getYouTubeVideoId(href)
      if (videoId) {
        return <YouTubeEmbed videoId={videoId} />
      }
    }
    if (!isSafeUrl(href)) {
      // Небезопасная схема — рендерим как простой текст
      return <span>{children}</span>
    }
    return (
      <a href={href} target="_blank" rel="noreferrer noopener" {...props}>
        {children}
      </a>
    )
  },
}
