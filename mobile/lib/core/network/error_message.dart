import 'package:dio/dio.dart';

import 'api_exception.dart';

String errorMessageFromDio(DioException error) {
  return apiExceptionFromDio(error).message;
}

ApiException apiExceptionFromDio(DioException error) {
  final data = error.response?.data;
  if (data is Map<String, dynamic>) {
    final payload = data['error'];
    if (payload is Map<String, dynamic>) {
      final message = payload['message'];
      if (message is String && message.isNotEmpty) {
        return ApiException(
          message,
          code: payload['code'] is String ? payload['code'] as String : null,
          statusCode: error.response?.statusCode,
        );
      }
    }
  }
  return ApiException(
    error.message ?? 'Không thể kết nối đến máy chủ.',
    statusCode: error.response?.statusCode,
  );
}
