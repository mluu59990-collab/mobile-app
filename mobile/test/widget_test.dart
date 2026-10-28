import 'package:expense_app/core/providers/app_providers.dart';
import 'package:expense_app/main.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  testWidgets('first launch shows the welcome screen', (tester) async {
    SharedPreferences.setMockInitialValues({});
    final preferences = await SharedPreferences.getInstance();

    await tester.pumpWidget(
      ProviderScope(
        overrides: [sharedPreferencesProvider.overrideWithValue(preferences)],
        child: const ExpenseApp(),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Bắt đầu'), findsOneWidget);
    expect(find.text('Quản lý chi tiêu\nthông minh hơn'), findsOneWidget);
  });
}
