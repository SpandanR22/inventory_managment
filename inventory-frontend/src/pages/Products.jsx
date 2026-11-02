import { useCallback, useEffect, useState } from 'react'
import { api } from '../api'
import { EmptyRow, ErrorBanner, Field, PageHeader } from '../components'
import { emptyToNull, formatCents, idOrNull, toCents } from '../format'

const EMPTY_FORM = { name: '', sku: '', price: '', quantity: '0', category_id: '', supplier_id: '', description: '' }

export default function Products() {
  const [products, setProducts] = useState([])
  const [categories, setCategories] = useState([])
  const [suppliers, setSuppliers] = useState([])
  const [loadError, setLoadError] = useState('')

  const [form, setForm] = useState(EMPTY_FORM)
  const [editingId, setEditingId] = useState(null) // null = creating
  const [formError, setFormError] = useState(null) // ApiError from the last save
  const [saving, setSaving] = useState(false)

  const load = useCallback(async () => {
    try {
      const [p, c, s] = await Promise.all([api.listProducts(), api.listCategories(), api.listSuppliers()])
      setProducts(p)
      setCategories(c)
      setSuppliers(s)
    } catch (err) {
      setLoadError(err.message)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const categoryName = (id) => categories.find((c) => c.id === id)?.name ?? '—'
  const supplierName = (id) => suppliers.find((s) => s.id === id)?.name ?? '—'

  const update = (field) => (event) => setForm({ ...form, [field]: event.target.value })

  function startEdit(product) {
    setEditingId(product.id)
    setFormError(null)
    setForm({
      name: product.name,
      sku: product.sku,
      price: (product.price_cents / 100).toFixed(2),
      quantity: String(product.quantity),
      category_id: product.category_id ?? '',
      supplier_id: product.supplier_id ?? '',
      description: product.description ?? '',
    })
  }

  function resetForm() {
    setEditingId(null)
    setForm(EMPTY_FORM)
    setFormError(null)
  }

  async function handleSubmit(event) {
    event.preventDefault()
    setSaving(true)
    setFormError(null)

    const body = {
      name: form.name,
      sku: form.sku,
      price_cents: toCents(form.price) || 0,
      category_id: idOrNull(form.category_id),
      supplier_id: idOrNull(form.supplier_id),
      description: emptyToNull(form.description),
    }

    try {
      if (editingId) {
        // Quantity is not sent on update: stock only changes via stock movements.
        await api.updateProduct(editingId, body)
      } else {
        await api.createProduct({ ...body, quantity: Number(form.quantity) || 0 })
      }
      resetForm()
      await load()
    } catch (err) {
      setFormError(err)
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(product) {
    if (!window.confirm(`Delete "${product.name}"? Its stock history is deleted too.`)) return
    try {
      await api.deleteProduct(product.id)
      if (editingId === product.id) resetForm()
      await load()
    } catch (err) {
      setLoadError(err.message)
    }
  }

  const fieldError = (name) => formError?.fields?.[name]

  return (
    <>
      <PageHeader title="Products" />
      <ErrorBanner message={loadError} />

      <form className="card form" onSubmit={handleSubmit}>
        <h3>{editingId ? 'Edit product' : 'Add product'}</h3>
        <ErrorBanner message={formError && Object.keys(formError.fields).length === 0 ? formError.message : ''} />

        <div className="form-grid">
          <Field label="Name" error={fieldError('name')}>
            <input value={form.name} onChange={update('name')} required />
          </Field>
          <Field label="SKU" error={fieldError('sku')}>
            <input value={form.sku} onChange={update('sku')} required />
          </Field>
          <Field label="Price (₹)" error={fieldError('price_cents')}>
            <input type="number" min="0" step="0.01" value={form.price} onChange={update('price')} required />
          </Field>
          <Field label={editingId ? 'Quantity (use Stock Movements)' : 'Opening quantity'} error={fieldError('quantity')}>
            <input type="number" min="0" step="1" value={form.quantity} onChange={update('quantity')} disabled={!!editingId} />
          </Field>
          <Field label="Category" error={fieldError('category_id')}>
            <select value={form.category_id} onChange={update('category_id')}>
              <option value="">— None —</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </Field>
          <Field label="Supplier" error={fieldError('supplier_id')}>
            <select value={form.supplier_id} onChange={update('supplier_id')}>
              <option value="">— None —</option>
              {suppliers.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
          </Field>
          <Field label="Description" error={fieldError('description')}>
            <textarea rows="2" value={form.description} onChange={update('description')} />
          </Field>
        </div>

        <div className="form-actions">
          <button type="submit" className="primary" disabled={saving}>
            {saving ? 'Saving…' : editingId ? 'Save changes' : 'Add product'}
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
              <th>SKU</th>
              <th>Category</th>
              <th>Supplier</th>
              <th className="num">Price</th>
              <th className="num">Qty</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {products.length === 0 && <EmptyRow colSpan={7}>No products yet. Add one above.</EmptyRow>}
            {products.map((p) => (
              <tr key={p.id} className={p.id === editingId ? 'selected' : ''}>
                <td>
                  <div>{p.name}</div>
                  {p.description && <div className="muted small">{p.description}</div>}
                </td>
                <td className="mono">{p.sku}</td>
                <td>{categoryName(p.category_id)}</td>
                <td>{supplierName(p.supplier_id)}</td>
                <td className="num">{formatCents(p.price_cents)}</td>
                <td className={p.quantity === 0 ? 'num danger' : 'num'}>{p.quantity}</td>
                <td className="row-actions">
                  <button onClick={() => startEdit(p)}>Edit</button>
                  <button className="danger" onClick={() => handleDelete(p)}>
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
