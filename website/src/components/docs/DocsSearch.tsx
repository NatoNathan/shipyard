import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'

export type SearchEntry = {
  title: string
  tagline: string
  description: string
  url: string
  section: string
}

function scoreEntry(entry: SearchEntry, query: string): number {
  const q = query.toLowerCase()
  const title = entry.title.toLowerCase()
  const tagline = entry.tagline.toLowerCase()
  const desc = entry.description.toLowerCase()

  if (title === q) return 100
  if (title.startsWith(q)) return 80
  if (title.includes(q)) return 60
  if (tagline.includes(q)) return 30
  if (desc.includes(q)) return 10
  return 0
}

function search(entries: SearchEntry[], query: string): SearchEntry[] {
  const q = query.trim()
  if (!q) return []
  return entries
    .map((entry) => ({ entry, score: scoreEntry(entry, q) }))
    .filter(({ score }) => score > 0)
    .sort((a, b) => b.score - a.score)
    .map(({ entry }) => entry)
    .slice(0, 8)
}

interface Props {
  searchIndex?: SearchEntry[]
}

export default function DocsSearch({ searchIndex = [] }: Props) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)
  const results = search(searchIndex, query)

  useEffect(() => {
    function handleKey(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setOpen((o) => !o)
      }
      if (e.key === 'Escape') setOpen(false)
    }
    window.addEventListener('keydown', handleKey)
    return () => window.removeEventListener('keydown', handleKey)
  }, [])

  useEffect(() => {
    if (open) {
      setTimeout(() => inputRef.current?.focus(), 50)
    } else {
      setQuery('')
    }
  }, [open])

  return (
    <>
      {/* Search trigger button */}
      <button
        onClick={() => setOpen(true)}
        className="w-full flex items-center gap-2 px-3 py-2 rounded-lg border border-[#072038] bg-[#020d18] text-sm text-[#94a3b8] hover:border-[#1a5080] hover:text-[#f0f9ff] transition-colors"
        aria-label="Search documentation"
      >
        <svg viewBox="0 0 20 20" fill="currentColor" width="14" height="14" aria-hidden="true">
          <path fillRule="evenodd" d="M9 3.5a5.5 5.5 0 100 11 5.5 5.5 0 000-11zM2 9a7 7 0 1112.452 4.391l3.328 3.329a.75.75 0 11-1.06 1.06l-3.329-3.328A7 7 0 012 9z" clipRule="evenodd" />
        </svg>
        <span className="flex-1 text-left">Search docs…</span>
        <kbd className="text-xs font-mono bg-[#041424] border border-[#0d3558] rounded px-1.5 py-0.5">⌘K</kbd>
      </button>

      {/* Modal — portalled to body to escape sidebar stacking context */}
      {open && createPortal(
        <div
          className="fixed inset-0 z-[200] bg-black/85 flex items-center justify-center px-4"
          onClick={() => setOpen(false)}
        >
          <div
            className="w-full max-w-2xl -mt-[8vh]"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="bg-[#041424] border border-[#0d3558] rounded-xl shadow-2xl overflow-hidden">
              {/* Search input */}
              <div className="flex items-center gap-3 px-4 py-3 border-b border-[#072038]">
                <svg viewBox="0 0 20 20" fill="currentColor" width="16" height="16" className="text-[#4fa3e0] shrink-0" aria-hidden="true">
                  <path fillRule="evenodd" d="M9 3.5a5.5 5.5 0 100 11 5.5 5.5 0 000-11zM2 9a7 7 0 1112.452 4.391l3.328 3.329a.75.75 0 11-1.06 1.06l-3.329-3.328A7 7 0 012 9z" clipRule="evenodd" />
                </svg>
                <input
                  ref={inputRef}
                  type="text"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Search docs…"
                  className="flex-1 bg-transparent text-[#f0f9ff] placeholder-[#475569] text-base outline-none"
                  aria-label="Search documentation"
                />
                {query && (
                  <button
                    onClick={() => setQuery('')}
                    className="text-[#475569] hover:text-[#94a3b8] text-sm transition-colors"
                  >
                    Clear
                  </button>
                )}
                <kbd className="text-xs font-mono text-[#475569] border border-[#0d3558] rounded px-1.5 py-0.5 shrink-0">Esc</kbd>
              </div>

              {/* Results */}
              {query.trim() === '' ? (
                <div className="px-4 py-8 text-center text-sm text-[#475569]">
                  Type to search…
                </div>
              ) : results.length === 0 ? (
                <div className="px-4 py-8 text-center text-sm text-[#475569]">
                  No results for <span className="text-[#94a3b8]">"{query}"</span>
                </div>
              ) : (
                <ul className="max-h-[60vh] overflow-y-auto divide-y divide-[#072038]">
                  {results.map((entry) => (
                    <li key={entry.url}>
                      <a
                        href={entry.url}
                        onClick={() => setOpen(false)}
                        className="flex items-start gap-3 px-4 py-3 hover:bg-[#072038] transition-colors group"
                      >
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2 mb-0.5">
                            <span className="font-mono font-semibold text-[#f0f9ff] group-hover:text-[#38bdf8] transition-colors text-sm">
                              {entry.title}
                            </span>
                            <span className="text-xs text-[#334155] bg-[#0d3558] rounded px-1.5 py-0.5 shrink-0">
                              {entry.section}
                            </span>
                          </div>
                          {entry.tagline && (
                            <p className="text-xs text-[#64748b] truncate">{entry.tagline}</p>
                          )}
                        </div>
                        <svg viewBox="0 0 20 20" fill="currentColor" width="14" height="14" className="text-[#334155] group-hover:text-[#38bdf8] transition-colors shrink-0 mt-0.5" aria-hidden="true">
                          <path fillRule="evenodd" d="M3 10a.75.75 0 01.75-.75h10.638L10.23 5.29a.75.75 0 111.04-1.08l5.5 5.25a.75.75 0 010 1.08l-5.5 5.25a.75.75 0 11-1.04-1.08l4.158-3.96H3.75A.75.75 0 013 10z" clipRule="evenodd" />
                        </svg>
                      </a>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>
        </div>,
        document.body
      )}
    </>
  )
}
