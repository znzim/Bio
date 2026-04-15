import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import process from "node:process";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const configuredBasePath = env.VITE_BASE_PATH?.trim() || '/'
  const basePath = configuredBasePath.endsWith('/')
    ? configuredBasePath
    : `${configuredBasePath}/`

  return {
    plugins: [vue()],
    base: basePath,
    build: {
      outDir: 'dist',
    },
  }
})
