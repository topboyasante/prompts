export default function Loading() {
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
      <div className="mt-8">
        <div className="mb-3 h-6 w-24 animate-pulse rounded bg-zinc-200 dark:bg-zinc-700" />
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-10 animate-pulse rounded bg-zinc-200 dark:bg-zinc-700" />
          ))}
        </div>
      </div>
    </div>
  )
}
