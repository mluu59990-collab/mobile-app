import 'package:flutter/material.dart';

class ExpenseCategory {
  const ExpenseCategory(this.key, this.label, this.icon);
  final String key;
  final String label;
  final IconData icon;
}

const expenseCategories = <ExpenseCategory>[
  ExpenseCategory('food', 'Ăn uống', Icons.restaurant),
  ExpenseCategory('drink', 'Đồ uống', Icons.local_cafe),
  ExpenseCategory('shopping', 'Mua sắm', Icons.shopping_bag),
  ExpenseCategory('transport', 'Di chuyển', Icons.directions_car),
  ExpenseCategory('bills', 'Hóa đơn', Icons.receipt_long),
  ExpenseCategory('entertainment', 'Giải trí', Icons.movie),
  ExpenseCategory('health', 'Sức khỏe', Icons.favorite),
  ExpenseCategory('other', 'Khác', Icons.more_horiz),
];

ExpenseCategory categoryFor(String? key) => expenseCategories.firstWhere(
  (category) => category.key == key,
  orElse: () => expenseCategories.last,
);

class Expense {
  const Expense({
    required this.id,
    required this.amount,
    required this.spentAt,
    this.note,
    this.category,
  });

  final String id;
  final int amount;
  final String? note;
  final String? category;
  final DateTime spentAt;

  factory Expense.fromJson(Map<String, dynamic> json) => Expense(
    id: json['id'] as String,
    amount: (json['amount'] as num).toInt(),
    note: json['note'] as String?,
    category: json['category'] as String?,
    spentAt: DateTime.parse(json['spent_at'] as String).toUtc(),
  );

  Map<String, dynamic> toJson() => {
    'amount': amount,
    'note': note,
    'category': category,
    'spent_at': spentAt.toUtc().toIso8601String(),
  };
}

class ExpenseInput {
  const ExpenseInput({
    required this.amount,
    required this.spentAt,
    this.note,
    this.category,
  });
  final int amount;
  final String? note;
  final String? category;
  final DateTime spentAt;

  Map<String, dynamic> toJson() => {
    'amount': amount,
    'note': note,
    'category': category,
    'spent_at': spentAt.toUtc().toIso8601String(),
  };
}
