import { useEffect, useState } from 'react'
import { api } from '../api'
import { EmptyRow, ErrorBanner, PageHeader } from '../components'
import { formatCents } from '../format'

export default function Dashboard() {
  const [summary, setSummary] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api.dashboard().then(setSummary).catch((err) => setError(err.message))
  }, [])

  if (error) return <ErrorBanner message={error} />
  if (!summary) return <p className="muted">Loading…</p>

  const { totals, low_stock: lowStock, by_category: byCategory, low_stock_threshold: threshold } = summary
  const maxValue = Math.max(1, ...byCategory.map((row) => row.stock_value_cents))

  return (
    <>
      <PageHeader title="Dashboard" />

      <section className="stats">
        <Stat label="Products" value={totals.products} />
        <Stat label="Units in stock" value={totals.units.toLocaleString('en-IN')} />
        <Stat label="Stock value" value={formatCents(totals.stock_value_cents)} />
        <Stat label="Categories" value={totals.categories} />
        <Stat label="Suppliers" value={totals.suppliers} />
        <Stat label="Low stock items" value={lowStock.length} warn={lowStock.length > 0} />
      </section>

      <div className="grid-2">
        <section className="card">
          <h3>Low stock (≤ {threshold} units)</h3>
          <table>
            <thead>
              <tr>
                <th>Product</th>
                <th>SKU</th>
                <th className="num">Qty</th>
              </tr>
            </thead>
            <tbody>
              {lowStock.length === 0 && <EmptyRow colSpan={3}>Everything is well stocked.</EmptyRow>}
              {lowStock.map((item) => (
                <tr key={item.id}>
                  <td>{item.name}</td>
                  <td className="mono">{item.sku}</td>
                  <td className={item.quantity === 0 ? 'num danger' : 'num warn'}>{item.quantity}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>

        <section className="card">
          <h3>Stock value by category</h3>
          <table>
            <thead>
              <tr>
                <th>Category</th>
                <th className="num">Products</th>
                <th className="num">Units</th>
                <th>Value</th>
              </tr>
            </thead>
            <tbody>
              {byCategory.length === 0 && <EmptyRow colSpan={4}>No products yet.</EmptyRow>}
              {byCategory.map((row) => (
                <tr key={row.category}>
                  <td>{row.category}</td>
                  <td className="num">{row.products}</td>
                  <td className="num">{row.units}</td>
                  <td>
                    <div className="bar-cell">
                      <div className="bar" style={{ width: `${(row.stock_value_cents / maxValue) * 100}%` }} />
                      <span>{formatCents(row.stock_value_cents)}</span>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      </div>
    </>
  )
}

function Stat({ label, value, warn = false }) {
  return (
    <div className={warn ? 'stat warn' : 'stat'}>
      <span className="stat-label">{label}</span>
      <span className="stat-value">{value}</span>
    </div>
  )
}
