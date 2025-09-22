export function Admin() {
  return (
    <div className="space-y-6">
      <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 className="text-xl font-semibold text-slate-900">Governance controls</h2>
        <p className="mt-2 text-sm text-slate-600">
          Configure tenant-wide LLM vendor access, retention policies, and audit exports. This scaffold provides the surface
          area for security and compliance workflows described in the BRD.
        </p>
        <div className="mt-6 grid gap-4 md:grid-cols-2">
          <div className="rounded-xl bg-slate-50 p-4 text-sm">
            <p className="font-semibold text-slate-700">LLM vendors</p>
            <ul className="mt-2 space-y-1 text-slate-600">
              <li>OpenAI · enabled · confidence threshold 0.7</li>
              <li>Anthropic · pending approval</li>
              <li>Google Vertex · disabled</li>
            </ul>
          </div>
          <div className="rounded-xl bg-slate-50 p-4 text-sm">
            <p className="font-semibold text-slate-700">Data residency</p>
            <p className="mt-2">Region: EU West · Retention: 30 days</p>
            <p className="mt-1 text-slate-500">Update values once legal reviews are complete.</p>
          </div>
        </div>
      </section>
      <section className="rounded-2xl border border-dashed border-slate-300 p-6 text-sm text-slate-500">
        Wire role-based access control and audit log exports here. Tie actions to the backend governance endpoints during Phase 3
        of the roadmap.
      </section>
    </div>
  )
}
