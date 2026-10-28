# Chạy backend và ứng dụng mobile

## Khởi động backend

Từ thư mục gốc, thiết lập mật khẩu PostgreSQL và JWT secret cục bộ. Không commit các giá trị này:

```sh
export POSTGRES_PASSWORD="$(openssl rand -hex 24)"
export JWT_SECRET="$(openssl rand -hex 32)"
```

Khởi động PostgreSQL dev, áp dụng migration, rồi chạy Gin API:

```sh
docker compose up -d db
go -C backend run ./cmd/migrate up
go -C backend run ./cmd/server
```

Backend mặc định nghe tại `http://localhost:8080`. Kiểm tra health:

```sh
curl http://localhost:8080/api/v1/health
```

Giữ terminal backend chạy trong lúc sử dụng app. `POSTGRES_PASSWORD` và `JWT_SECRET` phải có trong môi trường của terminal chạy backend. Cấu hình auth mặc định dùng access token 15 phút và refresh token 30 ngày.

## Chạy app

Trong terminal khác:

```sh
cd mobile
flutter pub get
flutter run --dart-define=API_BASE_URL=http://localhost:8080/api/v1
```

`localhost` là địa chỉ phù hợp khi chạy trên iOS Simulator và cũng là giá trị mặc định. Với Android Emulator, dùng địa chỉ host dành cho emulator:

```sh
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8080/api/v1
```

Nếu chạy trên thiết bị thật, thay `localhost` bằng địa chỉ IP LAN của máy chạy backend và bảo đảm thiết bị truy cập được cổng `8080`.

Đăng ký bằng email/mật khẩu sẽ tạo tài khoản và đăng nhập ngay. Access/refresh token được lưu bằng secure storage của nền tảng. Google/Facebook và các màn hình chi tiêu/thống kê chưa kết nối API.
