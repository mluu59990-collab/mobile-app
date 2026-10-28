import 'package:shared_preferences/shared_preferences.dart';

class TokenStorage {
  TokenStorage(this._preferences);

  static const _accessTokenKey = 'access_token';
  final SharedPreferences _preferences;

  String? get accessToken => _preferences.getString(_accessTokenKey);

  Future<void> saveAccessToken(String token) async {
    await _preferences.setString(_accessTokenKey, token);
  }

  Future<void> clear() async {
    await _preferences.remove(_accessTokenKey);
  }
}
