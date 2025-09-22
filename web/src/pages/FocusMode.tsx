import { useEffect, useMemo, useState } from 'react'
import { useFocusPlan } from '../api/hooks'
import { FocusSessionCard } from '../components/FocusSessionCard'

export function FocusMode() {
  const { data, loading, error } = useFocusPlan()
  const [preferences, setPreferences] = useState({
    notificationsMuted: false,
    batchingEnabled: true,
    autoSummaries: true,
  })
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null)
  const [remainingSeconds, setRemainingSeconds] = useState<number | null>(null)

  useEffect(() => {
    if (data?.controls) {
      setPreferences(data.controls)
    }
  }, [data])

  useEffect(() => {
    if (!activeSessionId || remainingSeconds === null || remainingSeconds <= 0) {
      return
    }

    const interval = window.setInterval(() => {
      setRemainingSeconds((prev) => {
        if (prev === null || prev <= 0) {
          return 0
        }
        return prev - 1
      })
    }, 1000)

    return () => window.clearInterval(interval)
  }, [activeSessionId, remainingSeconds])

  const activeSession = useMemo(
    () => data?.sessions.find((session) => session.id === activeSessionId) ?? null,
    [activeSessionId, data?.sessions],
  )

  const totalSeconds = activeSession ? activeSession.estimated * 60 : null
  const progress = activeSession && totalSeconds ? Math.min(100, Math.round(((totalSeconds - (remainingSeconds ?? 0)) / totalSeconds) * 100)) : 0

  const countdown = remainingSeconds !== null && remainingSeconds >= 0
    ? `${String(Math.floor(remainingSeconds / 60)).padStart(2, '0')}:${String(Math.abs(remainingSeconds % 60)).padStart(2, '0')}`
    : '--:--'

  const togglePreference = (key: keyof typeof preferences) => {
    setPreferences((prev) => ({
      ...prev,
      [key]: !prev[key],
    }))
  }

  const handleStart = (sessionId: string, estimated: number) => {
    setActiveSessionId(sessionId)
    setRemainingSeconds(estimated * 60)
  }

  if (loading) {
    return <div className="text-sm text-slate-500">Preparing focus mode…</div>
  }

  if (error) {
    return <div className="rounded-lg border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">{error}</div>
  }

  if (!data) {
    return null
  }

  return (
    <div className="space-y-8">
      <section className="rounded-2xl bg-gradient-to-r from-primary-500 to-primary-600 p-6 text-white shadow-lg">
        <p className="text-sm uppercase tracking-wide text-primary-100">Focus streak</p>
        <div className="mt-2 flex flex-wrap items-center gap-6">
          <div>
            <p className="text-4xl font-semibold">{data.metrics.streak} days</p>
            <p className="text-sm opacity-80">Goal: {data.metrics.goal} days</p>
          </div>
          <div className="text-sm opacity-90">
            <p>{data.metrics.clearedToday} emails cleared today</p>
            <p className="mt-1">Scheduled sessions keep you on track for Inbox Zero.</p>
          </div>
        </div>
      </section>

      <section className="grid gap-4 lg:grid-cols-[2fr,1fr]">
        <div className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-slate-900">Distraction controls</h2>
          <p className="mt-2 text-sm text-slate-600">
            Tailor the guardrails for this focus sprint. These toggles mirror the CODEx principle of keeping automation
            transparent while preserving momentum.
          </p>
          <ul className="mt-6 space-y-4">
            {[{
              key: 'notificationsMuted' as const,
              label: 'Silence non-urgent notifications',
              helper: 'Pause Slack and Teams alerts for non-escalation senders.',
            },
            {
              key: 'batchingEnabled' as const,
              label: 'Batch similar intents',
              helper: 'Group emails by category to reduce context switching.',
            },
            {
              key: 'autoSummaries' as const,
              label: 'Show AI summaries',
              helper: 'Surface key points and attachments before entering each email.',
            }].map((item) => (
              <li key={item.key} className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-sm font-medium text-slate-700">{item.label}</p>
                  <p className="text-xs text-slate-500">{item.helper}</p>
                </div>
                <button
                  type="button"
                  onClick={() => togglePreference(item.key)}
                  className={`relative inline-flex h-6 w-11 flex-shrink-0 items-center rounded-full transition ${preferences[item.key] ? 'bg-primary-500' : 'bg-slate-300'}`}
                >
                  <span
                    className={`inline-block h-5 w-5 transform rounded-full bg-white shadow transition ${preferences[item.key] ? 'translate-x-5' : 'translate-x-1'}`}
                  />
                </button>
              </li>
            ))}
          </ul>
        </div>
        <div className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-slate-900">Session tracker</h2>
          {activeSession ? (
            <div className="mt-4 space-y-4">
              <div>
                <p className="text-sm font-medium text-slate-700">{activeSession.label}</p>
                <p className="text-xs text-slate-500">{activeSession.emails} emails · {activeSession.estimated} minute block</p>
              </div>
              <div>
                <p className="text-3xl font-semibold text-primary-600">{countdown}</p>
                <p className="text-xs uppercase tracking-wide text-slate-500">Countdown</p>
              </div>
              <div className="h-2 rounded-full bg-slate-100">
                <div className="h-full rounded-full bg-primary-500 transition-all" style={{ width: `${progress}%` }} />
              </div>
              <p className="text-xs text-slate-500">
                {progress >= 100
                  ? 'Great work! Log automation overrides and share wins with your team.'
                  : 'Stay in flow—automations are handling low-priority noise while you clear critical items.'}
              </p>
            </div>
          ) : (
            <p className="mt-4 text-sm text-slate-600">
              Select a session to launch a timer, mute distractions, and apply batching rules. Your controls persist between
              focus streaks.
            </p>
          )}
        </div>
      </section>

      <section>
        <h2 className="text-xl font-semibold text-slate-900">Upcoming sessions</h2>
        <div className="mt-4 grid gap-4 md:grid-cols-2">
          {data.sessions.map((session) => (
            <FocusSessionCard
              key={session.id}
              {...session}
              active={session.id === activeSessionId}
              remainingSeconds={session.id === activeSessionId ? remainingSeconds : null}
              onStart={() => handleStart(session.id, session.estimated)}
            />
          ))}
        </div>
      </section>
    </div>
  )
}
