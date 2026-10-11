import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { CharacterLearnedSkill, CharacterSheet, ClassSkill } from '../api/types'

type Props = {
  sheet: CharacterSheet
  onRefresh: () => void
  onError: (error: unknown) => void
}

export function SkillTreeTab({ sheet, onRefresh, onError }: Props) {
  const { character, skills_loaded = [] } = sheet
  const [treeSkills, setTreeSkills] = useState<ClassSkill[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setLoading(true)
    api
      .listSkillTrees(character.class)
      .then((skills) => setTreeSkills(skills))
      .catch(onError)
      .finally(() => setLoading(false))
  }, [character.class, onError])

  const handleLearn = async (skillId: number) => {
    try {
      await api.learnSkill(character.id, skillId)
      onRefresh()
    } catch (err) {
      onError(err)
    }
  }

  const handleToggle = async (skillId: number, currentlyActive: boolean) => {
    try {
      await api.toggleSkillActive(character.id, skillId, !currentlyActive)
      onRefresh()
    } catch (err) {
      onError(err)
    }
  }

  const learnedMap = new Map<number, CharacterLearnedSkill>()
  skills_loaded.forEach((ls) => learnedMap.set(ls.skill_id, ls))

  const activeCount = skills_loaded.filter((s) => s.is_active).length

  return (
    <div className="space-y-4 text-xs">
      {/* Skill Points & Active Deck Status */}
      <div className="flex items-center justify-between p-2.5 bg-[#12151e] border border-[#2b3145] rounded">
        <div>
          <span className="font-bold text-[#66fcf1] uppercase">Skill Points: </span>
          <span className="font-extrabold text-white text-sm">{character.skill_points}</span>
        </div>
        <div>
          <span className="font-bold text-slate-300">Combat Deck: </span>
          <span className={`font-bold ${activeCount >= 10 ? 'text-amber-400' : 'text-[#2ed573]'}`}>
            {activeCount}/10 Active
          </span>
        </div>
      </div>

      {loading && <p className="text-slate-500 italic">Loading skill tree…</p>}

      {!loading && treeSkills.length === 0 && (
        <div className="p-3 bg-[#161a24] border border-[#262c3b] rounded text-slate-400 text-center">
          No skill tree available for class <strong>{character.class || 'Unknown'}</strong>.
          <p className="text-[10px] text-slate-500 mt-1">
            Choose Mage, Assassin, Warrior, Archer, or Cleric to unlock dedicated skill trees.
          </p>
        </div>
      )}

      {/* Tiers Display */}
      {[1, 2, 3, 4].map((tier) => {
        const tierSkills = treeSkills.filter((s) => s.tier === tier)
        if (tierSkills.length === 0) return null

        return (
          <div key={tier} className="space-y-1.5">
            <div className="text-[10px] font-bold uppercase tracking-wider text-slate-400 border-b border-[#232733] pb-1 flex justify-between">
              <span>{tier === 4 ? 'Tier IV (Ultimate)' : `Tier ${tier}`}</span>
              <span className="text-[9px] text-slate-500">Req Lv {tierSkills[0].req_level}+</span>
            </div>

            <div className="grid grid-cols-1 gap-2">
              {tierSkills.map((skill) => {
                const learned = learnedMap.get(skill.id)
                const isLearned = Boolean(learned)
                const isActive = learned?.is_active ?? false
                const meetsLevel = character.level >= skill.req_level
                const parentLearned = !skill.parent_skill_id || learnedMap.has(skill.parent_skill_id)
                const canLearn = !isLearned && meetsLevel && parentLearned && character.skill_points > 0

                return (
                  <div
                    key={skill.id}
                    className={`p-2.5 rounded border transition-all ${
                      isLearned
                        ? isActive
                          ? 'bg-[#152326] border-[#66fcf1]/60'
                          : 'bg-[#181c26] border-[#363f54]'
                        : 'bg-[#0f1118] border-[#202533] opacity-80'
                    }`}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div>
                        <div className="font-bold text-white flex items-center gap-1.5">
                          <span>{skill.name}</span>
                          {isLearned && (
                            <span className="text-[9px] px-1 py-0.2 bg-[#1f5156] text-[#66fcf1] rounded uppercase">
                              Learned
                            </span>
                          )}
                        </div>
                        <div className="text-[10px] text-slate-400 mt-0.5">
                          {skill.cost_resource && <span className="mr-2">⏳ {skill.cost_resource}</span>}
                          {skill.damage && (
                            <span className="text-amber-300 font-semibold mr-2">
                              ⚔️ {skill.damage} {skill.damage_type}
                            </span>
                          )}
                          {skill.cooldown_rounds > 0 && <span>Cooldown: {skill.cooldown_rounds}r</span>}
                        </div>
                      </div>

                      <div className="flex flex-col gap-1 shrink-0">
                        {!isLearned ? (
                          <button
                            disabled={!canLearn}
                            onClick={() => handleLearn(skill.id)}
                            className={`px-2 py-1 text-[10px] font-bold uppercase rounded border transition-colors ${
                              canLearn
                                ? 'bg-[#1f5156] border-[#66fcf1] text-[#66fcf1] hover:bg-[#45a29e] hover:text-black cursor-pointer'
                                : 'bg-[#191d26] border-[#2b3145] text-slate-500 cursor-not-allowed'
                            }`}
                          >
                            Learn (1 SP)
                          </button>
                        ) : (
                          <button
                            onClick={() => handleToggle(skill.id, isActive)}
                            className={`px-2 py-0.5 text-[10px] font-bold uppercase rounded border transition-colors ${
                              isActive
                                ? 'bg-[#1e4a38] border-[#2ed573] text-[#2ed573] hover:bg-red-950 hover:text-red-300 hover:border-red-600'
                                : 'bg-[#1e2333] border-[#3a445d] text-slate-300 hover:border-[#66fcf1] hover:text-white'
                            }`}
                          >
                            {isActive ? 'Active (Slot)' : 'Equip (Deck)'}
                          </button>
                        )}
                      </div>
                    </div>

                    <p className="text-[11px] text-slate-300 mt-2 mb-0 leading-relaxed">
                      {skill.description}
                    </p>

                    {!meetsLevel && !isLearned && (
                      <div className="text-[9px] text-amber-500 font-medium mt-1">
                        Locked: Requires character Level {skill.req_level}
                      </div>
                    )}
                    {meetsLevel && !parentLearned && !isLearned && (
                      <div className="text-[9px] text-amber-500 font-medium mt-1">
                        Locked: Unlock prerequisite skill first
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          </div>
        )
      })}
    </div>
  )
}
