import type { Character, CharacterLearnedSkill } from '../api/types'
import { playSpellSound } from '../utils/audio'

type Props = {
  character: Character | null
  activeSkills?: CharacterLearnedSkill[]
  onSendChat: (message: string) => void
  onRoll: (expression: string, label: string, characterId: number | null) => void
}

export function ActionBar({ character, activeSkills = [], onSendChat, onRoll }: Props) {
  if (!character) return null

  const charName = character.name
  const dexMod = character.modifiers?.dexterity ?? 0
  const wisMod = character.modifiers?.wisdom ?? 0

  const handleDash = () => {
    playSpellSound()
    const normal = character.speed || 30
    onSendChat(`⚡ **${charName}** uses **Dash Action**! Movement speed doubled from ${normal}ft to **${normal * 2}ft** this turn.`)
  }

  const handleDodge = () => {
    playSpellSound()
    onSendChat(`🛡️ **${charName}** takes the **Dodge Action**! Attacks against ${charName} have Disadvantage, and ${charName} gains Advantage on DEX saves until next turn.`)
  }

  const handleDisengage = () => {
    playSpellSound()
    onSendChat(`💨 **${charName}** uses **Disengage Action**! Movement will not provoke Opportunity Attacks this turn.`)
  }

  const handleHide = () => {
    const modStr = dexMod >= 0 ? `+${dexMod}` : String(dexMod)
    onRoll(`1d20${modStr}`, `Stealth Check (Hide) - ${charName}`, character.id)
  }

  const handleSearch = () => {
    const modStr = wisMod >= 0 ? `+${wisMod}` : String(wisMod)
    onRoll(`1d20${modStr}`, `Perception Check (Search) - ${charName}`, character.id)
  }

  const handleHelp = () => {
    const target = prompt('Who are you helping?', 'ally')
    if (!target) return
    onSendChat(`🤝 **${charName}** takes the **Help Action** to assist **${target}**! ${target}'s next attack or ability check has Advantage.`)
  }

  const handleReady = () => {
    const trigger = prompt('What is your readied trigger and action?', 'If enemy comes in range, attack')
    if (!trigger) return
    onSendChat(`⏳ **${charName}** readies an action: "*${trigger}*" (Uses Reaction when triggered).`)
  }

  return (
    <div className="action-bar">
      <span className="action-bar-label">5e Actions:</span>
      <button className="action-btn" title="Double movement speed" onClick={handleDash}>
        ⚡ Dash
      </button>
      <button className="action-btn" title="Attacks against you have disadvantage" onClick={handleDodge}>
        🛡️ Dodge
      </button>
      <button className="action-btn" title="Move without opportunity attacks" onClick={handleDisengage}>
        💨 Disengage
      </button>
      <button className="action-btn" title="Roll Stealth to hide" onClick={handleHide}>
        👁️ Hide
      </button>
      <button className="action-btn" title="Roll Perception to find hidden things" onClick={handleSearch}>
        🔍 Search
      </button>
      <button className="action-btn" title="Give ally advantage" onClick={handleHelp}>
        🤝 Help
      </button>
      <button className="action-btn" title="Ready an action on trigger" onClick={handleReady}>
        ⏳ Ready
      </button>

      {/* 10 Active Combat Skills Deck */}
      {activeSkills.filter((s) => s.is_active).map((learned) => (
        <button
          key={learned.skill.id}
          className="action-btn border-[#66fcf1] text-[#66fcf1]"
          title={`${learned.skill.name}: ${learned.skill.description}`}
          onClick={() => {
            playSpellSound()
            let msg = `✨ **${charName}** casts **${learned.skill.name}**!`
            if (learned.skill.damage) {
              msg += ` (Deals ${learned.skill.damage} ${learned.skill.damage_type})`
            }
            onSendChat(msg)
            if (learned.skill.damage && learned.skill.damage.includes('d')) {
              onRoll(learned.skill.damage, `${learned.skill.name} Damage`, character.id)
            }
          }}
        >
          🔮 {learned.skill.name}
        </button>
      ))}
    </div>
  )
}
