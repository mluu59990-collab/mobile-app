import 'package:expense_app/core/formatters/vnd_formatter.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('formats whole VND with dot groups and no decimals', () {
    expect(formatVnd(3450000), '3.450.000 ₫');
    expect(formatVnd(0), '0 ₫');
    expect(formatVnd(-25000), '-25.000 ₫');
  });
}
