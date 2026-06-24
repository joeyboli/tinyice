import { useState, useEffect } from 'preact/hooks'
import { Nav } from '@/components/Nav'
import { CodeBlock } from '@/components/CodeBlock'
import type { TinyIceBase } from '@/types'

const data = window.__TINYICE__ as TinyIceBase

const sections = [
  { id: 'getting-started', label: 'Getting Started', children: [
    { id: 'quick-start', label: 'Quick Start' },
    { id: 'installation', label: 'Installation' },
  ]},
  { id: 'streaming', label: 'Streaming', children: [
    { id: 'webrtc-source', label: 'WebRTC Source' },
    { id: 'http-listening', label: 'HTTP Listening' },
    { id: 'hls-playback', label: 'HLS Playback' },
    { id: 'webrtc-playback', label: 'WebRTC Playback' },
    { id: 'metadata-sse', label: 'Metadata SSE' },
    { id: 'podcast-feed', label: 'Podcast Feed' },
  ]},
  { id: 'reference', label: 'Reference', children: [
    { id: 'rest-api', label: 'REST API' },
    { id: 'api-tokens', label: 'API Tokens' },
    { id: 'live-listeners', label: 'Live Listeners' },
    { id: 'player-options', label: 'Player Options' },
    { id: 'api-docs', label: 'API Docs (Swagger)' },
    { id: 'embed-widget', label: 'Embed Widget' },
    { id: 'icecast-compat', label: 'Icecast Compat' },
  ]},
]

const endpoints = [
  { method: 'GET', path: '/{mount}', desc: 'Listen to audio stream (HTTP)' },
  { method: 'GET', path: '/{mount}/playlist.m3u8', desc: 'HLS playlist (audio + video)' },
  { method: 'GET', path: '/{mount}/feed.xml', desc: 'RSS podcast feed (recent tracks)' },
  { method: 'POST', path: '/webrtc/source-offer', desc: 'Start WebRTC source stream' },
  { method: 'POST', path: '/webrtc/offer', desc: 'WebRTC listener connection' },
  { method: 'GET', path: '/events', desc: 'Real-time metadata (SSE)' },
  { method: 'GET', path: '/api/streams', desc: 'List all stream mounts' },
  { method: 'POST', path: '/api/streams', desc: 'Create a stream mount' },
  { method: 'GET', path: '/api/autodj', desc: 'List AutoDJ instances' },
  { method: 'GET', path: '/api/stats', desc: 'Server statistics' },
  { method: 'GET', path: '/api/listeners', desc: 'Live listeners (IP, UA, transport)' },
  { method: 'GET', path: '/admin/listeners', desc: 'Live listeners (session auth)' },
  { method: 'GET', path: '/admin/insights/export', desc: 'Export listener history (CSV)' },
  { method: 'GET', path: '/api/relays', desc: 'List relay connections' },
  { method: 'GET', path: '/api/users', desc: 'List users' },
  { method: 'GET', path: '/api/settings', desc: 'Server configuration' },
  { method: 'GET', path: '/api/branding', desc: 'Branding settings' },
]

function methodColor(m: string) {
  const colors: Record<string, string> = {
    GET: 'bg-blue-500/15 text-blue-400 border-blue-500/20',
    POST: 'bg-green-500/15 text-green-400 border-green-500/20',
    PUT: 'bg-yellow-500/15 text-yellow-400 border-yellow-500/20',
    DELETE: 'bg-red-500/15 text-red-400 border-red-500/20',
  }
  return colors[m] || colors.GET
}

