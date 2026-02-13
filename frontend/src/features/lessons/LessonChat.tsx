import { useCallback, useEffect, useRef, useState } from 'react'
import { chat as chatApi, type ChatMessage } from '../../shared/api'
import { useAuth } from '../auth/AuthContext'
import type { Lesson as LessonType } from '../../shared/types'

const ROLE_PROMPT = `Ты — дружелюбный помощник по робототехнике для детей. Отвечай коротко и понятно, по делу урока. Помогай разобраться в шагах, коде и схемах. Если ребёнок спрашивает что-то не по теме урока, вежливо направь разговор к текущему уроку.`

function buildLessonContext(lesson: LessonType): string {
  const parts: string[] = [ROLE_PROMPT, '', '--- Текст текущего урока ---', '']
  parts.push(`Название: ${lesson.title}`)
  if (lesson.description) parts.push(`Описание: ${lesson.description}`)
  parts.push('')
  if (lesson.steps && lesson.steps.length > 0) {
    const sorted = [...lesson.steps].sort((a, b) => a.sort_order - b.sort_order)
    sorted.forEach((step, i) => {
      parts.push(`Шаг ${i + 1}: ${step.title}`)
      parts.push(step.content)
      parts.push('')
    })
  }
  if (lesson.checklist && lesson.checklist.length > 0) {
    parts.push('Чек-лист:')
    const sorted = [...lesson.checklist].sort((a, b) => a.sort_order - b.sort_order)
    sorted.forEach((item) => parts.push(`- ${item.title}`))
  }
  return parts.join('\n')
}

declare global {
  interface Window {
    SpeechRecognition?: new () => SpeechRecognition
    webkitSpeechRecognition?: new () => SpeechRecognition
  }
}

interface SpeechRecognition extends EventTarget {
  continuous: boolean
  interimResults: boolean
  lang: string
  start(): void
  stop(): void
  onresult: ((e: SpeechRecognitionEvent) => void) | null
  onerror: ((e: SpeechRecognitionErrorEvent) => void) | null
  onend: (() => void) | null
}

interface SpeechRecognitionEvent {
  results: SpeechRecognitionResultList
  resultIndex: number
}

interface SpeechRecognitionErrorEvent {
  error: string
}

