import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { CatalogPage } from './CatalogPage'

vi.mock('../../shared/api', () => ({
  modules: {
    list: vi.fn(),
  },
}))

vi.mock('../auth/AuthContext', () => ({
  useAuth: () => ({
    user: { id: 'u1', login: 'teacher1', name: 'Teacher', role: 'teacher' },
  }),
}))

const api = await import('../../shared/api')

describe('CatalogPage', () => {
  beforeEach(() => {
    vi.mocked(api.modules.list).mockResolvedValue([
      {
        id: 'm1',
        title: 'Введение',
        description: 'Первый модуль',
        sort_order: 0,
        created_at: '2025-01-01T00:00:00Z',
      },
    ])
  })

  it('shows catalog title and modules', async () => {
    render(
      <MemoryRouter>
        <CatalogPage />
      </MemoryRouter>
    )
    expect(await screen.findByRole('heading', { name: /каталог уроков/i })).toBeInTheDocument()
    const intros = await screen.findAllByText('Введение')
    expect(intros.length).toBeGreaterThanOrEqual(1)
  })
})
