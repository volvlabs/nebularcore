package tenant_test

import (
	"context"
	"errors"
	"testing"

	"github.com/volvlabs/nebularcore/modules/tenant"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// reading and farm are minimal tenant-bound models used only to exercise
// the plugin. Real callers embed core/model.Model; these skip it to stay
// SQLite-compatible in a fast, dependency-free unit test.

type reading struct {
	ID     uint `gorm:"primaryKey"`
	FarmID uint
	ZoneID string
	Value  float64
}

func (reading) TableName() string   { return "readings" }
func (reading) IsTenantBound() bool { return true }

type farm struct {
	ID       uint `gorm:"primaryKey"`
	Name     string
	Readings []reading `gorm:"foreignKey:FarmID"`
}

func (farm) TableName() string   { return "farms" }
func (farm) IsTenantBound() bool { return true }

// setupDB opens an in-memory SQLite db with two ATTACHed databases standing
// in for Postgres schemas ("farm_a", "farm_b"), each with its own copy of
// the farms/readings tables, and registers the tenant plugin on it.
//
// SQLite ATTACHes are per-connection, so the pool is pinned to a single
// connection — otherwise a later query could land on a connection that
// never ran the ATTACH and fail with "no such table" independent of
// anything the plugin does.
func setupDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	for _, schema := range []string{"farm_a", "farm_b"} {
		if err := db.Exec("ATTACH DATABASE ':memory:' AS " + schema).Error; err != nil {
			t.Fatalf("attach %s: %v", schema, err)
		}
		if err := db.Exec("CREATE TABLE " + schema + ".farms (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)").Error; err != nil {
			t.Fatalf("create farms table in %s: %v", schema, err)
		}
		if err := db.Exec("CREATE TABLE " + schema + ".readings (id INTEGER PRIMARY KEY AUTOINCREMENT, farm_id INTEGER, zone_id TEXT, value REAL)").Error; err != nil {
			t.Fatalf("create readings table in %s: %v", schema, err)
		}
	}

	if err := db.Use(tenant.NewPlugin()); err != nil {
		t.Fatalf("register plugin: %v", err)
	}
	return db
}

func ctxFor(schema string) context.Context {
	return tenant.WithTenant(context.Background(), schema, schema, schema)
}

func TestPlugin_CrossTenantIsolation(t *testing.T) {
	db := setupDB(t)

	mustCreate(t, db.WithContext(ctxFor("farm_a")), &reading{ZoneID: "z1", Value: 1})
	mustCreate(t, db.WithContext(ctxFor("farm_b")), &reading{ZoneID: "z1", Value: 2})

	var aRows, bRows []reading
	if err := db.WithContext(ctxFor("farm_a")).Find(&aRows).Error; err != nil {
		t.Fatalf("find farm_a: %v", err)
	}
	if err := db.WithContext(ctxFor("farm_b")).Find(&bRows).Error; err != nil {
		t.Fatalf("find farm_b: %v", err)
	}

	if len(aRows) != 1 || aRows[0].Value != 1 {
		t.Fatalf("farm_a query leaked or missing data: %+v", aRows)
	}
	if len(bRows) != 1 || bRows[0].Value != 2 {
		t.Fatalf("farm_b query leaked or missing data: %+v", bRows)
	}
}

func TestPlugin_UpdateDeleteIsolation(t *testing.T) {
	db := setupDB(t)
	mustCreate(t, db.WithContext(ctxFor("farm_a")), &reading{ZoneID: "z1", Value: 10})
	mustCreate(t, db.WithContext(ctxFor("farm_b")), &reading{ZoneID: "z1", Value: 20})

	if err := db.WithContext(ctxFor("farm_a")).Model(&reading{}).Where("zone_id = ?", "z1").Update("value", 99).Error; err != nil {
		t.Fatalf("update farm_a: %v", err)
	}

	var a, b reading
	if err := db.WithContext(ctxFor("farm_a")).First(&a, "zone_id = ?", "z1").Error; err != nil {
		t.Fatalf("read back farm_a: %v", err)
	}
	if err := db.WithContext(ctxFor("farm_b")).First(&b, "zone_id = ?", "z1").Error; err != nil {
		t.Fatalf("read back farm_b: %v", err)
	}
	if a.Value != 99 {
		t.Fatalf("farm_a update did not apply: got %v", a.Value)
	}
	if b.Value != 20 {
		t.Fatalf("farm_b row leaked the farm_a update: got %v", b.Value)
	}

	if err := db.WithContext(ctxFor("farm_a")).Delete(&reading{}, a.ID).Error; err != nil {
		t.Fatalf("delete farm_a: %v", err)
	}
	var remainingA []reading
	db.WithContext(ctxFor("farm_a")).Find(&remainingA)
	if len(remainingA) != 0 {
		t.Fatalf("expected farm_a empty after delete, got %d rows", len(remainingA))
	}
	var remainingB []reading
	db.WithContext(ctxFor("farm_b")).Find(&remainingB)
	if len(remainingB) != 1 {
		t.Fatalf("farm_a delete affected farm_b: got %d rows", len(remainingB))
	}
}

