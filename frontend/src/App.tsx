import { useState } from 'react'
import { DevicesPage } from './pages/DevicesPage'
import { DeviceGroupsPage } from './pages/DeviceGroupsPage'
import { LogsPage } from './pages/LogsPage'
import { ProfilesPage } from './pages/ProfilesPage'
import { SettingsPage } from './pages/SettingsPage'

type Page = 'devices' | 'groups' | 'profiles' | 'logs' | 'settings'

const nav: { id: Page; label: string }[] = [
  { id: 'devices',  label: 'Devices'  },
  { id: 'groups',   label: 'Groups'   },
  { id: 'profiles', label: 'Profiles' },
  { id: 'logs',     label: 'Logs'     },
  { id: 'settings', label: 'Settings' },
]

export default function App() {
  const [page, setPage] = useState<Page>('devices')

  return (
    <div className="min-h-screen bg-gray-950 text-gray-100">
      <header className="border-b border-gray-800 bg-gray-900">
        <div className="mx-auto max-w-5xl flex items-center justify-between px-6 h-14">
          <div className="flex items-center gap-2">
            <span className="text-indigo-400 font-bold text-lg">⬡</span>
            <span className="font-semibold text-white">DNS Manager</span>
          </div>
          <nav className="flex gap-1">
            {nav.map(n => (
              <button
                key={n.id}
                onClick={() => setPage(n.id)}
                className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                  page === n.id
                    ? 'bg-indigo-600 text-white'
                    : 'text-gray-400 hover:text-white hover:bg-gray-800'
                }`}
              >
                {n.label}
              </button>
            ))}
          </nav>
        </div>
      </header>

      <main className="mx-auto max-w-5xl px-6 py-8">
        {page === 'devices'  && <DevicesPage />}
        {page === 'groups'   && <DeviceGroupsPage />}
        {page === 'profiles' && <ProfilesPage />}
        {page === 'logs'     && <LogsPage />}
        {page === 'settings' && <SettingsPage />}
      </main>
    </div>
  )
}
