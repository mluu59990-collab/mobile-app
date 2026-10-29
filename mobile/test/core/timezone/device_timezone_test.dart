import 'package:expense_app/core/timezone/device_timezone.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('falls back to UTC when timezone lookup fails', () async {
    final timezone = await DeviceTimezone.load(
      lookup: () async => throw StateError('platform unavailable'),
    );

    expect(timezone.name, 'UTC');
  });
}
