import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:logger/logger.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'core/providers/app_providers.dart';
import 'core/providers/auth_session.dart';
import 'core/router/app_router.dart';
import 'core/storage/token_storage.dart';
import 'core/timezone/device_timezone.dart';
import 'core/ui/app_theme.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final preferences = await SharedPreferences.getInstance();
  final deviceTimezone = await DeviceTimezone.load();
  final tokenStorage = TokenStorage();
  var isAuthenticated = false;
  try {
    isAuthenticated = await tokenStorage.hasSession();
  } catch (_) {
    Logger().w('Could not read the saved session; starting signed out.');
  }
  final authSession = AuthSession(isAuthenticated);
  runApp(
    ProviderScope(
      overrides: [
        sharedPreferencesProvider.overrideWithValue(preferences),
        tokenStorageProvider.overrideWithValue(tokenStorage),
        authSessionProvider.overrideWithValue(authSession),
        deviceTimezoneProvider.overrideWithValue(deviceTimezone),
      ],
      child: const ExpenseApp(),
    ),
  );
}

class ExpenseApp extends ConsumerWidget {
  const ExpenseApp({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(appRouterProvider);
    return MaterialApp.router(
      title: 'Expense App',
      theme: AppTheme.light,
      darkTheme: AppTheme.dark,
      themeMode: ThemeMode.system,
      routerConfig: router,
    );
  }
}
