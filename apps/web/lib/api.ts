import axios from 'axios'
import type { Prompt, PromptVersion, User } from './types'

export const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/v1',
  withCredentials: true,
  headers: { 'Content-Type': 'application/json' },
})

apiClient.interceptors.response.use(
  r => r,
  err => Promise.reject(err.response?.data ?? err)
)

export const getMe = () =>
  apiClient
    .get<{ data: User }>('/me')
    .then(r => r.data.data)

export const searchPrompts = (q: string, limit = 20, offset = 0) =>
  apiClient
    .get<{ data: { items: Prompt[]; query: string; limit: number; offset: number } }>(
      '/prompts', { params: { q, limit, offset } }
    )
    .then(r => r.data.data)

export const getPrompt = (owner: string, name: string) =>
  apiClient
    .get<{ data: Prompt }>(`/prompts/${owner}/${name}`)
    .then(r => r.data.data)

export const listVersions = (owner: string, name: string) =>
  apiClient
    .get<{ data: { items: PromptVersion[] } }>(`/prompts/${owner}/${name}/versions`)
    .then(r => r.data.data)

export const createPrompt = (body: { name: string; description: string; tags?: string[] }) =>
  apiClient
    .post<{ data: Prompt }>('/prompts', body)
    .then(r => r.data.data)

export const uploadVersion = (promptId: string, formData: FormData) =>
  apiClient
    .post<{ data: PromptVersion }>(`/prompts/${promptId}/versions`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    .then(r => r.data.data)

export const publishPrompt = (body: { name: string; description: string; tags: string[]; content: string }) =>
  apiClient
    .post<{ data: PromptVersion }>('/prompts/publish', body)
    .then(r => r.data.data)

export const deletePrompt = (id: string) =>
  apiClient.delete(`/prompts/${id}`)

export const logout = () =>
  apiClient.post('/auth/logout')

export const getLoginURL = (provider: 'github' | 'google') =>
  `${process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/v1'}/auth/${provider}/login`

export const getDownloadURL = (owner: string, name: string, version: string) =>
  `${process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/v1'}/prompts/${owner}/${name}/versions/${version}/download`
