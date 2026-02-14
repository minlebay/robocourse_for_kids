import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { RequireTeacher } from './RequireTeacher'

vi.mock('../features/auth/AuthContext', () => ({
  useAuth: vi.fn(),
}))

const useAuth = await import('../features/auth/AuthContext').then((m) => m.useAuth) as ReturnType<
  typeof vi.fn
>

describe('RequireTeacher', () => {
  it('renders children when user is teacher', () => {
    useAuth.mockReturnValue({
      user: { id: '1', login: 't', name: 'Teacher', role: 'teacher' },
      loading: false,
    })
    render(
      <MemoryRouter>
        <RequireTeacher>
          <span>Dashboard</span>
        </RequireTeacher>
      </MemoryRouter>
    )
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
  })

  it('redirects to /login when not authenticated', () => {
    useAuth.mockReturnValue({ user: null, loading: false })
    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <RequireTeacher>
          <span>Dashboard</span>
        </RequireTeacher>
      </MemoryRouter>
    )
    expect(screen.queryByText('Dashboard')).not.toBeInTheDocument()
  })

  it('redirects to / when user is student', () => {
    useAuth.mockReturnValue({
      user: { id: '1', login: 's', name: 'Student', role: 'student' },
      loading: false,
    })
    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <RequireTeacher>
          <span>Dashboard</span>
        </RequireTeacher>
      </MemoryRouter>
    )
    expect(screen.queryByText('Dashboard')).not.toBeInTheDocument()
  })
})
