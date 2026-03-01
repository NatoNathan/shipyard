import { defineConfig } from 'astro/config'
import react from '@astrojs/react'
import mdx from '@astrojs/mdx'
import tailwindcss from '@tailwindcss/vite'
import pagefind from 'astro-pagefind'

export default defineConfig({
  site: 'https://shipyard.natonathan.dev',
  output: 'static',
  integrations: [react(), mdx(), pagefind()],
  vite: {
    plugins: [tailwindcss()],
  },
  markdown: {
    shikiConfig: {
      theme: 'one-dark-pro',
      wrap: false,
    },
  },
})
