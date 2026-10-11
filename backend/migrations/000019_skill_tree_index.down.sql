-- 000019_skill_tree_index.down.sql

ALTER TABLE character_skills
    DROP KEY idx_character_skills_deck;

ALTER TABLE class_skills
    DROP KEY idx_class_skills_tree;
