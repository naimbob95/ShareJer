import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte(), tailwindcss()],
  server: {
    port: 5173,
    // In dev, forward /api/* to the Go server so the browser stays same-origin
    // (no CORS). In prod, nginx does the same forwarding (see frontend/nginx.conf).
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
