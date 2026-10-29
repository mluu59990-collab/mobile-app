# AGENTS.md

## Mục tiêu

App quản lý tài chính cá nhân trên iOS/Android: xem tháng qua đã tiêu gì, tiêu bao nhiêu, vào lúc nào.

Điểm khác biệt: người dùng ghi chi tiêu bằng **ảnh**. Chụp món đồ/khoảnh khắc mua, nhập giá, app tự lưu ngày giờ và thống kê. Giai đoạn sau, AI đọc ảnh hóa đơn để tự lấy giá tiền và danh sách sản phẩm.

## Giai đoạn

1. **MVP:** đăng ký/đăng nhập (email + mật khẩu, Google, Facebook), thêm chi tiêu thủ công (giá, ghi chú, danh mục, ảnh), xem danh sách và tổng kết theo tháng.
2. **Ảnh:** upload ảnh lên S3/Cloudflare R2, gắn ảnh vào chi tiêu.
3. **AI:** trích xuất giá + sản phẩm từ ảnh hóa đơn (Vision LLM, PaddleOCR tùy chọn).
4. **Hoàn thiện:** Redis (cache, rate limiting), CI/CD, Sentry, TestFlight.

Chỉ làm giai đoạn hiện tại. Không tự thêm tính năng của giai đoạn sau.

## Stack

| Layer | Công nghệ |
|---|---|
| Mobile | Flutter + Dart, Riverpod (state), Dio (HTTP) |
| Backend | Go + Gin |
| ORM / DB | GORM + PostgreSQL |
| Auth | JWT access token + refresh token; đăng nhập email/mật khẩu, Google, Facebook |
| Social login (mobile) | `google_sign_in`, `flutter_facebook_auth` |
| AI | Vision-capable LLM; PaddleOCR (tùy chọn) |
| Storage | S3 / Cloudflare R2 |
| Cache | Redis (giai đoạn sau) |
| Hạ tầng | Docker + Compose, GitHub Actions |
| Giám sát / phát hành | Sentry, TestFlight + App Store Connect |

## Cấu trúc thư mục

```
expense-app/
├── backend/            # Go + Gin
│   ├── cmd/server/     # main.go
│   ├── internal/
│   │   ├── handler/    # HTTP handlers (Gin)
│   │   ├── service/    # business logic
│   │   ├── repository/ # truy cập DB (GORM)
│   │   ├── model/      # GORM models
│   │   ├── middleware/ # auth, logging, error
│   │   └── config/
│   └── migrations/
├── mobile/             # Flutter
│   └── lib/
│       ├── core/       # Dio client, theme, router, constants
│       └── features/   # auth, expense, report, ...
│           └── <feature>/{data,domain,presentation}
├── docs/               # ghi chú thiết kế
├── .agents/skills/     # skill cho Codex
├── docker-compose.yml
└── AGENTS.md
```

Backend đi theo luồng `handler -> service -> repository`. Handler không gọi DB trực tiếp.

## Quy ước

### Dữ liệu
- **Tiền:** lưu số nguyên VND (`int64` / `BIGINT`), không dùng float. Đơn vị là đồng, không có phần thập phân. Chỉ format thành "50.000 ₫" ở tầng UI.
- **Thời gian:** lưu UTC (`TIMESTAMPTZ`). Chuyển sang múi giờ máy khi hiển thị.
- **ID:** UUID.
- **Bảng `expenses`** có sẵn các cột cho AI: `source` (`manual` | `receipt_ai`), `raw_ai_result` (JSONB, nullable), `image_url` (nullable).

### API
- Mọi route nằm dưới `/api/v1`.
- Trả JSON; lỗi theo dạng `{"error": {"code": "...", "message": "..."}}`.
- Truy vấn tháng chi tiêu/báo cáo nhận `tz` IANA tùy chọn (mặc định UTC); biên tháng và ngày báo cáo theo múi giờ đó, dữ liệu vẫn lưu UTC.
- Route cần đăng nhập dùng middleware JWT. Access token ngắn hạn, refresh token dài hạn và có thể thu hồi.
- Không bao giờ trả `password_hash` hay token của người dùng khác.

### Đăng nhập (email, Google, Facebook)
- Luồng social login: app lấy token từ nhà cung cấp (Google: ID token; Facebook: access token) rồi gửi lên backend. **Backend tự xác thực token với Google/Facebook**, tìm hoặc tạo user, rồi cấp JWT + refresh token của chính app. Không tin dữ liệu user do client gửi.
- Endpoint: `POST /api/v1/auth/google`, `POST /api/v1/auth/facebook`, cùng với `register`, `login`, `refresh`, `logout`.
- Bảng `auth_identities` (`user_id`, `provider` = `password` | `google` | `facebook`, `provider_user_id`, unique theo `provider + provider_user_id`). Một user có thể liên kết nhiều provider.
- Cùng email ở nhiều provider: chỉ tự động liên kết khi email đã được provider xác minh; nếu không thì tạo tài khoản riêng hoặc yêu cầu xác nhận. Facebook có thể không trả email, nên `users.email` cho phép null.
- User đăng nhập social không có mật khẩu: `password_hash` cho phép null.
- Client ID/secret của Google và Facebook đọc từ biến môi trường.

### Code
- Go: `gofmt`, xử lý lỗi rõ ràng, không bỏ qua `error`. Truyền `context.Context` xuống service/repository.
- Dart: chạy `flutter analyze` sạch; dùng `const` khi có thể; state qua Riverpod, không đặt logic nghiệp vụ trong widget.
- Đặt tên: Go theo chuẩn Go; Dart `snake_case` cho file, `PascalCase` cho class.
- Thay đổi schema luôn qua migration, không sửa DB bằng tay.

### Bảo mật
- Secret (JWT secret, Google/Facebook client ID và secret, key R2, key LLM) đọc từ biến môi trường, không commit. Giữ `.env.example` không chứa giá trị thật.
- Ảnh hóa đơn là dữ liệu cá nhân: kiểm tra loại và dung lượng file khi upload, chỉ chủ sở hữu mới truy cập được.
- Ảnh dùng R2 private; DB lưu `image_key` (cột legacy `image_url` được giữ nhưng không dùng), còn `image_url` API là presigned URL tạm thời; không log URL ký hoặc credentials.

## Lệnh thường dùng

```bash
# Backend
docker compose up -d db          # chạy PostgreSQL
cd backend && go run ./cmd/server
cd backend && go test ./...

# Mobile
cd mobile && flutter pub get
cd mobile && flutter analyze
cd mobile && flutter test
cd mobile && flutter run
```

## Cách làm việc

- Thay đổi nhỏ, từng bước; chạy test/analyze sau mỗi thay đổi.
- Trước khi sửa lớn (đổi schema, đổi cấu trúc thư mục), nêu kế hoạch ngắn rồi mới làm.
- Không xóa file hay dữ liệu, không push/deploy nếu chưa được yêu cầu.
- Khi đổi API, cập nhật cả backend lẫn client Dio ở mobile trong cùng một thay đổi.
