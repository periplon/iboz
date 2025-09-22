import { useCallback, useEffect, useMemo, useState } from 'react'

type ApiState<T> = {
  data?: T
  loading: boolean
  error?: string
}

async function fetchJSON<T>(endpoint: string, init?: RequestInit): Promise<T> {
  const response = await fetch(endpoint, {
    headers: {
      'Content-Type': 'application/json',
    },
    ...init,
  })

  if (!response.ok) {
    const message = await response.text()
    throw new Error(message || `Request failed with status ${response.status}`)
  }

  return response.json() as Promise<T>
}

export function useApi<T>(endpoint: string): ApiState<T> {
  const [state, setState] = useState<ApiState<T>>({ loading: true })

  useEffect(() => {
    let cancelled = false

    setState({ loading: true })

    fetchJSON<T>(endpoint)
      .then((data) => {
        if (!cancelled) {
          setState({ data, loading: false })
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          const message = err instanceof Error ? err.message : 'Unknown error'
          setState({ error: message, loading: false })
        }
      })

    return () => {
      cancelled = true
    }
  }, [endpoint])

  return state
}

export function useDashboard() {
  return useApi<DashboardResponse>('/api/dashboard')
}

export function useFocusPlan() {
  return useApi<FocusPlanResponse>('/api/focus/plan')
}

export function useAutomations() {
  return useApi<AutomationsResponse>('/api/automations')
}

export function useAutomationTestRunner() {
  const [state, setState] = useState<ApiState<AutomationTestResponse>>({ loading: false })

  const runTest = useCallback(async (templateId: string, parameters: Record<string, unknown>) => {
    setState({ loading: true })
    try {
      const data = await fetchJSON<AutomationTestResponse>('/api/automations/test-run', {
        method: 'POST',
        body: JSON.stringify({ templateId, parameters }),
      })
      setState({ data, loading: false })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unknown error'
      setState({ error: message, loading: false })
    }
  }, [])

  return useMemo(
    () => ({
      ...state,
      runTest,
    }),
    [state, runTest],
  )
}

export function useEmailProvider() {
  return useApi<EmailProviderState>('/api/email/provider')
}

export function updateEmailProvider(config: EmailProviderConfig) {
  return fetchJSON<EmailProviderState>('/api/email/provider', {
    method: 'POST',
    body: JSON.stringify(config),
  })
}

export type EmailAuthenticatePayload = {
  method: EmailAuthMethod
  username: string
  appPassword?: string
  oauthToken?: string
}

export function authenticateEmailProvider(payload: EmailAuthenticatePayload) {
  return fetchJSON<EmailProviderState>('/api/email/provider/authenticate', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function fetchEmailMessages() {
  return fetchJSON<EmailMessagesResponse>('/api/email/messages')
}

type DashboardResponse = {
  summary: {
    inboxZeroTarget: number
    currentInbox: number
    automationRate: number
    timeSavedMinutes: number
  }
  focusSessions: Array<{
    id: string
    label: string
    start: string
    estimated: number
    emails: number
    llmSupport: boolean
    description: string
  }>
  queues: Array<{
    id: string
    label: string
    description: string
    count: number
    llmEnabled: boolean
  }>
  recommendations: Array<{
    id: string
    title: string
    description: string
    confidence: number
  }>
}

type FocusPlanResponse = {
  date: string
  sessions: Array<{
    id: string
    label: string
    estimated: number
    emails: number
    llmSupport: boolean
    description: string
  }>
  metrics: {
    clearedToday: number
    streak: number
    goal: number
  }
  controls: {
    notificationsMuted: boolean
    batchingEnabled: boolean
    autoSummaries: boolean
  }
}

type AutomationsResponse = {
  overview: {
    active: number
    automationCoverage: number
    avgTimeSaved: number
  }
  templates: Array<{
    id: string
    name: string
    description: string
    trigger: string
    conditions: string[]
    actions: string[]
    requiresApproval: boolean
    owner: string
    lastRun: string
  }>
}

type AutomationTestResponse = {
  templateId: string
  status: string
  summary: string
  parameters: Record<string, unknown>
  review: {
    requiresApproval: boolean
    confidence: number
  }
}

export type EmailProviderState = {
  config?: EmailProviderConfig
  auth?: EmailAuthState
  lastSync?: string
  messagesFetched: number
}

export type EmailProviderConfig = {
  provider: EmailProviderType
  displayName: string
  connection: EmailConnectionSettings
  syncWindowHours: number
  labelFilters: string[]
}

export type EmailProviderType = 'gmail' | 'outlook' | 'imap'

export type EmailConnectionSettings = {
  protocol: EmailConnectionProtocol
  host?: string
  port?: number
  useTls?: boolean
  apiBaseUrl?: string
}

export type EmailConnectionProtocol = 'api' | 'imap'

export type EmailAuthState = {
  method: EmailAuthMethod
  username?: string
  status: string
  updatedAt: string
}

export type EmailAuthMethod = 'oauth' | 'appPassword'

export type EmailMessagesResponse = {
  messages: EmailMessage[]
  syncedAt?: string
}

export type EmailMessage = {
  id: string
  subject: string
  sender: string
  receivedAt: string
  snippet: string
  labels: string[]
  importance: string
}
