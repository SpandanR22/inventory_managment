// Small building blocks shared by every page.

export function PageHeader({ title, children }) {
  return (
    <header className="page-header">
      <h2>{title}</h2>
      <div className="page-actions">{children}</div>
    </header>
  )
}

// Field shows a label, the input passed as children, and the backend's
// validation message for this field, if there is one.
export function Field({ label, error, children }) {
  return (
    <label className={error ? 'field has-error' : 'field'}>
      <span className="field-label">{label}</span>
      {children}
      {error && <span className="field-error">{error}</span>}
    </label>
  )
}

export function ErrorBanner({ message }) {
  if (!message) return null
  return (
    <div className="banner error" role="alert">
      {message}
    </div>
  )
}

export function EmptyRow({ colSpan, children }) {
  return (
    <tr>
      <td colSpan={colSpan} className="empty">
        {children}
      </td>
    </tr>
  )
}
