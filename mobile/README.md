# Expense App mobile

Flutter client scaffold using Riverpod, GoRouter, Dio, SharedPreferences, and AdaptiveTheme.

Set the API root at build or run time with a Dart define:

```sh
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8080/api/v1
```

The default is `http://localhost:8080/api/v1`. For an Android emulator, use `10.0.2.2` to reach services on the host machine.
