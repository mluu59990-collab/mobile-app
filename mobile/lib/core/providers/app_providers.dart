import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:logger/logger.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../network/dio_client.dart';
import '../storage/token_storage.dart';
import '../storage/app_launch_storage.dart';
import '../../features/auth/data/auth_repository.dart';
import '../timezone/device_timezone.dart';
import 'auth_session.dart';

final sharedPreferencesProvider = Provider<SharedPreferences>((ref) {
  throw StateError('SharedPreferences must be overridden during app startup.');
});

final tokenStorageProvider = Provider<TokenStorage>((ref) => TokenStorage());

final authSessionProvider = Provider<AuthSession>((ref) {
  final session = AuthSession();
  ref.onDispose(session.dispose);
  return session;
});

final appLaunchStorageProvider = Provider<AppLaunchStorage>(
  (ref) => AppLaunchStorage(ref.watch(sharedPreferencesProvider)),
);

final loggerProvider = Provider<Logger>((ref) => Logger());

final deviceTimezoneProvider = Provider<DeviceTimezone>(
  (ref) => const DeviceTimezone('UTC'),
);

final dioProvider = Provider<Dio>(
  (ref) => createDioClient(
    ref.watch(tokenStorageProvider),
    ref.watch(authSessionProvider),
    logger: ref.watch(loggerProvider),
  ),
);

final authRepositoryProvider = Provider<AuthRepository>(
  (ref) => AuthRepository(
    ref.watch(dioProvider),
    ref.watch(tokenStorageProvider),
    ref.watch(authSessionProvider),
  ),
);
