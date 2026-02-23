import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { chat as chatApi, type ChatMessage } from '../../shared/api'
import { CHAT_MESSAGE_MAX, validateChatMessage } from '../../shared/validation'
import { useAuth } from '../auth/AuthContext'

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

export function LessonChat({ lessonId }: { lessonId: string }) {
  const { user } = useAuth()
  const { t, i18n } = useTranslation()
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [loadingHistory, setLoadingHistory] = useState(true)
  const [historyError, setHistoryError] = useState(false)
  const [error, setError] = useState('')
  const [isRecording, setIsRecording] = useState(false)
  const recognitionRef = useRef<SpeechRecognition | null>(null)
  const transcriptRef = useRef('')
  const sendingRef = useRef(false)
  const speechLang = i18n.language === 'en' ? 'en-US' : 'ru-RU'

  useEffect(() => {
    if (!user) {
      setLoadingHistory(false)
      return
    }
    setHistoryError(false)
    chatApi
      .getHistory(lessonId)
      .then((res) =>
        setMessages(
          (res.messages || []).map((m, i) => ({ ...m, id: m.id ?? `hist-${i}-${m.role}` }))
        )
      )
      .catch(() => {
        setMessages([])
        setHistoryError(true)
      })
      .finally(() => setLoadingHistory(false))
  }, [user, lessonId])

  useEffect(() => {
    return () => {
      if (recognitionRef.current) {
        recognitionRef.current.stop()
        recognitionRef.current = null
      }
    }
  }, [])

  const clearChat = useCallback(async () => {
    if (!user) return
    setError('')
    try {
      await chatApi.clearHistory(lessonId)
      setMessages([])
    } catch (e) {
      setError(e instanceof Error ? e.message : t('chat.clearChatError'))
    }
  }, [user, lessonId, t])

  const send = useCallback(
    async (text: string) => {
      if (sendingRef.current || loading) return
      const trimmed = text.trim()
      if (!trimmed) return
      const msgErr = validateChatMessage(text)
      if (msgErr) {
        setError(t(msgErr.key, msgErr.params as Record<string, string | number>))
        return
      }
      setError('')
      sendingRef.current = true
      const userMsgId = `user-${Date.now()}`
      setMessages((prev) => [...prev, { role: 'user', text: trimmed, id: userMsgId }])
      setInput('')
      setLoading(true)
      try {
        const res = await chatApi.send(lessonId, trimmed)
        setMessages((prev) => [
          ...prev,
          { role: 'model', text: res.text || '', id: `model-${Date.now()}` },
        ])
      } catch (e) {
        setError(e instanceof Error ? e.message : t('chat.requestError'))
        setMessages((prev) => prev.slice(0, -1))
      } finally {
        setLoading(false)
        sendingRef.current = false
      }
    },
    [lessonId, loading, t]
  )

  const startVoice = useCallback(() => {
    const SR = window.SpeechRecognition || window.webkitSpeechRecognition
    if (!SR) {
      setError(t('chat.voiceUnsupported'))
      return
    }
    if (recognitionRef.current) return
    transcriptRef.current = ''
    const recognition = new SR() as SpeechRecognition
    recognition.continuous = true
    recognition.interimResults = true
    recognition.lang = speechLang
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
      const errCode = e?.error ?? 'unknown'
      if (errCode === 'not-allowed' || errCode === 'permission-denied') {
        setError(t('chat.voicePermission'))
      } else if (errCode === 'no-speech') {
        setError(t('chat.voiceNoSpeech'))
      } else {
        setError(t('chat.voiceError'))
      }
    }
    recognitionRef.current = recognition
    setIsRecording(true)
    setError('')
    try {
      recognition.start()
    } catch {
      recognitionRef.current = null
      setIsRecording(false)
      setError(t('chat.voiceStartError'))
    }
  }, [t, speechLang])

  const stopVoice = useCallback(() => {
    if (recognitionRef.current) {
      recognitionRef.current.stop()
      recognitionRef.current = null
    }
    setIsRecording(false)
  }, [])

  return (
    <section className="lesson-chat" aria-label={t('chat.title')}>
      <div className="lesson-chat-header">
        <h2>{t('chat.title')}</h2>
        {user && messages.length > 0 && (
          <button
            type="button"
            className="lesson-chat-clear"
            onClick={clearChat}
            disabled={loading}
            aria-label={t('chat.clearChat')}
          >
            {t('chat.clearChat')}
          </button>
        )}
      </div>
      <div className="lesson-chat-messages">
        {loadingHistory && (
          <p className="lesson-chat-placeholder">{t('chat.loadingHistory')}</p>
        )}
        {!loadingHistory && historyError && (
          <p className="lesson-chat-error">{t('errors.historyLoadFailed')}</p>
        )}
        {!loadingHistory && messages.length === 0 && !user && (
          <p className="lesson-chat-placeholder">
            {t('chat.loginToChat')}
          </p>
        )}
        {!loadingHistory && messages.length === 0 && user && !historyError && (
          <p className="lesson-chat-placeholder">
            {t('chat.placeholder')}
          </p>
        )}
        {messages.map((m) => (
          <div key={m.id} className={`lesson-chat-msg lesson-chat-msg-${m.role}`}>
            <span className="lesson-chat-msg-role">{m.role === 'user' ? t('chat.you') : t('chat.assistant')}</span>
            <div className="lesson-chat-msg-text">{m.text}</div>
          </div>
        ))}
        {loading && (
          <div className="lesson-chat-msg lesson-chat-msg-model">
            <span className="lesson-chat-msg-role">{t('chat.assistant')}</span>
            <div className="lesson-chat-msg-text lesson-chat-msg-loading">{t('chat.thinking')}</div>
          </div>
        )}
      </div>
      {error && <p className="lesson-chat-error">{error}</p>}
      <div className="lesson-chat-input-row">
        <input
          type="text"
          className="lesson-chat-input"
          placeholder={user ? t('chat.inputPlaceholder') : t('chat.inputPlaceholderGuest')}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => user && e.key === 'Enter' && !e.shiftKey && send(input)}
          disabled={loading || !user}
          maxLength={CHAT_MESSAGE_MAX}
          aria-label={t('chat.inputAria')}
        />
        <button
          type="button"
          className={`lesson-chat-voice ${isRecording ? 'recording' : ''}`}
          title={t('chat.voiceHold')}
          aria-label={t('chat.voiceAria')}
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
        <span className="lesson-chat-voice-hint" title={t('chat.voiceHold')}>
          {t('chat.voiceHold')}
        </span>
        <span className="lesson-chat-char-hint" aria-live="polite">
          {input.length}/{CHAT_MESSAGE_MAX}
        </span>
        <button
          type="button"
          className="lesson-chat-send"
          onClick={() => send(input)}
          disabled={loading || !input.trim() || !user}
          aria-label={t('chat.sendAria')}
        >
          {t('common.send')}
        </button>
      </div>
    </section>
  )
}
