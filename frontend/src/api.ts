import type { Device, LogEntry, Profile, SeenDevice, Settings } from './types'

async function req<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error ?? res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  // Settings
  getSettings: () => req<Settings>('GET', '/api/settings'),
  putSettings: (s: Settings) => req<Settings>('PUT', '/api/settings', s),

  // Upstreams
  getUpstreams: () => req<string[]>('GET', '/api/upstreams'),
  putUpstreams: (u: string[]) => req<string[]>('PUT', '/api/upstreams', u),

  // Profiles
  listProfiles: () => req<Profile[]>('GET', '/api/profiles'),
  createProfile: (p: Omit<Profile, ''>) => req<Profile>('POST', '/api/profiles', p),
  updateProfile: (name: string, p: Omit<Profile, 'name'>) =>
    req<Profile>('PUT', `/api/profiles/${encodeURIComponent(name)}`, p),
  deleteProfile: (name: string) =>
    req<void>('DELETE', `/api/profiles/${encodeURIComponent(name)}`),

  // Devices
  listDevices: () => req<Device[]>('GET', '/api/devices'),
  createDevice: (d: Device) => req<Device>('POST', '/api/devices', d),
  updateDevice: (mac: string, d: Omit<Device, 'mac'>) =>
    req<Device>('PUT', `/api/devices/${encodeURIComponent(mac)}`, d),
  deleteDevice: (mac: string) =>
    req<void>('DELETE', `/api/devices/${encodeURIComponent(mac)}`),

  // Logs
  listLogs: (limit = 200) => req<LogEntry[]>('GET', `/api/logs?limit=${limit}`),
  seenDevices: () => req<SeenDevice[]>('GET', '/api/seen-devices'),
}
