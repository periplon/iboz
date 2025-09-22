import { NavLink, Outlet, Route, Routes } from 'react-router-dom'
import { Dashboard } from './pages/Dashboard'
import { FocusMode } from './pages/FocusMode'
import { Automations } from './pages/Automations'
import { Analytics } from './pages/Analytics'
import { Admin } from './pages/Admin'

const navigation = [
  { name: 'Dashboard', to: '/' },
  { name: 'Focus Mode', to: '/focus' },
  { name: 'Automations', to: '/automations' },
  { name: 'Analytics', to: '/analytics' },
  { name: 'Admin', to: '/admin' },
]

export default function App() {
  return (
    <div className="min-h-screen bg-slate-100">
      <div className="flex min-h-screen">
        <aside className="hidden w-64 flex-shrink-0 border-r border-slate-200 bg-white px-6 py-8 lg:flex lg:flex-col">
          <div className="text-2xl font-semibold text-slate-900">
            Inbox <span className="text-primary-600">Zero</span>
          </div>
          <p className="mt-2 text-sm text-slate-500">Automation command center</p>
          <nav className="mt-10 space-y-1">
            {navigation.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.to === '/'}
                className={({ isActive }) =>
                  `flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition ${
                    isActive ? 'bg-primary-50 text-primary-700' : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900'
                  }`
                }
              >
                {item.name}
              </NavLink>
            ))}
          </nav>
          <div className="mt-auto rounded-xl bg-slate-50 p-4 text-xs text-slate-500">
            Confidence telemetry, audit trails, and multi-tenant controls are wired into backend roadmaps. This shell highlights
            where to connect future iterations.
          </div>
        </aside>
        <main className="flex-1">
          <header className="border-b border-slate-200 bg-white/80 backdrop-blur">
            <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
              <div>
                <h1 className="text-xl font-semibold text-slate-900">Inbox Zero Automation Platform</h1>
                <p className="text-sm text-slate-500">Transparent automation, focus-first workflows, and compliance by design.</p>
              </div>
              <div className="hidden items-center gap-3 text-sm text-slate-500 md:flex">
                <span className="rounded-full bg-emerald-100 px-3 py-1 font-medium text-emerald-700">Pilot tenants · 3</span>
                <span className="rounded-full bg-slate-100 px-3 py-1 font-medium">Version 0.1.0</span>
              </div>
            </div>
          </header>
          <div className="mx-auto max-w-6xl px-6 py-8">
            <Routes>
              <Route element={<Outlet />}>
                <Route index element={<Dashboard />} />
                <Route path="focus" element={<FocusMode />} />
                <Route path="automations" element={<Automations />} />
                <Route path="analytics" element={<Analytics />} />
                <Route path="admin" element={<Admin />} />
              </Route>
            </Routes>
          </div>
        </main>
      </div>
    </div>
  )
}
