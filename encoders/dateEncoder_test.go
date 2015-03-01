package encoders

import (
	"github.com/nupic-community/htm/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSimpleDateEncoding(t *testing.T) {

	p := NewDateEncoderParams()
	p.SeasonWidth = 3
	p.DayOfWeekWidth = 1
	p.WeekendWidth = 3
	p.TimeOfDayWidth = 5
	de := NewDateEncoder(p)

	// season is aaabbbcccddd (1 bit/month)  TODO should be <<3?
	// should be 000000000111 (centered on month 11 - Nov)
	seasonExpected := utils.Make1DBool([]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1})
	// week is SMTWTFS
	// differs from python implementation
	dayOfWeekExpected := utils.Make1DBool([]int{0, 0, 0, 0, 1, 0, 0})
	// not a weekend, so it should be "False"
	weekendExpected := utils.Make1DBool([]int{1, 1, 1, 0, 0, 0})
	// time of day has radius of 4 hours and w of 5 so each bit = 240/5 min = 48min
	// 14:55 is minute 14*60 + 55 = 895; 895/48 = bit 18.6
	// should be 30 bits total (30 * 48 minutes = 24 hours)
	timeOfDayExpected := utils.Make1DBool([]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0})

	d := time.Date(2010, 11, 4, 14, 55, 0, 0, time.UTC)
	encoded := de.Encode(d)
	t.Log(utils.Bool2Int(encoded))

	expected := append(seasonExpected, dayOfWeekExpected...)
	expected = append(expected, weekendExpected...)
	expected = append(expected, timeOfDayExpected...)

	assert.Equal(t, utils.Bool2Int(expected), utils.Bool2Int(encoded))

}
