package store

import (
	"path/filepath"
	"testing"
)

func TestApproveRegearConsumesInventoryAndCreatesShoppingList(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	st := New(db)
	member, err := st.Login("Blazor", "123")
	if err != nil {
		t.Fatal(err)
	}
	officer, err := st.Login("Blazor", "123")
	if err != nil {
		t.Fatal(err)
	}
	build, err := st.CreateBuild(officer, Build{
		Name:        "Test Tank",
		Role:        "Tank",
		SilverValue: 1800000,
		Items: []BuildItem{
			{Slot: "Main Hand", ItemName: "Incubus Mace", Tier: 7, Enchantment: 1, Quantity: 1},
			{Slot: "Cape", ItemName: "Martlock Cape", Tier: 7, Enchantment: 1, Quantity: 1},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := st.CreateRegear(member, RegearRequest{
		PlayerName:         "Blazor",
		BuildID:            build.ID,
		DeathScreenshotURL: "https://example.com/death.png",
		VodURL:             "https://albiononline.com/killboard/kill/1",
	})
	if err != nil {
		t.Fatal(err)
	}

	approved, err := st.UpdateRegearStatus(officer, req.ID, "Approved", "")
	if err != nil {
		t.Fatal(err)
	}
	if approved.Status != "Approved" {
		t.Fatalf("status = %s, want Approved", approved.Status)
	}
	if len(approved.Items) == 0 {
		t.Fatal("expected approved request items")
	}

	var martlockCape RegearRequestItem
	for _, item := range approved.Items {
		if item.ItemName == "Martlock Cape" {
			martlockCape = item
			break
		}
	}
	if martlockCape.QuantityMissing != 1 {
		t.Fatalf("martlock cape missing = %d, want 1", martlockCape.QuantityMissing)
	}

	list, err := st.GenerateShoppingList(officer)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) == 0 {
		t.Fatal("expected shopping list items")
	}
	if list.Items[0].QuantityNeeded < 1 {
		t.Fatalf("quantity needed = %d, want positive", list.Items[0].QuantityNeeded)
	}
}

func TestUpdateAndDeleteBuild(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	st := New(db)
	admin, err := st.Login("Blazor", "123")
	if err != nil {
		t.Fatal(err)
	}

	build, err := st.CreateBuild(admin, Build{
		Name:          "Starter DPS",
		Role:          "DPS",
		SilverValue:   100,
		ScreenshotURL: "data:image/png;base64,test",
		Items: []BuildItem{
			{Slot: "Main Hand", ItemName: "Bow", Tier: 4, Enchantment: 0, Quantity: 1},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := st.UpdateBuild(admin, build.ID, Build{
		Name:          "Updated DPS",
		Role:          "DPS",
		SilverValue:   200,
		ScreenshotURL: "data:image/png;base64,updated",
		Items: []BuildItem{
			{Slot: "Main Hand", ItemName: "Longbow", Tier: 6, Enchantment: 1, Quantity: 1},
			{Slot: "Cape", ItemName: "Thetford Cape", Tier: 6, Enchantment: 1, Quantity: 1},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Updated DPS" || updated.ScreenshotURL == "" || len(updated.Items) != 2 {
		t.Fatalf("unexpected updated build: %#v", updated)
	}

	if err := st.DeleteBuild(admin, build.ID); err != nil {
		t.Fatal(err)
	}
	builds, err := st.ListBuilds()
	if err != nil {
		t.Fatal(err)
	}
	if len(builds) != 0 {
		t.Fatalf("build count = %d, want 0", len(builds))
	}
}

func TestDeleteInventory(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	st := New(db)
	admin, err := st.Login("Blazor", "123")
	if err != nil {
		t.Fatal(err)
	}

	item, err := st.UpsertInventory(admin, InventoryItem{
		ItemName:          "Delete Me",
		EquivalentTier:    5,
		QuantityAvailable: 3,
		LowStockThreshold: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	items, err := st.ListInventory()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range items {
		if row.ItemName == item.ItemName {
			item.ID = row.ID
			break
		}
	}
	if item.ID == 0 {
		t.Fatal("expected inventory id")
	}
	if err := st.DeleteInventory(admin, item.ID); err != nil {
		t.Fatal(err)
	}
	items, err = st.ListInventory()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range items {
		if row.ID == item.ID {
			t.Fatalf("inventory item was not deleted: %#v", row)
		}
	}
}
