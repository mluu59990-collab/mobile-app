import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/providers/app_providers.dart';
import '../data/expense.dart';
import '../data/expense_repository.dart';
import '../data/monthly_report.dart';

String monthKey(DateTime date) =>
    '${date.year.toString().padLeft(4, '0')}-${date.month.toString().padLeft(2, '0')}';

String currentLocalMonth(DateTime localNow) => monthKey(localNow.toLocal());

final selectedExpenseMonthProvider =
    NotifierProvider<SelectedExpenseMonth, String>(SelectedExpenseMonth.new);

class SelectedExpenseMonth extends Notifier<String> {
  @override
  String build() => currentLocalMonth(DateTime.now());

  void select(String month) => state = month;

  void shift(int delta) {
    final parts = state.split('-').map(int.parse).toList();
    final date = DateTime(parts[0], parts[1] + delta);
    state = monthKey(date);
  }
}

final expenseRepositoryProvider = Provider<ExpenseRepository>(
  (ref) => ExpenseRepository(
    ref.watch(dioProvider),
    timezone: ref.watch(deviceTimezoneProvider).name,
  ),
);

final expensesForSelectedMonthProvider =
    FutureProvider.autoDispose<List<Expense>>(
      (ref) => ref
          .watch(expenseRepositoryProvider)
          .list(ref.watch(selectedExpenseMonthProvider)),
    );

final monthlyReportProvider = FutureProvider.autoDispose<MonthlyReport>(
  (ref) => ref
      .watch(expenseRepositoryProvider)
      .monthlyReport(ref.watch(selectedExpenseMonthProvider)),
);

final expenseByIdProvider = FutureProvider.autoDispose.family<Expense, String>(
  (ref, id) => ref.watch(expenseRepositoryProvider).get(id),
);

final expenseActionsProvider = Provider<ExpenseActions>(
  (ref) => ExpenseActions(ref, ref.watch(expenseRepositoryProvider)),
);

class ExpenseActions {
  ExpenseActions(this._ref, this._repository);
  final Ref _ref;
  final ExpenseRepository _repository;

  Future<Expense> create(ExpenseInput input) async {
    final result = await _repository.create(input);
    _invalidate();
    return result;
  }

  Future<Expense> update(String id, ExpenseInput input) async {
    final result = await _repository.update(id, input);
    _invalidate();
    _ref.invalidate(expenseByIdProvider(id));
    return result;
  }

  Future<void> delete(String id) async {
    await _repository.delete(id);
    _invalidate();
  }

  void _invalidate() {
    _ref.invalidate(expensesForSelectedMonthProvider);
    _ref.invalidate(monthlyReportProvider);
  }
}
