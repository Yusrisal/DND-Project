-- 000021_core_conditions.up.sql
--
-- The fifteen conditions the rules define, as global rows (campaign_id NULL) so
-- every campaign sees them and the DM can add their own alongside. Ids 1-15 are
-- reserved for them, which is what the down migration removes.
--
-- Durations are 0: a condition lasts until something ends it, unlike a buff.
INSERT INTO status_effects (id, campaign_id, name, kind, default_duration_rounds, color, effect_description) VALUES
(1, NULL, 'Blinded', 'condition', 0, '#2c3e50', 'You cannot see, and automatically fail any check that needs sight. Attacks against you have advantage, and your own attacks have disadvantage.'),
(2, NULL, 'Charmed', 'condition', 0, '#e056fd', 'You cannot attack whoever charmed you. They have advantage on social checks against you.'),
(3, NULL, 'Deafened', 'condition', 0, '#7f8c8d', 'You cannot hear, and automatically fail any check that needs hearing.'),
(4, NULL, 'Exhaustion', 'condition', 0, '#6e2c00', 'Stacks in six levels, each worse than the last: disadvantage on checks, then half speed, then disadvantage on attacks and saves, then half hit points, then speed 0, then death.'),
(5, NULL, 'Frightened', 'condition', 0, '#5b2c6f', 'Disadvantage on checks and attacks while the thing scaring you is in sight, and you cannot willingly move closer to it.'),
(6, NULL, 'Grappled', 'condition', 0, '#b9770e', 'Your speed is 0. It ends if the grappler is incapacitated or you are moved out of reach.'),
(7, NULL, 'Incapacitated', 'condition', 0, '#95a5a6', 'You can take no actions and no reactions.'),
(8, NULL, 'Invisible', 'condition', 0, '#3498db', 'You cannot be seen without special senses. Your attacks have advantage, attacks against you have disadvantage.'),
(9, NULL, 'Paralyzed', 'condition', 0, '#8e44ad', 'Incapacitated, cannot move or speak, and you fail Strength and Dexterity saves automatically. Attacks against you have advantage, and any hit from within 5 feet is a critical.'),
(10, NULL, 'Petrified', 'condition', 0, '#616a6b', 'Turned to stone. Incapacitated, unaware of your surroundings, resistant to all damage, and immune to poison and disease.'),
(11, NULL, 'Poisoned', 'condition', 0, '#27ae60', 'Disadvantage on attack rolls and ability checks.'),
(12, NULL, 'Prone', 'condition', 0, '#a04000', 'You can only crawl, and it costs extra movement to stand. Disadvantage on your attacks. Attacks against you from within 5 feet have advantage, from further away disadvantage.'),
(13, NULL, 'Restrained', 'condition', 0, '#c0392b', 'Speed 0, disadvantage on your attacks and on Dexterity saves. Attacks against you have advantage.'),
(14, NULL, 'Stunned', 'condition', 0, '#f39c12', 'Incapacitated, cannot move, and can barely speak. You fail Strength and Dexterity saves automatically, and attacks against you have advantage.'),
(15, NULL, 'Unconscious', 'condition', 0, '#17202a', 'Incapacitated, unaware, you drop what you hold and fall prone. You fail Strength and Dexterity saves automatically, attacks against you have advantage, and any hit from within 5 feet is a critical.');
