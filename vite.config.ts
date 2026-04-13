import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const configuredBasePath = env.VITE_BASE_PATH?.trim() || '/'
  const basePath = configuredBasePath.endsWith('/')
    ? configuredBasePath
    : `${configuredBasePath}/`

  return {
    plugins: [vue()],
    base: basePath,
    server: {
      proxy: {
        '/api': {
          target: 'http://127.0.0.1:8080',
          changeOrigin: true,
        },
      },
    },
    build: {
      outDir: 'dist',
    },
  }
})
