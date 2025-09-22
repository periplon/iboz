interface TestResultPanelProps {
  status?: string
  summary?: string
  confidence?: number
  requiresApproval?: boolean
  parameters?: Record<string, unknown>
  error?: string
}

export function TestResultPanel({ status, summary, confidence, requiresApproval, parameters, error }: TestResultPanelProps) {
  if (error) {
    return (
      <div className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">
        <p className="font-semibold">Simulation failed</p>
        <p className="mt-1">{error}</p>
      </div>
    )
  }

  if (!status) {
    return (
      <div className="rounded-xl border border-dashed border-slate-300 p-4 text-sm text-slate-500">
        Select an automation to view the simulation summary.
      </div>
    )
  }

  return (
    <div className="space-y-4 rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
      <div>
        <p className="text-sm font-semibold uppercase tracking-wide text-primary-600">{status}</p>
        <p className="mt-1 text-base text-slate-700">{summary}</p>
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Confidence</p>
          <p className="mt-1 text-xl font-semibold text-slate-900">{confidence !== undefined ? `${Math.round(confidence * 100)}%` : '—'}</p>
        </div>
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Approval</p>
          <p className="mt-1 text-xl font-semibold text-slate-900">{requiresApproval ? 'Required' : 'Auto-execute'}</p>
        </div>
      </div>
      {parameters && Object.keys(parameters).length > 0 ? (
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Parameters</p>
          <pre className="mt-2 rounded-lg bg-slate-900/95 p-4 text-xs text-slate-100">
            {JSON.stringify(parameters, null, 2)}
          </pre>
        </div>
      ) : null}
    </div>
  )
}
