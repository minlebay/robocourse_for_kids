import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { LessonChat } from './LessonChat'

vi.mock('../../shared/api', () => ({
  chat: {
    getHistory: vi.fn(),
    send: vi.fn(),
    clearHistory: vi.fn(),
  },
}))

vi.mock('../auth/AuthContext', () => ({
  useAuth: vi.fn(),
}))

const api = await import('../../shared/api')
const useAuth = (await import('../auth/AuthContext').then((m) => m.useAuth)) as ReturnType<typeof vi.fn>

describe('LessonChat', () => {
  beforeEach(() => {
    vi.mocked(api.chat.getHistory).mockResolvedValue({ messages: [] })
  })

  it('shows login hint when user is not authenticated', async () => {
    useAuth.mockReturnValue({ user: null })
    render(<LessonChat lessonId="lesson-1" />)
    expect(
      await screen.findByText(/войдите, чтобы общаться с помощником/i)
    ).toBeInTheDocument()
  })

  it('shows placeholder and send disabled when not authenticated', async () => {
    useAuth.mockReturnValue({ user: null })
    render(<LessonChat lessonId="lesson-1" />)
    await screen.findByPlaceholderText(/войдите, чтобы отправить/i)
    const sendBtn = screen.getByRole('button', { name: /отправить/i })
    expect(sendBtn).toBeDisabled()
  })

  it('shows character count and max length when authenticated', async () => {
    useAuth.mockReturnValue({
      user: { id: 'u1', login: 'u', name: 'U', role: 'student' },
    })
    render(<LessonChat lessonId="lesson-1" />)
    await screen.findByPlaceholderText(/введите вопрос/i)
    expect(screen.getByText('0/1000')).toBeInTheDocument()
  })

  it('sends message on button click when user authenticated', async () => {
    useAuth.mockReturnValue({
      user: { id: 'u1', login: 'u', name: 'U', role: 'student' },
    })
    vi.mocked(api.chat.send).mockResolvedValue({ text: 'Answer' })
    render(<LessonChat lessonId="lesson-1" />)
    await screen.findByPlaceholderText(/введите вопрос/i)
    const input = screen.getByRole('textbox', { name: /текст вопроса/i })
    fireEvent.change(input, { target: { value: 'Hello' } })
    const sendBtn = screen.getByRole('button', { name: /отправить/i })
    fireEvent.click(sendBtn)
    await waitFor(() => {
      expect(api.chat.send).toHaveBeenCalledWith('lesson-1', 'Hello')
    })
  })
})
