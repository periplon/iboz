interface QueueCardProps {
  label: string
  description: string
  count: number
  llmEnabled: boolean
}

export function QueueCard({ label, description, count, llmEnabled }: QueueCardProps) {
  return (
    <div className="flex flex-col justify-between rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition hover:-translate-y-1 hover:shadow-md">
      <div>
        <h3 className="text-lg font-semibold text-slate-900">{label}</h3>
        <p className="mt-1 text-sm text-slate-500">{description}</p>
      </div>
      <div className="mt-4 flex items-center justify-between">
        <span className="text-3xl font-bold text-slate-900">{count}</span>
        <span
          className={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${llmEnabled ? 'bg-primary-100 text-primary-700' : 'bg-slate-100 text-slate-500'}`}
        >
          {llmEnabled ? 'LLM Assisted' : 'Rules'}
        </span>
      </div>
    </div>
  )
}
