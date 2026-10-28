import 'package:dio/dio.dart';
import 'package:logger/logger.dart';

class SafeLoggingInterceptor extends Interceptor {
  SafeLoggingInterceptor(this._logger);

  final Logger _logger;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    _logger.d('${options.method} ${_safeUri(options.uri)}');
    handler.next(options);
  }

  @override
  void onResponse(
    Response<dynamic> response,
    ResponseInterceptorHandler handler,
  ) {
    _logger.d(
      '${response.statusCode} ${_safeUri(response.requestOptions.uri)}',
    );
    handler.next(response);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    _logger.w(
      '${err.response?.statusCode ?? 'network'} ${_safeUri(err.requestOptions.uri)}',
    );
    handler.next(err);
  }

  String _safeUri(Uri uri) => '${uri.origin}${uri.path}';
}
