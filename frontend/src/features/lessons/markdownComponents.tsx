import { getYouTubeVideoId, isYouTubeUrl, isSafeUrl } from './youtubeUtils'
import { YouTubeEmbed } from './YouTubeEmbed'

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
