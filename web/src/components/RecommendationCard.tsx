interface RecommendationCardProps {
  title: string
  description: string
  confidence: number
}

export function RecommendationCard({ title, description, confidence }: RecommendationCardProps) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
      <h3 className="text-base font-semibold text-slate-900">{title}</h3>
      <p className="mt-2 text-sm text-slate-600">{description}</p>
      <div className="mt-4 flex items-center justify-between text-sm font-medium text-slate-500">
        <span>Confidence</span>
        <span className="text-primary-600">{Math.round(confidence * 100)}%</span>
      </div>
      <div className="mt-2 h-2 rounded-full bg-slate-100">
        <div className="h-full rounded-full bg-primary-500" style={{ width: `${confidence * 100}%` }} />
      </div>
    </div>
  )
}
