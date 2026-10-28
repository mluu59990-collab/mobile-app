# Auth email API

Base path `/api/v1`. Request/response JSON. Email được trim và lowercase;
password giữ nguyên, dài 8–72 byte (giới hạn bcrypt). Email tối đa 254 byte.

| Route | Body | Thành công | Lỗi dự kiến |
|---|---|---|---|
| POST /auth/register | `{ "email": "demo@example.com", "password": "example-password" }` | 201, token pair | 400 invalid_request, 409 email_taken |
| POST /auth/login | email, password | 200, token pair | 400 invalid_request, 401 invalid_credentials |
| POST /auth/refresh | `{ "refresh_token": "..." }` | 200, token pair mới | 400 invalid_request, 401 invalid_refresh_token |
| POST /auth/logout | refresh_token | 204, không body | 400 invalid_request |
| GET /auth/me | header `Authorization: Bearer <access_token>` | 200, `{ "user": { "id": "UUID", "email": "..." } }` | 401 invalid_access_token |

Token pair: `{ "access_token": "JWT", "refresh_token": "opaque token", "token_type": "Bearer", "expires_in": 900, "user": { "id": "UUID", "email": "..." } }`.
`expires_in` là số giây theo JWT_ACCESS_TTL. Lỗi luôn theo
`{ "error": { "code": "...", "message": "..." } }`; lỗi hệ thống: 500 internal_error.

Access JWT HS256 mặc định 15 phút, có issuer/audience/subject/jti/iat/nbf/exp.
Refresh token ngẫu nhiên 256 bit, mặc định 30 ngày, DB chỉ lưu SHA-256.
Refresh đổi token trong transaction: token cũ không dùng lại được, kể cả gọi
đồng thời. Client cần tuần tự hóa refresh và lưu token mới ngay khi thành công.
Logout thu hồi refresh token được gửi, idempotent; access token đã cấp vẫn
hợp lệ đến hết hạn. Không trả password hash hoặc hash refresh trong API.
Register tạo user + password identity + refresh token cùng transaction.
Không tự liên kết tài khoản social có cùng email chưa xác minh.

Mobile chưa có source/package Flutter trong repo; contract này dành cho Dio khi
khởi tạo mobile. Social login, màn hình Flutter và Redis nằm ngoài thay đổi này.
