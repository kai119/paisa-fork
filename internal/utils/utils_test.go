package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/stretchr/testify/assert"
)

func mustLoadConfig(startDay int) {
	var yaml []byte
	if startDay > 1 {
		yaml = []byte(fmt.Sprintf("journal_path: main.ledger\ndb_path: main.db\nbudget:\n  start_day: %d\n", startDay))
	} else {
		yaml = []byte("journal_path: main.ledger\ndb_path: main.db\n")
	}
	if err := config.LoadConfig(yaml, ""); err != nil {
		panic(err)
	}
}

func dateStr(t time.Time) string {
	return t.Format("2006-01-02")
}

func ds(year int, month time.Month, day int) string {
	return fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
}

func d(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 12, 0, 0, 0, time.UTC)
}

func TestBudgetPeriodStart(t *testing.T) {
	// Default (start_day=1): behaves like BeginningOfMonth
	mustLoadConfig(1)
	assert.Equal(t, ds(2024, time.March, 1), dateStr(BudgetPeriodStart(d(2024, time.March, 10))))
	assert.Equal(t, ds(2024, time.March, 1), dateStr(BudgetPeriodStart(d(2024, time.March, 1))))

	// start_day=15: mid-month periods
	mustLoadConfig(15)
	// Mar 10 → period started Feb 15
	assert.Equal(t, ds(2024, time.February, 15), dateStr(BudgetPeriodStart(d(2024, time.March, 10))))
	// Mar 15 → period starts Mar 15
	assert.Equal(t, ds(2024, time.March, 15), dateStr(BudgetPeriodStart(d(2024, time.March, 15))))
	// Mar 20 → period starts Mar 15
	assert.Equal(t, ds(2024, time.March, 15), dateStr(BudgetPeriodStart(d(2024, time.March, 20))))
	// Jan 5 → period started Dec 15 of previous year
	assert.Equal(t, ds(2023, time.December, 15), dateStr(BudgetPeriodStart(d(2024, time.January, 5))))

	// BudgetPeriodEnd: period starting Feb 15 ends Mar 14
	assert.Equal(t, ds(2024, time.March, 14), dateStr(BudgetPeriodEnd(d(2024, time.March, 10))))

	// BudgetPeriodKey (shifted one month forward so budget entries dated the 1st
	// of a month always appear under the matching month label)
	// Mar 10 → period started Feb 15 → shifted key = Mar 15 → "2024-03"
	assert.Equal(t, "2024-03", BudgetPeriodKey(d(2024, time.March, 10)))
	// Mar 15 → period started Mar 15 → shifted key = Apr 15 → "2024-04"
	assert.Equal(t, "2024-04", BudgetPeriodKey(d(2024, time.March, 15)))
}

func TestBuildSubPath(t *testing.T) {
	path, err := BuildSubPath("/usr/home/john/paisa", "main.ledger")
	assert.Nil(t, err)
	assert.Equal(t, "/usr/home/john/paisa/main.ledger", path)

	path, err = BuildSubPath("/usr/home/john/paisa", "subfolder/main.ledger")
	assert.Nil(t, err)
	assert.Equal(t, "/usr/home/john/paisa/subfolder/main.ledger", path)

	path, err = BuildSubPath("/usr/home/john/paisa", "../../../subfolder/travel.ledger")
	assert.Error(t, err)

	path, err = BuildSubPath("/usr/home/john/paisa", "..")
	assert.Error(t, err)

	path, err = BuildSubPath("/usr/home/john/paisa", "./..")
	assert.Error(t, err)

	path, err = BuildSubPath("/usr/home/john/paisa", "./../test.ledger")
	assert.Error(t, err)
}
