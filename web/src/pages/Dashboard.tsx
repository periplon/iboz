import { MetricCard } from '../components/MetricCard'
import { QueueCard } from '../components/QueueCard'
import { RecommendationCard } from '../components/RecommendationCard'
import { useDashboard } from '../api/hooks'

export function Dashboard() {
  const { data, loading, error } = useDashboard()

  if (loading) {
    return <div className="text-sm text-slate-500">Loading dashboard…</div>
  }

  if (error) {
    return <div className="rounded-lg border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">{error}</div>
  }

  if (!data) {
    return null
  }

  const { summary, queues, recommendations } = data

  return (
    <div className="space-y-8">
      <section>
        <h2 className="text-xl font-semibold text-slate-900">Today&apos;s momentum</h2>
        <dl className="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <MetricCard
            label="Inbox"
            value={`${summary.currentInbox} / ${summary.inboxZeroTarget}`}
            helper="Messages remaining to hit Inbox Zero"
            accent="primary"
          />
          <MetricCard
            label="Automation rate"
            value={`${Math.round(summary.automationRate * 100)}%`}
            helper="Emails triaged without manual effort"
            accent="emerald"
          />
          <MetricCard
            label="Time saved"
            value={`${summary.timeSavedMinutes} min`}
            helper="Estimated reclaimed focus time today"
            accent="amber"
          />
          <MetricCard
            label="Focus potential"
            value={`${summary.currentInbox - summary.inboxZeroTarget > 0 ? summary.currentInbox - summary.inboxZeroTarget : 0}`}
            helper="Extra emails to schedule into focus mode"
          />
        </dl>
      </section>

      <section>
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold text-slate-900">Action queues</h2>
          <span className="text-xs font-semibold uppercase tracking-wide text-slate-500">Live classification snapshot</span>
        </div>
        <div className="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {queues.map((queue) => (
            <QueueCard key={queue.id} {...queue} />
          ))}
        </div>
      </section>

      <section>
        <h2 className="text-xl font-semibold text-slate-900">Automation recommendations</h2>
        <div className="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {recommendations.map((item) => (
            <RecommendationCard key={item.id} {...item} />
          ))}
        </div>
      </section>
    </div>
  )
}
