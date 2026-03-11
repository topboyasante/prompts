'use client'
import { use, useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getPrompt, listVersions, getDownloadURL, getMe, uploadVersion, deletePrompt } from '@/lib/api'
import type { ApiError } from '@/lib/types'

function NewVersionForm({ promptId }: { promptId: string }) {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)

  const mutation = useMutation({
    mutationFn: async (fd: FormData) => {
      const version = fd.get('version') as string
      const content = fd.get('content') as string
      const uploadFD = new FormData()
      uploadFD.append('version', version)
      uploadFD.append('tarball', new File([content], 'prompt.tar.gz', { type: 'application/gzip' }))
      return uploadVersion(promptId, uploadFD)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['versions'] })
      setOpen(false)
    },
  })

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    mutation.mutate(new FormData(e.currentTarget))
  }

  const apiError = mutation.error as ApiError | null
  const errorMessage = apiError?.error?.message

  if (!open) {
    return (
      <button
        onClick={() => setOpen(true)}
        className="mt-4 rounded-lg border border-zinc-200 px-3 py-1.5 text-sm text-zinc-600 hover:bg-zinc-50 dark:border-zinc-700 dark:text-zinc-400 dark:hover:bg-zinc-800"
      >
        + New Version
      </button>
    )
  }

  return (
    <form onSubmit={handleSubmit} className="mt-4 rounded-lg border border-zinc-200 p-4 dark:border-zinc-800 space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-zinc-900 dark:text-zinc-50">New Version</h3>
        <button type="button" onClick={() => setOpen(false)} className="text-sm text-zinc-400 hover:text-zinc-600">✕</button>
      </div>

      <div>
        <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300">
          Version <span className="text-red-500">*</span>
        </label>
        <input
          name="version" type="text" required pattern="^\d+\.\d+\.\d+$" placeholder="1.1.0"
          className="mt-1 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm outline-none focus:border-zinc-400 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50 dark:focus:border-zinc-500"
        />
        <p className="mt-1 text-xs text-zinc-400">Semver — e.g. 1.1.0</p>
      </div>

      <div>
        <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300">
          Prompt text <span className="text-red-500">*</span>
        </label>
        <textarea
          name="content" required rows={6}
          placeholder={'Write your prompt here.\nUse {{variable}} for inputs.'}
          className="mt-1 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 font-mono text-sm outline-none focus:border-zinc-400 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50 dark:focus:border-zinc-500"
        />
      </div>

      {errorMessage && <p className="text-sm text-red-500">{errorMessage}</p>}

      <div className="flex gap-2">
        <button
          type="submit" disabled={mutation.isPending}
          className="flex items-center gap-2 rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white hover:bg-zinc-700 disabled:opacity-50 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
        >
          {mutation.isPending && (
            <div className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent dark:border-zinc-900 dark:border-t-transparent" />
          )}
          Publish Version
        </button>
        <button type="button" onClick={() => setOpen(false)} className="rounded-lg border border-zinc-200 px-4 py-2 text-sm text-zinc-600 hover:bg-zinc-50 dark:border-zinc-700 dark:text-zinc-400 dark:hover:bg-zinc-800">
          Cancel
        </button>
      </div>
    </form>
  )
}

