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

## Cấu hình ảnh chi tiêu trên Cloudflare R2

Tạo bucket R2 riêng tư và API token S3 có quyền đọc, ghi, sao chép, xem metadata và xóa object trong bucket đó. Thêm các biến `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET` vào `.env`; không commit credentials. Nếu cả bốn biến để trống, backend vẫn chạy nhưng endpoint upload trả `503`. Nếu chỉ cấu hình một phần, backend từ chối khởi động.

`R2_PUT_URL_TTL` mặc định 10 phút và `R2_GET_URL_TTL` mặc định 15 phút. Backend chỉ nhận JPEG, PNG, WebP và giới hạn object được gắn vào chi tiêu ở 10 MiB. Presigned PUT được ký với `Content-Type` và `Content-Length`; kiểm thử AWS SDK for Go v2 xác nhận đúng giá trị Content-Length nằm trong signed headers, và client phải gửi đúng các header backend trả về. Chưa kiểm chứng request PUT với credentials/bucket R2 thật; HEAD sau upload vẫn là lớp xác thực kích thước thứ hai trước khi ảnh được gắn vào chi tiêu. Backend cũng đọc vài byte đầu của bản copy để xác nhận chữ ký ảnh.

Luồng tải: gọi `POST /api/v1/uploads/presign` với `content_type` và `size` (byte), PUT file lên `upload_url` với các `headers` trả về, sau đó gửi `image_key` nhận được vào POST/PATCH chi tiêu. Key tạm có dạng `tmp/users/<user_id>/<uuid>.<ext>` để lifecycle prefix áp dụng đồng nhất; key đã gắn có dạng `users/<user_id>/<uuid>.<ext>`. R2 lifecycle chỉ lọc theo prefix từ đầu key, không hỗ trợ wildcard theo thư mục lồng nhau. Tạo rule trong Dashboard tại bucket → Settings → Object lifecycle rules, prefix `tmp/`, xóa object sau 1 ngày; hoặc chạy:

```sh
npx wrangler r2 bucket lifecycle add <BUCKET> "Expire temporary expense images" "tmp/" --expire-days 1
```

R2 thường xóa object trong vòng 24 giờ sau thời điểm hết hạn. Bucket không được bật public access; API trả URL GET có chữ ký tạm thời khi đọc chi tiêu. Key tạm sẽ được xóa ngay sau khi bản copy đã xác minh; lifecycle là bước dọn dự phòng cho upload bỏ dở. Migration thêm `image_key` và **không bỏ cột DB `image_url` cũ**; ứng dụng không đọc/ghi cột legacy đó, còn `image_url` trong response được tính từ key.

Đăng ký bằng email/mật khẩu sẽ tạo tài khoản và đăng nhập ngay. Access/refresh token được lưu bằng secure storage của nền tảng.

## Hạn chế đã biết

- App đọc một lần múi giờ IANA của thiết bị lúc khởi động và gửi tham số `tz` cho API chi tiêu/báo cáo. Biên tháng và nhóm ngày thống kê theo múi giờ máy; nếu không đọc được timezone, app gửi `UTC`. Backend vẫn hỗ trợ UTC mặc định cho client cũ, còn bản ghi luôn lưu thời điểm UTC.
- Tìm kiếm và lọc danh mục trong lịch sử hiện thực hiện trên danh sách tháng đã tải; backend chỉ hỗ trợ lọc theo tháng.
- Ngân sách và các mục chưa có API hiển thị trạng thái “Sắp có”; ứng dụng không hiển thị số liệu mẫu cho các mục này.
