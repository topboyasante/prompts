'use client'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createPrompt } from '@/lib/api'
import type { ApiError } from '@/lib/types'

export default function NewPromptPage() {
  const router = useRouter()
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: createPrompt,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['prompts'] })
      router.push('/')
    },
  })

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const data = new FormData(e.currentTarget)
    const tags = (data.get('tags') as string)
      .split(',').map(t => t.trim()).filter(Boolean)
    mutation.mutate({
      name: data.get('name') as string,
      description: data.get('description') as string,
      tags,
    })
  }

  const apiError = mutation.error as ApiError | null
  const is401 = apiError?.error?.code === 'UNAUTHORIZED'
  const errorMessage = is401
    ? 'Please log in to create prompts'
    : apiError?.error?.message

  return (
    <div className="mx-auto max-w-lg px-4 py-10">
      <Link href="/" className="text-sm text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-200">
        ← Back
      </Link>

      <h1 className="mt-6 text-2xl font-bold text-zinc-900 dark:text-zinc-50">New Prompt</h1>

      <form onSubmit={handleSubmit} className="mt-6 space-y-5">
        <div>
          <label htmlFor="name" className="block text-sm font-medium text-zinc-700 dark:text-zinc-300">
            Name
          </label>
          <input
            id="name"
            name="name"
            type="text"
            required
            pattern="^[a-z0-9-]+$"
            placeholder="my-prompt"
            className="mt-1 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none focus:border-zinc-400 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50 dark:focus:border-zinc-500"
          />
          <p className="mt-1 text-xs text-zinc-400">Lowercase letters, numbers, hyphens</p>
        </div>

        <div>
          <label htmlFor="description" className="block text-sm font-medium text-zinc-700 dark:text-zinc-300">
            Description
          </label>
          <textarea
            id="description"
            name="description"
            rows={3}
            className="mt-1 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none focus:border-zinc-400 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50 dark:focus:border-zinc-500"
          />
        </div>

        <div>
          <label htmlFor="tags" className="block text-sm font-medium text-zinc-700 dark:text-zinc-300">
            Tags
          </label>
          <input
            id="tags"
            name="tags"
            type="text"
            placeholder="coding, gpt-4"
            className="mt-1 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none focus:border-zinc-400 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50 dark:focus:border-zinc-500"
          />
          <p className="mt-1 text-xs text-zinc-400">Comma-separated (e.g. coding, gpt-4)</p>
        </div>

        {errorMessage && (
          <p className="text-sm text-red-500">{errorMessage}</p>
        )}

        <button
          type="submit"
          disabled={mutation.isPending}
          className="flex w-full items-center justify-center gap-2 rounded-lg bg-zinc-900 px-4 py-2.5 text-sm font-medium text-white hover:bg-zinc-700 disabled:opacity-50 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
        >
          {mutation.isPending && (
            <div className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent dark:border-zinc-900 dark:border-t-transparent" />
          )}
          Create Prompt
        </button>
      </form>
    </div>
  )
}
