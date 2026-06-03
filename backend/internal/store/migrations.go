package store

import (
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

const postgresSchema = `
CREATE TABLE IF NOT EXISTS guild_roles (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) NOT NULL UNIQUE,
  permissions TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  player_name VARCHAR(100) NOT NULL UNIQUE,
  role_id INTEGER NOT NULL REFERENCES guild_roles(id),
  api_token VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL DEFAULT '123',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS builds (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE,
  role VARCHAR(50) NOT NULL CHECK(role IN ('Tank','Healer','DPS','Support')),
  silver_value BIGINT NOT NULL DEFAULT 0,
  screenshot_url VARCHAR(512) NOT NULL DEFAULT '',
  created_by INTEGER REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS build_items (
  id SERIAL PRIMARY KEY,
  build_id INTEGER NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
  slot VARCHAR(50) NOT NULL,
  item_name VARCHAR(100) NOT NULL,
  tier INTEGER NOT NULL DEFAULT 7,
  enchantment INTEGER NOT NULL DEFAULT 0,
  quantity INTEGER NOT NULL DEFAULT 1,
  UNIQUE(build_id, slot)
);

CREATE TABLE IF NOT EXISTS regear_requests (
  id SERIAL PRIMARY KEY,
  user_id INTEGER REFERENCES users(id),
  player_name VARCHAR(100) NOT NULL,
  request_date VARCHAR(50) NOT NULL,
  build_id INTEGER NOT NULL REFERENCES builds(id),
  death_screenshot_url VARCHAR(512) NOT NULL,
  vod_url VARCHAR(512) NOT NULL,
  notes TEXT,
  status VARCHAR(50) NOT NULL DEFAULT 'Pending' CHECK(status IN ('Pending','Approved','Denied','Completed')),
  silver_value BIGINT NOT NULL DEFAULT 0,
  reviewed_by INTEGER REFERENCES users(id),
  reviewed_at TIMESTAMP,
  pickup_location VARCHAR(255) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS regear_request_items (
  id SERIAL PRIMARY KEY,
  regear_request_id INTEGER NOT NULL REFERENCES regear_requests(id) ON DELETE CASCADE,
  item_name VARCHAR(100) NOT NULL,
  tier INTEGER NOT NULL,
  enchantment INTEGER NOT NULL,
  quantity_needed INTEGER NOT NULL,
  quantity_fulfilled INTEGER NOT NULL DEFAULT 0,
  quantity_missing INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS inventory (
  id SERIAL PRIMARY KEY,
  item_name VARCHAR(100) NOT NULL,
  equivalent_tier INTEGER NOT NULL,
  quantity_available INTEGER NOT NULL DEFAULT 0,
  low_stock_threshold INTEGER NOT NULL DEFAULT 5,
  last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(item_name, equivalent_tier)
);

CREATE TABLE IF NOT EXISTS recipes (
  id SERIAL PRIMARY KEY,
  output_item_name VARCHAR(100) NOT NULL,
  output_tier INTEGER NOT NULL,
  output_enchantment INTEGER NOT NULL DEFAULT 0,
  output_quantity INTEGER NOT NULL DEFAULT 1,
  notes TEXT,
  UNIQUE(output_item_name, output_tier, output_enchantment)
);

CREATE TABLE IF NOT EXISTS recipe_materials (
  id SERIAL PRIMARY KEY,
  recipe_id INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
  material_name VARCHAR(100) NOT NULL,
  tier INTEGER NOT NULL,
  quantity INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS shopping_lists (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'Open',
  generated_by INTEGER REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS shopping_list_items (
  id SERIAL PRIMARY KEY,
  shopping_list_id INTEGER NOT NULL REFERENCES shopping_lists(id) ON DELETE CASCADE,
  item_name VARCHAR(100) NOT NULL,
  equivalent_tier INTEGER NOT NULL,
  quantity_needed INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_logs (
  id SERIAL PRIMARY KEY,
  actor_user_id INTEGER REFERENCES users(id),
  action VARCHAR(50) NOT NULL,
  entity_type VARCHAR(50) NOT NULL,
  entity_id INTEGER,
  details TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS discord_outbox (
  id SERIAL PRIMARY KEY,
  event_type VARCHAR(50) NOT NULL,
  payload TEXT NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'Pending',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  sent_at TIMESTAMP
);

-- Seed initial roles and owner account
INSERT INTO guild_roles (id, name, permissions) VALUES
  (1, 'Member', 'submit_regear,view_own_regears'),
  (2, 'Officer', 'review_regears,manage_inventory,generate_shopping_lists'),
  (3, 'Admin', 'full_access'),
  (4, 'Owner', 'full_access')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('guild_roles', 'id'), COALESCE(max(id), 1)) FROM guild_roles;

INSERT INTO users (id, player_name, role_id, api_token, password) VALUES
  (1, 'Blazor', 4, 'blazor-admin-token', '123')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('users', 'id'), COALESCE(max(id), 1)) FROM users;
`

func Migrate(db *sql.DB) error {
	var version string
	isPostgres := false
	if err := db.QueryRow("SELECT version()").Scan(&version); err == nil {
		isPostgres = true
	}

	if isPostgres {
		return migratePostgres(db)
	}

	// SQLite migrations
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		return err
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, entry.Name()).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		sqlBytes, err := migrationFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s: %w", entry.Name(), err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, entry.Name()); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func migratePostgres(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version VARCHAR(100) PRIMARY KEY, applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		return err
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = 'postgres_init'`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(postgresSchema); err != nil {
		return fmt.Errorf("postgres_init migration failed: %w", err)
	}

	if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES ('postgres_init')`); err != nil {
		return err
	}

	return tx.Commit()
}
