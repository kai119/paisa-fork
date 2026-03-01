package server

import (
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}
	if err := db.AutoMigrate(&posting.Posting{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func loadTestConfig(t *testing.T) {
	t.Helper()
	yaml := []byte("journal_path: main.ledger\ndb_path: main.db\ndefault_currency: GBP\n")
	if err := config.LoadConfig(yaml, ""); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
}

func makePosting(account string, amount float64, date time.Time) posting.Posting {
	return posting.Posting{
		Account:   account,
		Commodity: "GBP",
		Amount:    decimal.NewFromFloat(amount),
		Quantity:  decimal.NewFromFloat(amount),
		Date:      date,
	}
}

// For GBP currency postings, NetInvestmentAmount = market value = current balance
// of Assets:Investments:* accounts (positive + negative postings summed).

func TestComputeNetworth_OnlyInvestmentAccountsInNetBalance(t *testing.T) {
	loadTestConfig(t)
	db := openTestDB(t)
	service.ClearInterestCache()
	utils.SetNow("2026-03-01")

	date := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	postings := []posting.Posting{
		makePosting("Assets:Investments:ISA", 1000, date),
		makePosting("Assets:Checking:Main", 500, date),
	}

	nw := computeNetworth(db, postings)

	// Net Investment = market value of Assets:Investments:* only = £1000
	assert.Equal(t, "1000", nw.NetInvestmentAmount.String())
	// Total investment (all inflows) still = £1500
	assert.Equal(t, "1500", nw.InvestmentAmount.String())
}

func TestComputeNetworth_NetBalanceReflectsWithdrawals(t *testing.T) {
	loadTestConfig(t)
	db := openTestDB(t)
	service.ClearInterestCache()
	utils.SetNow("2026-03-01")

	date := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	postings := []posting.Posting{
		makePosting("Assets:Investments:ISA", 1000, date),
		makePosting("Assets:Investments:ISA", -300, date),
		makePosting("Assets:Checking:Main", -200, date),
	}

	nw := computeNetworth(db, postings)

	// Net Investment = 1000 - 300 = 700 (Checking withdrawal excluded)
	assert.Equal(t, "700", nw.NetInvestmentAmount.String())
}

func TestComputeNetworth_NonInvestmentOnlyHasZeroNetBalance(t *testing.T) {
	loadTestConfig(t)
	db := openTestDB(t)
	service.ClearInterestCache()
	utils.SetNow("2026-03-01")

	date := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	postings := []posting.Posting{
		makePosting("Assets:Checking:Main", 5000, date),
		makePosting("Assets:Savings", 2000, date),
	}

	nw := computeNetworth(db, postings)

	assert.Equal(t, "0", nw.NetInvestmentAmount.String())
	assert.Equal(t, "7000", nw.InvestmentAmount.String())
}
