import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { RequireAuth } from './RequireAuth'

vi.mock('../features/auth/AuthContext', () => ({
  useAuth: vi.fn(),
}))

const useAuth = await import('../features/auth/AuthContext').then((m) => m.useAuth) as ReturnType<
  typeof vi.fn
>

describe('RequireAuth', () => {
  it('shows loading then renders children when user is present', async () => {
    useAuth.mockReturnValue({ user: { id: '1', login: 'u', name: 'U', role: 'student' }, loading: false })
    render(
      <MemoryRouter>
        <RequireAuth>
          <span>Protected</span>
        </RequireAuth>
      </MemoryRouter>
    )
    expect(screen.getByText('Protected')).toBeInTheDocument()
  })

  it('shows loading while loading', () => {
    useAuth.mockReturnValue({ user: null, loading: true })
    render(
      <MemoryRouter>
        <RequireAuth>
          <span>Protected</span>
        </RequireAuth>
      </MemoryRouter>
    )
    expect(screen.getByText('Загрузка...')).toBeInTheDocument()
    expect(screen.queryByText('Protected')).not.toBeInTheDocument()
  })

  it('redirects to /login when not authenticated', () => {
    useAuth.mockReturnValue({ user: null, loading: false })
    render(
      <MemoryRouter initialEntries={['/progress']}>
        <RequireAuth>
          <span>Protected</span>
        </RequireAuth>
      </MemoryRouter>
    )
    expect(screen.queryByText('Protected')).not.toBeInTheDocument()
  })
})
