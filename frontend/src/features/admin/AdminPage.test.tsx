import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { AdminPage } from './AdminPage'

vi.mock('../../shared/api', () => ({
  admin: {
    listUsers: vi.fn(),
    getStats: vi.fn(),
    getActivity: vi.fn(),
    blockUser: vi.fn(),
    deleteUser: vi.fn(),
    resetPassword: vi.fn(),
    createUser: vi.fn(),
  },
}))

vi.mock('../auth/AuthContext', () => ({
  useAuth: vi.fn(),
}))

vi.mock('../../shared/ErrorBoundary', () => ({
  ErrorBoundary: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('./TempPasswordModal', () => ({
  TempPasswordModal: ({ password, onClose }: { password: string; onClose: () => void }) => (
    <div role="dialog" aria-label="temp-password">
      <span>{password}</span>
      <button onClick={onClose}>Close</button>
    </div>
  ),
}))

vi.mock('./CreateUserModal', () => ({
  CreateUserModal: ({ onClose }: { onClose: () => void }) => (
    <div role="dialog" aria-label="create-user">
      <button onClick={onClose}>Close</button>
    </div>
  ),
}))

vi.mock('../../components/ConfirmModal', () => ({
  ConfirmModal: ({ open, onConfirm, onCancel, children }: {
    open: boolean
    onConfirm: () => void
    onCancel: () => void
    children: React.ReactNode
  }) =>
    open ? (
      <div role="dialog">
        {children}
        <button onClick={onConfirm}>ConfirmDelete</button>
        <button onClick={onCancel}>Cancel</button>
      </div>
    ) : null,
}))

const api = await import('../../shared/api')
const useAuth = (await import('../auth/AuthContext').then((m) => m.useAuth)) as ReturnType<typeof vi.fn>

const ADMIN_USER = { id: 'admin1', login: 'admin', name: 'Admin', role: 'administrator' }

const USERS = [
  { id: 'u1', login: 'alice', name: 'Alice', role: 'student', is_blocked: false },
  { id: 'u2', login: 'bob', name: 'Bob', role: 'teacher', is_blocked: true },
]

const STATS = { users_count: 2, modules_count: 3, lessons_count: 10 }

describe('AdminPage', () => {
  beforeEach(() => {
    useAuth.mockReturnValue({ user: ADMIN_USER })
    vi.mocked(api.admin.listUsers).mockResolvedValue(USERS)
    vi.mocked(api.admin.getStats).mockResolvedValue(STATS)
    vi.mocked(api.admin.getActivity).mockResolvedValue([])
  })

  it('renders admin panel heading', async () => {
    render(<AdminPage />)
    expect(await screen.findByRole('heading', { name: /панель администратора/i })).toBeInTheDocument()
  })

  it('renders user list after loading', async () => {
    render(<AdminPage />)
    expect(await screen.findByText('@alice')).toBeInTheDocument()
    expect(await screen.findByText('@bob')).toBeInTheDocument()
  })

  it('shows create user button', async () => {
    render(<AdminPage />)
    expect(await screen.findByRole('button', { name: /создать пользователя/i })).toBeInTheDocument()
  })

  it('opens create user modal on button click', async () => {
    render(<AdminPage />)
    const btn = await screen.findByRole('button', { name: /создать пользователя/i })
    fireEvent.click(btn)
    expect(await screen.findByRole('dialog', { name: /create-user/i })).toBeInTheDocument()
  })

  it('shows error when users fail to load', async () => {
    vi.mocked(api.admin.listUsers).mockRejectedValue(new Error('Нет доступа'))
    render(<AdminPage />)
    expect(await screen.findByText(/нет доступа/i)).toBeInTheDocument()
  })

  it('calls blockUser when block button is clicked', async () => {
    vi.mocked(api.admin.blockUser).mockResolvedValue(undefined)

    render(<AdminPage />)
    await screen.findByText('@alice')

    // Alice is not blocked — her button title is "Заблокировать"
    const aliceBlockBtn = screen.getByRole('button', { name: 'Заблокировать' })
    fireEvent.click(aliceBlockBtn)

    await waitFor(() => {
      expect(api.admin.blockUser).toHaveBeenCalledWith('u1', true)
    })
  })

  it('calls deleteUser after confirmation', async () => {
    vi.mocked(api.admin.deleteUser).mockResolvedValue(undefined)
    vi.mocked(api.admin.listUsers)
      .mockResolvedValueOnce(USERS)
      .mockResolvedValueOnce([USERS[1]])

    render(<AdminPage />)
    await screen.findByText('@alice')

    const deleteBtns = screen.getAllByTitle(/удалить/i)
    fireEvent.click(deleteBtns[0])

    const confirmBtn = await screen.findByRole('button', { name: /confirmdelete/i })
    fireEvent.click(confirmBtn)

    await waitFor(() => {
      expect(api.admin.deleteUser).toHaveBeenCalledWith('u1')
    })
  })

  it('shows temp password modal after reset', async () => {
    vi.mocked(api.admin.resetPassword).mockResolvedValue({ temp_password: 'TempPass123!' })

    render(<AdminPage />)
    await screen.findByText('@alice')

    const resetBtns = screen.getAllByTitle(/сбросить пароль/i)
    fireEvent.click(resetBtns[0])

    expect(await screen.findByRole('dialog', { name: /temp-password/i })).toBeInTheDocument()
    expect(await screen.findByText('TempPass123!')).toBeInTheDocument()
  })

  it('disables action buttons while pendingAction is active', async () => {
    // Simulate slow blockUser
    let resolveBlock: () => void
    vi.mocked(api.admin.blockUser).mockReturnValue(
      new Promise<undefined>((res) => { resolveBlock = () => res(undefined) })
    )

    render(<AdminPage />)
    await screen.findByText('@alice')

    const blockBtns = screen.getAllByTitle(/заблокировать/i)
    fireEvent.click(blockBtns[0])

    // While pending, alice's buttons should be disabled
    await waitFor(() => {
      expect(blockBtns[0]).toBeDisabled()
    })

    resolveBlock!()
  })
})
