import { useCallback, useEffect, useState } from 'react'
import { api } from '../api'
import { EmptyRow, ErrorBanner, Field, PageHeader } from '../components'
import { emptyToNull } from '../format'

const EMPTY_FORM = { name: '', description: '' }

export default function Categories() {
  const [categories, setCategories] = useState([])
  const [loadError, setLoadError] = useState('')
  const [form, setForm] = useState(EMPTY_FORM)
  const [editingId, setEditingId] = useState(null)
  const [formError, setFormError] = useState(null)

  const load = useCallback(async () => {
    try {
      setCategories(await api.listCategories())
    } catch (err) {
      setLoadError(err.message)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  function startEdit(category) {
    setEditingId(category.id)
    setFormError(null)
    setForm({ name: category.name, description: category.description ?? '' })
  }

  function resetForm() {
    setEditingId(null)
    setForm(EMPTY_FORM)
    setFormError(null)
  }

  async function handleSubmit(event) {
    event.preventDefault()
    const body = { name: form.name, description: emptyToNull(form.description) }
    try {
      if (editingId) {
        await api.updateCategory(editingId, body)
      } else {
        await api.createCategory(body)
      }
      resetForm()
      await load()
    } catch (err) {
      setFormError(err)
    }
  }

  async function handleDelete(category) {
    if (!window.confirm(`Delete "${category.name}"? Its products will become uncategorized.`)) return
    try {
      await api.deleteCategory(category.id)
      if (editingId === category.id) resetForm()
      await load()
    } catch (err) {
      setLoadError(err.message)
    }
  }

  return (
    <>
      <PageHeader title="Categories" />
      <ErrorBanner message={loadError} />

      <form className="card form" onSubmit={handleSubmit}>
        <h3>{editingId ? 'Edit category' : 'Add category'}</h3>
        <div className="form-grid">
          <Field label="Name" error={formError?.fields?.name}>
            <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
          </Field>
          <Field label="Description" error={formError?.fields?.description}>
            <input value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
          </Field>
        </div>
        <div className="form-actions">
          <button type="submit" className="primary">
            {editingId ? 'Save changes' : 'Add category'}
          </button>
          {editingId && (
            <button type="button" onClick={resetForm}>
              Cancel
            </button>
          )}
        </div>
      </form>

      <section className="card">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Description</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {categories.length === 0 && <EmptyRow colSpan={3}>No categories yet.</EmptyRow>}
            {categories.map((c) => (
              <tr key={c.id} className={c.id === editingId ? 'selected' : ''}>
                <td>{c.name}</td>
                <td className="muted">{c.description ?? '—'}</td>
                <td className="row-actions">
                  <button onClick={() => startEdit(c)}>Edit</button>
                  <button className="danger" onClick={() => handleDelete(c)}>
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </>
  )
}
