interface FocusSessionCardProps {
  label: string
  description: string
  estimated: number
  emails: number
  llmSupport: boolean
}

export function FocusSessionCard({ label, description, estimated, emails, llmSupport }: FocusSessionCardProps) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
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
    </div>
  )
}
