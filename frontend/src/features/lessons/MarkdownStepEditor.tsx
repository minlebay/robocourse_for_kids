import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { isSafeHttpUrl } from '../../shared/url'
import { getYouTubeVideoId } from './youtubeUtils'

type InsertMode = 'image' | 'youtube' | null

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
  placeholder,
  label,
  id,
}: MarkdownStepEditorProps) {
  const { t } = useTranslation()
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const [insertMode, setInsertMode] = useState<InsertMode>(null)
  const [insertUrl, setInsertUrl] = useState('')
  const [insertAlt, setInsertAlt] = useState('')
  const [insertError, setInsertError] = useState('')

  const actualPlaceholder = placeholder ?? t('editor.defaultPlaceholder')
  const actualLabel = label !== undefined ? label : t('editor.defaultLabel')

  const insert = (text: string) => {
    const newValue = insertAtCursor(textareaRef.current, text)
    onChange(newValue)
    textareaRef.current?.focus()
  }

  const openInsert = (mode: 'image' | 'youtube') => {
    setInsertMode(mode)
    setInsertUrl('')
    setInsertAlt('')
    setInsertError('')
  }

  const cancelInsert = () => {
    setInsertMode(null)
    setInsertUrl('')
    setInsertAlt('')
    setInsertError('')
    textareaRef.current?.focus()
  }

  const submitInsert = () => {
    const url = insertUrl.trim()
    if (!url || !isSafeHttpUrl(url)) {
      setInsertError(t('editor.urlHttpsOnly'))
      return
    }
    if (insertMode === 'image') {
      const alt = insertAlt.trim() || t('editor.imageAltDefault')
      insert(`\n\n![${alt}](${url})\n\n`)
      cancelInsert()
    } else if (insertMode === 'youtube') {
      const videoId = getYouTubeVideoId(url)
      if (!videoId) {
        setInsertError(t('editor.youTubeOnly'))
        return
      }
      insert(`\n\n[Видео](https://www.youtube.com/watch?v=${videoId})\n\n`)
      cancelInsert()
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      submitInsert()
    } else if (e.key === 'Escape') {
      cancelInsert()
    }
  }

  return (
    <div className="markdown-step-editor">
      <div className="markdown-step-editor-toolbar">
        {actualLabel && <label htmlFor={id}>{actualLabel}</label>}
        <span className="markdown-step-editor-hint">{t('editor.insertHint')}</span>
        <div className="markdown-step-editor-buttons">
          <button
            type="button"
            className="button-secondary markdown-insert-btn"
            onClick={() => openInsert('image')}
            title={t('editor.insertImageTitle')}
            disabled={insertMode !== null}
          >
            {t('editor.insertImage')}
          </button>
          <button
            type="button"
            className="button-secondary markdown-insert-btn"
            onClick={() => openInsert('youtube')}
            title={t('editor.insertYouTubeTitle')}
            disabled={insertMode !== null}
          >
            {t('editor.insertYouTube')}
          </button>
        </div>
      </div>

      {insertMode !== null && (
        <div className="markdown-insert-form">
          <div className="form-group">
            <label>
              {insertMode === 'image' ? t('editor.imageUrlLabel') : t('editor.youTubeUrlLabel')}
            </label>
            <input
              type="url"
              value={insertUrl}
              onChange={(e) => { setInsertUrl(e.target.value); setInsertError('') }}
              placeholder={
                insertMode === 'image'
                  ? t('editor.imageUrlPlaceholder')
                  : t('editor.youTubeUrlPlaceholder')
              }
              // eslint-disable-next-line jsx-a11y/no-autofocus
              autoFocus
              onKeyDown={handleKeyDown}
            />
          </div>
          {insertMode === 'image' && (
            <div className="form-group">
              <label>{t('editor.imageAltLabel')}</label>
              <input
                type="text"
                value={insertAlt}
                onChange={(e) => setInsertAlt(e.target.value)}
                placeholder={t('editor.imageAltDefault')}
                onKeyDown={handleKeyDown}
              />
            </div>
          )}
          {insertError && <p className="error" role="alert">{insertError}</p>}
          <div className="form-actions">
            <button type="button" className="button-primary" onClick={submitInsert}>
              {insertMode === 'image' ? t('editor.insertImageTitle') : t('editor.insertYouTubeTitle')}
            </button>
            <button type="button" className="button-secondary" onClick={cancelInsert}>
              {t('common.cancel')}
            </button>
          </div>
        </div>
      )}

      <textarea
        ref={textareaRef}
        id={id}
        rows={rows}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={actualPlaceholder}
      />
    </div>
  )
}
