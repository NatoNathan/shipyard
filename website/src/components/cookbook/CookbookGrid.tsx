import { useState } from 'react'
import {
  Accordion,
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from '@/components/ui/accordion'
import type { CookbookCategory, Recipe } from '@/data/cookbook'

interface Props {
  categories: CookbookCategory[]
}

function CopyIcon() {
  return (
    <svg viewBox="0 0 20 20" fill="currentColor" width="13" height="13" aria-hidden="true">
      <path d="M7 3.5A1.5 1.5 0 018.5 2h3.879a1.5 1.5 0 011.06.44l3.122 3.12A1.5 1.5 0 0117 6.622V12.5a1.5 1.5 0 01-1.5 1.5h-1v-3.379a3 3 0 00-.879-2.121L10.5 5.379A3 3 0 008.379 4.5H7v-1z" />
      <path d="M4.5 6A1.5 1.5 0 003 7.5v9A1.5 1.5 0 004.5 18h7a1.5 1.5 0 001.5-1.5v-5.879a1.5 1.5 0 00-.44-1.06L9.44 6.439A1.5 1.5 0 008.378 6H4.5z" />
    </svg>
  )
}

function CheckIcon() {
  return (
    <svg viewBox="0 0 20 20" fill="currentColor" width="13" height="13" aria-hidden="true">
      <path fillRule="evenodd" d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z" clipRule="evenodd" />
    </svg>
  )
}

function RecipeBlock({ recipe }: { recipe: Recipe }) {
  const [copied, setCopied] = useState(false)

  function copy() {
    navigator.clipboard.writeText(recipe.code).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  return (
    <div className="mt-1 mb-4">
      <p className="text-xs text-[#94a3b8] mb-2">{recipe.description}</p>
      <div className="relative rounded-lg bg-[#020d18] border border-[#0d3558] overflow-hidden">
        <div className="flex items-center justify-between px-3 py-2 border-b border-[#0d3558] bg-[#041424]">
          <span className="text-xs font-mono text-[#4fa3e0]">{recipe.language}</span>
          <button
            onClick={copy}
            className="flex items-center gap-1.5 px-2 py-1 rounded text-xs text-[#94a3b8] hover:text-[#f0f9ff] hover:bg-[#072038] transition-colors"
            aria-label="Copy code"
          >
            {copied ? (
              <>
                <CheckIcon />
                <span>Copied</span>
              </>
            ) : (
              <>
                <CopyIcon />
                <span>Copy</span>
              </>
            )}
          </button>
        </div>
        <pre className="p-4 overflow-x-auto text-xs text-[#93c5fd] leading-relaxed font-mono whitespace-pre">
          <code>{recipe.code}</code>
        </pre>
      </div>
    </div>
  )
}

export default function CookbookGrid({ categories }: Props) {
  return (
    <div className="space-y-10">
      {categories.map((cat) => (
        <section key={cat.id} id={cat.id}>
          <div className="mb-5">
            <h2 className="text-xl font-bold text-[#f0f9ff] mb-1">{cat.title}</h2>
            <p className="text-sm text-[#94a3b8]">{cat.description}</p>
          </div>
          <div className="rounded-xl border border-[#072038] bg-[#041424]/40 divide-y divide-[#072038]">
            <Accordion type="multiple">
              {cat.recipes.map((recipe) => (
                <AccordionItem key={recipe.id} value={recipe.id} className="border-none px-5">
                  <AccordionTrigger className="text-sm font-medium">
                    {recipe.title}
                  </AccordionTrigger>
                  <AccordionContent>
                    <RecipeBlock recipe={recipe} />
                  </AccordionContent>
                </AccordionItem>
              ))}
            </Accordion>
          </div>
        </section>
      ))}
    </div>
  )
}
