import { useState } from 'react'
import { useAutomations, useAutomationTestRunner } from '../api/hooks'
import { AutomationCard } from '../components/AutomationCard'
import { TestResultPanel } from '../components/TestResultPanel'
import { MetricCard } from '../components/MetricCard'

export function Automations() {
  const { data, loading, error } = useAutomations()
  const { data: testResult, error: testError, loading: testing, runTest } = useAutomationTestRunner()
  const [selected, setSelected] = useState<string | null>(null)

  const handleTest = (templateId: string) => {
    setSelected(templateId)
    runTest(templateId, { dryRun: true })
  }

  if (loading) {
    return <div className="text-sm text-slate-500">Loading automations…</div>
  }

  if (error) {
    return <div className="rounded-lg border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">{error}</div>
  }

  if (!data) {
    return null
  }

  return (
    <div className="space-y-8">
      <section>
        <h2 className="text-xl font-semibold text-slate-900">Automation operations</h2>
        <p className="mt-2 text-sm text-slate-600">
          Monitor throughput, coverage, and reclaimed time as you expand deterministic rules with LLM assists. These metrics
          connect directly to the BRD&apos;s success criteria for automation adoption.
        </p>
        <dl className="mt-4 grid gap-4 md:grid-cols-3">
          <MetricCard label="Active automations" value={data.overview.active} helper="Enabled across this tenant" />
          <MetricCard
            label="Coverage"
            value={`${Math.round(data.overview.automationCoverage * 100)}%`}
            helper="Emails triaged without manual input"
            accent="emerald"
          />
          <MetricCard
            label="Avg. time saved"
            value={`${data.overview.avgTimeSaved} min`}
            helper="Per automation run"
            accent="amber"
          />
        </dl>
      </section>

      <div className="grid gap-8 lg:grid-cols-[2fr,1fr]">
        <div className="space-y-4">
          {data.templates.map((template) => (
            <AutomationCard
              key={template.id}
              {...template}
              testing={testing && selected === template.id}
              onTest={() => handleTest(template.id)}
            />
          ))}
        </div>
        <div className="space-y-4">
          <h2 className="text-xl font-semibold text-slate-900">Simulation output</h2>
          <TestResultPanel
            status={testResult?.status}
            summary={testResult?.summary}
            confidence={testResult?.review.confidence}
            requiresApproval={testResult?.review.requiresApproval}
            parameters={testResult?.parameters}
            error={testError}
          />
        </div>
      </div>
    </div>
  )
}
