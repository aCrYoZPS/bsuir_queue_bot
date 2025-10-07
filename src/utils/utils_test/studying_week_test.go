package utilstest

import (
	"testing"
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/utils"
)

func TestCalculateWeek(t *testing.T) {
	tested_date := time.Date(2025, time.October, 22, 0, 0, 0, 0, time.Local)
	want := int8(4)
	result := utils.CalculateWeek(tested_date)
	if want != result {
		t.Errorf(`CalculateWeek(%s) = %d, want %d`, tested_date.String(), result, want)
	}

	tested_date = time.Date(2025, time.October, 22, 23, 59, 59, 59, time.Local)
	want = int8(4)
	result = utils.CalculateWeek(tested_date)

	if want != result {
		t.Errorf(`CalculateWeek(%s) = %d, want %d`, tested_date.String(), result, want)
	}
	tested_date = time.Date(2025, time.October, 20, 0, 0, 0, 0, time.Local)
	want = 4
	for i := range 100000 {
		if i%7 == 6 {
			want = (want % 4) + 1
		}
		tested_date = tested_date.AddDate(0, 0, 1)
		result = utils.CalculateWeek(tested_date)
		if want != result {
			t.Errorf(`CalculateWeek(%s) = %d, want %d`, tested_date.String(), result, want)
		}
	}
}
