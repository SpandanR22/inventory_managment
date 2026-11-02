import { useCallback, useEffect, useState } from 'react'
import { api } from '../api'
import { EmptyRow, ErrorBanner, Field, PageHeader } from '../components'
import { emptyToNull, formatDate } from '../format'

const EMPTY_FORM = { product_id: '', type: 'IN', quantity: '', note: '' }

const TYPE_HELP = {
  IN: 'Goods received. Adds to stock.',
  OUT: 'Goods sold or shipped. Removes from stock.',
  ADJUST: 'Correction after a stock count. Use a negative number to reduce stock.',
}

export default function StockMovements() {
  const [products, setProducts] = useState([])
  const [movements, setMovements] = useState([])
  const [filterProductId, setFilterProductId] = useState('')
  const [loadError, setLoadError] = useState('')
  const [form, setForm] = useState(EMPTY_FORM)
  const [formError, setFormError] = useState(null)
  const [success, setSuccess] = useState('')

  const load = useCallback(async () => {
    try {
      const [p, m] = await Promise.all([api.listProducts(), api.listMovements(filterProductId)])
      setProducts(p)
      setMovements(m)
    } catch (err) {
      setLoadError(err.message)
    }
  }, [filterProductId])

  useEffect(() => {
    load()
  }, [load])

  const update = (field) => (event) => setForm({ ...form, [field]: event.target.value })
  const selectedProduct = products.find((p) => String(p.id) === form.product_id)

  async function handleSubmit(event) {
    event.preventDefault()
    setFormError(null)
    setSuccess('')
    try {
      const movement = await api.createMovement({
        product_id: Number(form.product_id) || 0,
        type: form.type,
        quantity: Number(form.quantity) || 0,
        note: emptyToNull(form.note),
      })
      setSuccess(`Recorded. ${movement.product_name} now has ${movement.balance_after} units.`)
      setForm({ ...EMPTY_FORM, product_id: form.product_id, type: form.type })
      await load()
    } catch (err) {
      setFormError(err)
    }
  }

  return (
    <>
      <PageHeader title="Stock Movements" />
      <ErrorBanner message={loadError} />

      <form className="card form" onSubmit={handleSubmit}>
        <h3>Record a movement</h3>
        {success && <div className="banner success">{success}</div>}
        <div className="form-grid">
          <Field label="Product" error={formError?.fields?.product_id}>
            <select value={form.product_id} onChange={update('product_id')} required>
              <option value="">Select a product…</option>
              {products.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name} ({p.sku})
                </option>
              ))}
            </select>
          </Field>
          <Field label="Type" error={formError?.fields?.type}>
            <select value={form.type} onChange={update('type')}>
              <option value="IN">IN: received</option>
              <option value="OUT">OUT: sold / shipped</option>
              <option value="ADJUST">ADJUST: correction</option>
            </select>
          </Field>
          <Field label="Quantity" error={formError?.fields?.quantity}>
            <input type="number" step="1" value={form.quantity} onChange={update('quantity')} required />
          </Field>
          <Field label="Note" error={formError?.fields?.note}>
            <input value={form.note} onChange={update('note')} placeholder="e.g. Invoice #1042" />
          </Field>
        </div>
        <p className="muted small">
          {TYPE_HELP[form.type]}
          {selectedProduct && ` Current stock: ${selectedProduct.quantity}.`}
        </p>
        <div className="form-actions">
          <button type="submit" className="primary">
            Record movement
          </button>
        </div>
      </form>

      <section className="card">
        <div className="card-header">
          <h3>History</h3>
          <select value={filterProductId} onChange={(e) => setFilterProductId(e.target.value)}>
            <option value="">All products</option>
            {products.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </div>
        <table>
          <thead>
            <tr>
              <th>When</th>
              <th>Product</th>
              <th>Type</th>
              <th className="num">Change</th>
              <th className="num">Balance</th>
              <th>Note</th>
            </tr>
          </thead>
          <tbody>
            {movements.length === 0 && <EmptyRow colSpan={6}>No movements recorded yet.</EmptyRow>}
            {movements.map((m) => (
              <tr key={m.id}>
                <td className="muted">{formatDate(m.created_at)}</td>
                <td>{m.product_name}</td>
                <td>
                  <span className={`tag tag-${m.type.toLowerCase()}`}>{m.type}</span>
                </td>
                <td className={m.change > 0 ? 'num positive' : 'num danger'}>
                  {m.change > 0 ? `+${m.change}` : m.change}
                </td>
                <td className="num">{m.balance_after}</td>
                <td className="muted">{m.note ?? '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </>
  )
}