func TestPlugin_PreloadIsolation(t *testing.T) {
	db := setupDB(t)

	mustCreate(t, db.WithContext(ctxFor("farm_a")), &farm{Name: "Alpha"})
	mustCreate(t, db.WithContext(ctxFor("farm_b")), &farm{Name: "Beta"})

	var fa, fb farm
	db.WithContext(ctxFor("farm_a")).First(&fa)
	db.WithContext(ctxFor("farm_b")).First(&fb)

	// Deliberately give both farms the same numeric ID pattern by creating
	// one reading each, so a leak can't hide behind "the IDs never matched
	// anyway" — isolation must come from the schema, not the id space.
	mustCreate(t, db.WithContext(ctxFor("farm_a")), &reading{ZoneID: "za", Value: 1, FarmID: fa.ID})
	mustCreate(t, db.WithContext(ctxFor("farm_b")), &reading{ZoneID: "zb", Value: 2, FarmID: fb.ID})

	var loaded farm
	if err := db.WithContext(ctxFor("farm_a")).Preload("Readings").First(&loaded, fa.ID).Error; err != nil {
		t.Fatalf("preload farm_a: %v", err)
	}
	if len(loaded.Readings) != 1 || loaded.Readings[0].ZoneID != "za" {
		t.Fatalf("preload leaked or missed farm_a readings: %+v", loaded.Readings)
	}
}

func TestPlugin_FailsClosedWithoutTenant(t *testing.T) {
	db := setupDB(t)

	err := db.WithContext(context.Background()).Create(&reading{ZoneID: "z1", Value: 1}).Error
	if err == nil {
		t.Fatal("expected error when no tenant is in context, got nil (silently succeeded)")
	}
	if !errors.Is(err, tenant.ErrNoTenantInContext) {
		t.Fatalf("expected ErrNoTenantInContext, got: %v", err)
	}

	var rows []reading
	err = db.WithContext(context.Background()).Find(&rows).Error
	if err == nil {
		t.Fatal("expected error on Find with no tenant in context, got nil")
	}
	if !errors.Is(err, tenant.ErrNoTenantInContext) {
		t.Fatalf("expected ErrNoTenantInContext on Find, got: %v", err)
	}
}

func TestPlugin_DoesNotDoubleQualifyExplicitTable(t *testing.T) {
	db := setupDB(t)

	// Caller manually schema-qualifies via Table(), while the *context*
	// resolves to a different tenant (farm_b) — the plugin must respect
	// the caller's explicit table and must not silently redirect the
	// write to the context's tenant instead.
	err := db.WithContext(ctxFor("farm_b")).Table("farm_a.readings").Create(&reading{ZoneID: "manual", Value: 7}).Error
	if err != nil {
		t.Fatalf("manual table create failed: %v", err)
	}

	var aRows, bRows []reading
	if err := db.WithContext(ctxFor("farm_a")).Find(&aRows).Error; err != nil {
		t.Fatalf("find farm_a: %v", err)
	}
	if err := db.WithContext(ctxFor("farm_b")).Find(&bRows).Error; err != nil {
		t.Fatalf("find farm_b: %v", err)
	}

	foundInA := false
	for _, r := range aRows {
		if r.ZoneID == "manual" {
			foundInA = true
		}
	}
	if !foundInA {
		t.Fatal("expected row written via explicit Table(\"farm_a.readings\") to land in farm_a")
	}
	for _, r := range bRows {
		if r.ZoneID == "manual" {
			t.Fatal("row written via explicit Table(\"farm_a.readings\") leaked into farm_b (the context's tenant) instead")
		}
	}
}

func mustCreate(t *testing.T, db *gorm.DB, v interface{}) {
	t.Helper()
	if err := db.Create(v).Error; err != nil {
		t.Fatalf("create %#v: %v", v, err)
	}
}
