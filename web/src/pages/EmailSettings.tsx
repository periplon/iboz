import { useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import {
  authenticateEmailProvider,
  fetchEmailMessages,
  updateEmailProvider,
  useEmailProvider,
} from '../api/hooks'
import type {
  EmailAuthMethod,
  EmailMessage,
  EmailProviderConfig,
  EmailProviderState,
  EmailProviderType,
} from '../api/hooks'

const providerOptions: Array<{
  value: EmailProviderType
  label: string
  description: string
}> = [
  {
    value: 'gmail',
    label: 'Gmail / Google Workspace',
    description: 'OAuth connection with least privileged scopes for delegated sending and read access.',
  },
  {
    value: 'outlook',
    label: 'Microsoft 365 / Outlook',
    description: 'Azure AD application with Graph API permissions scoped to the automation mailbox.',
  },
  {
    value: 'imap',
    label: 'Generic IMAP',
    description: 'Use app passwords for legacy systems that do not support modern auth.',
  },
]

const oauthProviders: EmailProviderType[] = ['gmail', 'outlook']

function createDefaultConfig(provider: EmailProviderType = 'gmail'): EmailProviderConfig {
  if (provider === 'outlook') {
    return {
      provider,
      displayName: 'Outlook automation inbox',
      connection: {
        protocol: 'api',
        apiBaseUrl: 'https://graph.microsoft.com/v1.0',
      },
      syncWindowHours: 24,
      labelFilters: ['Inbox', 'Automation'],
    }
  }

  if (provider === 'imap') {
    return {
      provider,
      displayName: 'IMAP automation inbox',
      connection: {
        protocol: 'imap',
        host: 'imap.example.com',
        port: 993,
        useTls: true,
      },
      syncWindowHours: 24,
      labelFilters: ['INBOX'],
    }
  }

  return {
    provider: 'gmail',
    displayName: 'Gmail automation inbox',
    connection: {
      protocol: 'api',
      apiBaseUrl: 'https://gmail.googleapis.com',
    },
    syncWindowHours: 24,
    labelFilters: ['INBOX', 'Urgent'],
  }
}

function normaliseLabelInput(input: string) {
  return input
    .split(/\r?\n|,/)
    .map((item) => item.trim())
    .filter((item) => item.length > 0)
}

function formatTimestamp(timestamp?: string) {
  if (!timestamp) return '—'
  const date = new Date(timestamp)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleString()
}

export function EmailSettings() {
  const { data, loading, error } = useEmailProvider()
  const [state, setState] = useState<EmailProviderState | undefined>(undefined)
  const [form, setForm] = useState<EmailProviderConfig>(() => createDefaultConfig())
  const [labelInput, setLabelInput] = useState('INBOX\nUrgent')
  const [configStatus, setConfigStatus] = useState<string | undefined>()
  const [configError, setConfigError] = useState<string | undefined>()
  const [authMethod, setAuthMethod] = useState<EmailAuthMethod>('oauth')
  const [authUsername, setAuthUsername] = useState('')
  const [authSecret, setAuthSecret] = useState('')
  const [authStatus, setAuthStatus] = useState<string | undefined>()
  const [authError, setAuthError] = useState<string | undefined>()
  const [messages, setMessages] = useState<EmailMessage[]>([])
  const [syncedAt, setSyncedAt] = useState<string | undefined>(undefined)
  const [messageError, setMessageError] = useState<string | undefined>()
  const [fetchingMessages, setFetchingMessages] = useState(false)

  useEffect(() => {
    if (data) {
      setState(data)
    }
  }, [data])

  useEffect(() => {
    if (state?.config) {
      setForm({
        ...state.config,
        connection: { ...state.config.connection },
        labelFilters: [...state.config.labelFilters],
      })
      setLabelInput(state.config.labelFilters.join('\n') || '')
    }
  }, [state?.config])

  useEffect(() => {
    if (!state?.config && !loading) {
      const defaults = createDefaultConfig()
      setForm(defaults)
      setLabelInput(defaults.labelFilters.join('\n'))
    }
  }, [loading, state?.config])

  useEffect(() => {
    if (state?.auth) {
      setAuthMethod(state.auth.method)
      setAuthUsername(state.auth.username ?? '')
      setAuthStatus(`Connected • updated ${formatTimestamp(state.auth.updatedAt)}`)
    } else {
      const preferredMethod: EmailAuthMethod = state?.config && oauthProviders.includes(state.config.provider) ? 'oauth' : 'appPassword'
      setAuthMethod(preferredMethod)
      setAuthStatus(undefined)
    }
  }, [state?.auth, state?.config])

  const integrationStatus = useMemo(() => {
    if (loading && !state) {
      return 'Loading…'
    }
    if (!state?.config) {
      return 'Not configured'
    }
    if (state?.auth) {
      return `Connected via ${state.auth.method === 'oauth' ? 'OAuth' : 'app password'}`
    }
    return 'Awaiting authentication'
  }, [loading, state])

  const lastSyncDisplay = useMemo(() => formatTimestamp(state?.lastSync), [state?.lastSync])

  const handleProviderChange = (provider: EmailProviderType) => {
    setForm((prev) => {
      const defaults = createDefaultConfig(provider)
      return {
        ...defaults,
        syncWindowHours: prev.syncWindowHours || defaults.syncWindowHours,
        labelFilters: prev.labelFilters.length ? prev.labelFilters : defaults.labelFilters,
      }
    })
    const defaults = createDefaultConfig(provider)
    setLabelInput((current) => (current.trim().length > 0 ? current : defaults.labelFilters.join('\n')))
    setAuthMethod(oauthProviders.includes(provider) ? 'oauth' : 'appPassword')
    setAuthSecret('')
  }

  const handleConfigSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setConfigError(undefined)
    setConfigStatus(undefined)
    try {
      const payload: EmailProviderConfig = {
        ...form,
        labelFilters: normaliseLabelInput(labelInput),
        syncWindowHours: Math.max(0, Number(form.syncWindowHours) || 0),
      }
      const updated = await updateEmailProvider(payload)
      setState(updated)
      setConfigStatus('Connection settings saved')
      setMessages([])
      setSyncedAt(undefined)
    } catch (err) {
      setConfigError(err instanceof Error ? err.message : 'Failed to save configuration')
    }
  }

  const handleAuthSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setAuthError(undefined)
    setAuthStatus(undefined)
    try {
      const payload = {
        method: authMethod,
        username: authUsername,
        appPassword: authMethod === 'appPassword' ? authSecret : undefined,
        oauthToken: authMethod === 'oauth' ? authSecret : undefined,
      }
      const updated = await authenticateEmailProvider(payload)
      setState(updated)
      setAuthSecret('')
      setAuthStatus(`Connected • updated ${formatTimestamp(updated.auth?.updatedAt)}`)
    } catch (err) {
      setAuthError(err instanceof Error ? err.message : 'Authentication failed')
    }
  }

  const handleFetchMessages = async () => {
    setFetchingMessages(true)
    setMessageError(undefined)
    try {
      const response = await fetchEmailMessages()
      setMessages(response.messages)
      setSyncedAt(response.syncedAt)
      setState((prev) =>
        prev
          ? {
              ...prev,
              lastSync: response.syncedAt ?? prev.lastSync,
              messagesFetched: response.messages.length,
            }
          : prev,
      )
    } catch (err) {
      setMessageError(err instanceof Error ? err.message : 'Unable to fetch messages')
    } finally {
      setFetchingMessages(false)
    }
  }

  return (
    <div className="space-y-6">
      <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
          <div>
            <h2 className="text-xl font-semibold text-slate-900">Email provider integration</h2>
            <p className="mt-2 text-sm text-slate-600">
              Configure how the automation backend connects to your operational mailbox. The Go service validates settings and
              surfaces a synthesized inbox feed so teams can align on requirements before wiring real APIs.
            </p>
          </div>
          <div className="flex flex-col gap-2 text-sm text-slate-500">
            <span className="rounded-full bg-slate-100 px-4 py-1 text-center font-medium">{integrationStatus}</span>
            <span className="rounded-full bg-slate-100 px-4 py-1 text-center">Last sync · {lastSyncDisplay}</span>
          </div>
        </div>
        {error && (
          <div className="mt-4 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            Failed to load current configuration: {error}
          </div>
        )}
        {state?.config && (
          <dl className="mt-6 grid gap-4 sm:grid-cols-3">
            <div className="rounded-xl bg-slate-50 p-4">
              <dt className="text-xs uppercase tracking-wide text-slate-500">Provider</dt>
              <dd className="mt-1 text-sm font-semibold text-slate-900">{state.config.displayName}</dd>
              <p className="mt-1 text-xs text-slate-500">Protocol · {state.config.connection.protocol.toUpperCase()}</p>
            </div>
            <div className="rounded-xl bg-slate-50 p-4">
              <dt className="text-xs uppercase tracking-wide text-slate-500">Sync window</dt>
              <dd className="mt-1 text-sm font-semibold text-slate-900">{state.config.syncWindowHours} hours</dd>
              <p className="mt-1 text-xs text-slate-500">Labels · {state.config.labelFilters.join(', ') || 'All mail'}</p>
            </div>
            <div className="rounded-xl bg-slate-50 p-4">
              <dt className="text-xs uppercase tracking-wide text-slate-500">Messages fetched</dt>
              <dd className="mt-1 text-sm font-semibold text-slate-900">{state.messagesFetched}</dd>
              <p className="mt-1 text-xs text-slate-500">Updates when running the fetch endpoint</p>
            </div>
          </dl>
        )}
      </section>

      <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="flex items-start justify-between gap-6">
          <div>
            <h3 className="text-lg font-semibold text-slate-900">Connection settings</h3>
            <p className="mt-2 text-sm text-slate-600">
              Map the integration to your provider and specify the folders or labels that should flow into automations.
            </p>
          </div>
        </div>
        <form className="mt-6 space-y-5" onSubmit={handleConfigSubmit}>
          <div>
            <label className="text-sm font-medium text-slate-700" htmlFor="email-provider">
              Email provider
            </label>
            <select
              id="email-provider"
              className="mt-2 w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100"
              value={form.provider}
              onChange={(event) => handleProviderChange(event.target.value as EmailProviderType)}
            >
              {providerOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            <p className="mt-1 text-xs text-slate-500">
              {providerOptions.find((item) => item.value === form.provider)?.description}
            </p>
          </div>
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label className="text-sm font-medium text-slate-700" htmlFor="display-name">
                Display name
              </label>
              <input
                id="display-name"
                className="mt-2 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100"
                value={form.displayName}
                onChange={(event) => setForm({ ...form, displayName: event.target.value })}
                placeholder="Automation inbox"
                required
              />
            </div>
            <div>
              <label className="text-sm font-medium text-slate-700" htmlFor="sync-window">
                Sync window (hours)
              </label>
              <input
                id="sync-window"
                type="number"
                min={0}
                className="mt-2 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100"
                value={form.syncWindowHours}
                onChange={(event) =>
                  setForm({ ...form, syncWindowHours: Math.max(0, Number(event.target.value) || 0) })
                }
              />
            </div>
          </div>
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label className="text-sm font-medium text-slate-700" htmlFor="connection-protocol">
                Connection protocol
              </label>
              <select
                id="connection-protocol"
                className="mt-2 w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100"
                value={form.connection.protocol}
                onChange={(event) => {
                  const nextProtocol = event.target.value as 'api' | 'imap'
                  setForm((prev) => {
                    if (nextProtocol === 'imap') {
                      return {
                        ...prev,
                        connection: {
                          protocol: 'imap',
                          host: prev.connection.host ?? 'imap.example.com',
                          port: prev.connection.port ?? 993,
                          useTls: prev.connection.useTls ?? true,
                        },
                      }
                    }

                    return {
                      ...prev,
                      connection: {
                        protocol: 'api',
                        apiBaseUrl: prev.connection.apiBaseUrl ?? 'https://gmail.googleapis.com',
                      },
                    }
                  })
                }}
              >
                <option value="api">Provider API (OAuth)</option>
                <option value="imap">IMAP</option>
              </select>
            </div>
            {form.connection.protocol === 'api' ? (
              <div>
                <label className="text-sm font-medium text-slate-700" htmlFor="api-base-url">
                  API base URL
                </label>
                <input
                  id="api-base-url"
                  className="mt-2 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100"
                  value={form.connection.apiBaseUrl ?? ''}
                  onChange={(event) =>
                    setForm({ ...form, connection: { ...form.connection, apiBaseUrl: event.target.value } })
                  }
                  placeholder="https://gmail.googleapis.com"
                />
              </div>
            ) : (
              <div className="grid gap-4 sm:grid-cols-3">
                <div className="sm:col-span-2">
                  <label className="text-sm font-medium text-slate-700" htmlFor="imap-host">
                    IMAP host
                  </label>
                  <input
                    id="imap-host"
                    className="mt-2 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100"
                    value={form.connection.host ?? ''}
                    onChange={(event) =>
                      setForm({ ...form, connection: { ...form.connection, host: event.target.value } })
                    }
                    placeholder="imap.example.com"
                    required
                  />
                </div>
                <div>
                  <label className="text-sm font-medium text-slate-700" htmlFor="imap-port">
                    Port
                  </label>
                  <input
                    id="imap-port"
                    type="number"
                    min={1}
                    className="mt-2 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100"
                    value={form.connection.port ?? 993}
                    onChange={(event) =>
                      setForm({ ...form, connection: { ...form.connection, port: Number(event.target.value) || 993 } })
                    }
                  />
                </div>
                <div className="sm:col-span-3 flex items-center gap-2">
                  <input
                    id="imap-tls"
                    type="checkbox"
                    className="h-4 w-4 rounded border-slate-300 text-primary-600 focus:ring-primary-500"
                    checked={form.connection.useTls ?? true}
                    onChange={(event) =>
                      setForm({ ...form, connection: { ...form.connection, useTls: event.target.checked } })
                    }
                  />
                  <label className="text-sm text-slate-700" htmlFor="imap-tls">
                    Require TLS for mailbox connection
                  </label>
                </div>
              </div>
            )}
          </div>
          <div>
            <label className="text-sm font-medium text-slate-700" htmlFor="label-filters">
              Label or folder filters
            </label>
            <textarea
              id="label-filters"
              rows={3}
              className="mt-2 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100"
              value={labelInput}
              onChange={(event) => setLabelInput(event.target.value)}
              placeholder={'INBOX\nUrgent'}
            />
            <p className="mt-1 text-xs text-slate-500">One label per line. Use commas or line breaks.</p>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <button
              type="submit"
              className="inline-flex items-center justify-center rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-400"
            >
              Save connection
            </button>
            {configStatus && <span className="text-sm text-emerald-600">{configStatus}</span>}
            {configError && <span className="text-sm text-red-600">{configError}</span>}
          </div>
        </form>
      </section>

      <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <h3 className="text-lg font-semibold text-slate-900">Authentication</h3>
        <p className="mt-2 text-sm text-slate-600">
          Complete the OAuth consent flow or store an app password for systems that require basic authentication.
        </p>
        <form className="mt-6 space-y-4" onSubmit={handleAuthSubmit}>
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label className="text-sm font-medium text-slate-700" htmlFor="auth-method">
                Method
              </label>
              <select
                id="auth-method"
                className="mt-2 w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100"
                value={authMethod}
                onChange={(event) => setAuthMethod(event.target.value as EmailAuthMethod)}
              >
                <option value="oauth">OAuth token exchange</option>
                <option value="appPassword">App password / service credential</option>
              </select>
            </div>
            <div>
              <label className="text-sm font-medium text-slate-700" htmlFor="auth-username">
                Mailbox user
              </label>
              <input
                id="auth-username"
                className="mt-2 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100"
                value={authUsername}
                onChange={(event) => setAuthUsername(event.target.value)}
                placeholder="ops-team@example.com"
                required
              />
            </div>
          </div>
          <div>
            <label className="text-sm font-medium text-slate-700" htmlFor="auth-secret">
              {authMethod === 'oauth' ? 'OAuth access token' : 'App password'}
            </label>
            <input
              id="auth-secret"
              className="mt-2 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100"
              value={authSecret}
              onChange={(event) => setAuthSecret(event.target.value)}
              placeholder={authMethod === 'oauth' ? 'ya29.a0AWY...' : 'xxxx-xxxx-xxxx'}
              required
            />
            <p className="mt-1 text-xs text-slate-500">
              Secrets are hashed and only used to simulate validation in this prototype.
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <button
              type="submit"
              className="inline-flex items-center justify-center rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-slate-700 focus:outline-none focus:ring-2 focus:ring-slate-400"
            >
              Save authentication
            </button>
            {authStatus && <span className="text-sm text-emerald-600">{authStatus}</span>}
            {authError && <span className="text-sm text-red-600">{authError}</span>}
          </div>
        </form>
      </section>

      <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h3 className="text-lg font-semibold text-slate-900">Sample inbox preview</h3>
            <p className="mt-1 text-sm text-slate-600">
              Fetch a deterministic set of messages that mirror the automation-ready payload the backend returns once real
              provider connections are wired.
            </p>
          </div>
          <button
            type="button"
            onClick={handleFetchMessages}
            className="inline-flex items-center justify-center rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-400 disabled:cursor-not-allowed disabled:bg-primary-300"
            disabled={fetchingMessages}
          >
            {fetchingMessages ? 'Loading…' : 'Fetch latest messages'}
          </button>
        </div>
        {messageError && (
          <div className="mt-4 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">{messageError}</div>
        )}
        {syncedAt && (
          <p className="mt-4 text-xs uppercase tracking-wide text-slate-500">Last sync {formatTimestamp(syncedAt)}</p>
        )}
        <div className="mt-4 grid gap-4 md:grid-cols-2">
          {messages.map((message) => (
            <article key={message.id} className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
              <header className="flex items-start justify-between gap-3">
                <div>
                  <h4 className="text-sm font-semibold text-slate-900">{message.subject}</h4>
                  <p className="text-xs text-slate-500">From {message.sender}</p>
                </div>
                <span
                  className={`rounded-full px-3 py-1 text-xs font-medium ${
                    message.importance === 'high'
                      ? 'bg-rose-100 text-rose-700'
                      : 'bg-slate-100 text-slate-600'
                  }`}
                >
                  {message.importance === 'high' ? 'High' : 'Normal'}
                </span>
              </header>
              <p className="mt-3 text-sm text-slate-600">{message.snippet}</p>
              <footer className="mt-4 flex flex-wrap items-center gap-2 text-xs text-slate-500">
                <span className="rounded bg-slate-100 px-2 py-1">
                  {new Date(message.receivedAt).toLocaleString()}
                </span>
                {message.labels.map((label) => (
                  <span key={label} className="rounded bg-primary-50 px-2 py-1 text-primary-700">
                    {label}
                  </span>
                ))}
              </footer>
            </article>
          ))}
          {!messages.length && !fetchingMessages && (
            <div className="rounded-xl border border-dashed border-slate-300 p-6 text-sm text-slate-500">
              Use the fetch action once authentication succeeds to preview the deterministic message payload delivered by the
              backend stub.
            </div>
          )}
        </div>
      </section>
    </div>
  )
}
