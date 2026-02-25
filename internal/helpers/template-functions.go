package helpers

import "time"

//addTemplateFunctions registers custom template helper functions
func addTemplateFunctions() {
	//Format date as YYYY-MM-DD
	views.AddGlobal("humanDate", func(t time.Time) string {
		return HumanDate(t)
	})

	//Format date using custom layout
	views.AddGlobal("dateFromLayout", func(t time.Time, l string) string {
		return FormatDateWithLayout(t, 1)
	})

	//Check if date is after year 1
	views.AddGlobal("dateAfterYearOne", func(t time.Time) bool {
		return DateAfterY1(t)
	})

}

//HumanDate returns date in YYYY-MM-DD format
func HumanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

//FormatDateWithLayout formats time using provided Go layout
func FormatDateWithLayout(t time.Time, f string) string {
	return t.Format(f)
}

//DateAfterY1 checks if date is after Go's year 1 default
func DateAfterY1(t time.Time) bool {
	yearOne := time.Date(0001, 11, 17, 20, 34, 58, 651387237, time.UTC)
	return t.After(yearOne)
}
