interface AutomationCardProps {
  name: string
  description: string
  trigger: string
  actions: string[]
  requiresApproval: boolean
  onTest: () => void
  testing?: boolean
}

export function AutomationCard({ name, description, trigger, actions, requiresApproval, onTest, testing }: AutomationCardProps) {
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
        <div className="mt-4 text-sm">
          <p className="font-medium text-slate-500">Trigger</p>
          <p className="mt-1 rounded-lg bg-slate-100 px-3 py-2 font-mono text-xs text-slate-700">{trigger}</p>
        </div>
        <div className="mt-4 text-sm">
          <p className="font-medium text-slate-500">Actions</p>
          <ul className="mt-2 list-disc space-y-1 pl-5 text-slate-600">
            {actions.map((action) => (
              <li key={action}>{action}</li>
            ))}
          </ul>
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
