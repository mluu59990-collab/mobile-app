class CategoryTotal {
  const CategoryTotal({required this.category, required this.totalAmount});
  final String? category;
  final int totalAmount;
}

class DayTotal {
  const DayTotal({required this.date, required this.totalAmount});
  final String date;
  final int totalAmount;
}

class MonthlyReport {
  const MonthlyReport({
    required this.month,
    required this.totalAmount,
    required this.byCategory,
    required this.byDay,
  });
  final String month;
  final int totalAmount;
  final List<CategoryTotal> byCategory;
  final List<DayTotal> byDay;

  factory MonthlyReport.fromJson(Map<String, dynamic> json) => MonthlyReport(
    month: json['month'] as String,
    totalAmount: (json['total_amount'] as num).toInt(),
    byCategory: (json['by_category'] as List<dynamic>? ?? const []).map((item) {
      final row = item as Map<String, dynamic>;
      return CategoryTotal(
        category: row['category'] as String?,
        totalAmount: (row['total_amount'] as num).toInt(),
      );
    }).toList(),
    byDay: (json['by_day'] as List<dynamic>? ?? const []).map((item) {
      final row = item as Map<String, dynamic>;
      return DayTotal(
        date: row['date'] as String,
        totalAmount: (row['total_amount'] as num).toInt(),
      );
    }).toList(),
  );
}
