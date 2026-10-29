import 'package:dio/dio.dart';
import 'package:expense_app/core/network/api_exception.dart';
import 'package:expense_app/features/expense/data/expense_repository.dart';
import 'package:flutter_test/flutter_test.dart';

class _Adapter implements HttpClientAdapter {
  _Adapter(this.handler);
  final ResponseBody Function(RequestOptions) handler;
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>?>? requestStream,
    Future<void>? cancelFuture,
  ) async => handler(options);
  @override
  void close({bool force = false}) {}
}

ResponseBody jsonResponse(String body, {int status = 200}) =>
    ResponseBody.fromString(
      body,
      status,
      headers: {
        Headers.contentTypeHeader: ['application/json'],
      },
    );

void main() {
  test('parses expense list and report payloads', () async {
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost/api/v1'))
      ..httpClientAdapter = _Adapter(
        (request) => jsonResponse(
          request.path.endsWith('expenses')
              ? '{"expenses":[{"id":"id-1","amount":25000,"note":"Lunch","category":"food","spent_at":"2025-02-03T04:05:00Z"}]}'
              : '{"month":"2025-02","total_amount":25000,"by_category":[{"category":"food","total_amount":25000}],"by_day":[{"date":"2025-02-03","total_amount":25000}]}',
        ),
      );
    final repository = ExpenseRepository(dio, timezone: 'Asia/Ho_Chi_Minh');

    final expenses = await repository.list('2025-02');
    expect(expenses.single.amount, 25000);
    expect(expenses.single.spentAt.isUtc, isTrue);
    final report = await repository.monthlyReport('2025-02');
    expect(report.totalAmount, 25000);
    expect(report.byCategory.single.category, 'food');
    expect(report.byDay.single.totalAmount, 25000);
    expect(report.byDay.single.date, '2025-02-03');
  });

  test('adds the configured timezone to both month queries', () async {
    final queries = <String, Map<String, dynamic>>{};
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost/api/v1'))
      ..httpClientAdapter = _Adapter((request) {
        queries[request.path] = request.queryParameters;
        return jsonResponse(
          request.path.endsWith('expenses') ? '{"expenses":[]}' : '{"month":"2025-02","total_amount":0,"by_category":[],"by_day":[]}',
        );
      });
    final repository = ExpenseRepository(dio, timezone: 'Asia/Ho_Chi_Minh');

    await repository.list('2025-02');
    await repository.monthlyReport('2025-02');

    expect(queries['/expenses'], {
      'month': '2025-02',
      'tz': 'Asia/Ho_Chi_Minh',
    });
    expect(queries['/reports/monthly'], {
      'month': '2025-02',
      'tz': 'Asia/Ho_Chi_Minh',
    });
  });

  test('turns API error envelope into ApiException', () async {
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost/api/v1'))
      ..httpClientAdapter = _Adapter(
        (_) => jsonResponse(
          '{"error":{"code":"invalid_input","message":"Số tiền không hợp lệ"}}',
          status: 400,
        ),
      );
    await expectLater(
      ExpenseRepository(dio).list('2025-02'),
      throwsA(
        isA<ApiException>().having(
          (e) => e.message,
          'message',
          'Số tiền không hợp lệ',
        ),
      ),
    );
  });
}
