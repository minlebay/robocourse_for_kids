import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { modules } from '../../shared/api'
import type { Module } from '../../shared/types'

export function CatalogPage() {
  const [list, setList] = useState<Module[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    modules
      .list()
      .then(setList)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <p>Загрузка...</p>
  if (error) return <p className="error">{error}</p>

  return (
    <div className="catalog">
      <h1>Каталог уроков</h1>
      <div className="module-list">
        {list.map((mod) => (
          <section key={mod.id} className="module-card">
            <h2><Link to={`/modules/${mod.id}`}>{mod.title}</Link></h2>
            {mod.description && <p>{mod.description}</p>}
          </section>
        ))}
      </div>
    </div>
  )
}
