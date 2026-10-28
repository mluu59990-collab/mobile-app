package model

type CategoryExpenseTotal struct {
	Category    *string
	TotalAmount int64
}

type DailyExpenseTotal struct {
	Date        string
	TotalAmount int64
}

type MonthlyExpenseReport struct {
	TotalAmount int64
	ByCategory  []CategoryExpenseTotal
	ByDay       []DailyExpenseTotal
}
