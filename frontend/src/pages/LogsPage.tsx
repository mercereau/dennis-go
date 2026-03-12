import { useEffect, useRef, useState } from 'react'
import { api } from '../api'
import type { LogEntry } from '../types'

const actionColor = {
  ALLOW: 'text-green-400',
  BLOCK: 'text-red-400',
  ERROR: 'text-yellow-400',
}

export function LogsPage() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [filter, setFilter] = useState('')
  const [autoRefresh, setAutoRefresh] = useState(false)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const load = () => api.listLogs(500).then(setLogs)

  useEffect(() => { load() }, [])

  useEffect(() => {
    if (autoRefresh) {
      intervalRef.current = setInterval(load, 2000)
    } else {
      if (intervalRef.current) clearInterval(intervalRef.current)
    }
    return () => { if (intervalRef.current) clearInterval(intervalRef.current) }
  }, [autoRefresh])

  const filtered = filter
    ? logs.filter(l =>
        l.domain.includes(filter) ||
        l.mac.includes(filter) ||
        l.device.includes(filter) ||
        l.client_ip.includes(filter) ||
        l.action.includes(filter.toUpperCase())
      )
    : logs

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-white">Query Log</h1>
          <p className="text-sm text-gray-400 mt-0.5">Last 500 DNS queries</p>
        </div>
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-sm text-gray-400 cursor-pointer select-none">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={e => setAutoRefresh(e.target.checked)}
              className="rounded"
            />
            Auto-refresh
          </label>
          <button onClick={load} className="btn-ghost">Refresh</button>
        </div>
      </div>

      <div className="mb-4">
        <input
          className="field max-w-sm"
          placeholder="Filter by domain, MAC, device, IP, action…"
          value={filter}
          onChange={e => setFilter(e.target.value)}
        />
      </div>

      <div className="card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-xs font-mono">
            <thead>
              <tr className="border-b border-gray-800 text-left text-gray-500">
                <th className="px-3 py-2">Time</th>
                <th className="px-3 py-2">Action</th>
                <th className="px-3 py-2">Domain</th>
                <th className="px-3 py-2">Type</th>
                <th className="px-3 py-2">Client</th>
                <th className="px-3 py-2">Device</th>
                <th className="px-3 py-2">Profile</th>
                <th className="px-3 py-2">RCode</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-800/50">
              {filtered.length === 0 && (
                <tr><td colSpan={8} className="px-3 py-6 text-center text-gray-500">No entries</td></tr>
              )}
              {filtered.map(l => (
                <tr key={l.id} className="hover:bg-gray-800/30 transition-colors">
                  <td className="px-3 py-1.5 text-gray-500 whitespace-nowrap">
                    {new Date(l.time).toLocaleTimeString()}
                  </td>
                  <td className={`px-3 py-1.5 font-bold ${actionColor[l.action] ?? 'text-gray-400'}`}>
                    {l.action}
                  </td>
                  <td className="px-3 py-1.5 text-gray-200 max-w-xs truncate">{l.domain}</td>
                  <td className="px-3 py-1.5 text-gray-500">{l.type}</td>
                  <td className="px-3 py-1.5 text-gray-400">{l.client_ip}</td>
                  <td className="px-3 py-1.5 text-gray-300">{l.device || l.mac || '—'}</td>
                  <td className="px-3 py-1.5 text-gray-400">{l.profile || '—'}</td>
                  <td className="px-3 py-1.5 text-gray-500">{l.rcode || '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
