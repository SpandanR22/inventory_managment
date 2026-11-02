import { useState } from 'react'
import Categories from './pages/Categories'
import Dashboard from './pages/Dashboard'
import Products from './pages/Products'
import StockMovements from './pages/StockMovements'
import Suppliers from './pages/Suppliers'
import './App.css'

// A plain object instead of a router library: the app has five pages and
// no need for URLs per page yet.
const PAGES = {
  dashboard: { label: 'Dashboard', component: Dashboard },
  products: { label: 'Products', component: Products },
  stock: { label: 'Stock Movements', component: StockMovements },
  categories: { label: 'Categories', component: Categories },
  suppliers: { label: 'Suppliers', component: Suppliers },
}

export default function App() {
  const [page, setPage] = useState('dashboard')
  const Page = PAGES[page].component

  return (
    <div className="layout">
      <aside className="sidebar">
        <h1>Inventory</h1>
        <nav>
          {Object.entries(PAGES).map(([key, { label }]) => (
            <button key={key} className={key === page ? 'active' : ''} onClick={() => setPage(key)}>
              {label}
            </button>
          ))}
        </nav>
      </aside>
      <main className="content">
        {/* key={page} remounts the page on every switch, so it reloads its data. */}
        <Page key={page} />
      </main>
    </div>
  )
}
