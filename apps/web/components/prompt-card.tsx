import Link from 'next/link'
import type { Prompt } from '@/lib/types'

export function PromptCard({ prompt }: { prompt: Prompt }) {
  const card = (
    <div className="rounded-lg border border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
      <h3 className="font-semibold text-zinc-900 dark:text-zinc-50">{prompt.name}</h3>
      {prompt.description && (
        <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">{prompt.description}</p>
      )}
      {prompt.tags && prompt.tags.length > 0 && (
        <div className="mt-3 flex flex-wrap gap-1">
          {prompt.tags.map(tag => (
            <span
              key={tag}
              className="rounded-full bg-zinc-100 px-2 py-0.5 text-xs text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300"
            >
              {tag}
            </span>
          ))}
        </div>
      )}
      <span className="mt-2 block text-xs text-zinc-400">
        {prompt.download_count} download{prompt.download_count !== 1 ? 's' : ''}
      </span>
    </div>
  )

  if (!prompt.owner_username) return card
  return (
    <Link href={`/${prompt.owner_username}/${prompt.name}`} className="block">
      {card}
    </Link>
  )
}
