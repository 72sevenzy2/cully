import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

export default withMermaid(defineConfig({
  title: 'Cully',
  description: 'Guides and reference for Cully, your companion for better work with AI agents.',
  lang: 'en-US',
  appearance: false,
  cleanUrls: true,
  outDir: '../dist/docs',
  mermaid: {
    theme: 'base',
    securityLevel: 'strict',
    themeVariables: {
      background: '#fffefa',
      primaryColor: '#e4efe2',
      primaryTextColor: '#183b2d',
      primaryBorderColor: '#7eaf88',
      secondaryColor: '#f3e6d2',
      secondaryTextColor: '#183b2d',
      tertiaryColor: '#f7f5ef',
      tertiaryTextColor: '#183b2d',
      lineColor: '#527b5d',
      fontFamily: 'DM Sans, Arial, sans-serif'
    }
  },
  head: [
    ['link', { rel: 'icon', href: '/favicon.svg', type: 'image/svg+xml' }],
    ['meta', { name: 'theme-color', content: '#f7f5ef' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:site_name', content: 'Cully Docs' }]
  ],
  themeConfig: {
    logo: '/favicon.svg',
    siteTitle: 'Cully Docs',
    nav: [
      { text: 'Get started', link: '/quickstart' },
      { text: 'Local advisor', link: '/advisor' },
      { text: 'Shared memory', link: '/memory' },
      { text: 'Self-hosting', link: '/hosting' },
      { text: 'Website', link: 'https://cully.net' }
    ],
    sidebar: [
      {
        text: 'Start here',
        items: [
          { text: 'Overview', link: '/' },
          { text: 'Quickstart', link: '/quickstart' },
          { text: 'Installation', link: '/installation' },
          { text: 'Connect an agent', link: '/agents' }
        ]
      },
      {
        text: 'Use Cully',
        items: [
          { text: 'Local advisor', link: '/advisor' },
          { text: 'Shared memory', link: '/memory' },
          { text: 'Mem0 recall', link: '/mem0' }
        ]
      },
      {
        text: 'Operate Cully',
        items: [
          { text: 'Hosting', link: '/hosting' },
          { text: 'Configuration', link: '/configuration' },
          { text: 'Architecture', link: '/architecture' },
          { text: 'VM migration', link: '/migration' },
          { text: 'Website and docs hosting', link: '/website' }
        ]
      },
      {
        text: 'Contribute',
        items: [
          { text: 'Development', link: '/development' },
          { text: 'Roadmap', link: '/roadmap' },
          { text: 'Cully changelog', link: 'https://github.com/mcp-runtime/cully/blob/main/CHANGELOG.md' }
        ]
      }
    ],
    search: { provider: 'local' },
    socialLinks: [{ icon: 'github', link: 'https://github.com/mcp-runtime/cully' }],
    editLink: {
      pattern: 'https://github.com/mcp-runtime/cully/edit/main/docs/:path',
      text: 'Improve this page'
    },
    footer: {
      message: 'Built in the open. Apache 2.0 licensed.',
      copyright: 'Cully'
    }
  }
}))
