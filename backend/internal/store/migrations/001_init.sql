CREATE TABLE IF NOT EXISTS guild_roles (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  permissions TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  player_name TEXT NOT NULL UNIQUE,
  role_id INTEGER NOT NULL REFERENCES guild_roles(id),
  api_token TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS builds (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  role TEXT NOT NULL CHECK(role IN ('Tank','Healer','DPS','Support')),
  silver_value INTEGER NOT NULL DEFAULT 0,
  created_by INTEGER REFERENCES users(id),
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS build_items (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  build_id INTEGER NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
  slot TEXT NOT NULL,
  item_name TEXT NOT NULL,
  tier INTEGER NOT NULL DEFAULT 7,
  enchantment INTEGER NOT NULL DEFAULT 0,
  quantity INTEGER NOT NULL DEFAULT 1,
  UNIQUE(build_id, slot)
);

CREATE TABLE IF NOT EXISTS regear_requests (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER REFERENCES users(id),
  player_name TEXT NOT NULL,
  request_date TEXT NOT NULL,
  build_id INTEGER NOT NULL REFERENCES builds(id),
  death_screenshot_url TEXT NOT NULL,
  killboard_url TEXT NOT NULL,
  notes TEXT,
  status TEXT NOT NULL DEFAULT 'Pending' CHECK(status IN ('Pending','Approved','Denied','Completed')),
  silver_value INTEGER NOT NULL DEFAULT 0,
  reviewed_by INTEGER REFERENCES users(id),
  reviewed_at TEXT,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS regear_request_items (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  regear_request_id INTEGER NOT NULL REFERENCES regear_requests(id) ON DELETE CASCADE,
  item_name TEXT NOT NULL,
  tier INTEGER NOT NULL,
  enchantment INTEGER NOT NULL,
  quantity_needed INTEGER NOT NULL,
  quantity_fulfilled INTEGER NOT NULL DEFAULT 0,
  quantity_missing INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS inventory (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  item_name TEXT NOT NULL,
  tier INTEGER NOT NULL,
  enchantment INTEGER NOT NULL DEFAULT 0,
  quantity_available INTEGER NOT NULL DEFAULT 0,
  low_stock_threshold INTEGER NOT NULL DEFAULT 5,
  last_updated TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(item_name, tier, enchantment)
);

CREATE TABLE IF NOT EXISTS recipes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  output_item_name TEXT NOT NULL,
  output_tier INTEGER NOT NULL,
  output_enchantment INTEGER NOT NULL DEFAULT 0,
  output_quantity INTEGER NOT NULL DEFAULT 1,
  notes TEXT,
  UNIQUE(output_item_name, output_tier, output_enchantment)
);

CREATE TABLE IF NOT EXISTS recipe_materials (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  recipe_id INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
  material_name TEXT NOT NULL,
  tier INTEGER NOT NULL,
  quantity INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS shopping_lists (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'Open',
  generated_by INTEGER REFERENCES users(id),
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS shopping_list_items (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  shopping_list_id INTEGER NOT NULL REFERENCES shopping_lists(id) ON DELETE CASCADE,
  item_name TEXT NOT NULL,
  tier INTEGER NOT NULL,
  enchantment INTEGER NOT NULL,
  quantity_needed INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  actor_user_id INTEGER REFERENCES users(id),
  action TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id INTEGER,
  details TEXT,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS discord_outbox (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  event_type TEXT NOT NULL,
  payload TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'Pending',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  sent_at TEXT
);
