INSERT OR IGNORE INTO guild_roles (id, name, permissions) VALUES
  (1, 'Member', 'submit_regear,view_own_regears'),
  (2, 'Officer', 'review_regears,manage_inventory,generate_shopping_lists'),
  (3, 'Admin', 'full_access');

INSERT OR IGNORE INTO users (id, player_name, role_id, api_token) VALUES
  (1, 'AstraAdmin', 3, 'demo-admin'),
  (2, 'Brimstone', 2, 'demo-officer'),
  (3, 'Ironclad', 1, 'demo-member');

INSERT OR IGNORE INTO builds (id, name, role, silver_value, created_by) VALUES
  (1, 'Tank ZvZ', 'Tank', 1800000, 1),
  (2, 'Holy ZvZ', 'Healer', 1450000, 1),
  (3, 'Clap DPS', 'DPS', 2100000, 2),
  (4, 'Arcane Support', 'Support', 1600000, 2);

INSERT OR IGNORE INTO build_items (build_id, slot, item_name, tier, enchantment, quantity) VALUES
  (1, 'Main Hand', 'Incubus Mace', 7, 1, 1),
  (1, 'Off Hand', 'Sarcophagus', 7, 1, 1),
  (1, 'Helmet', 'Guardian Helmet', 7, 1, 1),
  (1, 'Armor', 'Judicator Armor', 7, 1, 1),
  (1, 'Shoes', 'Hunter Shoes', 7, 1, 1),
  (1, 'Cape', 'Martlock Cape', 7, 1, 1),
  (1, 'Food', 'Pork Omelette', 7, 0, 1),
  (1, 'Potion', 'Resistance Potion', 7, 0, 1),
  (2, 'Main Hand', 'Fallen Staff', 7, 1, 1),
  (2, 'Off Hand', 'Mistcaller', 7, 1, 1),
  (2, 'Helmet', 'Cleric Cowl', 7, 1, 1),
  (2, 'Armor', 'Cleric Robe', 7, 1, 1),
  (2, 'Shoes', 'Scholar Sandals', 7, 1, 1),
  (2, 'Cape', 'Lymhurst Cape', 7, 1, 1),
  (2, 'Food', 'Pork Omelette', 7, 0, 1),
  (2, 'Potion', 'Resistance Potion', 7, 0, 1),
  (3, 'Main Hand', 'Brimstone Staff', 7, 1, 1),
  (3, 'Off Hand', 'Tome of Spells', 7, 1, 1),
  (3, 'Helmet', 'Royal Hood', 7, 1, 1),
  (3, 'Armor', 'Scholar Robe', 7, 1, 1),
  (3, 'Shoes', 'Cleric Sandals', 7, 1, 1),
  (3, 'Cape', 'Thetford Cape', 7, 1, 1),
  (3, 'Food', 'Beef Stew', 7, 0, 1),
  (3, 'Potion', 'Gigantify Potion', 7, 0, 1),
  (4, 'Main Hand', 'Enigmatic Staff', 7, 1, 1),
  (4, 'Off Hand', 'Mistcaller', 7, 1, 1),
  (4, 'Helmet', 'Judicator Helmet', 7, 1, 1),
  (4, 'Armor', 'Knight Armor', 7, 1, 1),
  (4, 'Shoes', 'Scholar Sandals', 7, 1, 1),
  (4, 'Cape', 'Fort Sterling Cape', 7, 1, 1),
  (4, 'Food', 'Pork Omelette', 7, 0, 1),
  (4, 'Potion', 'Resistance Potion', 7, 0, 1);

INSERT OR IGNORE INTO inventory (item_name, tier, enchantment, quantity_available, low_stock_threshold) VALUES
  ('Incubus Mace', 7, 1, 3, 5),
  ('Sarcophagus', 7, 1, 2, 5),
  ('Guardian Helmet', 7, 1, 8, 5),
  ('Judicator Armor', 7, 1, 1, 5),
  ('Hunter Shoes', 7, 1, 6, 5),
  ('Martlock Cape', 7, 1, 0, 5),
  ('Fallen Staff', 7, 1, 4, 5),
  ('Lymhurst Cape', 7, 1, 3, 5),
  ('Brimstone Staff', 7, 1, 1, 5),
  ('Thetford Cape', 7, 1, 2, 5),
  ('Pork Omelette', 7, 0, 40, 20),
  ('Resistance Potion', 7, 0, 35, 20);

INSERT OR IGNORE INTO recipes (id, output_item_name, output_tier, output_enchantment, output_quantity, notes) VALUES
  (1, 'Judicator Armor', 7, 1, 1, 'Artifact armor recipe placeholder'),
  (2, 'Incubus Mace', 7, 1, 1, 'Artifact weapon recipe placeholder');

INSERT OR IGNORE INTO recipe_materials (recipe_id, material_name, tier, quantity) VALUES
  (1, 'Metal Bars', 7, 12),
  (1, 'Leather', 7, 8),
  (1, 'Artifact Components', 7, 2),
  (2, 'Metal Bars', 7, 16),
  (2, 'Artifact Components', 7, 3);
