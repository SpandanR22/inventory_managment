// Every call to the Go backend goes through request(), so error handling
// lives in one place.

const BASE_URL = '/api/v1'

// ApiError carries the HTTP status and, for 409/422 responses, the
// per-field messages the backend sends as {"error": "...", "fields": {...}}.
export class ApiError extends Error {
  constructor(message, status, fields) {
    super(message)
    this.status = status
    this.fields = fields ?? {}
  }
}

async function request(path, { method = 'GET', body } = {}) {
  const response = await fetch(BASE_URL + path, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  })

  if (response.status === 204) {
    return null
  }

  const data = await response.json().catch(() => null)
  if (!response.ok) {
    throw new ApiError(data?.error ?? `Request failed with status ${response.status}`, response.status, data?.fields)
  }
  return data
}

export const api = {
  dashboard: () => request('/dashboard'),

  listProducts: async () => (await request('/products')).products,
  createProduct: (product) => request('/products', { method: 'POST', body: product }),
  updateProduct: (id, product) => request(`/products/${id}`, { method: 'PUT', body: product }),
  deleteProduct: (id) => request(`/products/${id}`, { method: 'DELETE' }),

  listCategories: async () => (await request('/categories')).categories,
  createCategory: (category) => request('/categories', { method: 'POST', body: category }),
  updateCategory: (id, category) => request(`/categories/${id}`, { method: 'PUT', body: category }),
  deleteCategory: (id) => request(`/categories/${id}`, { method: 'DELETE' }),

  listSuppliers: async () => (await request('/suppliers')).suppliers,
  createSupplier: (supplier) => request('/suppliers', { method: 'POST', body: supplier }),
  updateSupplier: (id, supplier) => request(`/suppliers/${id}`, { method: 'PUT', body: supplier }),
  deleteSupplier: (id) => request(`/suppliers/${id}`, { method: 'DELETE' }),

  listMovements: async (productId) => {
    const query = productId ? `?product_id=${productId}` : ''
    return (await request(`/stock-movements${query}`)).movements
  },
  createMovement: (movement) => request('/stock-movements', { method: 'POST', body: movement }),
}
