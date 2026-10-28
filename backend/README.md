# Backend

Tại gốc dự án, sao chép `.env.example` thành `.env` và điền
`POSTGRES_PASSWORD` và `JWT_SECRET` (ít nhất 32 byte ngẫu nhiên, không commit `.env`).
Có thể tạo secret bằng `openssl rand -hex 32`. Sau đó:

```sh
docker compose up -d db
set -a
. ./.env
set +a
cd backend
go run ./cmd/migrate up
go run ./cmd/server
```

Go đọc biến môi trường của process, không tự đọc file `.env`.
`HTTP_ADDR` mặc định `:8080`; `POSTGRES_SSLMODE=disable` dành cho DB local.
Health: `curl http://localhost:8080/api/v1/health`.
Server kiểm tra kết nối DB khi khởi động; health hiện chỉ báo server còn hoạt động.

Migration dùng golang-migrate, chạy từ `backend/`:

```sh
go run ./cmd/migrate up
go run ./cmd/migrate down
```

`up` áp dụng các migration còn thiếu; `down` rollback **một** migration.
Migration đầu tiên khi rollback sẽ xóa ba bảng và dữ liệu của chúng.
Chỉ kiểm tra up/down trên DB riêng. Có thể chọn thư mục bằng `-path`.
GORM không tự tạo hay sửa schema.

Schema dùng UUID, tiền VND nguyên dương (`BIGINT`), thời gian `TIMESTAMPTZ`.
`spent_at` là thời điểm chi tiêu; `created_at` là thời điểm tạo bản ghi.
Email không unique để cho phép tài khoản riêng khi provider chưa xác minh email;
định danh đăng nhập được ràng buộc unique theo `(provider, provider_user_id)`.
`updated_at` cần được cập nhật khi ghi dữ liệu (GORM hỗ trợ quy ước này).

```sh
go test ./...
```

## Auth email

Đã có register, login, refresh, logout và route JWT `/api/v1/auth/me`.
Xem [contract auth](../docs/auth-api.md) để biết body, status và error codes.
Migration `000002` thêm bảng `refresh_tokens`; không chỉnh migration ban đầu.
Refresh token chỉ lưu SHA-256; refresh dùng row lock và transaction để đổi token
một lần. Logout thu hồi refresh token của phiên được gửi; access JWT vẫn dùng
được đến hết TTL. Client phải tuần tự hóa refresh.

`JWT_SECRET` bắt buộc khi chạy server, không cần cho lệnh migration.
`JWT_ACCESS_TTL` mặc định `15m` (từ `1s` đến `1h`),
`REFRESH_TOKEN_TTL` mặc định `720h` và phải dài hơn access TTL.
Issuer/audience cấu hình bằng `JWT_ISSUER`, `JWT_AUDIENCE`.
Không tự liên kết social account theo email. Chưa thêm Redis/rate limiting
(theo giai đoạn dự án); cần xử lý giới hạn đăng nhập trước khi mở public.

```sh
# Unit tests (integration test được skip nếu chưa cấp DB test)
go test ./...
# DB riêng đã chạy migration, tuyệt đối không dùng DB production
AUTH_TEST_DATABASE_URL='postgres://USER:PASSWORD@127.0.0.1:PORT/TEST_DB?sslmode=disable' go test ./... -count=1
```

Integration test chạy HTTP handlers với PostgreSQL thật: register/login, email
trùng không tạo user mồ côi, refresh đồng thời chỉ một request thành công,
rollback transaction khi ghi token mới lỗi, token hết hạn/bị thu hồi, logout
lặp lại, phiên khác vẫn hoạt động và middleware JWT. Test tạo dữ liệu riêng
bằng email UUID, không xóa dữ liệu hoặc tự migration DB được truyền vào.
Repo hiện chưa có package Flutter/Dio để tích hợp mobile.

## Google login

`POST /api/v1/auth/google` nhận `{"id_token":"<Google ID token>"}`.
Server xác minh chữ ký bằng Google JWK, kiểm tra `aud`, `iss`, `exp` và `sub`
(issuer chấp nhận `accounts.google.com` hoặc `https://accounts.google.com`).
Audience lấy từ biến môi trường `GOOGLE_CLIENT_ID`; nếu chưa cấu hình endpoint
trả `503 google_login_unavailable`. Payload giả/không hợp lệ trả `401`.
Google `sub` là định danh provider duy nhất. Email tự động liên kết với tài
khoản hiện có khi đã xác minh và là Gmail hoặc Google Workspace; email khác
được giữ riêng. Email chưa xác minh không được lưu hay dùng để liên kết.
Thành công trả session JWT/refresh của ứng dụng, không trả ID token Google.
Test route dùng verifier mock; không cần gọi Google trong test.

## Facebook login

`POST /api/v1/auth/facebook` nhận `{"access_token":"<Facebook user access token>"}`.
Server gọi Graph API `debug_token` bằng `FACEBOOK_APP_ID` và
`FACEBOOK_APP_SECRET`, kiểm tra token hợp lệ, đúng app và có user ID; sau đó
lấy `/me?fields=id,email` bằng token người dùng và đối chiếu user ID. Thành
công trả JWT/refresh của app; token sai trả `401`, thiếu cấu hình trả `503`.
Facebook không phải lúc nào cũng trả email: trường `user.email` trong response
có thể là `null`. Email do `/me` trả về không được dùng để tự liên kết tài khoản
khác. Test giả lập verifier và Graph API; không cần Facebook credentials.

## Expenses

Các route `/api/v1/expenses` đều yêu cầu `Authorization: Bearer <access_token>`.
`POST` tạo với `amount` (VND nguyên dương), `note`, `category`, `spent_at` tùy
chọn; `GET` liệt kê, `GET /:id` xem một khoản, `PATCH /:id` sửa các trường được
gửi, `DELETE /:id` xóa. Bỏ `spent_at` khi tạo sẽ dùng thời gian UTC hiện tại;
gửi note/category là chuỗi rỗng để xóa giá trị. `GET ?month=YYYY-MM` lọc theo
tháng UTC, bao gồm đầu tháng và loại trừ đầu tháng tiếp theo. Mọi thao tác truy
vấn đều ràng buộc user ID từ access token; ID của user khác trả 404.

## Monthly reports

`GET /api/v1/reports/monthly?month=YYYY-MM` yêu cầu access token. Response có
`total_amount`, `by_category` và `by_day`; các tổng là số nguyên VND.
Danh mục không có giá trị được trả `category: null`; ngày chỉ gồm ngày có chi
tiêu. Tháng được tính theo UTC, cùng quy tắc biên thời gian với danh sách chi
tiêu. Các truy vấn tổng hợp chạy trong cùng snapshot và chỉ tính chi tiêu của
user từ access token.
