-- 000019_skill_tree_index.up.sql

-- ListClassSkills filters on class_name and sorts by tier, req_level, name.
-- The existing (class_name, tier) index leaves the rest to a filesort over all
-- 250 rows on every skill tree load.
ALTER TABLE class_skills
    ADD KEY idx_class_skills_tree (class_name, tier, req_level, name);

-- CharacterSkills reads the deck for one character on every sheet load.
ALTER TABLE character_skills
    ADD KEY idx_character_skills_deck (character_id, is_active);
