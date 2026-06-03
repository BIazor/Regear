package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

type queryable interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

type Store struct {
	rawDB      *sql.DB
	db         queryable
	isPostgres bool
}

type wrappedDB struct {
	q queryable
	s *Store
}

func (w wrappedDB) Query(query string, args ...any) (*sql.Rows, error) {
	return w.q.Query(w.s.query(query), args...)
}

func (w wrappedDB) QueryRow(query string, args ...any) *sql.Row {
	return w.q.QueryRow(w.s.query(query), args...)
}

func (w wrappedDB) Exec(query string, args ...any) (sql.Result, error) {
	return w.q.Exec(w.s.query(query), args...)
}

type wrappedTx struct {
	*sql.Tx
	s *Store
}

func (w wrappedTx) Query(query string, args ...any) (*sql.Rows, error) {
	return w.Tx.Query(w.s.query(query), args...)
}

func (w wrappedTx) QueryRow(query string, args ...any) *sql.Row {
	return w.Tx.QueryRow(w.s.query(query), args...)
}

func (w wrappedTx) Exec(query string, args ...any) (sql.Result, error) {
	return w.Tx.Exec(w.s.query(query), args...)
}

func (s *Store) query(q string) string {
	if s.isPostgres {
		var sb strings.Builder
		paramIndex := 1
		for i := 0; i < len(q); i++ {
			c := q[i]
			if c == '?' {
				sb.WriteString(fmt.Sprintf("$%d", paramIndex))
				paramIndex++
			} else {
				sb.WriteByte(c)
			}
		}
		return sb.String()
	}
	return q
}

