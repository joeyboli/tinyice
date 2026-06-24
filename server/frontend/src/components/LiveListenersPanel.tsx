import { signal } from '@preact/signals'
import { useEffect } from 'preact/hooks'
import { api } from '../lib/api'
import type { LiveListener } from '../types'

const listeners = signal<LiveListener[]>([])
const loading = signal(true)
const mountFilter = signal('')
const transportFilter = signal('')

function formatDuration(secs: number): string {
  if (secs < 60) return `${secs}s`
  const m = Math.floor(secs / 60)
  const s = secs % 60
  if (m < 60) return `${m}m ${s}s`
  const h = Math.floor(m / 60)
  return `${h}h ${m % 60}m`
}

function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1048576) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1048576).toFixed(1)} MB`
}

function shortUA(ua: string): string {
  if (!ua) return '—'
  if (ua.length <= 48) return ua
  return ua.slice(0, 45) + '…'
}

function locationLabel(l: LiveListener): string {
  if (l.city && l.country) return `${l.city}, ${l.country}`
  if (l.country) return l.country
  if (l.country_iso) return l.country_iso
  return '—'
}

async function loadListeners() {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (mountFilter.value) params.set('mount', mountFilter.value)
    if (transportFilter.value) params.set('transport', transportFilter.value)
    const qs = params.toString()
    const res = await api.get<{ listeners: LiveListener[]; total: number }>(
      `/api/listeners${qs ? `?${qs}` : ''}`,
    )
    listeners.value = res.listeners || []
  } catch {
    listeners.value = []
  }
  loading.value = false
}

export function applyListenersSSE(data: { listeners?: LiveListener[] }) {
  if (!mountFilter.value && !transportFilter.value && data.listeners) {
    listeners.value = data.listeners
    loading.value = false
  }
}

interface Props {
  mounts: string[]
}

export function LiveListenersPanel({ mounts }: Props) {
  useEffect(() => {
    loadListeners()
    const id = setInterval(() => { void loadListeners() }, 15000)
    return () => clearInterval(id)
  }, [])

  useEffect(() => {
    void loadListeners()
  }, [mountFilter.value, transportFilter.value])

  const uniqueMounts = ['', ...Array.from(new Set(mounts)).sort()]

  return (
    <div class="rounded-lg border border-border bg-surface-raised overflow-hidden mb-6">
      <div class="px-4 py-3 border-b border-border flex flex-wrap items-center justify-between gap-3">
        <div>
          <span class="font-mono text-[10px] tracking-widest uppercase text-text-tertiary block">
            Live Listeners
          </span>
          <span class="text-text-secondary text-xs mt-0.5">
            {listeners.value.length} connected now
          </span>
        </div>
        <div class="flex items-center gap-2">
          <select
            value={mountFilter.value}
            onChange={(e) => { mountFilter.value = (e.target as HTMLSelectElement).value }}
            class="bg-surface-overlay border border-border rounded px-2 py-1 font-mono text-[10px] text-text-secondary"
          >
            {uniqueMounts.map((m) => (
              <option key={m || 'all'} value={m}>{m || 'All mounts'}</option>
            ))}
          </select>
          <select
            value={transportFilter.value}
            onChange={(e) => { transportFilter.value = (e.target as HTMLSelectElement).value }}
            class="bg-surface-overlay border border-border rounded px-2 py-1 font-mono text-[10px] text-text-secondary"
          >
            <option value="">All transports</option>
            <option value="http">HTTP</option>
            <option value="hls">HLS</option>
            <option value="webrtc">WebRTC</option>
            <option value="whep">WHEP</option>
          </select>
        </div>
      </div>

      {loading.value && listeners.value.length === 0 ? (
        <div class="px-4 py-8 text-center text-text-tertiary text-sm">Loading listeners…</div>
      ) : listeners.value.length === 0 ? (
        <div class="px-4 py-8 text-center text-text-tertiary text-sm">No active listeners</div>
      ) : (
        <div class="overflow-x-auto">
          <table class="w-full min-w-[720px]">
            <thead>
              <tr class="border-b border-border text-left">
                <th class="px-4 py-2 font-mono text-[9px] tracking-widest text-text-tertiary uppercase">Mount</th>
                <th class="px-4 py-2 font-mono text-[9px] tracking-widest text-text-tertiary uppercase">IP</th>
                <th class="px-4 py-2 font-mono text-[9px] tracking-widest text-text-tertiary uppercase">Location</th>
                <th class="px-4 py-2 font-mono text-[9px] tracking-widest text-text-tertiary uppercase">Transport</th>
                <th class="px-4 py-2 font-mono text-[9px] tracking-widest text-text-tertiary uppercase">User Agent</th>
                <th class="px-4 py-2 font-mono text-[9px] tracking-widest text-text-tertiary uppercase">Connected</th>
                <th class="px-4 py-2 font-mono text-[9px] tracking-widest text-text-tertiary uppercase text-right">Sent</th>
              </tr>
            </thead>
            <tbody>
              {listeners.value.map((l) => (
                <tr key={l.id} class="border-b border-border/50 hover:bg-surface-hover/50">
                  <td class="px-4 py-2.5 font-mono text-xs text-text-primary">{l.mount}</td>
                  <td class="px-4 py-2.5 font-mono text-xs text-text-secondary">{l.ip}</td>
                  <td class="px-4 py-2.5 font-mono text-xs text-text-tertiary">{locationLabel(l)}</td>
                  <td class="px-4 py-2.5">
                    <span class="font-mono text-[10px] tracking-wider uppercase px-1.5 py-0.5 rounded bg-surface-overlay text-text-secondary">
                      {l.transport}
                    </span>
                  </td>
                  <td class="px-4 py-2.5 font-mono text-[11px] text-text-tertiary max-w-[220px] truncate" title={l.user_agent}>
                    {shortUA(l.user_agent)}
                  </td>
                  <td class="px-4 py-2.5 font-mono text-xs text-text-tertiary tabular-nums">
                    {formatDuration(l.connected_seconds)}
                  </td>
                  <td class="px-4 py-2.5 font-mono text-xs text-text-tertiary text-right tabular-nums">
                    {formatBytes(l.bytes_sent)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
