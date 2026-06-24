import { signal } from '@preact/signals'

export const passwordReveal = signal<{ title: string; label: string; password: string } | null>(null)
const copied = signal(false)

function fallbackCopy(text: string) {
  const el = document.createElement('textarea')
  el.value = text
  el.style.position = 'fixed'
  el.style.left = '-9999px'
  document.body.appendChild(el)
  el.select()
  document.execCommand('copy')
  document.body.removeChild(el)
}

export function showPasswordReveal(title: string, label: string, password: string) {
  copied.value = false
  passwordReveal.value = { title, label, password }
}

async function copyPassword() {
  const pw = passwordReveal.value?.password
  if (!pw) return
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(pw)
    } else {
      fallbackCopy(pw)
    }
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    fallbackCopy(pw)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  }
}

export function PasswordRevealModal() {
  if (!passwordReveal.value) return null
  const { title, label, password } = passwordReveal.value

  return (
    <div class="fixed inset-0 z-[60] flex items-center justify-center bg-black/60">
      <div class="bg-surface-overlay border border-border rounded-xl p-6 max-w-md w-full mx-4">
        <h2 class="text-lg font-bold text-text-primary mb-2">{title}</h2>
        <p class="text-sm text-danger mb-4">Copy this password now. It won't be shown again.</p>

        <label class="font-mono text-[10px] tracking-[2px] text-text-tertiary uppercase block mb-1.5">{label}</label>
        <div class="flex gap-2 mb-6">
          <code class="flex-1 bg-[rgba(255,255,255,0.03)] border border-border rounded-lg px-4 py-2.5 font-mono text-sm text-text-primary break-all select-all">
            {password}
          </code>
          <button
            type="button"
            onClick={copyPassword}
            class={`border font-mono text-xs px-4 py-2.5 rounded-lg shrink-0 transition-colors ${
              copied.value
                ? 'border-green-500/30 text-green-400'
                : 'border-border text-text-secondary hover:border-border-hover'
            }`}
          >
            {copied.value ? 'COPIED' : 'COPY'}
          </button>
        </div>

        <div class="flex justify-end">
          <button
            type="button"
            onClick={() => { passwordReveal.value = null; copied.value = false }}
            class="bg-accent text-surface-base font-mono font-bold text-xs tracking-[1px] px-4 py-2.5 rounded-lg"
          >
            DONE
          </button>
        </div>
      </div>
    </div>
  )
}