func Open(dataSourceName string) (*sql.DB, error) {
	if strings.HasPrefix(dataSourceName, "postgres://") || strings.HasPrefix(dataSourceName, "postgresql://") {
		db, err := sql.Open("postgres", dataSourceName)
		if err != nil {
			return nil, err
		}
		db.SetMaxOpenConns(20)
		db.SetMaxIdleConns(5)
		return db, db.Ping()
	}

	db, err := sql.Open("sqlite", dataSourceName+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	return db, db.Ping()
}

func New(db *sql.DB) *Store {
	var isPostgres bool
	var version string
	if err := db.QueryRow("SELECT version()").Scan(&version); err == nil {
		isPostgres = true
	}
	s := &Store{rawDB: db, isPostgres: isPostgres}
	s.db = wrappedDB{q: db, s: s}
	return s
}

type User struct {
	ID         int64  `json:"id"`
	PlayerName string `json:"playerName"`
	Role       string `json:"role"`
	APIToken   string `json:"apiToken,omitempty"`
}

type BuildItem struct {
	ID          int64  `json:"id"`
	Slot        string `json:"slot"`
	ItemName    string `json:"itemName"`
	Tier        int    `json:"tier"`
	Enchantment int    `json:"enchantment"`
	Quantity    int    `json:"quantity"`
}

type Build struct {
	ID            int64       `json:"id"`
	Name          string      `json:"name"`
	Role          string      `json:"role"`
	SilverValue   int64       `json:"silverValue"`
	ScreenshotURL string      `json:"screenshotUrl"`
	Items         []BuildItem `json:"items"`
}

type RegearRequest struct {
	ID                 int64               `json:"id"`
	UserID             int64               `json:"userId"`
	PlayerName         string              `json:"playerName"`
	RequestDate        string              `json:"requestDate"`
	BuildID            int64               `json:"buildId"`
	BuildName          string              `json:"buildName"`
	DeathScreenshotURL string              `json:"deathScreenshotUrl"`
	VodURL             string              `json:"vodUrl"`
	Notes              string              `json:"notes"`
	Status             string              `json:"status"`
	SilverValue        int64               `json:"silverValue"`
	PickupLocation     string              `json:"pickupLocation"`
	Items              []RegearRequestItem `json:"items,omitempty"`
	CreatedAt          string              `json:"createdAt"`
}

type RegearRequestItem struct {
	ItemName          string `json:"itemName"`
	Tier              int    `json:"tier"`
	Enchantment       int    `json:"enchantment"`
	QuantityNeeded    int    `json:"quantityNeeded"`
	QuantityFulfilled int    `json:"quantityFulfilled"`
	QuantityMissing   int    `json:"quantityMissing"`
}

type InventoryItem struct {
	ID                int64  `json:"id"`
	ItemName          string `json:"itemName"`
	EquivalentTier    int    `json:"equivalentTier"`
	QuantityAvailable int    `json:"quantityAvailable"`
	LowStockThreshold int    `json:"lowStockThreshold"`
	LastUpdated       string `json:"lastUpdated"`
}

type ShoppingList struct {
	ID        int64              `json:"id"`
	Name      string             `json:"name"`
	Status    string             `json:"status"`
	CreatedAt string             `json:"createdAt"`
	Items     []ShoppingListItem `json:"items"`
}

type ShoppingListItem struct {
	ItemName       string `json:"itemName"`
	EquivalentTier int    `json:"equivalentTier"`
	QuantityNeeded int    `json:"quantityNeeded"`
}


type Dashboard struct {
	PendingRegears       int64              `json:"pendingRegears"`
	ApprovedRegears      int64              `json:"approvedRegears"`
	DeniedRegears        int64              `json:"deniedRegears"`
	PendingSilverValue   int64              `json:"pendingSilverValue"`
	MostRequestedItems   []ShoppingListItem `json:"mostRequestedItems"`
	LowStockItems        []InventoryItem    `json:"lowStockItems"`
	RecentRegears        []RegearRequest    `json:"recentRegears"`
	TotalInventoryItems  int64              `json:"totalInventoryItems"`
	OpenShortageQuantity int64              `json:"openShortageQuantity"`
}

type MemberHistory struct {
	ID                int64  `json:"id"`
	PlayerName        string `json:"playerName"`
	Role              string `json:"role"`
	Requested         int64  `json:"requested"`
	Approved          int64  `json:"approved"`
	SilverValue       int64  `json:"silverValue"`
	LastRequestStatus string `json:"lastRequestStatus"`
}

func (s *Store) Login(playerName, password string) (User, error) {
	var u User
	err := s.db.QueryRow(`
		SELECT u.id, u.player_name, r.name, u.api_token
		FROM users u JOIN guild_roles r ON r.id = u.role_id
		WHERE lower(u.player_name) = lower(?) AND u.password = ?`, playerName, password).
		Scan(&u.ID, &u.PlayerName, &u.Role, &u.APIToken)
	return u, err
}

func (s *Store) CreateUser(playerName, password string) (User, error) {
	// Generate a simple API token
	apiToken := "user-" + playerName + "-token"
	
	// Insert new user with Member role (role_id = 1)
	var userID int64
	err := s.db.QueryRow(`
		INSERT INTO users (player_name, role_id, password, api_token)
		VALUES (?, 1, ?, ?) RETURNING id`, playerName, password, apiToken).Scan(&userID)
	if err != nil {
		return User{}, err
	}
	
	// Return the created user
	var u User
	err = s.db.QueryRow(`
		SELECT u.id, u.player_name, r.name, u.api_token
		FROM users u JOIN guild_roles r ON r.id = u.role_id
		WHERE u.id = ?`, userID).
		Scan(&u.ID, &u.PlayerName, &u.Role, &u.APIToken)
	return u, err
}

func (s *Store) UserByToken(token string) (User, error) {
	var u User
	err := s.db.QueryRow(`
		SELECT u.id, u.player_name, r.name, u.api_token
		FROM users u JOIN guild_roles r ON r.id = u.role_id
		WHERE u.api_token = ?`, token).
		Scan(&u.ID, &u.PlayerName, &u.Role, &u.APIToken)
	return u, err
}

func (s *Store) DefaultUser() (User, error) {
	var u User
	err := s.db.QueryRow(`
		SELECT u.id, u.player_name, r.name, u.api_token
		FROM users u JOIN guild_roles r ON r.id = u.role_id
		WHERE u.id = 1`).
		Scan(&u.ID, &u.PlayerName, &u.Role, &u.APIToken)
	return u, err
}

func (s *Store) ListBuilds() ([]Build, error) {
	rows, err := s.db.Query(`SELECT id, name, role, silver_value, screenshot_url FROM builds ORDER BY name`)
	if err != nil {
		return nil, err
	}
	builds := []Build{}
	for rows.Next() {
		var b Build
		if err := rows.Scan(&b.ID, &b.Name, &b.Role, &b.SilverValue, &b.ScreenshotURL); err != nil {
			rows.Close()
			return nil, err
		}
		builds = append(builds, b)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for i := range builds {
		items, err := s.BuildItems(builds[i].ID)
		if err != nil {
			return nil, err
		}
		builds[i].Items = items
	}
	return builds, nil
}

func (s *Store) BuildItems(buildID int64) ([]BuildItem, error) {
	rows, err := s.db.Query(`SELECT id, slot, item_name, tier, enchantment, quantity FROM build_items WHERE build_id = ? ORDER BY id`, buildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []BuildItem{}
	for rows.Next() {
		var item BuildItem
		if err := rows.Scan(&item.ID, &item.Slot, &item.ItemName, &item.Tier, &item.Enchantment, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CreateBuild(actor User, b Build) (Build, error) {
	if strings.TrimSpace(b.Name) == "" {
		return b, errors.New("build name is required")
	}
	tx, err := s.rawDB.Begin()
	if err != nil {
		return b, err
	}
	defer tx.Rollback()
	wTx := wrappedDB{q: tx, s: s}
	err = wTx.QueryRow(`INSERT INTO builds (name, role, silver_value, screenshot_url, created_by) VALUES (?, ?, ?, ?, ?) RETURNING id`, b.Name, b.Role, b.SilverValue, b.ScreenshotURL, actor.ID).Scan(&b.ID)
	if err != nil {
		return b, err
	}
	for _, item := range b.Items {
		if _, err := wTx.Exec(`INSERT INTO build_items (build_id, slot, item_name, tier, enchantment, quantity) VALUES (?, ?, ?, ?, ?, ?)`,
			b.ID, item.Slot, item.ItemName, item.Tier, item.Enchantment, item.Quantity); err != nil {
			return b, err
		}
	}
	_, _ = wTx.Exec(`INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, details) VALUES (?, 'create', 'build', ?, ?)`, actor.ID, b.ID, b.Name)
	if err := tx.Commit(); err != nil {
		return b, err
	}
	return s.Build(b.ID)
}

func (s *Store) Build(id int64) (Build, error) {
	var b Build
	err := s.db.QueryRow(`SELECT id, name, role, silver_value, screenshot_url FROM builds WHERE id = ?`, id).Scan(&b.ID, &b.Name, &b.Role, &b.SilverValue, &b.ScreenshotURL)
	if err != nil {
		return b, err
	}
	b.Items, err = s.BuildItems(id)
	return b, err
}

func (s *Store) UpdateBuild(actor User, id int64, b Build) (Build, error) {
	if strings.TrimSpace(b.Name) == "" {
		return b, errors.New("build name is required")
	}
	rawTx, err := s.rawDB.Begin()
	if err != nil {
		return b, err
	}
	defer rawTx.Rollback()
	tx := wrappedTx{Tx: rawTx, s: s}

	res, err := tx.Exec(`UPDATE builds SET name = ?, role = ?, silver_value = ?, screenshot_url = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, b.Name, b.Role, b.SilverValue, b.ScreenshotURL, id)
	if err != nil {
		return b, err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return b, sql.ErrNoRows
	}
	if _, err := tx.Exec(`DELETE FROM build_items WHERE build_id = ?`, id); err != nil {
		return b, err
	}
	for _, item := range b.Items {
		if strings.TrimSpace(item.ItemName) == "" {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO build_items (build_id, slot, item_name, tier, enchantment, quantity) VALUES (?, ?, ?, ?, ?, ?)`,
			id, item.Slot, item.ItemName, item.Tier, item.Enchantment, item.Quantity); err != nil {
			return b, err
		}
	}
	_, _ = tx.Exec(`INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, details) VALUES (?, 'update', 'build', ?, ?)`, actor.ID, id, b.Name)
	if err := tx.Commit(); err != nil {
		return b, err
	}
	return s.Build(id)
}

func (s *Store) DeleteBuild(actor User, id int64) error {
	var used int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM regear_requests WHERE build_id = ?`, id).Scan(&used); err != nil {
		return err
	}
	if used > 0 {
		return errors.New("build has regear history and cannot be deleted")
	}
	rawTx, err := s.rawDB.Begin()
	if err != nil {
		return err
	}
	defer rawTx.Rollback()
	tx := wrappedTx{Tx: rawTx, s: s}
	if _, err := tx.Exec(`DELETE FROM build_items WHERE build_id = ?`, id); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM builds WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	_, _ = tx.Exec(`INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id) VALUES (?, 'delete', 'build', ?)`, actor.ID, id)
	return tx.Commit()
}

func (s *Store) CreateRegear(actor User, r RegearRequest) (RegearRequest, error) {
	var silver int64
	if err := s.db.QueryRow(`SELECT silver_value FROM builds WHERE id = ?`, r.BuildID).Scan(&silver); err != nil {
		return r, err
	}
	if r.RequestDate == "" {
		r.RequestDate = time.Now().Format("2006-01-02")
	}
	playerName := strings.TrimSpace(r.PlayerName)
	if playerName == "" {
		playerName = actor.PlayerName
	}
	err := s.db.QueryRow(`
		INSERT INTO regear_requests (user_id, player_name, request_date, build_id, death_screenshot_url, vod_url, notes, silver_value)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
		actor.ID, playerName, r.RequestDate, r.BuildID, r.DeathScreenshotURL, r.VodURL, r.Notes, silver).Scan(&r.ID)
	if err != nil {
		return r, err
	}
	return s.Regear(r.ID)
}

func (s *Store) ListRegears(actor User) ([]RegearRequest, error) {
	query := `
		SELECT rr.id, rr.user_id, rr.player_name, rr.request_date, rr.build_id, b.name, rr.death_screenshot_url,
		       rr.vod_url, COALESCE(rr.notes, ''), rr.status, rr.silver_value, COALESCE(rr.pickup_location, ''), rr.created_at
		FROM regear_requests rr JOIN builds b ON b.id = rr.build_id
		WHERE rr.status IN ('Pending', 'Approved')`
	args := []any{}
	if actor.Role == "Member" {
		query += ` AND (rr.user_id = ? OR lower(rr.player_name) = lower(?))`
		args = append(args, actor.ID, actor.PlayerName)
	}
	query += ` ORDER BY rr.created_at DESC`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRegears(rows)
}

func (s *Store) Regear(id int64) (RegearRequest, error) {
	rows, err := s.db.Query(`
		SELECT rr.id, rr.user_id, rr.player_name, rr.request_date, rr.build_id, b.name, rr.death_screenshot_url,
		       rr.vod_url, COALESCE(rr.notes, ''), rr.status, rr.silver_value, COALESCE(rr.pickup_location, ''), rr.created_at
		FROM regear_requests rr JOIN builds b ON b.id = rr.build_id WHERE rr.id = ?`, id)
	if err != nil {
		return RegearRequest{}, err
	}
	defer rows.Close()
	regears, err := scanRegears(rows)
	if err != nil || len(regears) == 0 {
		if err != nil {
			return RegearRequest{}, err
		}
		return RegearRequest{}, sql.ErrNoRows
	}
	regears[0].Items, _ = s.RegearItems(id)
	return regears[0], nil
}

func scanRegears(rows *sql.Rows) ([]RegearRequest, error) {
	regears := []RegearRequest{}
	for rows.Next() {
		var r RegearRequest
		if err := rows.Scan(&r.ID, &r.UserID, &r.PlayerName, &r.RequestDate, &r.BuildID, &r.BuildName, &r.DeathScreenshotURL, &r.VodURL, &r.Notes, &r.Status, &r.SilverValue, &r.PickupLocation, &r.CreatedAt); err != nil {
			return nil, err
		}
		regears = append(regears, r)
	}
	return regears, rows.Err()
}

func (s *Store) RegearItems(id int64) ([]RegearRequestItem, error) {
	rows, err := s.db.Query(`SELECT item_name, tier, enchantment, quantity_needed, quantity_fulfilled, quantity_missing FROM regear_request_items WHERE regear_request_id = ? ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []RegearRequestItem{}
	for rows.Next() {
		var item RegearRequestItem
		if err := rows.Scan(&item.ItemName, &item.Tier, &item.Enchantment, &item.QuantityNeeded, &item.QuantityFulfilled, &item.QuantityMissing); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) UpdateRegearStatus(actor User, id int64, status string, pickupLocation string) (RegearRequest, error) {
	if !validStatus(status) {
		return RegearRequest{}, errors.New("invalid status")
	}
	rawTx, err := s.rawDB.Begin()
	if err != nil {
		return RegearRequest{}, err
	}
	defer rawTx.Rollback()
	tx := wrappedTx{Tx: rawTx, s: s}

	var oldStatus string
	var buildID int64
	var userID int64
	if err := tx.QueryRow(`SELECT status, build_id, user_id FROM regear_requests WHERE id = ?`, id).Scan(&oldStatus, &buildID, &userID); err != nil {
		return RegearRequest{}, err
	}

	isOfficer := actor.Role == "Officer" || actor.Role == "Admin" || actor.Role == "Owner"
	if !isOfficer {
		if userID != actor.ID {
			return RegearRequest{}, errors.New("permission denied: not your request")
		}
		if status != "Completed" {
			return RegearRequest{}, errors.New("permission denied: can only mark completed")
		}
		pickupLocation = ""
	}

	if status == "Approved" && oldStatus != "Approved" {
		if err := fulfillBuildItems(tx, id, buildID); err != nil {
			return RegearRequest{}, err
		}
	}

	if status == "Completed" && oldStatus != "Completed" {
		if err := deductMissingItems(tx, id); err != nil {
			return RegearRequest{}, err
		}
	}

	if status == "Denied" && oldStatus == "Approved" {
		if err := returnFulfilledItems(tx, id); err != nil {
			return RegearRequest{}, err
		}
	}

	var query string
	var args []any
	if isOfficer {
		query = `
			UPDATE regear_requests
			SET status = ?, reviewed_by = ?, reviewed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP,
			    pickup_location = CASE WHEN ? != '' THEN ? ELSE pickup_location END
			WHERE id = ?`
		args = []any{status, actor.ID, pickupLocation, pickupLocation, id}
	} else {
		query = `
			UPDATE regear_requests
			SET status = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`
		args = []any{status, id}
	}

	_, err = tx.Exec(query, args...)
	if err != nil {
		return RegearRequest{}, err
	}
	_, _ = tx.Exec(`INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, details) VALUES (?, 'status_change', 'regear_request', ?, ?)`, actor.ID, id, status)
	if err := tx.Commit(); err != nil {
		return RegearRequest{}, err
	}
	return s.Regear(id)
}

func fulfillBuildItems(tx queryable, regearID, buildID int64) error {
	_, _ = tx.Exec(`DELETE FROM regear_request_items WHERE regear_request_id = ?`, regearID)
	rows, err := tx.Query(`SELECT item_name, tier, enchantment, quantity FROM build_items WHERE build_id = ?`, buildID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type buildItem struct {
		name       string
		tier, ench int
		needed     int
	}
	var items []buildItem

	for rows.Next() {
		var item buildItem
		if err := rows.Scan(&item.name, &item.tier, &item.ench, &item.needed); err != nil {
			return err
		}
		items = append(items, item)
	}
	rows.Close()

	for _, item := range items {
		var available int
		err := tx.QueryRow(`SELECT quantity_available FROM inventory WHERE item_name = ? AND equivalent_tier = ?`, item.name, item.tier+item.ench).Scan(&available)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		fulfilled := min(available, item.needed)
		missing := item.needed - fulfilled
		if fulfilled > 0 {
			if _, err := tx.Exec(`UPDATE inventory SET quantity_available = quantity_available - ?, last_updated = CURRENT_TIMESTAMP WHERE item_name = ? AND equivalent_tier = ?`, fulfilled, item.name, item.tier+item.ench); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(`
			INSERT INTO regear_request_items (regear_request_id, item_name, tier, enchantment, quantity_needed, quantity_fulfilled, quantity_missing)
			VALUES (?, ?, ?, ?, ?, ?, ?)`, regearID, item.name, item.tier, item.ench, item.needed, fulfilled, missing); err != nil {
			return err
		}
	}
	return nil
}

func deductMissingItems(tx queryable, regearID int64) error {
	rows, err := tx.Query(`SELECT item_name, tier, enchantment, quantity_missing FROM regear_request_items WHERE regear_request_id = ?`, regearID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type missingItem struct {
		name       string
		tier, ench int
		missingQty int
	}
	var items []missingItem

	for rows.Next() {
		var mi missingItem
		if err := rows.Scan(&mi.name, &mi.tier, &mi.ench, &mi.missingQty); err != nil {
			return err
		}
		if mi.missingQty > 0 {
			items = append(items, mi)
		}
	}
	rows.Close()

	for _, mi := range items {
		var available int
		err := tx.QueryRow(`SELECT quantity_available FROM inventory WHERE item_name = ? AND equivalent_tier = ?`, mi.name, mi.tier+mi.ench).Scan(&available)
		if err == nil {
			deduct := min(available, mi.missingQty)
			if deduct > 0 {
				if _, err := tx.Exec(`UPDATE inventory SET quantity_available = quantity_available - ?, last_updated = CURRENT_TIMESTAMP WHERE item_name = ? AND equivalent_tier = ?`, deduct, mi.name, mi.tier+mi.ench); err != nil {
					return err
				}
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if _, err := tx.Exec(`UPDATE regear_request_items SET quantity_fulfilled = quantity_fulfilled + ?, quantity_missing = quantity_missing - ? WHERE regear_request_id = ? AND item_name = ? AND tier = ? AND enchantment = ?`, mi.missingQty, mi.missingQty, regearID, mi.name, mi.tier, mi.ench); err != nil {
			return err
		}
	}
	return nil
}

func returnFulfilledItems(tx queryable, regearID int64) error {
	rows, err := tx.Query(`SELECT item_name, tier, enchantment, quantity_fulfilled FROM regear_request_items WHERE regear_request_id = ?`, regearID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type fulfilledItem struct {
		name       string
		tier, ench int
		fulfilled  int
	}
	var items []fulfilledItem

	for rows.Next() {
		var fi fulfilledItem
		if err := rows.Scan(&fi.name, &fi.tier, &fi.ench, &fi.fulfilled); err != nil {
			return err
		}
		if fi.fulfilled > 0 {
			items = append(items, fi)
		}
	}
	rows.Close()

	for _, fi := range items {
		_, err := tx.Exec(`UPDATE inventory SET quantity_available = quantity_available + ?, last_updated = CURRENT_TIMESTAMP WHERE item_name = ? AND equivalent_tier = ?`, fi.fulfilled, fi.name, fi.tier+fi.ench)
		if err != nil {
			return err
		}
	}
	return nil
}

func validStatus(status string) bool {
	return status == "Pending" || status == "Approved" || status == "Denied" || status == "Completed"
}

func (s *Store) ListInventory() ([]InventoryItem, error) {
	rows, err := s.db.Query(`SELECT id, item_name, equivalent_tier, quantity_available, low_stock_threshold, last_updated FROM inventory ORDER BY item_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []InventoryItem{}
	for rows.Next() {
		var item InventoryItem
		if err := rows.Scan(&item.ID, &item.ItemName, &item.EquivalentTier, &item.QuantityAvailable, &item.LowStockThreshold, &item.LastUpdated); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) UpsertInventory(actor User, item InventoryItem) (InventoryItem, error) {
	_, err := s.db.Exec(`
		INSERT INTO inventory (item_name, equivalent_tier, quantity_available, low_stock_threshold, last_updated)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(item_name, equivalent_tier) DO UPDATE SET
		  quantity_available = excluded.quantity_available,
		  low_stock_threshold = excluded.low_stock_threshold,
		  last_updated = CURRENT_TIMESTAMP`,
		item.ItemName, item.EquivalentTier, item.QuantityAvailable, item.LowStockThreshold)
	if err != nil {
		return item, err
	}
	_, _ = s.db.Exec(`INSERT INTO audit_logs (actor_user_id, action, entity_type, details) VALUES (?, 'upsert', 'inventory', ?)`, actor.ID, item.ItemName)
	return item, nil
}

func (s *Store) UpdateInventory(actor User, id int64, item InventoryItem) (InventoryItem, error) {
	_, err := s.db.Exec(`UPDATE inventory SET quantity_available = ?, low_stock_threshold = ?, last_updated = CURRENT_TIMESTAMP WHERE id = ?`, item.QuantityAvailable, item.LowStockThreshold, id)
	if err != nil {
		return item, err
	}
	_, _ = s.db.Exec(`INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, details) VALUES (?, 'update', 'inventory', ?, ?)`, actor.ID, id, fmt.Sprint(item.QuantityAvailable))
	return item, nil
}

func (s *Store) DeleteInventory(actor User, id int64) error {
	res, err := s.db.Exec(`DELETE FROM inventory WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	_, _ = s.db.Exec(`INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id) VALUES (?, 'delete', 'inventory', ?)`, actor.ID, id)
	return nil
}
func (s *Store) GenerateShoppingList(actor User) (ShoppingList, error) {
	rawTx, err := s.rawDB.Begin()
	if err != nil {
		return ShoppingList{}, err
	}
	defer rawTx.Rollback()
	tx := wrappedTx{Tx: rawTx, s: s}
	name := "Open shortages " + time.Now().Format("2006-01-02 15:04")
	var id int64
	err = tx.QueryRow(`INSERT INTO shopping_lists (name, generated_by) VALUES (?, ?) RETURNING id`, name, actor.ID).Scan(&id)
	if err != nil {
		return ShoppingList{}, err
	}
	rows, err := tx.Query(`
		SELECT item_name, (tier + enchantment) AS equivalent_tier, SUM(quantity_missing)
		FROM regear_request_items rri
		JOIN regear_requests rr ON rr.id = rri.regear_request_id
		WHERE rr.status = 'Approved' AND quantity_missing > 0
		GROUP BY item_name, (tier + enchantment)
		ORDER BY SUM(quantity_missing) DESC, item_name`)
	if err != nil {
		return ShoppingList{}, err
	}
	defer rows.Close()

	var items []ShoppingListItem
	for rows.Next() {
		var item ShoppingListItem
		if err := rows.Scan(&item.ItemName, &item.EquivalentTier, &item.QuantityNeeded); err != nil {
			return ShoppingList{}, err
		}
		items = append(items, item)
	}
	rows.Close()

	for _, item := range items {
		if _, err := tx.Exec(`INSERT INTO shopping_list_items (shopping_list_id, item_name, equivalent_tier, quantity_needed) VALUES (?, ?, ?, ?)`,
			id, item.ItemName, item.EquivalentTier, item.QuantityNeeded); err != nil {
			return ShoppingList{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return ShoppingList{}, err
	}
	return s.ShoppingList(id)
}

func (s *Store) LatestShoppingList() (ShoppingList, error) {
	var id int64
	if err := s.db.QueryRow(`SELECT id FROM shopping_lists ORDER BY created_at DESC, id DESC LIMIT 1`).Scan(&id); err != nil {
		return ShoppingList{}, err
	}
	return s.ShoppingList(id)
}

func (s *Store) ShoppingList(id int64) (ShoppingList, error) {
	var list ShoppingList
	if err := s.db.QueryRow(`SELECT id, name, status, created_at FROM shopping_lists WHERE id = ?`, id).Scan(&list.ID, &list.Name, &list.Status, &list.CreatedAt); err != nil {
		return list, err
	}
	list.Items = []ShoppingListItem{}
	rows, err := s.db.Query(`SELECT item_name, equivalent_tier, quantity_needed FROM shopping_list_items WHERE shopping_list_id = ? ORDER BY quantity_needed DESC`, id)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var item ShoppingListItem
		if err := rows.Scan(&item.ItemName, &item.EquivalentTier, &item.QuantityNeeded); err != nil {
			return list, err
		}
		list.Items = append(list.Items, item)
	}
	return list, rows.Err()
}

func (s *Store) Dashboard(actor User) (Dashboard, error) {
	d := Dashboard{
		MostRequestedItems: []ShoppingListItem{},
		LowStockItems:      []InventoryItem{},
		RecentRegears:      []RegearRequest{},
	}
	
	if actor.Role == "Member" {
		_ = s.db.QueryRow(`SELECT COUNT(*) FROM regear_requests WHERE status = 'Pending' AND (user_id = ? OR lower(player_name) = lower(?))`, actor.ID, actor.PlayerName).Scan(&d.PendingRegears)
		_ = s.db.QueryRow(`SELECT COUNT(*) FROM regear_requests WHERE status = 'Approved' AND (user_id = ? OR lower(player_name) = lower(?))`, actor.ID, actor.PlayerName).Scan(&d.ApprovedRegears)
		_ = s.db.QueryRow(`SELECT COUNT(*) FROM regear_requests WHERE status = 'Denied' AND (user_id = ? OR lower(player_name) = lower(?))`, actor.ID, actor.PlayerName).Scan(&d.DeniedRegears)
		_ = s.db.QueryRow(`SELECT COALESCE(SUM(silver_value),0) FROM regear_requests WHERE status = 'Pending' AND (user_id = ? OR lower(player_name) = lower(?))`, actor.ID, actor.PlayerName).Scan(&d.PendingSilverValue)
		d.TotalInventoryItems = 0
		d.OpenShortageQuantity = 0
		d.RecentRegears, _ = s.recentRegears(actor)
	} else {
		_ = s.db.QueryRow(`SELECT COUNT(*) FROM regear_requests WHERE status = 'Pending'`).Scan(&d.PendingRegears)
		_ = s.db.QueryRow(`SELECT COUNT(*) FROM regear_requests WHERE status = 'Approved'`).Scan(&d.ApprovedRegears)
		_ = s.db.QueryRow(`SELECT COUNT(*) FROM regear_requests WHERE status = 'Denied'`).Scan(&d.DeniedRegears)
		_ = s.db.QueryRow(`SELECT COALESCE(SUM(silver_value),0) FROM regear_requests WHERE status = 'Pending'`).Scan(&d.PendingSilverValue)
		_ = s.db.QueryRow(`SELECT COALESCE(SUM(quantity_available),0) FROM inventory`).Scan(&d.TotalInventoryItems)
		_ = s.db.QueryRow(`SELECT COALESCE(SUM(quantity_missing),0) FROM regear_request_items rri JOIN regear_requests rr ON rr.id = rri.regear_request_id WHERE rr.status = 'Approved'`).Scan(&d.OpenShortageQuantity)
		d.LowStockItems, _ = s.lowStock()
		d.MostRequestedItems, _ = s.mostRequested()
		d.RecentRegears, _ = s.recentRegears(actor)
	}
	return d, nil
}

func (s *Store) lowStock() ([]InventoryItem, error) {
	rows, err := s.db.Query(`SELECT id, item_name, equivalent_tier, quantity_available, low_stock_threshold, last_updated FROM inventory WHERE quantity_available <= low_stock_threshold ORDER BY quantity_available, item_name LIMIT 8`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []InventoryItem{}
	for rows.Next() {
		var item InventoryItem
		if err := rows.Scan(&item.ID, &item.ItemName, &item.EquivalentTier, &item.QuantityAvailable, &item.LowStockThreshold, &item.LastUpdated); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) mostRequested() ([]ShoppingListItem, error) {
	rows, err := s.db.Query(`
		SELECT bi.item_name, (bi.tier + bi.enchantment) AS equivalent_tier, COUNT(*)
		FROM regear_requests rr JOIN build_items bi ON bi.build_id = rr.build_id
		GROUP BY bi.item_name, (bi.tier + bi.enchantment)
		ORDER BY COUNT(*) DESC, bi.item_name LIMIT 8`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ShoppingListItem{}
	for rows.Next() {
		var item ShoppingListItem
		if err := rows.Scan(&item.ItemName, &item.EquivalentTier, &item.QuantityNeeded); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) recentRegears(actor User) ([]RegearRequest, error) {
	query := `
		SELECT rr.id, rr.user_id, rr.player_name, rr.request_date, rr.build_id, b.name, rr.death_screenshot_url,
		       rr.vod_url, COALESCE(rr.notes, ''), rr.status, rr.silver_value, COALESCE(rr.pickup_location, ''), rr.created_at
		FROM regear_requests rr JOIN builds b ON b.id = rr.build_id`
	args := []any{}
	if actor.Role == "Member" {
		query += ` WHERE rr.user_id = ? OR lower(rr.player_name) = lower(?)`
		args = append(args, actor.ID, actor.PlayerName)
	}
	query += ` ORDER BY rr.created_at DESC LIMIT 6`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRegears(rows)
}

func (s *Store) MemberHistory() ([]MemberHistory, error) {
	rows, err := s.db.Query(`
		SELECT u.id, u.player_name, r.name,
		       COUNT(rr.id),
		       COALESCE(SUM(CASE WHEN rr.status IN ('Approved','Completed') THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN rr.status IN ('Approved','Completed') THEN rr.silver_value ELSE 0 END), 0),
		       (SELECT status FROM regear_requests rr2 WHERE rr2.user_id = u.id ORDER BY created_at DESC LIMIT 1)
		FROM users u
		JOIN guild_roles r ON r.id = u.role_id
		LEFT JOIN regear_requests rr ON rr.user_id = u.id
		GROUP BY u.id, u.player_name, r.id, r.name
		ORDER BY r.id DESC, u.player_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	history := []MemberHistory{}
	for rows.Next() {
		var row MemberHistory
		var lastStatus sql.NullString
		if err := rows.Scan(&row.ID, &row.PlayerName, &row.Role, &row.Requested, &row.Approved, &row.SilverValue, &lastStatus); err != nil {
			return nil, err
		}
		if lastStatus.Valid {
			row.LastRequestStatus = lastStatus.String
		}
		history = append(history, row)
	}
	return history, rows.Err()
}

func (s *Store) UpdateUserRole(actor User, targetID int64, newRole string) error {
	var actorRoleID int
	_ = s.db.QueryRow(`SELECT id FROM guild_roles WHERE name = ?`, actor.Role).Scan(&actorRoleID)
	
	var targetRoleID int
	err := s.db.QueryRow(`SELECT role_id FROM users WHERE id = ?`, targetID).Scan(&targetRoleID)
	if err != nil {
		return err
	}
	
	var newRoleID int
	err = s.db.QueryRow(`SELECT id FROM guild_roles WHERE name = ?`, newRole).Scan(&newRoleID)
	if err != nil {
		return err
	}
	
	if actorRoleID <= targetRoleID {
		return fmt.Errorf("insufficient permissions to edit this user")
	}
	if actorRoleID <= newRoleID {
		return fmt.Errorf("insufficient permissions to assign this role")
	}
	
	_, err = s.db.Exec(`UPDATE users SET role_id = ? WHERE id = ?`, newRoleID, targetID)
	return err
}

func (s *Store) DeleteUser(actor User, targetID int64) error {
	var actorRoleID int
	_ = s.db.QueryRow(`SELECT id FROM guild_roles WHERE name = ?`, actor.Role).Scan(&actorRoleID)
	
	var targetRoleID int
	err := s.db.QueryRow(`SELECT role_id FROM users WHERE id = ?`, targetID).Scan(&targetRoleID)
	if err != nil {
		return err
	}
	
	if actorRoleID <= targetRoleID {
		return fmt.Errorf("insufficient permissions to delete this user")
	}
	
	rawTx, err := s.rawDB.Begin()
	if err != nil {
		return err
	}
	defer rawTx.Rollback()
	tx := wrappedTx{Tx: rawTx, s: s}
	
	// Nullify references in regear_requests, builds, shopping_lists, audit_logs
	if _, err := tx.Exec(`UPDATE regear_requests SET user_id = NULL WHERE user_id = ?`, targetID); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE regear_requests SET reviewed_by = NULL WHERE reviewed_by = ?`, targetID); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE builds SET created_by = NULL WHERE created_by = ?`, targetID); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE shopping_lists SET generated_by = NULL WHERE generated_by = ?`, targetID); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE audit_logs SET actor_user_id = NULL WHERE actor_user_id = ?`, targetID); err != nil {
		return err
	}
	
	// Finally delete the user
	if _, err := tx.Exec(`DELETE FROM users WHERE id = ?`, targetID); err != nil {
		return err
	}
	
	return tx.Commit()
}
