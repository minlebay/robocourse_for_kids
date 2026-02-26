import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { LessonPage } from './LessonPage'
import type { Lesson, Module } from '../../shared/types'

// Mock all API calls used by hooks and nested components
vi.mock('../../shared/api', () => ({
  lessons: {
    get: vi.fn(),
    delete: vi.fn(),
    update: vi.fn(),
    setReaction: vi.fn(),
    deleteReaction: vi.fn(),
    // LessonComments uses these via lessonsApi
    getComments: vi.fn(),
    addComment: vi.fn(),
    deleteComment: vi.fn(),
    setCommentReaction: vi.fn(),
    deleteCommentReaction: vi.fn(),
  },
  modules: {
    get: vi.fn(),
  },
  progress: {
    get: vi.fn(),
    setLesson: vi.fn(),
    setChecklist: vi.fn(),
  },
  chat: {
    getHistory: vi.fn(),
    send: vi.fn(),
    clearHistory: vi.fn(),
  },
  comments: {
    list: vi.fn(),
    create: vi.fn(),
    delete: vi.fn(),
    setReaction: vi.fn(),
    deleteReaction: vi.fn(),
  },
}))

vi.mock('../auth/AuthContext', () => ({
  useAuth: vi.fn(),
}))

vi.mock('./PageWithToc', () => ({
  PageWithToc: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

vi.mock('../../components', () => ({
  ConfirmModal: () => null,
}))

const api = await import('../../shared/api')
const useAuth = (await import('../auth/AuthContext').then((m) => m.useAuth)) as ReturnType<typeof vi.fn>

const LESSON: Lesson = {
  id: 'lesson1',
  title: 'Мигаем светодиодом',
  description: 'Первый шаг в Arduino',
  lesson_type: 'theory',
  module_id: 'mod1',
  sort_order: 0,
  created_at: '2025-01-01T00:00:00Z',
  steps: [
    { id: 's1', lesson_id: 'lesson1', title: 'Подключение', content: 'Подключите LED к пину 13.', sort_order: 0 },
  ],
  checklist: [],
  likes_count: 5,
  dislikes_count: 1,
  user_reaction: undefined,
  tags: ['arduino', 'led'],
}

const MODULE: Module = {
  id: 'mod1',
  title: 'Введение в Arduino',
  description: 'Базовый курс',
  sort_order: 0,
  created_at: '2025-01-01T00:00:00Z',
  lessons: [LESSON],
}

function renderLessonPage() {
  return render(
    <MemoryRouter initialEntries={['/lessons/lesson1']}>
      <Routes>
        <Route path="/lessons/:id" element={<LessonPage />} />
      </Routes>
    </MemoryRouter>
  )
}

describe('LessonPage', () => {
  beforeEach(() => {
    useAuth.mockReturnValue({ user: null })
    vi.mocked(api.lessons.get).mockResolvedValue(LESSON)
    vi.mocked(api.lessons.getComments).mockResolvedValue([])
    vi.mocked(api.modules.get).mockResolvedValue(MODULE)
    vi.mocked(api.progress.get).mockResolvedValue({ lessons: [], checklist: [] })
    vi.mocked(api.chat.getHistory).mockResolvedValue({ messages: [] })
  })

  it('renders lesson title after loading', async () => {
    renderLessonPage()
    expect(await screen.findByRole('heading', { name: /мигаем светодиодом/i })).toBeInTheDocument()
  })

  it('renders step content', async () => {
    renderLessonPage()
    expect(await screen.findByRole('heading', { name: /подключение/i })).toBeInTheDocument()
  })

  it('renders lesson tags', async () => {
    renderLessonPage()
    // Tags render as "Теги: arduino, led" in a <p> element
    expect(await screen.findByText(/arduino, led/i)).toBeInTheDocument()
  })

  it('shows not found message when lesson is missing', async () => {
    vi.mocked(api.lessons.get).mockRejectedValue(new Error('Урок не найден'))
    renderLessonPage()
    expect(await screen.findByText(/урок не найден/i)).toBeInTheDocument()
  })

  it('shows edit button for teacher', async () => {
    useAuth.mockReturnValue({ user: { id: 't1', login: 'teacher', name: 'T', role: 'teacher' } })
    vi.mocked(api.progress.get).mockResolvedValue({ lessons: [], checklist: [] })
    renderLessonPage()
    expect(await screen.findByRole('button', { name: /редактировать урок/i })).toBeInTheDocument()
  })

  it('hides edit button for student', async () => {
    useAuth.mockReturnValue({ user: { id: 's1', login: 'student', name: 'S', role: 'student' } })
    vi.mocked(api.progress.get).mockResolvedValue({ lessons: [], checklist: [] })
    renderLessonPage()
    await screen.findByRole('heading', { name: /мигаем светодиодом/i })
    expect(screen.queryByRole('button', { name: /редактировать урок/i })).not.toBeInTheDocument()
  })

  it('shows likes count', async () => {
    renderLessonPage()
    // 5 лайков показываются рядом с кнопкой
    expect(await screen.findByText('5')).toBeInTheDocument()
  })
})
