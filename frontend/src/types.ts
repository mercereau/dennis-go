export interface Settings {
  listen: string
  default_profile: string
}

export interface Profile {
  name: string
  block: string[]
  allow_only: string[]
}

export interface Device {
  mac: string
  name: string
  profile: string
}

export interface LogEntry {
  id: number
  time: string
  client_ip: string
  mac: string
  device: string
  profile: string
  domain: string
  type: string
  action: 'ALLOW' | 'BLOCK' | 'ERROR'
  rcode: string
}

export interface SeenDevice {
  mac: string
  client_ip: string
  last_seen: string
  query_count: number
  registered: boolean
  name: string
  profile: string
}
