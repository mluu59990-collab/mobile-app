import 'package:flutter/foundation.dart';

class AuthSession extends ChangeNotifier {
  AuthSession([this._isAuthenticated = false]);

  bool _isAuthenticated;

  bool get isAuthenticated => _isAuthenticated;

  void markAuthenticated() => _setAuthenticated(true);

  void markUnauthenticated() => _setAuthenticated(false);

  void _setAuthenticated(bool value) {
    if (_isAuthenticated == value) return;
    _isAuthenticated = value;
    notifyListeners();
  }
}
