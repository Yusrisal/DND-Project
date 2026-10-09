/**
 * Web Audio API procedural sound synthesizer for VTT sound effects and ambient soundscape.
 * Requires 0 external audio assets or npm packages.
 */

let audioCtx: AudioContext | null = null
let ambientOsc1: OscillatorNode | null = null
let ambientOsc2: OscillatorNode | null = null
let ambientGain: GainNode | null = null

function getAudioContext(): AudioContext | null {
  if (typeof window === 'undefined') return null
  if (!audioCtx) {
    const AudioContextClass =
      window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
    if (AudioContextClass) {
      audioCtx = new AudioContextClass()
    }
  }
  if (audioCtx && audioCtx.state === 'suspended') {
    audioCtx.resume().catch(() => {})
  }
  return audioCtx
}

export function playDiceSound() {
  const ctx = getAudioContext()
  if (!ctx) return

  const now = ctx.currentTime
  for (let i = 0; i < 5; i++) {
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    const time = now + i * 0.04 + Math.random() * 0.02

    osc.type = 'triangle'
    osc.frequency.setValueAtTime(120 + Math.random() * 260, time)
    osc.frequency.exponentialRampToValueAtTime(50, time + 0.05)

    gain.gain.setValueAtTime(0.2, time)
    gain.gain.exponentialRampToValueAtTime(0.001, time + 0.05)

    osc.connect(gain)
    gain.connect(ctx.destination)

    osc.start(time)
    osc.stop(time + 0.06)
  }
}

export function playHitSound() {
  const ctx = getAudioContext()
  if (!ctx) return

  const now = ctx.currentTime
  const osc = ctx.createOscillator()
  const gain = ctx.createGain()

  osc.type = 'sawtooth'
  osc.frequency.setValueAtTime(320, now)
  osc.frequency.exponentialRampToValueAtTime(80, now + 0.12)

  gain.gain.setValueAtTime(0.3, now)
  gain.gain.exponentialRampToValueAtTime(0.001, now + 0.14)

  osc.connect(gain)
  gain.connect(ctx.destination)

  osc.start(now)
  osc.stop(now + 0.15)
}

export function playCritSound() {
  const ctx = getAudioContext()
  if (!ctx) return

  const now = ctx.currentTime
  const notes = [523.25, 659.25, 783.99, 1046.5] // C5, E5, G5, C6 fanfare
  notes.forEach((freq, idx) => {
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    const startTime = now + idx * 0.06

    osc.type = 'triangle'
    osc.frequency.setValueAtTime(freq, startTime)

    gain.gain.setValueAtTime(0.25, startTime)
    gain.gain.exponentialRampToValueAtTime(0.001, startTime + 0.35)

    osc.connect(gain)
    gain.connect(ctx.destination)

    osc.start(startTime)
    osc.stop(startTime + 0.36)
  })
}

export function playMissSound() {
  const ctx = getAudioContext()
  if (!ctx) return

  const now = ctx.currentTime
  const osc = ctx.createOscillator()
  const gain = ctx.createGain()

  osc.type = 'sine'
  osc.frequency.setValueAtTime(140, now)
  osc.frequency.exponentialRampToValueAtTime(40, now + 0.18)

  gain.gain.setValueAtTime(0.15, now)
  gain.gain.exponentialRampToValueAtTime(0.001, now + 0.18)

  osc.connect(gain)
  gain.connect(ctx.destination)

  osc.start(now)
  osc.stop(now + 0.2)
}

export function playHealSound() {
  const ctx = getAudioContext()
  if (!ctx) return

  const now = ctx.currentTime
  const notes = [440, 554.37, 659.25, 880] // A major chime
  notes.forEach((freq, idx) => {
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    const time = now + idx * 0.07

    osc.type = 'sine'
    osc.frequency.setValueAtTime(freq, time)

    gain.gain.setValueAtTime(0.2, time)
    gain.gain.exponentialRampToValueAtTime(0.001, time + 0.4)

    osc.connect(gain)
    gain.connect(ctx.destination)

    osc.start(time)
    osc.stop(time + 0.45)
  })
}

export function playSpellSound() {
  const ctx = getAudioContext()
  if (!ctx) return

  const now = ctx.currentTime
  const osc = ctx.createOscillator()
  const gain = ctx.createGain()

  osc.type = 'sine'
  osc.frequency.setValueAtTime(300, now)
  osc.frequency.exponentialRampToValueAtTime(900, now + 0.2)
  osc.frequency.exponentialRampToValueAtTime(400, now + 0.35)

  gain.gain.setValueAtTime(0.22, now)
  gain.gain.exponentialRampToValueAtTime(0.001, now + 0.38)

  osc.connect(gain)
  gain.connect(ctx.destination)

  osc.start(now)
  osc.stop(now + 0.4)
}

export function playPingSound() {
  const ctx = getAudioContext()
  if (!ctx) return

  const now = ctx.currentTime
  const osc = ctx.createOscillator()
  const gain = ctx.createGain()

  osc.type = 'sine'
  osc.frequency.setValueAtTime(950, now)
  osc.frequency.exponentialRampToValueAtTime(1400, now + 0.08)

  gain.gain.setValueAtTime(0.18, now)
  gain.gain.exponentialRampToValueAtTime(0.001, now + 0.3)

  osc.connect(gain)
  gain.connect(ctx.destination)

  osc.start(now)
  osc.stop(now + 0.32)
}

