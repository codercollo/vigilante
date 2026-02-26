package helpers

import "time"

//Package helper provides utility functions for working with time values
//and registers template functions for formating dates in Jet templates

func addTemplateFunctions() {
	//Add a gloabal template function to format a time in YYYY-MM-DD format
	views.AddGlobal("humanDate", func(t time.Time) string {
		return HumanDate(t)
	})

	//Add a global template function to format a time using a custom Go layout
	views.AddGlobal("dateFromLayout", func(t time.Time, l string) string {
		return FormatDateWithLayout(t, l)
	})

	//Add a global template function to check if a date is after year 1
	views.AddGlobal("dateAfterYearOne", func(t time.Time) bool {
		return DateAfterY1(t)
	})
}

// HumanDate formats a time in YYYY-MM-DD format
func HumanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// FormatDateWithLayout formats a time with provided (go compliant) format string, and returns it as a string
func FormatDateWithLayout(t time.Time, f string) string {
	return t.Format(f)
}

// DateAfterY1 is used to verify that a date is after the year 1 (since go hates nulls)
func DateAfterY1(t time.Time) bool {
	yearOne := time.Date(0001, 11, 17, 20, 34, 58, 651387237, time.UTC)
	return t.After(yearOne)
}
