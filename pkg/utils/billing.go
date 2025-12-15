package utils

import (
	"time"
)

// CalculateBillingPeriod calculates the billing period for an expense based on payment method closing day
func CalculateBillingPeriod(expenseDate time.Time, closingDay int64) (start, end time.Time) {
	expenseDay := expenseDate.Day()
	expenseYear := expenseDate.Year()
	expenseMonth := expenseDate.Month()

	closingDayInt := int(closingDay)

	if expenseDay < closingDayInt {
		// Expense is before closing day, so it's in the current billing period
		// Billing period starts on the 1st of current month (or day after previous closing)
		start = time.Date(expenseYear, expenseMonth, 1, 0, 0, 0, 0, expenseDate.Location())
		
		// Billing period ends on closing day of current month
		lastDayOfMonth := time.Date(expenseYear, expenseMonth+1, 0, 0, 0, 0, 0, expenseDate.Location()).Day()
		endDay := closingDayInt
		if endDay > lastDayOfMonth {
			endDay = lastDayOfMonth
		}
		end = time.Date(expenseYear, expenseMonth, endDay, 23, 59, 59, 999999999, expenseDate.Location())
	} else {
		// Expense is on or after closing day, so it's in the next billing period
		// Billing period starts on the day after closing day of current month
		start = time.Date(expenseYear, expenseMonth, closingDayInt+1, 0, 0, 0, 0, expenseDate.Location())
		
		// Billing period ends on closing day of next month
		nextMonth := expenseMonth + 1
		nextYear := expenseYear
		if nextMonth > 12 {
			nextMonth = 1
			nextYear++
		}
		lastDayOfNextMonth := time.Date(nextYear, nextMonth+1, 0, 0, 0, 0, 0, expenseDate.Location()).Day()
		endDay := closingDayInt
		if endDay > lastDayOfNextMonth {
			endDay = lastDayOfNextMonth
		}
		end = time.Date(nextYear, nextMonth, endDay, 23, 59, 59, 999999999, expenseDate.Location())
	}

	return start, end
}

// GetBillingPeriodForMonth returns the billing period dates for a given month and closing day
func GetBillingPeriodForMonth(year int, month time.Month, closingDay int) (start, end time.Time) {
	// Previous period ends on closing day of previous month
	prevMonth := month - 1
	prevYear := year
	if prevMonth < 1 {
		prevMonth = 12
		prevYear--
	}
	
	lastDayOfPrevMonth := time.Date(prevYear, prevMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()
	endDayPrev := closingDay
	if endDayPrev > lastDayOfPrevMonth {
		endDayPrev = lastDayOfPrevMonth
	}
	
	// Current period starts day after previous closing
	start = time.Date(year, month, closingDay+1, 0, 0, 0, 0, time.UTC)
	if closingDay == lastDayOfPrevMonth {
		start = time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	}
	
	// Current period ends on closing day of current month
	lastDayOfMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	endDay := closingDay
	if endDay > lastDayOfMonth {
		endDay = lastDayOfMonth
	}
	end = time.Date(year, month, endDay, 23, 59, 59, 999999999, time.UTC)
	
	return start, end
}

