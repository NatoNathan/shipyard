export interface NavItem {
  title: string
  slug: string
}

export interface NavSection {
  title: string
  items: NavItem[]
}

export const sidebarNav: NavSection[] = [
  {
    title: 'Getting Started',
    items: [
      { title: 'Configuration', slug: 'configuration' },
      { title: 'Consignment Format', slug: 'consignment-format' },
      { title: 'Tag Generation', slug: 'tag-generation' },
    ],
  },
  {
    title: 'Command Reference',
    items: [
      { title: 'add', slug: 'reference/add' },
      { title: 'completion', slug: 'reference/completion' },
      { title: 'config show', slug: 'reference/config-show' },
      { title: 'init', slug: 'reference/init' },
      { title: 'prerelease', slug: 'reference/prerelease' },
      { title: 'promote', slug: 'reference/promote' },
      { title: 'release', slug: 'reference/release' },
      { title: 'release-notes', slug: 'reference/release-notes' },
      { title: 'remove', slug: 'reference/remove' },
      { title: 'snapshot', slug: 'reference/snapshot' },
      { title: 'status', slug: 'reference/status' },
      { title: 'upgrade', slug: 'reference/upgrade' },
      { title: 'validate', slug: 'reference/validate' },
      { title: 'version', slug: 'reference/version' },
    ],
  },
]
