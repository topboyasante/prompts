'use client'
import { useQuery, keepPreviousData } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { searchPrompts } from '@/lib/api'
import { PromptCard } from './prompt-card'

export function SearchBar() {
  const [input, setInput] = useState('')
  const [query, setQuery] = useState('')

  useEffect(() => {
    const id = setTimeout(() => setQuery(input), 300)
    return () => clearTimeout(id)
  }, [input])

  const { data, isLoading, isFetching } = useQuery({
    queryKey: ['prompts', query],
    queryFn: () => searchPrompts(query),
    placeholderData: keepPreviousData,
  })

  const prompts = data?.items ?? []

  return (
    <div className="w-full">
      <div className="relative">
        <input
          type="text"
          value={input}
          onChange={e => setInput(e.target.value)}
          placeholder="Search prompts..."
          className="w-full rounded-lg border border-zinc-200 bg-white px-4 py-2.5 text-sm text-zinc-900 placeholder-zinc-400 outline-none focus:border-zinc-400 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50 dark:placeholder-zinc-500 dark:focus:border-zinc-500"
        />
        {isFetching && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2">
            <div className="h-4 w-4 animate-spin rounded-full border-2 border-zinc-300 border-t-zinc-600 dark:border-zinc-600 dark:border-t-zinc-300" />
          </div>
        )}
      </div>

      <div className="mt-4">
        {isLoading ? (
          <div className="grid gap-3 sm:grid-cols-2">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="h-24 animate-pulse rounded-lg bg-zinc-100 dark:bg-zinc-800" />
            ))}
          </div>
        ) : prompts.length === 0 ? (
          <p className="text-center text-sm text-zinc-500 dark:text-zinc-400">
            {query ? 'No prompts found.' : 'No prompts yet.'}
          </p>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {prompts.map(prompt => (
              <PromptCard key={prompt.id} prompt={prompt} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
