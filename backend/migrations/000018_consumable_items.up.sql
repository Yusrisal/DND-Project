-- 000018_consumable_items.up.sql

-- Consumable items: Health Potion, Mana Potion, Portal Scroll, and 1-use combat spell scrolls
INSERT INTO compendium_items (name, category, rarity, cost_cp, weight, attack_bonus, damage, damage_type, description, req_level, req_class, req_str, req_dex, req_con, req_int, req_wis, req_cha, teaches_skill_id, set_name, set_piece, set_bonus_2pc, set_bonus_4pc) VALUES
-- Potions
('Lesser Health Potion', 'potion', 'common', 500, 0.5, 0, '', '', 'Consumable potion. Restores 2d4+2 Hit Points immediately when used.', 1, '', 0, 0, 0, 0, 0, 0, NULL, '', '', '', ''),
('Health Potion', 'potion', 'common', 1500, 0.5, 0, '', '', 'Consumable potion. Restores 4d4+4 Hit Points immediately when used.', 5, '', 0, 0, 0, 0, 0, 0, NULL, '', '', '', ''),
('Greater Health Potion', 'potion', 'uncommon', 4000, 0.5, 0, '', '', 'Consumable potion. Restores 8d4+8 Hit Points immediately when used.', 15, '', 0, 0, 0, 0, 0, 0, NULL, '', '', '', ''),
('Superior Health Potion', 'potion', 'rare', 12000, 0.5, 0, '', '', 'Consumable potion. Restores 10d4+20 Hit Points immediately when used.', 30, '', 0, 0, 0, 0, 0, 0, NULL, '', '', '', ''),
('Supreme Health Potion', 'potion', 'very rare', 30000, 0.5, 0, '', '', 'Consumable potion. Restores 20d4+40 Hit Points immediately when used.', 45, '', 0, 0, 0, 0, 0, 0, NULL, '', '', '', ''),
('Elixir of Restoration', 'potion', 'rare', 10000, 0.5, 0, '', '', 'Consumable potion. Removes harmful conditions and heals 25 Hit Points.', 10, '', 0, 0, 0, 0, 0, 0, NULL, '', '', '', ''),

-- Portal & Escape Scrolls
('Town Portal Scroll', 'scroll', 'common', 1000, 0.1, 0, '', '', 'Consumable scroll. Opens a magical rift returning party to safety or resting area.', 1, '', 0, 0, 0, 0, 0, 0, NULL, '', '', '', ''),
('Dungeon Escape Scroll', 'scroll', 'uncommon', 3000, 0.1, 0, '', '', 'Consumable scroll. Instantly escapes dangerous encounters back to refuge.', 10, '', 0, 0, 0, 0, 0, 0, NULL, '', '', '', ''),

-- 1-Use Offensive & Utility Spell Scrolls
('Scroll of Fireball (1-use)', 'scroll', 'uncommon', 5000, 0.1, 0, '8d6', 'fire', '1-use scroll. Ignites an explosive fiery burst dealing 8d6 fire damage.', 5, '', 0, 0, 0, 0, 0, 0, NULL, '', '', '', ''),
('Scroll of Lightning Bolt (1-use)', 'scroll', 'uncommon', 5000, 0.1, 0, '8d6', 'lightning', '1-use scroll. Unleases a searing stroke of lightning dealing 8d6 lightning damage.', 5, '', 0, 0, 0, 0, 0, 0, NULL, '', '', '', ''),
('Scroll of Ice Storm (1-use)', 'scroll', 'rare', 15000, 0.1, 0, '2d8+4d6', 'cold', '1-use scroll. Hailstones batter the area dealing 2d8 bludgeoning and 4d6 cold damage.', 15, '', 0, 0, 0, 0, 0, 0, NULL, '', '', '', ''),
('Scroll of Chain Lightning (1-use)', 'scroll', 'very rare', 40000, 0.1, 0, '10d8', 'lightning', '1-use scroll. Arc of lightning leaping between enemies for 10d8 lightning damage.', 30, '', 0, 0, 0, 0, 0, 0, NULL, '', '', '', ''),
('Scroll of Meteor Swarm (1-use)', 'scroll', 'legendary', 100000, 0.1, 0, '20d6+20d6', 'fire', '1-use scroll. Blazing orbs plunge from sky dealing apocalyptic fire damage.', 45, '', 0, 0, 0, 0, 0, 0, NULL, '', '', '', '');
