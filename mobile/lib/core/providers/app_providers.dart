import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:logger/logger.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../network/dio_client.dart';
import '../storage/app_launch_storage.dart';
import '../storage/token_storage.dart';

final sharedPreferencesProvider = Provider<SharedPreferences>((ref) {
  throw StateError('SharedPreferences must be overridden during app startup.');
});

final tokenStorageProvider = Provider<TokenStorage>(
  (ref) => TokenStorage(ref.watch(sharedPreferencesProvider)),
);

final appLaunchStorageProvider = Provider<AppLaunchStorage>(
  (ref) => AppLaunchStorage(ref.watch(sharedPreferencesProvider)),
);

final loggerProvider = Provider<Logger>((ref) => Logger());

final dioProvider = Provider<Dio>(
  (ref) => createDioClient(
    ref.watch(tokenStorageProvider),
    logger: ref.watch(loggerProvider),
  ),
);
