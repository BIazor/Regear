-- Rename old table
ALTER TABLE shopping_list_items RENAME TO shopping_list_items_old;

-- Create new table
CREATE TABLE shopping_list_items (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  shopping_list_id INTEGER NOT NULL REFERENCES shopping_lists(id) ON DELETE CASCADE,
  item_name TEXT NOT NULL,
  equivalent_tier INTEGER NOT NULL,
  quantity_needed INTEGER NOT NULL
);

-- Copy data, converting tier + enchantment to equivalent_tier
INSERT INTO shopping_list_items (shopping_list_id, item_name, equivalent_tier, quantity_needed)
SELECT shopping_list_id, item_name, (tier + enchantment) AS equivalent_tier, quantity_needed
FROM shopping_list_items_old;

-- Drop old table
DROP TABLE shopping_list_items_old;
