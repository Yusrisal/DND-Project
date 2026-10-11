-- 000018_consumable_items.down.sql

DELETE FROM compendium_items WHERE name IN (
    'Lesser Health Potion',
    'Health Potion',
    'Greater Health Potion',
    'Superior Health Potion',
    'Supreme Health Potion',
    'Elixir of Restoration',
    'Town Portal Scroll',
    'Dungeon Escape Scroll',
    'Scroll of Fireball (1-use)',
    'Scroll of Lightning Bolt (1-use)',
    'Scroll of Ice Storm (1-use)',
    'Scroll of Chain Lightning (1-use)',
    'Scroll of Meteor Swarm (1-use)'
);
