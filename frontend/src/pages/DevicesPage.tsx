import { useEffect, useState } from 'react'
import { api } from '../api'
import type { Device, DeviceGroup, Profile, SeenDevice } from '../types'
import { Modal } from '../components/Modal'

const empty: Device = { mac: '', name: '', profile: '' }

export function DevicesPage() {
  const [devices, setDevices]   = useState<Device[]>([])
  const [profiles, setProfiles] = useState<Profile[]>([])
  const [groups, setGroups]     = useState<DeviceGroup[]>([])
  const [seen, setSeen]         = useState<SeenDevice[]>([])
  const [editing, setEditing]   = useState<Device | null>(null)
  const [isNew, setIsNew]       = useState(false)
  const [error, setError]       = useState('')

  const load = async () => {
    const [d, p, g, s] = await Promise.all([
      api.listDevices(), api.listProfiles(), api.listDeviceGroups(), api.seenDevices(),
    ])
    setDevices(d)
    setProfiles(p)
    setGroups(g)
    setSeen(s)
  }

  useEffect(() => { load() }, [])

  const groupFor = (mac: string) => groups.find(g => g.devices.includes(mac))

  const openNew = (prefill?: Partial<Device>) => {
    setEditing({ ...empty, ...prefill })
    setIsNew(true)
    setError('')
  }
  const openEdit = (d: Device) => { setEditing({ ...d }); setIsNew(false); setError('') }
  const close = () => { setEditing(null); setError('') }

  const save = async () => {
    if (!editing) return
    try {
      if (isNew) {
        await api.createDevice(editing)
      } else {
        await api.updateDevice(editing.mac, { name: editing.name, profile: editing.profile })
      }
      await load()
      close()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to save')
    }
  }

  const remove = async (mac: string) => {
    if (!confirm(`Remove device ${mac}?`)) return
    await api.deleteDevice(mac)
    await load()
  }

  const unregistered = seen.filter(s => !s.registered)

  return (
    <div className="space-y-8">
      {/* Registered devices */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <div>
            <h1 className="text-2xl font-bold text-white">Devices</h1>
            <p className="text-sm text-gray-400 mt-0.5">Map MAC addresses to filter profiles</p>
          </div>
          <button onClick={() => openNew()} className="btn-primary">+ Add Device</button>
        </div>

        <div className="card overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-800 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                <th className="px-4 py-3">MAC Address</th>
                <th className="px-4 py-3">Name</th>
                <th className="px-4 py-3">Profile</th>
                <th className="px-4 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-800">
              {devices.length === 0 && (
                <tr><td colSpan={4} className="px-4 py-8 text-center text-gray-500">No devices registered</td></tr>
              )}
              {devices.map(d => {
                const group = !d.profile ? groupFor(d.mac) : undefined
                return (
                  <tr key={d.mac} className="hover:bg-gray-800/50 transition-colors">
                    <td className="px-4 py-3 font-mono text-gray-300">{d.mac}</td>
                    <td className="px-4 py-3 text-white">{d.name || <span className="text-gray-500">—</span>}</td>
                    <td className="px-4 py-3">
                      {d.profile
                        ? <span className="badge">{d.profile}</span>
                        : group
                          ? <span className="text-xs text-indigo-300">via <span className="font-medium">{group.name}</span></span>
                          : <span className="text-gray-500">—</span>}
                    </td>
                    <td className="px-4 py-3 text-right space-x-2">
                      <button onClick={() => openEdit(d)} className="btn-ghost">Edit</button>
                      <button onClick={() => remove(d.mac)} className="btn-danger">Remove</button>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>

      {/* Seen but unregistered */}
      {unregistered.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold text-white mb-3">
            Seen on network
            <span className="ml-2 text-sm font-normal text-gray-500">unregistered devices making DNS queries</span>
          </h2>
          <div className="card overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-800 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  <th className="px-4 py-3">MAC Address</th>
                  <th className="px-4 py-3">Last IP</th>
                  <th className="px-4 py-3">Queries</th>
                  <th className="px-4 py-3">Last Seen</th>
                  <th className="px-4 py-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-800">
                {unregistered.map(s => (
                  <tr key={s.mac} className="hover:bg-gray-800/50 transition-colors">
                    <td className="px-4 py-3 font-mono text-gray-400">{s.mac}</td>
                    <td className="px-4 py-3 text-gray-400">{s.client_ip}</td>
                    <td className="px-4 py-3 text-gray-400">{s.query_count.toLocaleString()}</td>
                    <td className="px-4 py-3 text-gray-500 text-xs">
                      {new Date(s.last_seen).toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => openNew({ mac: s.mac })}
                        className="btn-primary text-xs px-3 py-1"
                      >
                        Register
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {editing && (
        <Modal title={isNew ? 'Add Device' : 'Edit Device'} onClose={close}>
          <div className="space-y-4">
            <div>
              <label className="field-label">MAC Address</label>
              <input
                className="field"
                placeholder="aa:bb:cc:dd:ee:ff"
                value={editing.mac}
                readOnly={!isNew}
                onChange={e => setEditing({ ...editing, mac: e.target.value })}
              />
            </div>
            <div>
              <label className="field-label">Name</label>
              <input
                className="field"
                placeholder="my-laptop"
                value={editing.name}
                onChange={e => setEditing({ ...editing, name: e.target.value })}
              />
            </div>
            <div>
              <label className="field-label">Profile</label>
              <select
                className="field"
                value={editing.profile}
                onChange={e => setEditing({ ...editing, profile: e.target.value })}
              >
                <option value="">— no profile (use group or default) —</option>
                {profiles.map(p => <option key={p.name} value={p.name}>{p.name}</option>)}
              </select>
            </div>
            {error && <p className="text-sm text-red-400">{error}</p>}
            <div className="flex justify-end gap-2 pt-2">
              <button onClick={close} className="btn-ghost">Cancel</button>
              <button onClick={save} className="btn-primary">Save</button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
