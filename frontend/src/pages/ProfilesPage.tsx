import { useEffect, useState } from 'react'
import { api } from '../api'
import type { Profile } from '../types'
import { Modal } from '../components/Modal'
import { PatternInput } from '../components/PatternInput'

const emptyProfile: Profile = { name: '', block: [], allow_only: [] }

export function ProfilesPage() {
  const [profiles, setProfiles] = useState<Profile[]>([])
  const [editing, setEditing] = useState<Profile | null>(null)
  const [isNew, setIsNew] = useState(false)
  const [expanded, setExpanded] = useState<string | null>(null)
  const [error, setError] = useState('')

  const load = () => api.listProfiles().then(setProfiles)
  useEffect(() => { load() }, [])

  const openNew = () => { setEditing({ ...emptyProfile }); setIsNew(true); setError('') }
  const openEdit = (p: Profile) => { setEditing({ ...p, block: [...p.block], allow_only: [...p.allow_only] }); setIsNew(false); setError('') }
  const close = () => { setEditing(null); setError('') }

  const save = async () => {
    if (!editing) return
    try {
      if (isNew) {
        await api.createProfile(editing)
      } else {
        await api.updateProfile(editing.name, { block: editing.block, allow_only: editing.allow_only })
      }
      await load()
      close()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to save')
    }
  }

  const remove = async (name: string) => {
    if (!confirm(`Delete profile "${name}"?`)) return
    await api.deleteProfile(name)
    await load()
  }

  const toggle = (name: string) => setExpanded(expanded === name ? null : name)

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-white">Profiles</h1>
          <p className="text-sm text-gray-400 mt-0.5">Define block and allow-only rules</p>
        </div>
        <button onClick={openNew} className="btn-primary">+ Add Profile</button>
      </div>

      <div className="space-y-2">
        {profiles.length === 0 && (
          <div className="card px-4 py-8 text-center text-gray-500">No profiles defined</div>
        )}
        {profiles.map(p => (
          <div key={p.name} className="card overflow-hidden">
            <div
              className="flex items-center justify-between px-4 py-3 cursor-pointer hover:bg-gray-800/50 transition-colors"
              onClick={() => toggle(p.name)}
            >
              <div className="flex items-center gap-3">
                <span className="text-white font-medium">{p.name}</span>
                <div className="flex gap-1.5">
                  {p.block.length > 0 && (
                    <span className="text-xs bg-red-900/50 text-red-300 rounded px-1.5 py-0.5">
                      {p.block.length} blocked
                    </span>
                  )}
                  {p.allow_only.length > 0 && (
                    <span className="text-xs bg-green-900/50 text-green-300 rounded px-1.5 py-0.5">
                      {p.allow_only.length} allow-only
                    </span>
                  )}
                  {p.block.length === 0 && p.allow_only.length === 0 && (
                    <span className="text-xs text-gray-500">no rules</span>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <button onClick={e => { e.stopPropagation(); openEdit(p) }} className="btn-ghost">Edit</button>
                <button onClick={e => { e.stopPropagation(); remove(p.name) }} className="btn-danger">Delete</button>
                <span className="text-gray-500 ml-1">{expanded === p.name ? '▲' : '▼'}</span>
              </div>
            </div>

            {expanded === p.name && (
              <div className="border-t border-gray-800 px-4 py-3 grid grid-cols-2 gap-4">
                <div>
                  <p className="text-xs font-medium uppercase tracking-wider text-gray-500 mb-2">Blocked domains</p>
                  {p.block.length === 0
                    ? <p className="text-xs text-gray-600">None</p>
                    : <ul className="space-y-0.5">{p.block.map(pat => (
                        <li key={pat} className="font-mono text-xs text-red-300">{pat}</li>
                      ))}</ul>}
                </div>
                <div>
                  <p className="text-xs font-medium uppercase tracking-wider text-gray-500 mb-2">Allow-only domains</p>
                  {p.allow_only.length === 0
                    ? <p className="text-xs text-gray-600">None</p>
                    : <ul className="space-y-0.5">{p.allow_only.map(pat => (
                        <li key={pat} className="font-mono text-xs text-green-300">{pat}</li>
                      ))}</ul>}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>

      {editing && (
        <Modal title={isNew ? 'Add Profile' : `Edit Profile: ${editing.name}`} onClose={close}>
          <div className="space-y-4">
            {isNew && (
              <div>
                <label className="field-label">Profile Name</label>
                <input
                  className="field"
                  placeholder="kids"
                  value={editing.name}
                  onChange={e => setEditing({ ...editing, name: e.target.value })}
                />
              </div>
            )}
            <PatternInput
              label="Block patterns"
              patterns={editing.block}
              onChange={block => setEditing({ ...editing, block })}
              placeholder="**.tiktok.com"
            />
            <PatternInput
              label="Allow-only patterns (if set, everything else is blocked)"
              patterns={editing.allow_only}
              onChange={allow_only => setEditing({ ...editing, allow_only })}
              placeholder="**.apple.com"
            />
            <p className="text-xs text-gray-500">
              Use <code className="text-gray-400">**.example.com</code> to match the domain and all subdomains,
              or <code className="text-gray-400">*.example.com</code> for subdomains only.
            </p>
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
