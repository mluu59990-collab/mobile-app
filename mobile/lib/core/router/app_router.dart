import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/app_providers.dart';
import '../../features/auth/pages/login_page.dart';
import '../../features/home/pages/home_page.dart';
import '../../features/expense/pages/add_expense_page.dart';
import '../../features/settings/pages/settings_page.dart';
import '../../features/splash/pages/splash_page.dart';
import '../../features/welcome/pages/welcome_page.dart';

final appRouterProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/splash',
    routes: [
      GoRoute(path: '/splash', builder: (context, state) => const SplashPage()),
      GoRoute(
        path: '/welcome',
        builder: (context, state) => const WelcomePage(),
      ),
      GoRoute(path: '/login', builder: (context, state) => const LoginPage()),
      GoRoute(path: '/home', builder: (context, state) => const HomePage()),
      GoRoute(
        path: '/expense/new',
        builder: (context, state) => const AddExpensePage(),
      ),
      GoRoute(
        path: '/settings',
        builder: (context, state) => const SettingsPage(),
      ),
    ],
    redirect: (context, state) {
      final path = state.uri.path;
      if (path == '/splash') return null;
      final storage = ref.read(appLaunchStorageProvider);
      final token = ref.read(tokenStorageProvider).accessToken;
      if (token != null && token.isNotEmpty && path != '/home') return '/home';
      if (!storage.hasSeenWelcome && path != '/welcome') return '/welcome';
      if (storage.hasSeenWelcome && path == '/welcome') return '/login';
      if (path == '/home' && (token == null || token.isEmpty)) return '/login';
      return null;
    },
  );
});
