import { useEffect, useRef, useState } from 'react'

export default function DocsSearch() {
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const initialized = useRef(false)

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
    if (!open || initialized.current) return
    if (typeof window === 'undefined') return

    // PagefindUI is loaded via script tag in BaseLayout
    const win = window as unknown as { PagefindUI?: new (opts: object) => void }
    if (!win.PagefindUI) return

    initialized.current = true
    new win.PagefindUI({
      element: '#pagefind-search',
      showSubResults: true,
      resetStyles: false,
    })
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
          <path fill-rule="evenodd" d="M9 3.5a5.5 5.5 0 100 11 5.5 5.5 0 000-11zM2 9a7 7 0 1112.452 4.391l3.328 3.329a.75.75 0 11-1.06 1.06l-3.329-3.328A7 7 0 012 9z" clip-rule="evenodd" />
        </svg>
        <span className="flex-1 text-left">Search docs…</span>
        <kbd className="text-xs font-mono bg-[#041424] border border-[#0d3558] rounded px-1.5 py-0.5">⌘K</kbd>
      </button>

      {/* Modal overlay */}
      {open && (
        <div
          className="fixed inset-0 z-50 flex items-start justify-center pt-20 px-4"
          onClick={(e) => e.target === e.currentTarget && setOpen(false)}
        >
          <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setOpen(false)} />
          <div
            ref={containerRef}
            className="relative w-full max-w-xl bg-[#041424] border border-[#0d3558] rounded-xl shadow-2xl overflow-hidden"
          >
            <div id="pagefind-search" className="p-2" />
          </div>
        </div>
      )}
    </>
  )
}
