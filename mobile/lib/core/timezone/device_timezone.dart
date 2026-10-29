import 'package:flutter_timezone/flutter_timezone.dart';

typedef TimezoneLookup = Future<String> Function();

class DeviceTimezone {
  const DeviceTimezone(this.name);

  final String name;

  static Future<DeviceTimezone> load({TimezoneLookup? lookup}) async {
    try {
      final name = await (lookup ?? _lookupDeviceTimezone)();
      if (name.trim().isEmpty) return const DeviceTimezone('UTC');
      return DeviceTimezone(name);
    } catch (_) {
      return const DeviceTimezone('UTC');
    }
  }

  static Future<String> _lookupDeviceTimezone() async =>
      (await FlutterTimezone.getLocalTimezone()).identifier;
}
