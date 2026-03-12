import { useEffect, useState } from 'react'
import { api } from '../api'
import type { Settings } from '../types'

export function SettingsPage() {
  const [settings, setSettings] = useState<Settings>({ listen: '', default_profile: '' })
  const [upstreams, setUpstreams] = useState<string[]>([])
  const [newUpstream, setNewUpstream] = useState('')
  const [saved, setSaved] = useState(false)
  const [error, setError] = useState('')

  const load = async () => {
    const [s, u] = await Promise.all([api.getSettings(), api.getUpstreams()])
    setSettings(s)
    setUpstreams(u)
  }

  useEffect(() => { load() }, [])

  const saveSettings = async () => {
    try {
      await api.putSettings(settings)
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to save')
    }
  }

  const addUpstream = async () => {
    const val = newUpstream.trim()
    if (!val || upstreams.includes(val)) return
    const next = [...upstreams, val]
    await api.putUpstreams(next)
    setUpstreams(next)
    setNewUpstream('')
  }

  const removeUpstream = async (addr: string) => {
    const next = upstreams.filter(u => u !== addr)
    await api.putUpstreams(next)
    setUpstreams(next)
  }

  const moveUpstream = async (index: number, dir: -1 | 1) => {
    const next = [...upstreams]
    const swap = index + dir
    if (swap < 0 || swap >= next.length) return
    ;[next[index], next[swap]] = [next[swap], next[index]]
    await api.putUpstreams(next)
    setUpstreams(next)
  }

  return (
    <div className="max-w-2xl space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-white">Settings</h1>
        <p className="text-sm text-gray-400 mt-0.5">Server configuration</p>
      </div>

      {/* Server settings */}
      <div className="card px-6 py-5 space-y-4">
        <h2 className="font-semibold text-white">Server</h2>
        <div>
          <label className="field-label">Listen Address</label>
          <input
            className="field"
            value={settings.listen}
            onChange={e => setSettings({ ...settings, listen: e.target.value })}
            placeholder=":53"
          />
          <p className="text-xs text-gray-500 mt-1">Restart required for changes to take effect</p>
        </div>
        <div>
          <label className="field-label">Default Profile</label>
          <input
            className="field"
            value={settings.default_profile}
            onChange={e => setSettings({ ...settings, default_profile: e.target.value })}
            placeholder="default"
          />
          <p className="text-xs text-gray-500 mt-1">Applied to devices not listed in the Devices table</p>
        </div>
        {error && <p className="text-sm text-red-400">{error}</p>}
        <div className="flex items-center gap-3">
          <button onClick={saveSettings} className="btn-primary">Save</button>
          {saved && <span className="text-sm text-green-400">Saved</span>}
        </div>
      </div>

      {/* Upstream DNS */}
      <div className="card px-6 py-5 space-y-4">
        <h2 className="font-semibold text-white">Upstream DNS Servers</h2>
        <p className="text-sm text-gray-400">Tried in order — first successful response wins</p>

        <ul className="space-y-2">
          {upstreams.map((u, i) => (
            <li key={u} className="flex items-center gap-2 rounded-lg bg-gray-800 px-3 py-2">
              <span className="text-xs text-gray-500 w-5 text-right">{i + 1}</span>
              <span className="font-mono text-sm text-gray-200 flex-1">{u}</span>
              <button
                onClick={() => moveUpstream(i, -1)}
                disabled={i === 0}
                className="text-gray-500 hover:text-white disabled:opacity-20 transition-colors px-1"
                title="Move up"
              >↑</button>
              <button
                onClick={() => moveUpstream(i, 1)}
                disabled={i === upstreams.length - 1}
                className="text-gray-500 hover:text-white disabled:opacity-20 transition-colors px-1"
                title="Move down"
              >↓</button>
              <button onClick={() => removeUpstream(u)} className="btn-danger text-xs">Remove</button>
            </li>
          ))}
        </ul>

        <div className="flex gap-2">
          <input
            className="field flex-1"
            placeholder="1.1.1.1:53"
            value={newUpstream}
            onChange={e => setNewUpstream(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') addUpstream() }}
          />
          <button onClick={addUpstream} className="btn-primary">Add</button>
        </div>
      </div>
    </div>
  )
}
