import { useState } from 'react'

interface CopyButtonProps {
  text: string
  className?: string
}

export const CopyButton = ({ text, className }: CopyButtonProps) => {
  const [copied, setCopied] = useState(false)

  const copy = async () => {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <button
      onClick={copy}
      aria-label={copied ? 'Copied!' : 'Copy to clipboard'}
      className={className}
      style={{
        padding: '0.2rem 0.55rem',
        fontSize: '0.7rem',
        lineHeight: '1.5',
        color: copied ? '#38bdf8' : '#94a3b8',
        background: 'transparent',
        border: `1px solid ${copied ? '#38bdf8' : '#0d3558'}`,
        borderRadius: '4px',
        cursor: 'pointer',
        transition: 'color 0.15s, border-color 0.15s',
        fontFamily: 'ui-sans-serif, system-ui, sans-serif',
      }}
    >
      {copied ? 'Copied!' : 'Copy'}
    </button>
  )
}
