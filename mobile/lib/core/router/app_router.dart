import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/app_providers.dart';
import '../../features/auth/pages/login_page.dart';
import '../../features/auth/pages/register_page.dart';
import '../../features/home/pages/home_page.dart';
import '../../features/expense/pages/add_expense_page.dart';
import '../../features/settings/pages/settings_page.dart';
import '../../features/splash/pages/splash_page.dart';
import '../../features/welcome/pages/welcome_page.dart';

final appRouterProvider = Provider<GoRouter>((ref) {
  final session = ref.watch(authSessionProvider);
  final router = GoRouter(
    initialLocation: '/splash',
    refreshListenable: session,
    routes: [
      GoRoute(path: '/splash', builder: (context, state) => const SplashPage()),
      GoRoute(
        path: '/welcome',
        builder: (context, state) => const WelcomePage(),
      ),
      GoRoute(path: '/login', builder: (context, state) => const LoginPage()),
      GoRoute(
        path: '/register',
        builder: (context, state) => const RegisterPage(),
      ),
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
      final isAuthPage = {
        '/splash',
        '/welcome',
        '/login',
        '/register',
      }.contains(path);
      if (!session.isAuthenticated && !isAuthPage) return '/login';
      if (!session.isAuthenticated && path != '/login') return '/login';
      if (session.isAuthenticated && isAuthPage) return '/home';
      return null;
    },
  );
  ref.onDispose(router.dispose);
  return router;
});
