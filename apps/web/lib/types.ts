export interface User {
  id: string
  username: string
  email: string
  avatar_url: string
  created_at: string
}

export interface Prompt {
  id: string
  name: string
  description: string
  owner_id: string
  owner_username?: string
  tags: string[]
  created_at: string
  download_count: number
}

export interface PromptVersion {
  id: string
  prompt_id: string
  version: string
  tarball_url: string
  created_at: string
}

export interface ApiError {
  error: {
    code: string
    message: string
    request_id: string
    details?: Record<string, string>
  }
}
