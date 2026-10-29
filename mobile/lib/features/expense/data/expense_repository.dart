import 'package:dio/dio.dart';

import '../../../core/network/error_message.dart';
import 'expense.dart';
import 'monthly_report.dart';

class ExpenseRepository {
  ExpenseRepository(this._dio, {this.timezone = 'UTC'});
  final Dio _dio;
  final String timezone;

  Future<List<Expense>> list(String month) async => _guard(() async {
    final response = await _dio.get(
      '/expenses',
      queryParameters: {'month': month, 'tz': timezone},
    );
    final items =
        (response.data as Map<String, dynamic>)['expenses'] as List<dynamic>;
    return items
        .map((item) => Expense.fromJson(item as Map<String, dynamic>))
        .toList();
  });

  Future<Expense> get(String id) async => _guard(() async {
    final response = await _dio.get('/expenses/$id');
    return Expense.fromJson(response.data as Map<String, dynamic>);
  });

  Future<Expense> create(ExpenseInput input) async => _guard(() async {
    final response = await _dio.post('/expenses', data: input.toJson());
    return Expense.fromJson(response.data as Map<String, dynamic>);
  });

  Future<Expense> update(String id, ExpenseInput input) async => _guard(
    () async {
      final response = await _dio.patch('/expenses/$id', data: input.toJson());
      return Expense.fromJson(response.data as Map<String, dynamic>);
    },
  );

  Future<void> delete(String id) async => _guard(() async {
    await _dio.delete('/expenses/$id');
  });

  Future<MonthlyReport> monthlyReport(String month) async => _guard(() async {
    final response = await _dio.get(
      '/reports/monthly',
      queryParameters: {'month': month, 'tz': timezone},
    );
    return MonthlyReport.fromJson(response.data as Map<String, dynamic>);
  });

  Future<T> _guard<T>(Future<T> Function() request) async {
    try {
      return await request();
    } on DioException catch (error) {
      throw apiExceptionFromDio(error);
    }
  }
}
