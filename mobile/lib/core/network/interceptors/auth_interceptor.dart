import 'package:dio/dio.dart';

import '../../storage/token_storage.dart';

class AuthInterceptor extends Interceptor {
  AuthInterceptor(this._tokenStorage);

  final TokenStorage _tokenStorage;
  static const _publicPaths = {
    '/auth/login',
    '/auth/register',
    '/auth/google',
    '/auth/facebook',
    '/auth/refresh',
  };

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final token = _tokenStorage.accessToken;
    if (token != null &&
        token.isNotEmpty &&
        !_publicPaths.contains(options.path)) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }
}
