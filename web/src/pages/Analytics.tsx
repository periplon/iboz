export function Analytics() {
  return (
    <div className="space-y-6">
      <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 className="text-xl font-semibold text-slate-900">Operational KPIs</h2>
        <p className="mt-2 text-sm text-slate-600">
          Analytics integrates automation throughput, focus session engagement, and override frequency. Hook this view into the
          observability layer to align with the experiment roadmap outlined in the EDD.
        </p>
        <div className="mt-6 grid gap-4 md:grid-cols-3">
          <div className="rounded-xl bg-slate-50 p-4 text-sm">
            <p className="font-semibold text-slate-700">Inbox Zero attainment</p>
            <p className="mt-2 text-3xl font-semibold text-primary-600">85%</p>
            <p className="mt-1 text-slate-500">Target per pilot tenant</p>
          </div>
          <div className="rounded-xl bg-slate-50 p-4 text-sm">
            <p className="font-semibold text-slate-700">Automation adoption</p>
            <p className="mt-2 text-3xl font-semibold text-primary-600">68%</p>
            <p className="mt-1 text-slate-500">LLM-augmented vs deterministic</p>
          </div>
          <div className="rounded-xl bg-slate-50 p-4 text-sm">
            <p className="font-semibold text-slate-700">Average time saved</p>
            <p className="mt-2 text-3xl font-semibold text-primary-600">2.1h</p>
            <p className="mt-1 text-slate-500">Per user this week</p>
          </div>
        </div>
      </section>
      <section className="rounded-2xl border border-dashed border-slate-300 p-6 text-sm text-slate-500">
        Instrument dashboards here once telemetry events (automation_triggered, focus_session_started, override_submitted) are
        streaming into your analytics pipeline.
      </section>
    </div>
  )
}
