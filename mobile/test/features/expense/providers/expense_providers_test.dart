import 'package:expense_app/features/expense/providers/expense_providers.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('default month uses the device local date', () {
    final localNow = DateTime(2026, 11, 1, 0, 10);

    expect(currentLocalMonth(localNow), '2026-11');
    expect(currentLocalMonth(DateTime.now()), monthKey(DateTime.now()));
  });
}
