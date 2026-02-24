import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { LessonComments } from './LessonComments'

vi.mock('../../shared/api', () => ({
  lessons: {
    getComments: vi.fn(),
    addComment: vi.fn(),
    deleteComment: vi.fn(),
    setCommentReaction: vi.fn(),
    deleteCommentReaction: vi.fn(),
  },
}))

vi.mock('../auth/AuthContext', () => ({
  useAuth: vi.fn(),
}))

const { lessons: lessonsApi } = await import('../../shared/api')
const mockGetComments = lessonsApi.getComments as ReturnType<typeof vi.fn>
const mockAddComment = lessonsApi.addComment as ReturnType<typeof vi.fn>
const mockDeleteComment = lessonsApi.deleteComment as ReturnType<typeof vi.fn>

const { useAuth } = await import('../auth/AuthContext')
const mockUseAuth = useAuth as ReturnType<typeof vi.fn>

const mockUser = { id: 'u1', login: 'alice', name: 'Alice', role: 'student' as const, created_at: '' }

const comment1 = {
  id: 'c1',
  lesson_id: 'l1',
  user_id: 'u1',
  user_name: 'Alice',
  text: 'Отличный урок!',
  created_at: '2024-01-01T10:00:00Z',
  likes_count: 2,
  dislikes_count: 0,
}

beforeEach(() => {
  vi.clearAllMocks()
  mockUseAuth.mockReturnValue({ user: mockUser })
})

describe('LessonComments', () => {
  it('shows loading state initially', () => {
    mockGetComments.mockReturnValue(new Promise(() => {}))
    render(<LessonComments lessonId="l1" />)
    expect(screen.getByText(/загрузка/i)).toBeInTheDocument()
  })

  it('shows empty state when no comments', async () => {
    mockGetComments.mockResolvedValue([])
    render(<LessonComments lessonId="l1" />)
    expect(await screen.findByText(/пока нет комментариев/i)).toBeInTheDocument()
  })

  it('renders list of comments', async () => {
    mockGetComments.mockResolvedValue([comment1])
    render(<LessonComments lessonId="l1" />)
    expect(await screen.findByText('Отличный урок!')).toBeInTheDocument()
    expect(screen.getByText('Alice')).toBeInTheDocument()
  })

  it('shows login prompt when user is not authenticated', async () => {
    mockUseAuth.mockReturnValue({ user: null })
    mockGetComments.mockResolvedValue([])
    render(<LessonComments lessonId="l1" />)
    await waitFor(() => expect(screen.queryByText(/загрузка/i)).not.toBeInTheDocument())
    expect(screen.getByText(/войдите/i)).toBeInTheDocument()
    expect(screen.queryByRole('form')).not.toBeInTheDocument()
  })

  it('adds a comment on submit', async () => {
    const newComment = { ...comment1, id: 'c2', text: 'Новый комментарий' }
    mockGetComments.mockResolvedValue([comment1])
    mockAddComment.mockResolvedValue(newComment)
    render(<LessonComments lessonId="l1" />)
    await screen.findByText('Отличный урок!')
    const textarea = screen.getByPlaceholderText(/напишите/i)
    fireEvent.change(textarea, { target: { value: 'Новый комментарий' } })
    fireEvent.click(screen.getByRole('button', { name: /отправить/i }))
    expect(await screen.findByText('Новый комментарий')).toBeInTheDocument()
    expect(mockAddComment).toHaveBeenCalledWith('l1', 'Новый комментарий')
  })

  it('shows error when loading comments fails', async () => {
    mockGetComments.mockRejectedValue(new Error('Ошибка сети'))
    render(<LessonComments lessonId="l1" />)
    expect(await screen.findByText('Ошибка сети')).toBeInTheDocument()
  })

  it('delete button shown only for own comments', async () => {
    const otherComment = { ...comment1, id: 'c2', user_id: 'u2', user_name: 'Bob', text: 'Чужой' }
    mockGetComments.mockResolvedValue([comment1, otherComment])
    render(<LessonComments lessonId="l1" />)
    await screen.findByText('Отличный урок!')
    // Для своего комментария — кнопка удаления есть
    const deleteButtons = screen.getAllByTitle(/удалить комментарий/i)
    expect(deleteButtons).toHaveLength(1)
  })

  it('removes comment from list after delete', async () => {
    mockGetComments.mockResolvedValue([comment1])
    mockDeleteComment.mockResolvedValue(undefined)
    render(<LessonComments lessonId="l1" />)
    await screen.findByText('Отличный урок!')
    fireEvent.click(screen.getByTitle(/удалить комментарий/i))
    await waitFor(() => expect(screen.queryByText('Отличный урок!')).not.toBeInTheDocument())
  })
})
