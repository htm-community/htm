package encoders

import (
	"fmt"
	//"github.com/cznic/mathutil"
	//"github.com/zacg/floats"
	//"github.com/nupic-community/htm"
	//"github.com/nupic-community/htm/utils"
	//"github.com/zacg/ints"
	//"math"
)

/*
	Params for the date encoder
*/
type DateEncoderParams struct {
	HolidayWidth    int
	HolidayRadius   float64
	SeasonWidth     int
	SeasonRadius    float64
	DayOfWeekWidth  int
	DayOfWeekRadius float64
	WeekendWidth    int
	WeekendRadius   float64
	TimeOfDayWidth  int
	TimeOfDayRadius float64
	//CustomDays     int
	Name string
}

func NewDateEncoderParams() *DateEncoderParams {
	p := new(DateEncoderParams)

	//set defaults
	p.SeasonRadius = 91.5 //days
	p.DayOfWeekRadius = 1
	p.TimeOfDayRadius = 4
	p.WeekendRadius = 1
	p.HolidayRadius = 1

	return p
}

/*
	Date encoder encodes a datetime to a SDR. Params allow for tuning
	for specific date attributes
*/
type DateEncoder struct {
	DateEncoderParams
	seasonEncoder    *ScalerEncoder
	holidayEncoder   *ScalerEncoder
	dayOfWeekEncoder *ScalerEncoder
	weekendEncoder   *ScalerEncoder
	timeOfDayEncoder *ScalerEncoder

	width           int
	seasonOffset    int
	weekendOffset   int
	dayOfWeekOffset int
	holidayOffset   int
	timeOfDayOffset int
	Description     string
}

/*
	Intializes a new date encoder
*/
func NewDateEncoder(params *DateEncoderParams) *DateEncoderParams {
	se := new(DateEncoder)

	se.width = 0

	if params.SeasonWidth != 0 {
		// Ignore leapyear differences -- assume 366 days in a year
		// Radius = 91.5 days = length of season
		// Value is number of days since beginning of year (0 - 355)

		sep := NewScalerEncoderParams(params.SeasonWidth, 0, 366)
		sep.Name = "Season"
		sep.Periodic = true
		se.seasonEncoder = NewScalerEncoder(sep)
		se.seasonOffset = se.params.Width
		se.width += se.seasonEncoder.params.Width
		se.Description += fmt.Sprintf("season %v", se.seasonOffset)
	}

	if params.DayOfWeekWidth != 0 {
		// Value is day of week (floating point)
		// Radius is 1 day

		sep := NewScalerEncoderParams(params.DayOfWeekWidth, 0, 7)
		sep.Radius = params.DayOfWeekRadius
		sep.Name = "day of week"
		se.dayOfWeekEncoder = NewScalerEncoder(sep)
		se.dayOfWeekOffset = se.dayOfWeekEncoder.getWidth()
		se.Description += fmt.Sprintf(" day of week: %v", se.dayOfWeekOffset)
		se.width += se.dayOfWeekEncoder.Width
	}

	if params.WeekendWidth != 0 {
		// Binary value. Not sure if this makes sense. Also is somewhat redundant
		// with dayOfWeek
		//Append radius if it was not provided

		sep := NewScalerEncoderParams(params.WeekendWidth, 0, 1)
		sep.Name = "weekend"
		sep.Radius = params.WeekendRadius
		se.weekendEncoder = NewScalerEncoder(sep)
		se.width += se.weekendEncoder.getWidth()
		se.weekendOffset = se.weekendEncoder.getWidth()
		se.Description += fmt.Sprintf("weekend: %v", se.weekendOffset)

	}

	if params.HolidayWidth > 0 {
		// A "continuous" binary value. = 1 on the holiday itself and smooth ramp
		// 0->1 on the day before the holiday and 1->0 on the day after the holiday.

		sep := NewScalerEncoderParams(params.HolidayWidth, 0, 1)
		sep.Name = "holiday"
		sep.Radius = params.HolidayRadius
		se.holidayEncoder = NewScalerEncoder(sep)
		se.width += se.holidaEncoder.getWidth()
		se.holidayOffset = se.holidayEncoder.getWidth()
		se.description += fmt.Sprintf(" holiday %v", se.holidayOffset)
	}

	if params.TimeOfDayWidth > 0 {
		// Value is time of day in hours
		// Radius = 4 hours, e.g. morning, afternoon, evening, early night,
		// late night, etc.

		sep := NewScalerEncoderParams(params.TimeOfDayWidth, 0, 24)
		sep.Name = "time of day"
		sep.Radius = params.TimeOfDayRadius
		sep.Periodic = true
		se.timeOfDayEncoder = NewScalerEncoder(sep)
		se.width += se.timeOfDayEncoder.getWidth()
		se.timeOfDayOffset = se.timeOfDayEncoder.Width
		se.Description += fmt.Sprintf(" time of day: %v ", se.timeOfDayOffset)
	}

	return se
}
