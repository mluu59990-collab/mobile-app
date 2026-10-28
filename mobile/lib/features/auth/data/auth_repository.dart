import 'package:dio/dio.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/network/error_message.dart';
import '../../../core/providers/auth_session.dart';
import '../../../core/storage/token_storage.dart';

class AuthRepository {
  const AuthRepository(this._dio, this._tokenStorage, this._session);

  final Dio _dio;
  final TokenStorage _tokenStorage;
  final AuthSession _session;

  Future<void> login({required String email, required String password}) =>
      _authenticate('login', email: email, password: password);

  Future<void> register({required String email, required String password}) =>
      _authenticate('register', email: email, password: password);

  Future<void> logout() async {
    try {
      final refreshToken = await _tokenStorage.readRefreshToken();
      if (refreshToken != null && refreshToken.isNotEmpty) {
        await _dio.post<void>(
          '/auth/logout',
          data: {'refresh_token': refreshToken},
        );
      }
    } on DioException {
      // Always clear local credentials even if the server cannot be reached.
    } finally {
      try {
        await _tokenStorage.clear();
      } finally {
        _session.markUnauthenticated();
      }
    }
  }

  Future<void> _authenticate(
    String endpoint, {
    required String email,
    required String password,
  }) async {
    try {
      final response = await _dio.post<dynamic>(
        '/auth/$endpoint',
        data: {'email': email.trim(), 'password': password},
      );
      final data = response.data;
      if (data is! Map<String, dynamic> ||
          data['access_token'] is! String ||
          data['refresh_token'] is! String ||
          (data['access_token'] as String).isEmpty ||
          (data['refresh_token'] as String).isEmpty) {
        throw const ApiException('Phản hồi đăng nhập từ máy chủ không hợp lệ.');
      }

      await _tokenStorage.saveTokens(
        accessToken: data['access_token'] as String,
        refreshToken: data['refresh_token'] as String,
      );
      _session.markAuthenticated();
    } on DioException catch (error) {
      throw apiExceptionFromDio(error);
    }
  }
}
