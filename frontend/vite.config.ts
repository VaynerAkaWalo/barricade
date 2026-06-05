import { defineConfig } from 'vite'
import react, { reactCompilerPreset } from '@vitejs/plugin-react'
import babel from '@rolldown/plugin-babel'

export default defineConfig({
  resolve: {
    dedupe: ['react', 'react-dom'],
  },
  plugins: [
    react(),
    babel({ presets: [reactCompilerPreset()] })
  ],

  build: {
    outDir: "../backend/public",
    emptyOutDir: true
  },

  server: {
    proxy: {
      '/login': 'http://localhost:3000',
      '/users': 'http://localhost:3000',
      '/authenticate': 'http://localhost:3000',
      '/health': 'http://localhost:3000',
      '/logout': 'http://localhost:3000',
    },
  }
})