export function Developers() {
  const [activeSection, setActiveSection] = useState('quick-start')

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            setActiveSection(entry.target.id)
          }
        }
      },
      { rootMargin: '-80px 0px -60% 0px', threshold: 0 }
    )

    const allIds = sections.flatMap(s => [s.id, ...s.children.map(c => c.id)])
    for (const id of allIds) {
      const el = document.getElementById(id)
      if (el) observer.observe(el)
    }

    return () => observer.disconnect()
  }, [])

  const scrollTo = (id: string) => {
    const el = document.getElementById(id)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'start' })
    }
  }

  return (
    <div class="min-h-screen bg-surface-base">
      <Nav branding={data.branding} pageTitle={data.pageTitle} />

      <div class="flex pt-14">
        {/* Sidebar */}
        <aside class="fixed top-14 left-0 bottom-0 w-[210px] border-r border-border overflow-y-auto px-5 py-6">
          <div class="flex items-center gap-2 mb-6">
            <svg class="w-4 h-4 text-accent" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="16 18 22 12 16 6" />
              <polyline points="8 6 2 12 8 18" />
            </svg>
            <span class="font-mono text-[11px] tracking-widest text-accent font-bold">DEVELOPERS</span>
          </div>

          {sections.map((section) => (
            <div key={section.id} class="mb-5">
              <div class="font-mono text-[9px] tracking-widest text-text-tertiary uppercase mb-2 px-3">
                {section.label}
              </div>
              {section.children.map((child) => (
                <button
                  key={child.id}
                  onClick={() => scrollTo(child.id)}
                  class={`w-full text-left px-3 py-1.5 font-mono text-[11px] tracking-wide transition-colors rounded-r-md border-l-2 ${
                    activeSection === child.id
                      ? 'border-accent text-text-primary bg-accent-subtle'
                      : 'border-transparent text-text-tertiary hover:text-text-secondary hover:border-border-hover'
                  }`}
                >
                  {child.label}
                </button>
              ))}
            </div>
          ))}
        </aside>

        {/* Content */}
        <main class="flex-1 ml-[210px] min-w-0">
          <div class="max-w-3xl mx-auto px-8 py-10 space-y-16">

            {/* Quick Start */}
            <section id="quick-start">
              <h1 class="font-heading text-2xl font-bold text-text-primary mb-2">Quick Start</h1>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                Get streaming audio in under a minute. TinyIce uses WebRTC for ultra-low-latency source connections
                and standard HTTP for listener playback.
              </p>

              <CodeBlock tabs={[{
                label: 'TypeScript',
                language: 'typescript',
                code: `// Capture microphone audio
const stream = await navigator.mediaDevices
  .getUserMedia({ audio: true })

// Create WebRTC peer connection
const pc = new RTCPeerConnection()
stream.getTracks().forEach(t => pc.addTrack(t, stream))

// Signal to TinyIce server
const offer = await pc.createOffer()
await pc.setLocalDescription(offer)

const res = await fetch('/webrtc/source-offer?mount=/live', {
  method: 'POST',
  body: JSON.stringify(offer),
})
const answer = await res.json()
await pc.setRemoteDescription(answer)
// You're now streaming!`,
              }]} />

              {/* Flow cards */}
              <div class="flex items-center justify-center gap-3 mt-8">
                {[
                  { num: '01', label: 'Connect' },
                  { num: '02', label: 'Stream' },
                  { num: '03', label: 'Listen' },
                ].map((step, i) => (
                  <>
                    <div key={step.num} class="flex-1 rounded-lg border border-border bg-surface-raised p-4 text-center">
                      <div class="font-mono text-[10px] tracking-widest text-accent mb-1">{step.num}</div>
                      <div class="font-heading text-sm font-semibold text-text-primary">{step.label}</div>
                    </div>
                    {i < 2 && (
                      <svg class="w-4 h-4 text-text-tertiary flex-none" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M5 12h14M12 5l7 7-7 7" />
                      </svg>
                    )}
                  </>
                ))}
              </div>
            </section>

            {/* Installation */}
            <section id="installation">
              <h2 class="font-heading text-xl font-bold text-text-primary mb-2">Installation</h2>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                TinyIce is a single binary with zero dependencies. Download the latest release or build from source.
              </p>

              <CodeBlock tabs={[
                {
                  label: 'Binary',
                  language: 'bash',
                  code: `# Download latest release
curl -L "https://github.com/DatanoiseTV/tinyice/releases/latest/download/tinyice-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)" -o tinyice
chmod +x tinyice

# Run the server
./tinyice --port 8000`,
                },
                {
                  label: 'From Source',
                  language: 'bash',
                  code: `# Clone and build
git clone https://github.com/DatanoiseTV/tinyice.git
cd tinyice
make build

# Run the server
./tinyice --port 8000`,
                },
              ]} />
            </section>

            {/* WebRTC Source */}
            <section id="webrtc-source">
              <h2 class="font-heading text-xl font-bold text-text-primary mb-2">WebRTC Source</h2>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                Source connections use WebRTC for real-time audio ingest. Create a peer connection,
                add your audio tracks, and exchange SDP with the server.
              </p>

              <CodeBlock tabs={[{
                label: 'TypeScript',
                language: 'typescript',
                code: `// Capture microphone audio
const stream = await navigator.mediaDevices
  .getUserMedia({ audio: true })

// Create WebRTC peer connection
const pc = new RTCPeerConnection()
stream.getTracks().forEach(t => pc.addTrack(t, stream))

// Signal to TinyIce server
const offer = await pc.createOffer()
await pc.setLocalDescription(offer)

const res = await fetch('/webrtc/source-offer?mount=/live', {
  method: 'POST',
  body: JSON.stringify(offer),
})
const answer = await res.json()
await pc.setRemoteDescription(answer)`,
              }]} />
            </section>

            {/* HTTP Listening */}
            <section id="http-listening">
              <h2 class="font-heading text-xl font-bold text-text-primary mb-2">HTTP Listening</h2>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                Listeners connect via standard HTTP. Simply point an audio element or player at the mount URL.
              </p>

              <CodeBlock tabs={[
                {
                  label: 'TypeScript',
                  language: 'typescript',
                  code: `// Play a stream with the Audio API
const audio = new Audio('https://your-server.com/live')
audio.play()

// Or with an HTML element
// <audio src="https://your-server.com/live" controls />`,
                },
              ]} />
            </section>

            {/* HLS Playback */}
            <section id="hls-playback">
              <h2 class="font-heading text-xl font-bold text-text-primary mb-2">HLS Playback</h2>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                Video streams (RTMP/SRT ingest) are available as HLS at <code class="font-code text-accent">/{'{mount}'}/playlist.m3u8</code>.
                Use hls.js in browsers that lack native HLS support.
              </p>

              <CodeBlock tabs={[
                {
                  label: 'curl',
                  language: 'bash',
                  code: `# Fetch the live HLS playlist
curl -s https://your-server.com/live/playlist.m3u8

# Play with ffplay
ffplay https://your-server.com/live/playlist.m3u8`,
                },
                {
                  label: 'TypeScript',
                  language: 'typescript',
                  code: `import Hls from 'hls.js'

const video = document.querySelector('video')!
const src = 'https://your-server.com/live/playlist.m3u8'

if (Hls.isSupported()) {
  const hls = new Hls()
  hls.loadSource(src)
  hls.attachMedia(video)
} else if (video.canPlayType('application/vnd.apple.mpegurl')) {
  video.src = src
}
video.play()`,
                },
              ]} />
            </section>

            {/* WebRTC Playback */}
            <section id="webrtc-playback">
              <h2 class="font-heading text-xl font-bold text-text-primary mb-2">WebRTC Playback</h2>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                For ultra-low-latency listening, use WebRTC playback. The server pushes audio tracks
                directly to the client peer connection.
              </p>

              <CodeBlock tabs={[{
                label: 'TypeScript',
                language: 'typescript',
                code: `const pc = new RTCPeerConnection()
pc.ontrack = (e) => {
  const audio = document.createElement('audio')
  audio.srcObject = e.streams[0]
  audio.play()
}

const offer = await pc.createOffer()
await pc.setLocalDescription(offer)

const res = await fetch('/webrtc/offer?mount=/live', {
  method: 'POST',
  body: JSON.stringify(offer),
})
const answer = await res.json()
await pc.setRemoteDescription(answer)`,
              }]} />
            </section>

            {/* Metadata SSE */}
            <section id="metadata-sse">
              <h2 class="font-heading text-xl font-bold text-text-primary mb-2">Metadata (SSE)</h2>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                Subscribe to real-time metadata updates via Server-Sent Events. Get notified when
                track information changes on any mount.
              </p>

              <CodeBlock tabs={[{
                label: 'TypeScript',
                language: 'typescript',
                code: `const source = new EventSource('/events')
source.addEventListener('metadata', (e) => {
  const { mount, title, artist } = JSON.parse(e.data)
  console.log(\`Now playing on \${mount}: \${title} \u2014 \${artist}\`)
})`,
              }]} />
            </section>

            {/* Podcast Feed */}
            <section id="podcast-feed">
              <h2 class="font-heading text-xl font-bold text-text-primary mb-2">Podcast Feed (RSS)</h2>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                Each mount exposes a public RSS feed of recently played tracks at <code class="font-code text-accent">/{'{mount}'}/feed.xml</code>.
                Useful for podcast apps, station websites, and syndication.
              </p>

              <CodeBlock tabs={[{
                label: 'curl',
                language: 'bash',
                code: `curl -s https://your-server.com/live/feed.xml`,
              }]} />
            </section>

            {/* REST API */}
            <section id="rest-api">
              <h2 class="font-heading text-xl font-bold text-text-primary mb-2">REST API</h2>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                Core endpoints for interacting with TinyIce programmatically.
              </p>

              <div class="space-y-2">
                {endpoints.map((ep) => (
                  <div
                    key={ep.path}
                    class="group flex items-center gap-4 rounded-lg border border-border hover:border-accent/30 bg-surface-raised p-4 transition-colors cursor-pointer"
                  >
                    <span class={`inline-flex items-center justify-center px-2 py-0.5 rounded text-[10px] font-mono font-bold tracking-wider border ${methodColor(ep.method)}`}>
                      {ep.method}
                    </span>
                    <code class="font-code text-[13px] text-text-primary flex-1">{ep.path}</code>
                    <span class="text-text-tertiary text-xs hidden sm:block">{ep.desc}</span>
                    <svg class="w-4 h-4 text-text-tertiary group-hover:text-accent transition-colors flex-none" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M9 18l6-6-6-6" />
                    </svg>
                  </div>
                ))}
              </div>
            </section>

            {/* Live Listeners API */}
            <section id="live-listeners">
              <h2 class="font-heading text-xl font-bold text-text-primary mb-2">Live Listeners</h2>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                Query connected listeners with IP, user agent, transport, geo location, and bytes sent.
                The admin dashboard also receives live updates via the <code class="font-code">listeners</code> SSE event.
              </p>

              <CodeBlock tabs={[
                {
                  label: 'curl',
                  language: 'bash',
                  code: `# All live listeners
curl -b cookies.txt https://your-server.com/api/listeners

# Filter by mount and transport
curl -b cookies.txt "https://your-server.com/api/listeners?mount=/live&transport=http"`,
                },
                {
                  label: 'Response',
                  language: 'json',
                  code: `{
  "total": 1,
  "listeners": [{
    "id": "203.0.113.10:51234-1719240000000000",
    "mount": "/live",
    "ip": "203.0.113.10",
    "user_agent": "VLC/3.0.20",
    "transport": "http",
    "connected_at": "2026-06-24T12:00:00Z",
    "connected_seconds": 183,
    "bytes_sent": 2457600,
    "country_iso": "US",
    "country": "United States",
    "city": "Chicago"
  }]
}`,
                },
              ]} />
            </section>

            {/* API Tokens */}
            <section id="api-tokens">
              <h2 class="font-heading text-xl font-bold text-text-primary mb-2">API Tokens</h2>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                Create bearer tokens in Admin → API Tokens. Pass the token in the <code class="font-code">Authorization</code> header
                for programmatic access to the REST API.
              </p>

              <CodeBlock tabs={[
                {
                  label: 'curl',
                  language: 'bash',
                  code: `curl -H "Authorization: Bearer ti_YOUR_TOKEN" \\
  https://your-server.com/api/streams`,
                },
                {
                  label: 'Python',
                  language: 'python',
                  code: `import requests

headers = {"Authorization": "Bearer ti_YOUR_TOKEN"}
r = requests.get("https://your-server.com/api/streams", headers=headers)
print(r.json())`,
                },
              ]} />
            </section>

            {/* Player Options */}
            <section id="player-options">
              <h2 class="font-heading text-xl font-bold text-text-primary mb-2">Player Customization</h2>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                Customize the built-in player and embed widget with query parameters.
              </p>

              <CodeBlock tabs={[{
                label: 'URL params',
                language: 'bash',
                code: `# Hide stats overlay
/player/live?stats=0

# Autoplay on load
/player/live?autoplay=1

# Custom accent color
/player/live?accent=%23e91e63

# WebRTC mode (sub-second latency)
/player/live?mode=webrtc

# Combine options
/embed/live?autoplay=1&stats=0&accent=%233498db`,
              }]} />
            </section>

            {/* API Docs */}
            <section id="api-docs">
              <h2 class="font-heading text-xl font-bold text-text-primary mb-2">Interactive API Docs</h2>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                Explore the full API interactively with Swagger UI. All endpoints are documented
                with request/response schemas, parameter descriptions, and example values.
              </p>

              <div class="flex flex-col gap-3">
                <a
                  href="/api/docs"
                  target="_blank"
                  class="group flex items-center gap-4 rounded-lg border border-accent/30 bg-accent/5 hover:bg-accent/10 p-5 transition-colors"
                >
                  <div class="w-10 h-10 rounded-lg bg-accent/15 flex items-center justify-center flex-shrink-0">
                    <svg class="w-5 h-5 text-accent" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                      <polyline points="14 2 14 8 20 8" />
                      <line x1="16" y1="13" x2="8" y2="13" />
                      <line x1="16" y1="17" x2="8" y2="17" />
                    </svg>
                  </div>
                  <div class="flex-1">
                    <div class="font-heading text-sm font-semibold text-text-primary">Swagger UI</div>
                    <div class="text-text-tertiary text-xs mt-0.5">Interactive API explorer with try-it-out</div>
                  </div>
                  <svg class="w-4 h-4 text-text-tertiary group-hover:text-accent transition-colors" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                    <polyline points="15 3 21 3 21 9" />
                    <line x1="10" y1="14" x2="21" y2="3" />
                  </svg>
                </a>

                <a
                  href="/api/openapi.yaml"
                  target="_blank"
                  class="group flex items-center gap-4 rounded-lg border border-border hover:border-accent/30 bg-surface-raised p-5 transition-colors"
                >
                  <div class="w-10 h-10 rounded-lg bg-surface-overlay flex items-center justify-center flex-shrink-0">
                    <svg class="w-5 h-5 text-text-secondary" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                      <polyline points="16 18 22 12 16 6" />
                      <polyline points="8 6 2 12 8 18" />
                    </svg>
                  </div>
                  <div class="flex-1">
                    <div class="font-heading text-sm font-semibold text-text-primary">OpenAPI Spec</div>
                    <div class="text-text-tertiary text-xs mt-0.5">Download the raw OpenAPI 3.0 YAML specification</div>
                  </div>
                  <svg class="w-4 h-4 text-text-tertiary group-hover:text-accent transition-colors" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                    <polyline points="7 10 12 15 17 10" />
                    <line x1="12" y1="15" x2="12" y2="3" />
                  </svg>
                </a>
              </div>
            </section>

            {/* Embed Widget */}
            <section id="embed-widget">
              <h2 class="font-heading text-xl font-bold text-text-primary mb-2">Embed Widget</h2>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                Embed a minimal player widget on any website with a single iframe tag.
              </p>

              <CodeBlock tabs={[{
                label: 'HTML',
                language: 'bash',
                code: `<iframe
  src="https://your-server.com/embed/live"
  width="100%"
  height="80"
  frameborder="0"
  allow="autoplay"
/>`,
              }]} />
            </section>

            {/* Icecast Compat */}
            <section id="icecast-compat">
              <h2 class="font-heading text-xl font-bold text-text-primary mb-2">Icecast Compatibility</h2>
              <p class="text-text-secondary text-sm leading-relaxed mb-6">
                TinyIce implements the Icecast source protocol, so existing broadcasting software
                (like BUTT, Mixxx, or OBS) can connect without changes. Point your source client
                at the server address and mount path.
              </p>

              <CodeBlock tabs={[{
                label: 'Config',
                language: 'json',
                code: `{
  "host": "your-server.com",
  "port": 8080,
  "mount": "/live",
  "username": "source",
  "password": "your-password"
}`,
              }]} />
            </section>

          </div>

          {/* Footer */}
          <footer class="border-t border-border">
            <div class="max-w-3xl mx-auto px-8 py-6 flex items-center justify-between">
              <span class="font-mono text-[9px] tracking-widest text-text-tertiary/50 uppercase">
                TINYICE // PURE GO
              </span>
              <span class="font-mono text-[9px] tracking-widest text-text-tertiary/50">
                {data.version || 'dev'}
              </span>
            </div>
          </footer>
        </main>
      </div>
    </div>
  )
}
