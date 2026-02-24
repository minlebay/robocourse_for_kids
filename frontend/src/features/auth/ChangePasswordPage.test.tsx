import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { ChangePasswordPage } from './ChangePasswordPage'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-router-dom')>()
  return { ...actual, useNavigate: () => mockNavigate }
})

vi.mock('../../shared/api', () => ({
  auth: {
    changePassword: vi.fn(),
    me: vi.fn(),
  },
}))

vi.mock('./AuthContext', () => ({
  useAuth: vi.fn(),
}))

const { auth: authApi } = await import('../../shared/api')
const mockChangePassword = authApi.changePassword as ReturnType<typeof vi.fn>
const mockMe = authApi.me as ReturnType<typeof vi.fn>

const { useAuth } = await import('./AuthContext')
const mockUseAuth = useAuth as ReturnType<typeof vi.fn>

const mockUser = { id: '1', login: 'alice', name: 'Alice', role: 'student' as const, created_at: '' }
const mockSetUser = vi.fn()

beforeEach(() => {
  vi.clearAllMocks()
  mockUseAuth.mockReturnValue({ user: mockUser, setUser: mockSetUser })
})

function setup() {
  render(
    <MemoryRouter>
      <ChangePasswordPage />
    </MemoryRouter>
  )
  return {
    currentInput: () => screen.getByLabelText(/текущий пароль/i),
    newInput: () => screen.getByLabelText(/новый пароль/i),
    confirmInput: () => screen.getByLabelText(/подтвердите/i),
    submitBtn: () => screen.getByRole('button', { name: /сменить/i }),
  }
}

describe('ChangePasswordPage', () => {
  it('renders normal title when must_change_password is not set', () => {
    setup()
    expect(screen.getByRole('heading', { name: /смена пароля/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /назад/i })).toBeInTheDocument()
  })

  it('renders forced title and hides back button when must_change_password=true', () => {
    mockUseAuth.mockReturnValue({
      user: { ...mockUser, must_change_password: true },
      setUser: mockSetUser,
    })
    setup()
    expect(screen.getByRole('heading', { name: /требуется смена/i })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /назад/i })).not.toBeInTheDocument()
  })

  it('shows error when new password is too short', async () => {
    const { currentInput, newInput, confirmInput, submitBtn } = setup()
    fireEvent.change(currentInput(), { target: { value: 'oldpass' } })
    fireEvent.change(newInput(), { target: { value: '123' } })
    fireEvent.change(confirmInput(), { target: { value: '123' } })
    fireEvent.click(submitBtn())
    await waitFor(() =>
      expect(screen.getByRole('alert')).toHaveTextContent(/6.*72/)
    )
    expect(mockChangePassword).not.toHaveBeenCalled()
  })

  it('shows error when passwords do not match', async () => {
    const { currentInput, newInput, confirmInput, submitBtn } = setup()
    fireEvent.change(currentInput(), { target: { value: 'oldpass' } })
    fireEvent.change(newInput(), { target: { value: 'newpass1' } })
    fireEvent.change(confirmInput(), { target: { value: 'newpass2' } })
    fireEvent.click(submitBtn())
    await waitFor(() =>
      expect(screen.getByRole('alert')).toHaveTextContent(/не совпадают/i)
    )
    expect(mockChangePassword).not.toHaveBeenCalled()
  })

  it('shows server error when changePassword fails', async () => {
    mockChangePassword.mockRejectedValue(new Error('Неверный текущий пароль'))
    const { currentInput, newInput, confirmInput, submitBtn } = setup()
    fireEvent.change(currentInput(), { target: { value: 'wrong' } })
    fireEvent.change(newInput(), { target: { value: 'newpass1' } })
    fireEvent.change(confirmInput(), { target: { value: 'newpass1' } })
    fireEvent.click(submitBtn())
    await waitFor(() =>
      expect(screen.getByRole('alert')).toHaveTextContent(/неверный/i)
    )
    expect(mockNavigate).not.toHaveBeenCalled()
  })

  it('on success: clears must_change_password, shows success, navigates', async () => {
    vi.useFakeTimers()
    mockChangePassword.mockResolvedValue(undefined)
    mockMe.mockResolvedValue({ ...mockUser, must_change_password: false })
    const { currentInput, newInput, confirmInput, submitBtn } = setup()
    fireEvent.change(currentInput(), { target: { value: 'oldpass' } })
    fireEvent.change(newInput(), { target: { value: 'newpass1' } })
    fireEvent.change(confirmInput(), { target: { value: 'newpass1' } })
    fireEvent.click(submitBtn())
    // vi.runAllTimersAsync сбрасывает все таймеры + промисы разом
    await vi.runAllTimersAsync()
    expect(mockSetUser).toHaveBeenCalledWith(expect.any(Function))
    expect(screen.getByRole('status')).toBeInTheDocument()
    expect(mockNavigate).toHaveBeenCalledWith('/', { replace: true })
    vi.useRealTimers()
  })

  it('still navigates if me() fails after successful changePassword', async () => {
    vi.useFakeTimers()
    mockChangePassword.mockResolvedValue(undefined)
    mockMe.mockRejectedValue(new Error('network error'))
    const { currentInput, newInput, confirmInput, submitBtn } = setup()
    fireEvent.change(currentInput(), { target: { value: 'oldpass' } })
    fireEvent.change(newInput(), { target: { value: 'newpass1' } })
    fireEvent.change(confirmInput(), { target: { value: 'newpass1' } })
    fireEvent.click(submitBtn())
    await vi.runAllTimersAsync()
    expect(screen.getByRole('status')).toBeInTheDocument()
    expect(mockNavigate).toHaveBeenCalledWith('/', { replace: true })
    // Ошибки быть не должно — me() упал, но changePassword прошёл
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    vi.useRealTimers()
  })
})
