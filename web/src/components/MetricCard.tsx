import type { ReactNode } from 'react'

interface MetricCardProps {
  label: string
  value: ReactNode
  helper?: ReactNode
  accent?: 'primary' | 'emerald' | 'amber'
}

const accentClasses: Record<NonNullable<MetricCardProps['accent']>, string> = {
  primary: 'from-primary-100 to-primary-50 text-primary-800',
  emerald: 'from-emerald-100 to-emerald-50 text-emerald-800',
  amber: 'from-amber-100 to-amber-50 text-amber-800',
}

export function MetricCard({ label, value, helper, accent = 'primary' }: MetricCardProps) {
  return (
    <div className={`rounded-xl bg-gradient-to-br p-5 shadow-sm ${accentClasses[accent]}`}>
      <dt className="text-sm font-medium uppercase tracking-wide opacity-70">{label}</dt>
      <dd className="mt-2 text-3xl font-semibold">{value}</dd>
      {helper ? <p className="mt-2 text-sm opacity-80">{helper}</p> : null}
    </div>
  )
}
