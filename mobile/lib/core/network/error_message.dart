import 'package:dio/dio.dart';

import 'api_exception.dart';

String errorMessageFromDio(DioException error) {
  final data = error.response?.data;
  if (data is Map<String, dynamic>) {
    final payload = data['error'];
    if (payload is Map<String, dynamic> && payload['message'] is String) {
      return payload['message'] as String;
    }
  }
  return ApiException(
    error.message ?? 'Không thể kết nối đến máy chủ.',
    statusCode: error.response?.statusCode,
  ).message;
}
