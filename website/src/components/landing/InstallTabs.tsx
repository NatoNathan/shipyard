import { useState } from 'react'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'

interface InstallMethod {
  id: string
  label: string
  command: string
  comment?: string
}

const methods: InstallMethod[] = [
  {
    id: 'curl',
    label: 'curl',
    command: 'curl -fsSL https://raw.githubusercontent.com/NatoNathan/shipyard/main/install.sh | sh',
    comment: '# Install latest release',
  },
  {
    id: 'brew',
    label: 'Homebrew',
    command: 'brew install natonathan/tap/shipyard',
    comment: '# macOS / Linux',
  },
  {
    id: 'go',
    label: 'go install',
    command: 'go install github.com/NatoNathan/shipyard/cmd/shipyard@latest',
    comment: '# Requires Go 1.21+',
  },
  {
    id: 'npm',
    label: 'npm',
    command: 'npm install -g shipyard-cli',
    comment: '# or: npx shipyard-cli',
  },
  {
    id: 'docker',
    label: 'Docker',
    command: 'docker pull ghcr.io/natonathan/shipyard:latest',
    comment: '# ghcr.io/natonathan/shipyard:latest',
  },
]

function CopyIcon() {
  return (
    <svg viewBox="0 0 20 20" fill="currentColor" width="14" height="14" aria-hidden="true">
      <path d="M7 3.5A1.5 1.5 0 018.5 2h3.879a1.5 1.5 0 011.06.44l3.122 3.12A1.5 1.5 0 0117 6.622V12.5a1.5 1.5 0 01-1.5 1.5h-1v-3.379a3 3 0 00-.879-2.121L10.5 5.379A3 3 0 008.379 4.5H7v-1z" />
      <path d="M4.5 6A1.5 1.5 0 003 7.5v9A1.5 1.5 0 004.5 18h7a1.5 1.5 0 001.5-1.5v-5.879a1.5 1.5 0 00-.44-1.06L9.44 6.439A1.5 1.5 0 008.378 6H4.5z" />
    </svg>
  )
}

function CheckIcon() {
  return (
    <svg viewBox="0 0 20 20" fill="currentColor" width="14" height="14" aria-hidden="true">
      <path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z" clip-rule="evenodd" />
    </svg>
  )
}

export default function InstallTabs() {
  const [copied, setCopied] = useState<string | null>(null)

  function copy(id: string, text: string) {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(id)
      setTimeout(() => setCopied(null), 2000)
    })
  }

  return (
    <Tabs defaultValue="curl">
      <TabsList className="w-full justify-start overflow-x-auto">
        {methods.map((m) => (
          <TabsTrigger key={m.id} value={m.id}>
            {m.label}
          </TabsTrigger>
        ))}
      </TabsList>
      {methods.map((m) => (
        <TabsContent key={m.id} value={m.id}>
          <div className="relative rounded-lg bg-[#041424] border border-[#0d3558] text-left">
            <button
              onClick={() => copy(m.id, m.command)}
              className="absolute top-3 right-3 p-1.5 rounded text-[#94a3b8] hover:text-[#f0f9ff] hover:bg-[#072038] transition-colors"
              aria-label="Copy command"
            >
              {copied === m.id ? <CheckIcon /> : <CopyIcon />}
            </button>
            <div className="px-4 py-3 font-mono text-sm overflow-x-auto">
              {m.comment && (
                <div className="text-[#4fa3e0] mb-1 text-xs">{m.comment}</div>
              )}
              <div className="text-[#93c5fd] pr-8">
                <span className="text-[#38bdf8] select-none">$ </span>
                {m.command}
              </div>
            </div>
          </div>
        </TabsContent>
      ))}
    </Tabs>
  )
}
