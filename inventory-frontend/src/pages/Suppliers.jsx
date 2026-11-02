import { useCallback, useEffect, useState } from 'react'
import { api } from '../api'
import { EmptyRow, ErrorBanner, Field, PageHeader } from '../components'
import { emptyToNull } from '../format'

const EMPTY_FORM = { name: '', email: '', phone: '' }

export default function Suppliers() {
  const [suppliers, setSuppliers] = useState([])
  const [loadError, setLoadError] = useState('')
  const [form, setForm] = useState(EMPTY_FORM)
  const [editingId, setEditingId] = useState(null)
  const [formError, setFormError] = useState(null)

  const load = useCallback(async () => {
    try {
      setSuppliers(await api.listSuppliers())
    } catch (err) {
      setLoadError(err.message)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const update = (field) => (event) => setForm({ ...form, [field]: event.target.value })

  function startEdit(supplier) {
    setEditingId(supplier.id)
    setFormError(null)
    setForm({ name: supplier.name, email: supplier.email ?? '', phone: supplier.phone ?? '' })
  }

  function resetForm() {
    setEditingId(null)
    setForm(EMPTY_FORM)
    setFormError(null)
  }

  async function handleSubmit(event) {
    event.preventDefault()
    const body = { name: form.name, email: emptyToNull(form.email), phone: emptyToNull(form.phone) }
    try {
      if (editingId) {
        await api.updateSupplier(editingId, body)
      } else {
        await api.createSupplier(body)
      }
      resetForm()
      await load()
    } catch (err) {
      setFormError(err)
    }
  }

  async function handleDelete(supplier) {
    if (!window.confirm(`Delete "${supplier.name}"? Its products will have no supplier.`)) return
    try {
      await api.deleteSupplier(supplier.id)
      if (editingId === supplier.id) resetForm()
      await load()
    } catch (err) {
      setLoadError(err.message)
    }
  }

  return (
    <>
      <PageHeader title="Suppliers" />
      <ErrorBanner message={loadError} />

      <form className="card form" onSubmit={handleSubmit}>
        <h3>{editingId ? 'Edit supplier' : 'Add supplier'}</h3>
        <div className="form-grid">
          <Field label="Name" error={formError?.fields?.name}>
            <input value={form.name} onChange={update('name')} required />
          </Field>
          <Field label="Email" error={formError?.fields?.email}>
            <input type="email" value={form.email} onChange={update('email')} />
          </Field>
          <Field label="Phone" error={formError?.fields?.phone}>
            <input value={form.phone} onChange={update('phone')} />
          </Field>
        </div>
        <div className="form-actions">
          <button type="submit" className="primary">
            {editingId ? 'Save changes' : 'Add supplier'}
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
              <th>Email</th>
              <th>Phone</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {suppliers.length === 0 && <EmptyRow colSpan={4}>No suppliers yet.</EmptyRow>}
            {suppliers.map((s) => (
              <tr key={s.id} className={s.id === editingId ? 'selected' : ''}>
                <td>{s.name}</td>
                <td>{s.email ?? '—'}</td>
                <td>{s.phone ?? '—'}</td>
                <td className="row-actions">
                  <button onClick={() => startEdit(s)}>Edit</button>
                  <button className="danger" onClick={() => handleDelete(s)}>
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
