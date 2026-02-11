import { useRef } from 'react'
import { getYouTubeVideoId } from './YouTubeEmbed'

type MarkdownStepEditorProps = {
  value: string
  onChange: (value: string) => void
  rows?: number
  placeholder?: string
  label?: string
  id?: string
}

function insertAtCursor(textarea: HTMLTextAreaElement | null, text: string): string {
  if (!textarea) return text
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const before = textarea.value.slice(0, start)
  const after = textarea.value.slice(end)
  return before + text + after
}

export function MarkdownStepEditor({
  value,
  onChange,
  rows = 6,
  placeholder = 'Контент (Markdown)',
  label = 'Контент (Markdown)',
  id,
}: MarkdownStepEditorProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const insert = (text: string) => {
    const newValue = insertAtCursor(textareaRef.current, text)
    onChange(newValue)
    textareaRef.current?.focus()
  }

  const insertImage = () => {
    const url = window.prompt('URL картинки:')
    if (!url?.trim()) return
    const alt = window.prompt('Описание (alt):', 'Изображение') ?? 'Изображение'
    insert(`\n\n![${alt}](${url.trim()})\n\n`)
  }

  const insertYouTube = () => {
    const url = window.prompt('Ссылка на YouTube (youtube.com или youtu.be):')
    if (!url?.trim()) return
    const videoId = getYouTubeVideoId(url.trim())
    if (!videoId) {
      window.alert('Укажите ссылку только с YouTube.')
      return
    }
    insert(`\n\n[Видео](https://www.youtube.com/watch?v=${videoId})\n\n`)
  }

  return (
    <div className="markdown-step-editor">
      <div className="markdown-step-editor-toolbar">
        {label && <label htmlFor={id}>{label}</label>}
        <span className="markdown-step-editor-hint">Вставить:</span>
        <div className="markdown-step-editor-buttons">
          <button type="button" className="button-secondary markdown-insert-btn" onClick={insertImage} title="Вставить картинку">
            Картинка
          </button>
          <button type="button" className="button-secondary markdown-insert-btn" onClick={insertYouTube} title="Вставить YouTube">
            YouTube
          </button>
        </div>
      </div>
      <textarea
        ref={textareaRef}
        id={id}
        rows={rows}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
      />
    </div>
  )
}
