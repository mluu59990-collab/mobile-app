import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:expense_app/core/network/interceptors/auth_interceptor.dart';
import 'package:expense_app/core/providers/auth_session.dart';
import 'package:expense_app/core/storage/token_storage.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('AuthInterceptor', () {
    test(
      'refreshes once and retries the failed request with the new token',
      () async {
        final storage = _MemoryTokenStorage(
          accessToken: 'expired-access',
          refreshToken: 'refresh-one',
        );
        final session = AuthSession(true);
        var refreshCalls = 0;
        var resourceCalls = 0;
        final refreshDio = _dio((options) async {
          refreshCalls++;
          expect(options.uri.path, endsWith('/auth/refresh'));
          expect(options.data, {'refresh_token': 'refresh-one'});
          return _jsonResponse(200, {
            'access_token': 'fresh-access',
            'refresh_token': 'refresh-two',
          });
        });
        final apiDio = _dio((options) async {
          resourceCalls++;
          if (options.headers['Authorization'] == 'Bearer fresh-access') {
            return _jsonResponse(200, {'ok': true});
          }
          return _jsonResponse(401, {
            'error': {'code': 'invalid_access_token', 'message': 'Expired'},
          });
        });
        _installInterceptor(apiDio, refreshDio, storage, session);

        final response = await apiDio.get<Map<String, dynamic>>('/expenses');

        expect(response.data, {'ok': true});
        expect(refreshCalls, 1);
        expect(resourceCalls, 2);
        expect(await storage.readAccessToken(), 'fresh-access');
        expect(await storage.readRefreshToken(), 'refresh-two');
        expect(session.isAuthenticated, isTrue);
      },
    );

    test('clears tokens and signs out when refresh fails', () async {
      final storage = _MemoryTokenStorage(
        accessToken: 'expired-access',
        refreshToken: 'invalid-refresh',
      );
      final session = AuthSession(true);
      var refreshCalls = 0;
      final refreshDio = _dio((_) async {
        refreshCalls++;
        return _jsonResponse(401, {
          'error': {'code': 'invalid_refresh_token', 'message': 'Expired'},
        });
      });
      final apiDio = _dio((_) async => _jsonResponse(401, {'error': {}}));
      _installInterceptor(apiDio, refreshDio, storage, session);

      await expectLater(
        apiDio.get<void>('/expenses'),
        throwsA(isA<DioException>()),
      );

      expect(refreshCalls, 1);
      expect(await storage.readAccessToken(), isNull);
      expect(await storage.readRefreshToken(), isNull);
      expect(session.isAuthenticated, isFalse);
    });

    test('concurrent 401 responses share one refresh request', () async {
      final storage = _MemoryTokenStorage(
        accessToken: 'expired-access',
        refreshToken: 'refresh-one',
      );
      final session = AuthSession(true);
      var refreshCalls = 0;
      var resourceCalls = 0;
      final refreshDio = _dio((_) async {
        refreshCalls++;
        await Future<void>.delayed(const Duration(milliseconds: 25));
        return _jsonResponse(200, {
          'access_token': 'fresh-access',
          'refresh_token': 'refresh-two',
        });
      });
      final apiDio = _dio((options) async {
        resourceCalls++;
        if (options.headers['Authorization'] == 'Bearer fresh-access') {
          return _jsonResponse(200, {'ok': true});
        }
        await Future<void>.delayed(
          Duration(milliseconds: options.uri.path.endsWith('/two') ? 60 : 10),
        );
        return _jsonResponse(401, {
          'error': {'code': 'expired'},
        });
      });
      _installInterceptor(apiDio, refreshDio, storage, session);

      final responses = await Future.wait([
        apiDio.get<Map<String, dynamic>>('/expenses/one'),
        apiDio.get<Map<String, dynamic>>('/expenses/two'),
      ]);

      expect(responses, hasLength(2));
      expect(
        responses.every((response) => response.data?['ok'] == true),
        isTrue,
      );
      expect(refreshCalls, 1);
      expect(resourceCalls, 4);
      expect(await storage.readAccessToken(), 'fresh-access');
    });

    test('does not refresh requests to login or refresh endpoints', () async {
      final storage = _MemoryTokenStorage(
        accessToken: 'access',
        refreshToken: 'refresh',
      );
      final session = AuthSession(true);
      var refreshCalls = 0;
      final refreshDio = _dio((_) async {
        refreshCalls++;
        return _jsonResponse(200, {
          'access_token': 'new-access',
          'refresh_token': 'new-refresh',
        });
      });
      final apiDio = _dio((_) async => _jsonResponse(401, {'error': {}}));
      _installInterceptor(apiDio, refreshDio, storage, session);

      for (final path in ['/auth/login', '/auth/refresh']) {
        await expectLater(
          apiDio.post<void>(path),
          throwsA(isA<DioException>()),
        );
      }

      expect(refreshCalls, 0);
      expect(session.isAuthenticated, isTrue);
    });
  });
}

Dio _dio(Future<ResponseBody> Function(RequestOptions) respond) {
  final dio = Dio(BaseOptions(baseUrl: 'https://api.example.test/api/v1'));
  dio.httpClientAdapter = _FakeAdapter(respond);
  return dio;
}

void _installInterceptor(
  Dio apiDio,
  Dio refreshDio,
  _MemoryTokenStorage storage,
  AuthSession session,
) {
  apiDio.interceptors.add(
    AuthInterceptor(apiDio, refreshDio, storage, session),
  );
}

ResponseBody _jsonResponse(int statusCode, Object body) =>
    ResponseBody.fromString(
      jsonEncode(body),
      statusCode,
      headers: {
        Headers.contentTypeHeader: ['application/json'],
      },
    );

class _FakeAdapter implements HttpClientAdapter {
  _FakeAdapter(this.respond);

  final Future<ResponseBody> Function(RequestOptions) respond;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) => respond(options);

  @override
  void close({bool force = false}) {}
}

class _MemoryTokenStorage extends TokenStorage {
  _MemoryTokenStorage({this.accessToken, this.refreshToken});

  String? accessToken;
  String? refreshToken;

  @override
  Future<String?> readAccessToken() async => accessToken;

  @override
  Future<String?> readRefreshToken() async => refreshToken;

  @override
  Future<void> saveTokens({
    required String accessToken,
    required String refreshToken,
  }) async {
    this.accessToken = accessToken;
    this.refreshToken = refreshToken;
  }

  @override
  Future<void> clear() async {
    accessToken = null;
    refreshToken = null;
  }
}
