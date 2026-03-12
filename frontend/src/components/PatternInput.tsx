import { useState, type KeyboardEvent } from 'react'

interface Props {
  label: string
  patterns: string[]
  onChange: (patterns: string[]) => void
  placeholder?: string
}

export function PatternInput({ label, patterns, onChange, placeholder }: Props) {
  const [input, setInput] = useState('')

  const add = () => {
    const val = input.trim()
    if (val && !patterns.includes(val)) {
      onChange([...patterns, val])
    }
    setInput('')
  }

  const remove = (p: string) => onChange(patterns.filter(x => x !== p))

  const onKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') { e.preventDefault(); add() }
    if (e.key === 'Backspace' && input === '' && patterns.length > 0) {
      remove(patterns[patterns.length - 1])
    }
  }

  return (
    <div>
      <label className="block text-sm font-medium text-gray-400 mb-2">{label}</label>
      <div className="min-h-[2.5rem] rounded-lg border border-gray-700 bg-gray-800 px-2 py-1.5 flex flex-wrap gap-1.5 focus-within:border-indigo-500 transition-colors">
        {patterns.map(p => (
          <span key={p} className="inline-flex items-center gap-1 rounded-md bg-gray-700 px-2 py-0.5 text-xs text-gray-200 font-mono">
            {p}
            <button type="button" onClick={() => remove(p)} className="text-gray-400 hover:text-white leading-none">&times;</button>
          </span>
        ))}
        <input
          type="text"
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={onKeyDown}
          onBlur={add}
          placeholder={patterns.length === 0 ? (placeholder ?? 'Type and press Enter') : ''}
          className="flex-1 min-w-[12rem] bg-transparent text-sm text-white outline-none placeholder:text-gray-600"
        />
      </div>
    </div>
  )
}
