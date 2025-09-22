import { useFocusPlan } from '../api/hooks'
import { FocusSessionCard } from '../components/FocusSessionCard'

export function FocusMode() {
  const { data, loading, error } = useFocusPlan()

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

      <section>
        <h2 className="text-xl font-semibold text-slate-900">Upcoming sessions</h2>
        <div className="mt-4 grid gap-4 md:grid-cols-2">
          {data.sessions.map((session) => (
            <FocusSessionCard key={session.id} {...session} />
          ))}
        </div>
      </section>
    </div>
  )
}
