import 'package:shared_preferences/shared_preferences.dart';

class AppLaunchStorage {
  AppLaunchStorage(this._preferences);

  static const _welcomeSeenKey = 'welcome_seen';
  final SharedPreferences _preferences;

  bool get hasSeenWelcome => _preferences.getBool(_welcomeSeenKey) ?? false;

  Future<void> markWelcomeSeen() async {
    await _preferences.setBool(_welcomeSeenKey, true);
  }
}
