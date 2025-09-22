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

type DashboardResponse = {
  summary: {
    inboxZeroTarget: number
    currentInbox: number
    automationRate: number
    timeSavedMinutes: number
  }
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
}

type AutomationsResponse = {
  templates: Array<{
    id: string
    name: string
    description: string
    trigger: string
    actions: string[]
    requiresApproval: boolean
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
