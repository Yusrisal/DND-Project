import { useState } from 'react'
import {
  playChestSound,
  playCritSound,
  playDiceSound,
  playDoorSound,
  playDragonRoar,
  playHealSound,
  playThunderSound,
  playTrapSound,
  setAmbientSound,
} from '../utils/audio'

export function SoundboardPanel() {
  const [ambient, setAmbient] = useState<'none' | 'dungeon' | 'tavern' | 'storm'>('none')

  const handleAmbient = (type: 'none' | 'dungeon' | 'tavern' | 'storm') => {
    setAmbient(type)
    setAmbientSound(type)
  }

  return (
    <section className="soundboard-panel flex flex-col gap-3">
      <header className="flex items-center justify-between">
        <h2>🔊 Soundboard & Ambience</h2>
      </header>

      {/* Ambient Soundscapes */}
      <div className="card p-3 flex flex-col gap-2">
        <strong className="text-xs text-muted uppercase">Continuous Ambience</strong>
        <div className="grid grid-cols-2 gap-2">
          <button
            className={`btn text-xs py-1.5 ${ambient === 'dungeon' ? 'bg-accent text-white border-accent' : ''}`}
            onClick={() => handleAmbient(ambient === 'dungeon' ? 'none' : 'dungeon')}
          >
            🕯️ Dungeon Drone
          </button>
          <button
            className={`btn text-xs py-1.5 ${ambient === 'tavern' ? 'bg-accent text-white border-accent' : ''}`}
            onClick={() => handleAmbient(ambient === 'tavern' ? 'none' : 'tavern')}
          >
            🍺 Tavern Murmur
          </button>
          <button
            className={`btn text-xs py-1.5 ${ambient === 'storm' ? 'bg-accent text-white border-accent' : ''}`}
            onClick={() => handleAmbient(ambient === 'storm' ? 'none' : 'storm')}
          >
            ⛈️ Storm Winds
          </button>
          <button
            className={`btn text-xs py-1.5 ${ambient === 'none' ? 'opacity-50' : 'text-danger'}`}
            onClick={() => handleAmbient('none')}
          >
            ⏹️ Stop Ambience
          </button>
        </div>
      </div>

      {/* Sound Effects */}
      <div className="card p-3 flex flex-col gap-2">
        <strong className="text-xs text-muted uppercase">1-Click SFX Triggers</strong>
        <div className="grid grid-cols-2 gap-2">
          <button className="btn text-xs py-1.5" onClick={() => playDoorSound(true)}>
            🚪 Open Door
          </button>
          <button className="btn text-xs py-1.5" onClick={() => playDoorSound(false)}>
            🚪 Close Door
          </button>
          <button className="btn text-xs py-1.5" onClick={playChestSound}>
            📦 Open Chest
          </button>
          <button className="btn text-xs py-1.5" onClick={playTrapSound}>
            ⚠️ Trap Trigger
          </button>
          <button className="btn text-xs py-1.5" onClick={playThunderSound}>
            ⚡ Thunderstrike
          </button>
          <button className="btn text-xs py-1.5" onClick={playDragonRoar}>
            🐉 Dragon Roar
          </button>
          <button className="btn text-xs py-1.5" onClick={playCritSound}>
            🎺 Crit Fanfare
          </button>
          <button className="btn text-xs py-1.5" onClick={playHealSound}>
            ✨ Heal Chimes
          </button>
          <button className="btn text-xs py-1.5 col-span-2" onClick={playDiceSound}>
            🎲 Dice Clatter
          </button>
        </div>
      </div>
    </section>
  )
}
