import 'package:dio/dio.dart';

import '../../providers/auth_session.dart';
import '../../storage/token_storage.dart';

class AuthInterceptor extends Interceptor {
  AuthInterceptor(
    this._apiDio,
    this._refreshDio,
    this._tokenStorage,
    this._session,
  );

  static const _retriedKey = 'auth_retry_attempted';
  static const _excludedAuthPaths = [
    '/auth/login',
    '/auth/register',
    '/auth/refresh',
    '/auth/logout',
  ];

  final Dio _apiDio;
  final Dio _refreshDio;
  final TokenStorage _tokenStorage;
  final AuthSession _session;
  Future<_TokenPair>? _refreshInFlight;

  @override
  void onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final accessToken = await _tokenStorage.readAccessToken();
    if (accessToken != null &&
        accessToken.isNotEmpty &&
        !_isExcludedAuthPath(options.uri.path)) {
      options.headers['Authorization'] = 'Bearer $accessToken';
    }
    handler.next(options);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    final request = err.requestOptions;
    if (err.response?.statusCode != 401 ||
        request.extra[_retriedKey] == true ||
        _isExcludedAuthPath(request.uri.path)) {
      handler.next(err);
      return;
    }

    late final _TokenPair tokens;
    try {
      final currentAccessToken = await _tokenStorage.readAccessToken();
      final sentAccessToken = _bearerToken(request.headers['Authorization']);
      if (currentAccessToken != null &&
          currentAccessToken.isNotEmpty &&
          sentAccessToken != null &&
          sentAccessToken != currentAccessToken) {
        tokens = _TokenPair(
          accessToken: currentAccessToken,
          refreshToken: await _tokenStorage.readRefreshToken() ?? '',
        );
      } else {
        final refreshToken = await _tokenStorage.readRefreshToken();
        if (refreshToken == null || refreshToken.isEmpty) {
          throw const _RefreshFailed();
        }
        tokens = await _refreshTokens(refreshToken);
      }
    } catch (_) {
      await _signOut();
      handler.next(err);
      return;
    }

    try {
      final retryRequest = request.copyWith(
        headers: {
          ...request.headers,
          'Authorization': 'Bearer ${tokens.accessToken}',
        },
        extra: {...request.extra, _retriedKey: true},
      );
      final response = await _apiDio.fetch<dynamic>(retryRequest);
      handler.resolve(response);
    } on DioException catch (retryError) {
      if (retryError.response?.statusCode == 401) await _signOut();
      handler.next(retryError);
    }
  }

  Future<void> _signOut() async {
    try {
      await _tokenStorage.clear();
    } catch (_) {
      // Still move the router to login if secure storage itself is unavailable.
    } finally {
      _session.markUnauthenticated();
    }
  }

  bool _isExcludedAuthPath(String path) {
    final normalizedPath = path.endsWith('/')
        ? path.substring(0, path.length - 1)
        : path;
    return _excludedAuthPaths.any(normalizedPath.endsWith);
  }

  String? _bearerToken(Object? authorization) {
    if (authorization is! String || !authorization.startsWith('Bearer ')) {
      return null;
    }
    return authorization.substring('Bearer '.length);
  }

  Future<_TokenPair> _refreshTokens(String refreshToken) {
    final pending = _refreshInFlight;
    if (pending != null) return pending;

    final refresh = _performRefresh(refreshToken);
    _refreshInFlight = refresh;
    return refresh.whenComplete(() {
      if (identical(_refreshInFlight, refresh)) _refreshInFlight = null;
    });
  }

  Future<_TokenPair> _performRefresh(String refreshToken) async {
    final response = await _refreshDio.post<dynamic>(
      '/auth/refresh',
      data: {'refresh_token': refreshToken},
    );
    final data = response.data;
    if (data is! Map<String, dynamic> ||
        data['access_token'] is! String ||
        data['refresh_token'] is! String) {
      throw const _RefreshFailed();
    }

    final tokens = _TokenPair(
      accessToken: data['access_token'] as String,
      refreshToken: data['refresh_token'] as String,
    );
    if (tokens.accessToken.isEmpty || tokens.refreshToken.isEmpty) {
      throw const _RefreshFailed();
    }
    await _tokenStorage.saveTokens(
      accessToken: tokens.accessToken,
      refreshToken: tokens.refreshToken,
    );
    return tokens;
  }
}

class _TokenPair {
  const _TokenPair({required this.accessToken, required this.refreshToken});

  final String accessToken;
  final String refreshToken;
}

class _RefreshFailed implements Exception {
  const _RefreshFailed();
}
