const rupees = new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR' })

// Prices are stored as integer paise (cents) to avoid floating-point errors.
export const formatCents = (cents) => rupees.format(cents / 100)

export const toCents = (rupeeString) => Math.round(parseFloat(rupeeString) * 100)

export const formatDate = (iso) =>
  new Date(iso).toLocaleString('en-IN', { dateStyle: 'medium', timeStyle: 'short' })

// Optional text fields: send null rather than "" so the database stores NULL.
export const emptyToNull = (value) => (value.trim() === '' ? null : value.trim())

// Optional <select> ids: "" means "none".
export const idOrNull = (value) => (value === '' ? null : Number(value))
