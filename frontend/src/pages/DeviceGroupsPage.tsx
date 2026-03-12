import { useEffect, useState } from 'react'
import { api } from '../api'
import type { Device, DeviceGroup, Profile, Schedule } from '../types'
import { Modal } from '../components/Modal'

const emptyGroup: DeviceGroup = { name: '', profile: '', devices: [], schedules: [] }
const emptySchedule: Schedule = { profile: '', start: '00:00', end: '06:00' }

export function DeviceGroupsPage() {
  const [groups, setGroups]     = useState<DeviceGroup[]>([])
  const [profiles, setProfiles] = useState<Profile[]>([])
  const [devices, setDevices]   = useState<Device[]>([])
  const [editing, setEditing]   = useState<DeviceGroup | null>(null)
  const [isNew, setIsNew]       = useState(false)
  const [expanded, setExpanded] = useState<string | null>(null)
  const [error, setError]       = useState('')

  const load = async () => {
    const [g, p, d] = await Promise.all([api.listDeviceGroups(), api.listProfiles(), api.listDevices()])
    setGroups(g)
    setProfiles(p)
    setDevices(d)
  }
  useEffect(() => { load() }, [])

  const openNew = () => {
    setEditing({ ...emptyGroup, devices: [], schedules: [] })
    setIsNew(true)
    setError('')
  }
  const openEdit = (g: DeviceGroup) => {
    setEditing({ ...g, devices: [...g.devices], schedules: g.schedules.map(s => ({ ...s })) })
    setIsNew(false)
    setError('')
  }
  const close = () => { setEditing(null); setError('') }

  const save = async () => {
    if (!editing) return
    try {
      if (isNew) {
        await api.createDeviceGroup(editing)
      } else {
        await api.updateDeviceGroup(editing.name, {
          profile: editing.profile,
          devices: editing.devices,
          schedules: editing.schedules,
        })
      }
      await load()
      close()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to save')
    }
  }

  const remove = async (name: string) => {
    if (!confirm(`Delete group "${name}"?`)) return
    await api.deleteDeviceGroup(name)
    await load()
  }

  const toggle = (name: string) => setExpanded(expanded === name ? null : name)

  const toggleDevice = (mac: string) => {
    if (!editing) return
    const devs = editing.devices.includes(mac)
      ? editing.devices.filter(m => m !== mac)
      : [...editing.devices, mac]
    setEditing({ ...editing, devices: devs })
  }

  const addSchedule = () => {
    if (!editing) return
    setEditing({ ...editing, schedules: [...editing.schedules, { ...emptySchedule }] })
  }

  const updateSchedule = (i: number, s: Schedule) => {
    if (!editing) return
    setEditing({ ...editing, schedules: editing.schedules.map((ex, idx) => idx === i ? s : ex) })
  }

  const removeSchedule = (i: number) => {
    if (!editing) return
    setEditing({ ...editing, schedules: editing.schedules.filter((_, idx) => idx !== i) })
  }

  const deviceLabel = (mac: string) => devices.find(d => d.mac === mac)?.name || mac

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-white">Device Groups</h1>
          <p className="text-sm text-gray-400 mt-0.5">Apply a profile to multiple devices, with optional time-based schedules</p>
        </div>
        <button onClick={openNew} className="btn-primary">+ Add Group</button>
      </div>

      <div className="space-y-2">
        {groups.length === 0 && (
          <div className="card px-4 py-8 text-center text-gray-500">No device groups defined</div>
        )}
        {groups.map(g => (
          <div key={g.name} className="card overflow-hidden">
            <div
              className="flex items-center justify-between px-4 py-3 cursor-pointer hover:bg-gray-800/50 transition-colors"
              onClick={() => toggle(g.name)}
            >
              <div className="flex items-center gap-3">
                <span className="text-white font-medium">{g.name}</span>
                <span className="badge">{g.profile}</span>
                <span className="text-xs text-gray-500">
                  {g.devices.length} {g.devices.length === 1 ? 'device' : 'devices'}
                </span>
                {g.schedules.length > 0 && (
                  <span className="text-xs bg-indigo-900/50 text-indigo-300 rounded px-1.5 py-0.5">
                    {g.schedules.length} {g.schedules.length === 1 ? 'schedule' : 'schedules'}
                  </span>
                )}
              </div>
              <div className="flex items-center gap-2">
                <button onClick={e => { e.stopPropagation(); openEdit(g) }} className="btn-ghost">Edit</button>
                <button onClick={e => { e.stopPropagation(); remove(g.name) }} className="btn-danger">Delete</button>
                <span className="text-gray-500 ml-1">{expanded === g.name ? '▲' : '▼'}</span>
              </div>
            </div>

            {expanded === g.name && (
              <div className="border-t border-gray-800 px-4 py-3 space-y-4">
                <div>
                  <p className="text-xs font-medium uppercase tracking-wider text-gray-500 mb-2">Devices</p>
                  {g.devices.length === 0
                    ? <p className="text-xs text-gray-600">No devices</p>
                    : <div className="flex flex-wrap gap-1.5">
                        {g.devices.map(mac => (
                          <span key={mac} className="text-xs font-mono bg-gray-800 text-gray-300 rounded px-2 py-0.5">
                            {deviceLabel(mac) !== mac
                              ? <>{deviceLabel(mac)} <span className="text-gray-500">({mac})</span></>
                              : mac}
                          </span>
                        ))}
                      </div>
                  }
                </div>

                {g.schedules.length > 0 && (
                  <div>
                    <p className="text-xs font-medium uppercase tracking-wider text-gray-500 mb-2">Schedules</p>
                    <div className="space-y-1">
                      {g.schedules.map((s, i) => (
                        <div key={i} className="flex items-center gap-2 text-xs">
                          <span className="font-mono text-gray-400">{s.start} – {s.end}</span>
                          <span className="text-gray-600">→</span>
                          <span className="badge">{s.profile}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        ))}
      </div>

      {editing && (
        <Modal title={isNew ? 'Add Group' : `Edit Group: ${editing.name}`} onClose={close}>
          <div className="space-y-4">
            {isNew && (
              <div>
                <label className="field-label">Group Name</label>
                <input
                  className="field"
                  placeholder="kids-devices"
                  value={editing.name}
                  onChange={e => setEditing({ ...editing, name: e.target.value })}
                />
              </div>
            )}

            <div>
              <label className="field-label">Default Profile</label>
              <select
                className="field"
                value={editing.profile}
                onChange={e => setEditing({ ...editing, profile: e.target.value })}
              >
                <option value="">— select a profile —</option>
                {profiles.map(p => <option key={p.name} value={p.name}>{p.name}</option>)}
              </select>
            </div>

            <div>
              <label className="field-label">Devices</label>
              {devices.length === 0
                ? <p className="text-xs text-gray-500">No devices registered yet</p>
                : <div className="space-y-1 max-h-36 overflow-y-auto rounded border border-gray-700 p-2">
                    {devices.map(d => (
                      <label key={d.mac} className="flex items-center gap-2 cursor-pointer py-0.5">
                        <input
                          type="checkbox"
                          checked={editing.devices.includes(d.mac)}
                          onChange={() => toggleDevice(d.mac)}
                          className="rounded border-gray-600 bg-gray-800 text-indigo-500"
                        />
                        <span className="text-sm text-white">{d.name || d.mac}</span>
                        {d.name && <span className="text-xs text-gray-500 font-mono">{d.mac}</span>}
                      </label>
                    ))}
                  </div>
              }
            </div>

            <div>
              <div className="flex items-center justify-between mb-2">
                <label className="field-label" style={{ marginBottom: 0 }}>Schedules</label>
                <button onClick={addSchedule} className="btn-ghost text-xs px-2 py-0.5">+ Add</button>
              </div>
              {editing.schedules.length === 0
                ? <p className="text-xs text-gray-500">No schedules — default profile applies at all times</p>
                : <div className="space-y-2">
                    {editing.schedules.map((s, i) => (
                      <div key={i} className="flex items-center gap-2">
                        <input
                          type="time"
                          className="field py-1.5 w-28"
                          value={s.start}
                          onChange={e => updateSchedule(i, { ...s, start: e.target.value })}
                        />
                        <span className="text-gray-500 text-xs">to</span>
                        <input
                          type="time"
                          className="field py-1.5 w-28"
                          value={s.end}
                          onChange={e => updateSchedule(i, { ...s, end: e.target.value })}
                        />
                        <select
                          className="field py-1.5 flex-1"
                          value={s.profile}
                          onChange={e => updateSchedule(i, { ...s, profile: e.target.value })}
                        >
                          <option value="">— profile —</option>
                          {profiles.map(p => <option key={p.name} value={p.name}>{p.name}</option>)}
                        </select>
                        <button onClick={() => removeSchedule(i)} className="btn-danger text-xs px-2 py-1">×</button>
                      </div>
                    ))}
                    <p className="text-xs text-gray-500">Windows crossing midnight (e.g. 22:00–06:00) are handled correctly.</p>
                  </div>
              }
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
