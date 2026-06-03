-- Rename old table
ALTER TABLE inventory RENAME TO inventory_old;

-- Create new table
CREATE TABLE inventory (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  item_name TEXT NOT NULL,
  equivalent_tier INTEGER NOT NULL,
  quantity_available INTEGER NOT NULL DEFAULT 0,
  low_stock_threshold INTEGER NOT NULL DEFAULT 5,
  last_updated TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(item_name, equivalent_tier)
);

-- Copy data, converting tier + enchantment to equivalent_tier
INSERT INTO inventory (item_name, equivalent_tier, quantity_available, low_stock_threshold, last_updated)
SELECT item_name, (tier + enchantment) AS equivalent_tier, SUM(quantity_available), MAX(low_stock_threshold), MAX(last_updated)
FROM inventory_old
GROUP BY item_name, equivalent_tier;

-- Drop old table
DROP TABLE inventory_old;