export function LessonChat({ lesson }: { lesson: LessonType }) {
  const { user } = useAuth()
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [loadingHistory, setLoadingHistory] = useState(true)
  const [error, setError] = useState('')
  const [isRecording, setIsRecording] = useState(false)
  const recognitionRef = useRef<SpeechRecognition | null>(null)
  const transcriptRef = useRef('')
  const lessonContext = useRef(buildLessonContext(lesson)).current

  useEffect(() => {
    if (!user) {
      setLoadingHistory(false)
      return
    }
    chatApi
      .getHistory(lesson.id)
      .then((res) => setMessages(res.messages || []))
      .catch(() => setMessages([]))
      .finally(() => setLoadingHistory(false))
  }, [user, lesson.id])

  const clearChat = useCallback(async () => {
    if (!user) return
    setError('')
    try {
      await chatApi.clearHistory(lesson.id)
      setMessages([])
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось очистить чат')
    }
  }, [user, lesson.id])

  const send = useCallback(
    async (text: string) => {
      const trimmed = text.trim()
      if (!trimmed || loading) return
      setError('')
      const userMessage: ChatMessage = { role: 'user', text: trimmed }
      setMessages((prev) => [...prev, userMessage])
      setInput('')
      setLoading(true)
      try {
        const history: ChatMessage[] = [...messages, userMessage]
        const res = await chatApi.send(lesson.id, lessonContext, history)
        setMessages((prev) => [...prev, { role: 'model', text: res.text || '' }])
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Ошибка запроса')
        setMessages((prev) => prev.slice(0, -1))
      } finally {
        setLoading(false)
      }
    },
    [lesson.id, lessonContext, messages, loading]
  )

  const startVoice = useCallback(() => {
    const SR = window.SpeechRecognition || window.webkitSpeechRecognition
    if (!SR) {
      setError('Голосовой ввод не поддерживается. Используйте Chrome или Edge.')
      return
    }
    if (recognitionRef.current) return
    transcriptRef.current = ''
    const recognition = new SR() as SpeechRecognition
    recognition.continuous = true
    recognition.interimResults = true
    recognition.lang = 'ru-RU'
    recognition.onresult = (e: SpeechRecognitionEvent) => {
      const results = e.results
      for (let i = e.resultIndex; i < results.length; i++) {
        const result = results.item ? results.item(i) : results[i]
        if (!result) continue
        const first = result.length ? (result[0] as { transcript?: string }) : null
        const transcript = first?.transcript?.trim() ?? ''
        if (result.isFinal && transcript) {
          transcriptRef.current = (transcriptRef.current + ' ' + transcript).trim()
        }
      }
    }
    recognition.onend = () => {
      recognitionRef.current = null
      setIsRecording(false)
      const text = transcriptRef.current.trim()
      if (text) {
        setInput((prev) => (prev ? `${prev} ${text}` : text))
      }
    }
    recognition.onerror = (e: SpeechRecognitionErrorEvent) => {
      recognitionRef.current = null
      setIsRecording(false)
      const err = e?.error ?? 'unknown'
      if (err === 'not-allowed' || err === 'permission-denied') {
        setError('Разрешите доступ к микрофону в настройках браузера.')
      } else if (err === 'no-speech') {
        setError('Речь не распознана. Попробуйте ещё раз.')
      } else {
        setError('Ошибка голосового ввода. Попробуйте Chrome или Edge.')
      }
    }
    recognitionRef.current = recognition
    setIsRecording(true)
    setError('')
    try {
      recognition.start()
    } catch (err) {
      recognitionRef.current = null
      setIsRecording(false)
      setError('Не удалось запустить распознавание речи.')
    }
  }, [])

  const stopVoice = useCallback(() => {
    if (recognitionRef.current) {
      recognitionRef.current.stop()
      recognitionRef.current = null
    }
    setIsRecording(false)
  }, [])

  return (
    <section className="lesson-chat" aria-label="Чат с помощником по уроку">
      <div className="lesson-chat-header">
        <h2>Спроси у помощника</h2>
        {user && messages.length > 0 && (
          <button
            type="button"
            className="lesson-chat-clear"
            onClick={clearChat}
            disabled={loading}
            aria-label="Очистить чат"
          >
            Очистить чат
          </button>
        )}
      </div>
      <div className="lesson-chat-messages">
        {loadingHistory && (
          <p className="lesson-chat-placeholder">Загрузка истории…</p>
        )}
        {!loadingHistory && messages.length === 0 && !user && (
          <p className="lesson-chat-placeholder">
            Войдите, чтобы общаться с помощником и сохранять историю чата.
          </p>
        )}
        {!loadingHistory && messages.length === 0 && user && (
          <p className="lesson-chat-placeholder">
            Напиши вопрос по уроку или зажми кнопку микрофона и скажи команду голосом.
          </p>
        )}
        {messages.map((m, i) => (
          <div key={i} className={`lesson-chat-msg lesson-chat-msg-${m.role}`}>
            <span className="lesson-chat-msg-role">{m.role === 'user' ? 'Ты' : 'Помощник'}</span>
            <div className="lesson-chat-msg-text">{m.text}</div>
          </div>
        ))}
        {loading && (
          <div className="lesson-chat-msg lesson-chat-msg-model">
            <span className="lesson-chat-msg-role">Помощник</span>
            <div className="lesson-chat-msg-text lesson-chat-msg-loading">Думаю...</div>
          </div>
        )}
      </div>
      {error && <p className="lesson-chat-error">{error}</p>}
      <div className="lesson-chat-input-row">
        <input
          type="text"
          className="lesson-chat-input"
          placeholder={user ? 'Введите вопрос...' : 'Войдите, чтобы отправить'}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => user && e.key === 'Enter' && !e.shiftKey && send(input)}
          disabled={loading || !user}
          aria-label="Текст вопроса"
        />
        <button
          type="button"
          className={`lesson-chat-voice ${isRecording ? 'recording' : ''}`}
          title="Зажми и говори — отпусти, чтобы вставить текст"
          aria-label="Голосовой ввод: зажми и говори"
          disabled={!user}
          onPointerDown={(e) => {
            e.preventDefault()
            if (user) startVoice()
          }}
          onPointerUp={stopVoice}
          onPointerLeave={stopVoice}
          onTouchStart={(e) => {
            e.preventDefault()
            if (user) startVoice()
          }}
          onTouchEnd={(e) => {
            e.preventDefault()
            stopVoice()
          }}
          onContextMenu={(e) => e.preventDefault()}
        >
          {isRecording ? '🔴' : '🎤'}
        </button>
        <span className="lesson-chat-voice-hint" title="Зажми кнопку, говори, отпусти — текст появится в поле">
          Зажми и говори
        </span>
        <button
          type="button"
          className="lesson-chat-send"
          onClick={() => send(input)}
          disabled={loading || !input.trim() || !user}
          aria-label="Отправить"
        >
          Отправить
        </button>
      </div>
    </section>
  )
}
