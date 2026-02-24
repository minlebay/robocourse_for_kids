import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { LessonView } from './LessonView'
import type { Lesson } from '../../shared/types'

// LessonView рендерит LessonChat и LessonComments — мокаем, чтобы не тянуть их зависимости
vi.mock('./LessonChat', () => ({
  LessonChat: () => <div data-testid="lesson-chat" />,
}))

vi.mock('./LessonComments', () => ({
  LessonComments: () => <div data-testid="lesson-comments" />,
}))

vi.mock('../auth/AuthContext', () => ({
  useAuth: () => ({ user: null }),
}))

const baseLesson: Lesson = {
  id: 'l1',
  module_id: 'm1',
  title: 'Урок о роботах',
  description: 'Описание урока',
  lesson_type: 'theory',
  sort_order: 1,
  created_at: '',
}

const defaultProps = {
  lesson: baseLesson,
  status: 'not_started' as const,
  isAuthenticated: false,
  isChecklistDone: () => false,
  onStatusChange: vi.fn(),
  onChecklistToggle: vi.fn(),
  onReactionLike: vi.fn(),
  onReactionDislike: vi.fn(),
  onClearReaction: vi.fn(),
}

describe('LessonView', () => {
  it('renders lesson title and description', () => {
    render(<LessonView {...defaultProps} />)
    expect(screen.getByRole('heading', { level: 1, name: 'Урок о роботах' })).toBeInTheDocument()
    expect(screen.getByText('Описание урока')).toBeInTheDocument()
  })

  it('renders lesson type label', () => {
    render(<LessonView {...defaultProps} />)
    expect(screen.getByText(/теория/i)).toBeInTheDocument()
  })

  it('renders steps in sort_order', () => {
    const lesson: Lesson = {
      ...baseLesson,
      steps: [
        { id: 's2', lesson_id: 'l1', title: 'Шаг 2', content: '', sort_order: 2 },
        { id: 's1', lesson_id: 'l1', title: 'Шаг 1', content: '', sort_order: 1 },
      ],
    }
    render(<LessonView {...defaultProps} lesson={lesson} />)
    const headings = screen.getAllByRole('heading', { level: 3 })
    expect(headings[0]).toHaveTextContent('Шаг 1')
    expect(headings[1]).toHaveTextContent('Шаг 2')
  })

  it('does not render steps section when steps are empty', () => {
    render(<LessonView {...defaultProps} />)
    expect(screen.queryByText(/шаги/i)).not.toBeInTheDocument()
  })

  it('renders checklist items', () => {
    const lesson: Lesson = {
      ...baseLesson,
      checklist: [
        { id: 'c1', lesson_id: 'l1', title: 'Подключи Arduino', sort_order: 1 },
        { id: 'c2', lesson_id: 'l1', title: 'Запусти код', sort_order: 2 },
      ],
    }
    render(<LessonView {...defaultProps} lesson={lesson} isAuthenticated />)
    expect(screen.getByLabelText('Подключи Arduino')).toBeInTheDocument()
    expect(screen.getByLabelText('Запусти код')).toBeInTheDocument()
  })

  it('calls onChecklistToggle when checklist checkbox changes', () => {
    const onChecklistToggle = vi.fn()
    const lesson: Lesson = {
      ...baseLesson,
      checklist: [{ id: 'c1', lesson_id: 'l1', title: 'Задание', sort_order: 1 }],
    }
    render(<LessonView {...defaultProps} lesson={lesson} isAuthenticated onChecklistToggle={onChecklistToggle} />)
    fireEvent.click(screen.getByLabelText('Задание'))
    expect(onChecklistToggle).toHaveBeenCalledWith('c1', true)
  })

  it('disables checklist when not authenticated', () => {
    const lesson: Lesson = {
      ...baseLesson,
      checklist: [{ id: 'c1', lesson_id: 'l1', title: 'Задание', sort_order: 1 }],
    }
    render(<LessonView {...defaultProps} lesson={lesson} isAuthenticated={false} />)
    expect(screen.getByLabelText('Задание')).toBeDisabled()
  })

  it('calls onReactionLike when like button clicked without reaction', () => {
    const onReactionLike = vi.fn()
    render(<LessonView {...defaultProps} isAuthenticated onReactionLike={onReactionLike} />)
    fireEvent.click(screen.getByTitle('Нравится'))
    expect(onReactionLike).toHaveBeenCalled()
  })

  it('calls onClearReaction when like button clicked with active like', () => {
    const onClearReaction = vi.fn()
    const lesson: Lesson = { ...baseLesson, user_reaction: 'like' }
    render(<LessonView {...defaultProps} lesson={lesson} isAuthenticated onClearReaction={onClearReaction} />)
    fireEvent.click(screen.getByTitle('Нравится'))
    expect(onClearReaction).toHaveBeenCalled()
  })

  it('renders LessonChat and LessonComments', () => {
    render(<LessonView {...defaultProps} />)
    expect(screen.getByTestId('lesson-chat')).toBeInTheDocument()
    expect(screen.getByTestId('lesson-comments')).toBeInTheDocument()
  })
})
