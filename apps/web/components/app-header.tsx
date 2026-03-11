'use client'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getLoginURL, getMe, logout } from '@/lib/api'

export function AppHeader() {
  const router = useRouter()
  const queryClient = useQueryClient()

  const { data: user, isLoading } = useQuery({
    queryKey: ['me'],
    queryFn: getMe,
    retry: false,
  })

  async function handleLogout() {
    await logout()
    queryClient.clear()
    router.push('/')
    router.refresh()
  }

  return (
    <header className="border-b border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-950">
      <div className="mx-auto flex max-w-3xl items-center justify-between px-4 py-3">
        <Link href="/" className="text-lg font-bold tracking-tight text-zinc-900 dark:text-zinc-50">
          Prompts
        </Link>

        <div className="flex items-center gap-2">
          {isLoading ? (
            <div className="h-8 w-24 animate-pulse rounded-lg bg-zinc-100 dark:bg-zinc-800" />
          ) : user ? (
            <>
              {user.avatar_url && (
                <img src={user.avatar_url} alt={user.username} className="h-7 w-7 rounded-full" />
              )}
              <span className="text-sm text-zinc-700 dark:text-zinc-300">{user.username}</span>
              <Link
                href="/publish"
                className="rounded-lg bg-zinc-900 px-3 py-1.5 text-sm font-medium text-white hover:bg-zinc-700 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
              >
                + Publish
              </Link>
              <button
                onClick={handleLogout}
                className="rounded-lg border border-zinc-200 px-3 py-1.5 text-sm text-zinc-600 hover:bg-zinc-50 dark:border-zinc-700 dark:text-zinc-400 dark:hover:bg-zinc-800"
              >
                Log out
              </button>
            </>
          ) : (
            <a
              href={getLoginURL('github')}
              className="rounded-lg border border-zinc-200 bg-white px-3 py-1.5 text-sm font-medium text-zinc-700 hover:bg-zinc-50 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-300 dark:hover:bg-zinc-800"
            >
              Sign in with GitHub
            </a>
          )}
        </div>
      </div>
    </header>
  )
}
