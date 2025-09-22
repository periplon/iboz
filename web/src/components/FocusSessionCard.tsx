interface FocusSessionCardProps {
  label: string
  description: string
  estimated: number
  emails: number
  llmSupport: boolean
  active?: boolean
  remainingSeconds?: number | null
  onStart?: () => void
}

export function FocusSessionCard({
  label,
  description,
  estimated,
  emails,
  llmSupport,
  active,
  remainingSeconds,
  onStart,
}: FocusSessionCardProps) {
  const minutesRemaining = remainingSeconds !== undefined && remainingSeconds !== null ? Math.ceil(remainingSeconds / 60) : null

  return (
    <div className={`rounded-xl border bg-white p-5 shadow-sm transition ${active ? 'border-primary-300 ring-2 ring-primary-200' : 'border-slate-200'}`}>
      <div className="flex items-start justify-between">
        <div>
          <h3 className="text-lg font-semibold text-slate-900">{label}</h3>
          <p className="mt-1 text-sm text-slate-600">{description}</p>
        </div>
        <span className={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${llmSupport ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-100 text-slate-500'}`}>
          {llmSupport ? 'LLM Assisted' : 'Rules First'}
        </span>
      </div>
      <dl className="mt-4 grid grid-cols-2 gap-4 text-sm">
        <div>
          <dt className="font-medium text-slate-500">Estimated Focus</dt>
          <dd className="mt-1 text-base font-semibold text-slate-900">{estimated} min</dd>
        </div>
        <div>
          <dt className="font-medium text-slate-500">Emails Batched</dt>
          <dd className="mt-1 text-base font-semibold text-slate-900">{emails}</dd>
        </div>
      </dl>
      {active ? (
        <div className="mt-4 rounded-lg bg-primary-50 p-3 text-sm text-primary-700">
          <p className="font-semibold">In progress</p>
          <p className="mt-1">
            {minutesRemaining !== null && minutesRemaining > 0 ? `${minutesRemaining} minute${minutesRemaining === 1 ? '' : 's'} remaining` : 'Session complete — log actions and celebrate!'}
          </p>
        </div>
      ) : null}
      {onStart ? (
        <button
          type="button"
          onClick={onStart}
          className="mt-5 inline-flex items-center justify-center rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow transition hover:bg-primary-500"
        >
          {active ? 'Resume session' : 'Start session'}
        </button>
      ) : null}
    </div>
  )
}
