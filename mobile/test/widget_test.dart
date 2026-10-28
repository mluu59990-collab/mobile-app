import 'package:expense_app/core/providers/app_providers.dart';
import 'package:expense_app/core/providers/auth_session.dart';
import 'package:expense_app/main.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  testWidgets('signed-out launch shows the login screen', (tester) async {
    SharedPreferences.setMockInitialValues({});
    final preferences = await SharedPreferences.getInstance();

    await tester.pumpWidget(
      ProviderScope(
        overrides: [sharedPreferencesProvider.overrideWithValue(preferences)],
        child: const ExpenseApp(),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Đăng nhập'), findsWidgets);
    expect(find.text('Chưa có tài khoản? '), findsOneWidget);
  });

  testWidgets('authenticated launch opens the home screen', (tester) async {
    SharedPreferences.setMockInitialValues({});
    final preferences = await SharedPreferences.getInstance();
    final authSession = AuthSession(true);

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(preferences),
          authSessionProvider.overrideWithValue(authSession),
        ],
        child: const ExpenseApp(),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Giao dịch gần đây'), findsOneWidget);
  });
}