export function playDoorSound(open: boolean) {
  const ctx = getAudioContext()
  if (!ctx) return

  const now = ctx.currentTime
  const osc = ctx.createOscillator()
  const gain = ctx.createGain()

  osc.type = 'triangle'
  if (open) {
    osc.frequency.setValueAtTime(90, now)
    osc.frequency.linearRampToValueAtTime(160, now + 0.25)
  } else {
    osc.frequency.setValueAtTime(180, now)
    osc.frequency.exponentialRampToValueAtTime(60, now + 0.2)
  }

  gain.gain.setValueAtTime(0.25, now)
  gain.gain.exponentialRampToValueAtTime(0.001, now + 0.3)

  osc.connect(gain)
  gain.connect(ctx.destination)

  osc.start(now)
  osc.stop(now + 0.32)
}

export function playChestSound() {
  const ctx = getAudioContext()
  if (!ctx) return

  const now = ctx.currentTime
  // Metallic latch click then wooden creak
  const osc1 = ctx.createOscillator()
  const gain1 = ctx.createGain()
  osc1.type = 'square'
  osc1.frequency.setValueAtTime(800, now)
  gain1.gain.setValueAtTime(0.2, now)
  gain1.gain.exponentialRampToValueAtTime(0.001, now + 0.05)
  osc1.connect(gain1)
  gain1.connect(ctx.destination)
  osc1.start(now)
  osc1.stop(now + 0.06)

  const osc2 = ctx.createOscillator()
  const gain2 = ctx.createGain()
  osc2.type = 'sawtooth'
  osc2.frequency.setValueAtTime(110, now + 0.06)
  osc2.frequency.linearRampToValueAtTime(240, now + 0.3)
  gain2.gain.setValueAtTime(0.18, now + 0.06)
  gain2.gain.exponentialRampToValueAtTime(0.001, now + 0.35)
  osc2.connect(gain2)
  gain2.connect(ctx.destination)
  osc2.start(now + 0.06)
  osc2.stop(now + 0.36)
}

export function playTrapSound() {
  const ctx = getAudioContext()
  if (!ctx) return

  const now = ctx.currentTime
  const osc = ctx.createOscillator()
  const gain = ctx.createGain()

  osc.type = 'sawtooth'
  osc.frequency.setValueAtTime(600, now)
  osc.frequency.exponentialRampToValueAtTime(70, now + 0.25)

  gain.gain.setValueAtTime(0.35, now)
  gain.gain.exponentialRampToValueAtTime(0.001, now + 0.28)

  osc.connect(gain)
  gain.connect(ctx.destination)

  osc.start(now)
  osc.stop(now + 0.3)
}

export function playThunderSound() {
  const ctx = getAudioContext()
  if (!ctx) return

  const now = ctx.currentTime
  const osc = ctx.createOscillator()
  const gain = ctx.createGain()

  osc.type = 'sawtooth'
  osc.frequency.setValueAtTime(100, now)
  osc.frequency.exponentialRampToValueAtTime(30, now + 0.8)

  gain.gain.setValueAtTime(0.4, now)
  gain.gain.exponentialRampToValueAtTime(0.001, now + 1.2)

  osc.connect(gain)
  gain.connect(ctx.destination)

  osc.start(now)
  osc.stop(now + 1.3)
}

export function playDragonRoar() {
  const ctx = getAudioContext()
  if (!ctx) return

  const now = ctx.currentTime
  const osc1 = ctx.createOscillator()
  const osc2 = ctx.createOscillator()
  const gain = ctx.createGain()

  osc1.type = 'sawtooth'
  osc1.frequency.setValueAtTime(90, now)
  osc1.frequency.linearRampToValueAtTime(180, now + 0.3)
  osc1.frequency.exponentialRampToValueAtTime(45, now + 1.0)

  osc2.type = 'square'
  osc2.frequency.setValueAtTime(65, now)
  osc2.frequency.linearRampToValueAtTime(130, now + 0.3)
  osc2.frequency.exponentialRampToValueAtTime(35, now + 1.0)

  gain.gain.setValueAtTime(0.35, now)
  gain.gain.exponentialRampToValueAtTime(0.001, now + 1.1)

  osc1.connect(gain)
  osc2.connect(gain)
  gain.connect(ctx.destination)

  osc1.start(now)
  osc2.start(now)
  osc1.stop(now + 1.15)
  osc2.stop(now + 1.15)
}

export function setAmbientSound(type: 'none' | 'dungeon' | 'tavern' | 'storm') {
  const ctx = getAudioContext()
  if (!ctx) return

  if (ambientGain) {
    ambientGain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + 0.5)
    setTimeout(() => {
      ambientOsc1?.stop()
      ambientOsc2?.stop()
      ambientOsc1 = null
      ambientOsc2 = null
      ambientGain = null
    }, 550)
  }

  if (type === 'none') return

  const now = ctx.currentTime + 0.1
  const gain = ctx.createGain()
  gain.gain.setValueAtTime(0.001, now)
  gain.gain.linearRampToValueAtTime(0.08, now + 1.0)

  const osc1 = ctx.createOscillator()
  const osc2 = ctx.createOscillator()

  if (type === 'dungeon') {
    osc1.type = 'sine'
    osc1.frequency.setValueAtTime(55, now) // Deep drone (A1)
    osc2.type = 'triangle'
    osc2.frequency.setValueAtTime(82.4, now) // E2
  } else if (type === 'tavern') {
    osc1.type = 'triangle'
    osc1.frequency.setValueAtTime(130.8, now) // C3 warm chord
    osc2.type = 'sine'
    osc2.frequency.setValueAtTime(196, now) // G3
  } else if (type === 'storm') {
    osc1.type = 'sawtooth'
    osc1.frequency.setValueAtTime(45, now)
    osc2.type = 'sine'
    osc2.frequency.setValueAtTime(70, now)
  }

  osc1.connect(gain)
  osc2.connect(gain)
  gain.connect(ctx.destination)

  osc1.start(now)
  osc2.start(now)

  ambientOsc1 = osc1
  ambientOsc2 = osc2
  ambientGain = gain
}
