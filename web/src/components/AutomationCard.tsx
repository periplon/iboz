interface AutomationCardProps {
  name: string
  description: string
  trigger: string
  conditions: string[]
  actions: string[]
  requiresApproval: boolean
  onTest: () => void
  testing?: boolean
  owner: string
  lastRun: string
}

export function AutomationCard({
  name,
  description,
  trigger,
  conditions,
  actions,
  requiresApproval,
  onTest,
  testing,
  owner,
  lastRun,
}: AutomationCardProps) {
  const lastRunDisplay = new Date(lastRun).toLocaleString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    month: 'short',
    day: 'numeric',
  })

  return (
    <div className="flex flex-col justify-between rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
      <div>
        <div className="flex items-start justify-between gap-4">
          <div>
            <h3 className="text-lg font-semibold text-slate-900">{name}</h3>
            <p className="mt-2 text-sm text-slate-600">{description}</p>
          </div>
          <span className={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${requiresApproval ? 'bg-amber-100 text-amber-700' : 'bg-emerald-100 text-emerald-700'}`}>
            {requiresApproval ? 'Approval Required' : 'Auto Execute'}
          </span>
        </div>
        <div className="mt-5 grid gap-4 md:grid-cols-3">
          <div className="text-sm">
            <p className="font-medium text-slate-500">Trigger</p>
            <p className="mt-1 rounded-lg bg-slate-100 px-3 py-2 font-mono text-xs text-slate-700">{trigger}</p>
          </div>
          <div className="text-sm">
            <p className="font-medium text-slate-500">Conditions</p>
            <ul className="mt-2 space-y-1 text-slate-600">
              {conditions.map((condition) => (
                <li key={condition} className="rounded-md bg-slate-50 px-3 py-2 font-mono text-xs">
                  {condition}
                </li>
              ))}
            </ul>
          </div>
          <div className="text-sm">
            <p className="font-medium text-slate-500">Actions</p>
            <ul className="mt-2 list-disc space-y-1 pl-5 text-slate-600">
              {actions.map((action) => (
                <li key={action}>{action}</li>
              ))}
            </ul>
          </div>
        </div>
        <div className="mt-5 flex flex-col gap-2 text-xs text-slate-500 sm:flex-row sm:items-center sm:justify-between">
          <span>Managed by {owner}</span>
          <span>Last simulated {lastRunDisplay}</span>
        </div>
      </div>
      <button
        onClick={onTest}
        disabled={testing}
        className="mt-5 inline-flex items-center justify-center rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow transition hover:bg-primary-500 disabled:cursor-not-allowed disabled:bg-slate-400"
      >
        {testing ? 'Simulating…' : 'Simulate Run'}
      </button>
    </div>
  )
}