export default function PromptPage({ params }: { params: Promise<{ owner: string; name: string }> }) {
  const { owner, name } = use(params)
  const router = useRouter()
  const queryClient = useQueryClient()

  const { data: user } = useQuery({ queryKey: ['me'], queryFn: getMe, retry: false })

  const { data: prompt, isLoading: pLoading, isError } = useQuery({
    queryKey: ['prompt', owner, name],
    queryFn: () => getPrompt(owner, name),
  })

  const { data: versions, isLoading: vLoading } = useQuery({
    queryKey: ['versions', owner, name],
    queryFn: () => listVersions(owner, name),
    enabled: !!prompt,
  })

  const deleteMutation = useMutation({
    mutationFn: () => deletePrompt(prompt!.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['prompts'] })
      router.push('/')
    },
  })

  function handleDelete() {
    if (!window.confirm('Delete this prompt? This cannot be undone.')) return
    deleteMutation.mutate()
  }

  const isOwner = !!user && !!prompt && user.id === prompt.owner_id

  if (pLoading) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-10">
        <div className="mb-6 h-4 w-20 animate-pulse rounded bg-zinc-200 dark:bg-zinc-700" />
        <div className="mb-4 h-8 w-1/2 animate-pulse rounded bg-zinc-200 dark:bg-zinc-700" />
        <div className="mb-2 h-4 w-1/4 animate-pulse rounded bg-zinc-200 dark:bg-zinc-700" />
        <div className="mt-6 h-16 animate-pulse rounded bg-zinc-200 dark:bg-zinc-700" />
        <div className="mt-4 flex gap-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-5 w-14 animate-pulse rounded-full bg-zinc-200 dark:bg-zinc-700" />
          ))}
        </div>
      </div>
    )
  }

  if (isError || !prompt) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-10">
        <Link href="/" className="text-sm text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-200">← Back</Link>
        <p className="mt-8 text-center text-zinc-500 dark:text-zinc-400">Prompt not found.</p>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-2xl px-4 py-10">
      <Link href="/" className="text-sm text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-200">
        ← Back
      </Link>

      <div className="mt-6">
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-50">{prompt.name}</h1>
        <div className="mt-1 flex items-center gap-2 text-sm text-zinc-500 dark:text-zinc-400">
          {prompt.owner_username && <span>by {prompt.owner_username}</span>}
          <span>·</span>
          <span>{new Date(prompt.created_at).toLocaleDateString()}</span>
        </div>
        <span className="mt-1 block text-sm text-zinc-500">{prompt.download_count} downloads</span>
        {prompt.description && (
          <p className="mt-4 text-zinc-700 dark:text-zinc-300">{prompt.description}</p>
        )}
        {prompt.tags && prompt.tags.length > 0 && (
          <div className="mt-4 flex flex-wrap gap-1">
            {prompt.tags.map(tag => (
              <span key={tag} className="rounded-full bg-zinc-100 px-2 py-0.5 text-xs text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300">
                {tag}
              </span>
            ))}
          </div>
        )}
      </div>

      <div className="mt-8">
        <h2 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50">Versions</h2>

        {vLoading ? (
          <div className="mt-3 space-y-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="h-10 animate-pulse rounded bg-zinc-200 dark:bg-zinc-700" />
            ))}
          </div>
        ) : !versions?.items || versions.items.length === 0 ? (
          <p className="mt-3 text-sm text-zinc-500 dark:text-zinc-400">No versions published yet.</p>
        ) : (
          <ul className="mt-3 divide-y divide-zinc-100 dark:divide-zinc-800">
            {versions.items.map(v => (
              <li key={v.id} className="flex items-center justify-between py-3">
                <span className="font-mono text-sm text-zinc-800 dark:text-zinc-200">v{v.version}</span>
                <a
                  href={getDownloadURL(owner, name, v.version)}
                  className="text-sm text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-200"
                >
                  Download
                </a>
              </li>
            ))}
          </ul>
        )}

        {isOwner && (
          <div className="flex items-center gap-3">
            <NewVersionForm promptId={prompt.id} />
            <button
              onClick={handleDelete}
              disabled={deleteMutation.isPending}
              className="mt-4 flex items-center gap-2 rounded-lg border border-red-200 px-3 py-1.5 text-sm text-red-600 hover:bg-red-50 disabled:opacity-50 dark:border-red-800 dark:text-red-400 dark:hover:bg-red-950"
            >
              {deleteMutation.isPending && (
                <div className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-red-400 border-t-transparent" />
              )}
              Delete
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
