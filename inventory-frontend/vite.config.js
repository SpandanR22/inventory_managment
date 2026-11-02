import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    // Forward /api requests to the Go backend. The browser only ever talks to
    // the Vite server, so there is no cross-origin request and no CORS setup.
    // Set API_URL to point at a backend on another port.
    proxy: {
      '/api': process.env.API_URL ?? 'http://localhost:8086',
    },
  },
})
