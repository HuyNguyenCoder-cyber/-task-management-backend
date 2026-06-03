# Task Management System Backend

Backend cho **Task Management System** được xây dựng bằng **Golang**, chạy local bằng **Docker / Docker Compose**, và có **CI trên GitHub Actions** để tự động kiểm tra chất lượng code trước khi merge.

Hệ thống hỗ trợ:
- Quản lý người dùng, dự án, task và bình luận
- Xác thực JWT cho các API nghiệp vụ
- Redis cache cho một số luồng đọc và queue worker cho notification
- WebSocket để đẩy event realtime

---

## Mục Lục

- [Kiến Trúc](#kiến-trúc)
- [Yêu Cầu Môi Trường](#yêu-cầu-môi-trường)
- [Setup & Run Local](#setup--run-local)
- [Chạy Test & Vet](#chạy-test--vet)
- [API Usage](#api-usage)
- [Git Workflow Cho Team](#git-workflow-cho-team)
- [CI](#ci)

---

## Kiến Trúc

Dự án được tổ chức theo hướng **tách biệt business logic khỏi hạ tầng kỹ thuật**:

```text
.
├── cmd/
│   └── api/                 # Entry point khởi động HTTP server
├── internal/                # Toàn bộ logic ứng dụng, chỉ dùng bên trong repo
│   ├── app/                 # Wiring: khởi tạo DB, Redis, handlers, router
│   ├── auth/                # JWT service, context helper
│   ├── config/              # Load biến môi trường và config runtime
│   ├── database/            # Kết nối MySQL / Redis
│   ├── domain/              # Entity/domain model
│   ├── dto/                 # Request/response schema
│   ├── handler/             # HTTP handlers
│   ├── middleware/          # Auth, logger, recovery, request id
│   ├── repository/          # Truy cập dữ liệu qua GORM
│   ├── service/             # Business rules
│   ├── websocket/           # Realtime event broadcast
│   └── worker/              # Background worker xử lý notification
├── migrations/              # SQL migration versioned
├── web/                     # Static frontend tối giản phục vụ kèm backend
└── .github/workflows/       # CI GitHub Actions
```

### Vai Trò Từng Lớp

- **`cmd/`**: điểm vào của ứng dụng, giữ phần bootstrap gọn nhất có thể.
- **`internal/handler`**: nhận request, validate input cơ bản, trả response theo chuẩn thống nhất.
- **`internal/service`**: chứa luật nghiệp vụ như quyền truy cập, trạng thái task, validation domain.
- **`internal/repository`**: giao tiếp MySQL qua GORM, không chứa logic nghiệp vụ.
- **`internal/domain`**: mô tả object nghiệp vụ lõi như `User`, `Project`, `Task`, `Comment`.
- **`internal/dto`**: tách riêng schema API khỏi model lưu trữ.
- **`internal/middleware`**: xử lý cross-cutting concerns như auth, logging, recovery.
- **`internal/websocket`** và **`internal/worker`**: phục vụ realtime / background processing.
- **`migrations/`**: lưu schema SQL để khởi tạo database.
- **`.github/workflows/`**: tự động chạy kiểm tra khi có Pull Request.

### Lợi Ích Của Cách Tổ Chức Này

- Dễ test từng lớp riêng biệt
- Dễ thay thế hạ tầng như DB / cache mà không ảnh hưởng business logic
- Dễ mở rộng theo hướng module hóa
- Hạn chế coupling giữa HTTP layer và persistence layer

---

## Yêu Cầu Môi Trường

- **Docker Desktop** hoặc Docker Engine + Docker Compose
- **Go** để chạy test / vet local
- File **`.env`** với `JWT_SECRET` hợp lệ

Ví dụ cấu hình `.env`:

```env
APP_PORT=8080

DB_HOST=localhost
DB_PORT=3306
DB_USER=app_user
DB_PASSWORD=app_password
DB_NAME=task_management
REDIS_HOST=localhost
REDIS_PORT=6379

JWT_SECRET=your-dev-secret-key
JWT_EXPIRES_IN_HOURS=24
```

> Lưu ý: repo có thư mục `migrations/` chứa schema SQL. Nếu database của bạn đang trống, hãy apply các file `*.up.sql` trước khi gọi API nghiệp vụ.

---

## Setup & Run Local

### 1) Clone source code

```bash
git clone <repository-url>
cd task-management-backend
```

### 2) Kiểm tra file `.env`

Đảm bảo file `.env` có `JWT_SECRET` và các thông số DB/Redis phù hợp với môi trường local.

### 3) Khởi động toàn bộ stack bằng Docker Compose

```bash
docker compose up --build -d
```

Lệnh này sẽ khởi tạo:
- `mysql` cho database
- `redis` cho cache / worker queue
- `api` cho HTTP backend
- `worker` cho background job

### 4) Kiểm tra trạng thái container

```bash
docker compose ps
```

Nếu muốn xem log:

```bash
docker compose logs -f api
docker compose logs -f worker
```

### 5) Xác nhận service đã lên

- Health check: `GET http://localhost:8080/health`
- Ping nhanh: `GET http://localhost:8080/ping`
- UI tĩnh: `http://localhost:8080/`

### 6) Tắt toàn bộ môi trường local

```bash
docker compose down
```

Nếu muốn xoá luôn volume MySQL để reset dữ liệu:

```bash
docker compose down -v
```

---

## Chạy Test & Vet

Trước khi push code, hãy chạy tối thiểu:

```bash
go test ./... -v
go vet ./...
```

### Khuyến nghị quy trình local

1. Bật database và Redis bằng Docker Compose.
2. Chạy `go test ./... -v` để kiểm tra unit test và integration test.
3. Chạy `go vet ./...` để phát hiện lỗi logic / anti-pattern ở mức static analysis.
4. Chỉ push khi cả hai lệnh đều pass.

> CI trên GitHub Actions cũng kiểm tra lại các bước này khi mở Pull Request.

---

## API Usage

### Quy Ước Chung

- **Base URL**: `http://localhost:8080`
- **Response format**:

```json
{
  "success": true,
  "message": "Human readable message",
  "data": {}
}
```

Mẫu lỗi:

```json
{
  "success": false,
  "message": "Invalid request body",
  "error": "..."
}
```

- Các API protected dùng header:

```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

- Task status hợp lệ:
  - `todo`
  - `in_progress`
  - `done`

- Project member role:
  - `owner`
  - `member`

---

### 1) Auth

| Method | Endpoint | Auth | Mô tả |
|---|---|---:|---|
| `POST` | `/auth/register` | No | Đăng ký tài khoản mới |
| `POST` | `/auth/login` | No | Đăng nhập và nhận JWT |

#### `POST /auth/register`

**Request**

```json
{
  "email": "alice@example.com",
  "password": "123456",
  "full_name": "Alice Nguyen"
}
```

**Response 201**

```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": 1,
    "email": "alice@example.com",
    "full_name": "Alice Nguyen",
    "created_at": "2026-06-02T09:00:00Z"
  }
}
```

#### `POST /auth/login`

**Request**

```json
{
  "email": "alice@example.com",
  "password": "123456"
}
```

**Response 200**

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "email": "alice@example.com",
      "full_name": "Alice Nguyen"
    }
  }
}
```

---

### 2) Users

| Method | Endpoint | Auth | Mô tả |
|---|---|---:|---|
| `GET` | `/users` | No | Danh sách tất cả user |
| `GET` | `/users/:id` | No | Lấy user theo ID |
| `PUT` | `/users/:id` | No | Cập nhật user |
| `DELETE` | `/users/:id` | No | Xóa user |

#### `GET /users`

**Response 200**

```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 1,
      "email": "alice@example.com",
      "full_name": "Alice Nguyen",
      "created_at": "2026-06-02T09:00:00Z",
      "updated_at": "2026-06-02T09:00:00Z"
    }
  ]
}
```

#### `GET /users/:id`

**Response 200**

```json
{
  "success": true,
  "message": "User retrieved successfully",
  "data": {
    "id": 1,
    "email": "alice@example.com",
    "full_name": "Alice Nguyen",
    "created_at": "2026-06-02T09:00:00Z",
    "updated_at": "2026-06-02T09:00:00Z"
  }
}
```

#### `PUT /users/:id`

**Request**

```json
{
  "email": "alice.new@example.com",
  "full_name": "Alice New"
}
```

**Response 200**

```json
{
  "success": true,
  "message": "User updated successfully",
  "data": {
    "id": 1,
    "email": "alice.new@example.com",
    "full_name": "Alice New",
    "created_at": "2026-06-02T09:00:00Z",
    "updated_at": "2026-06-02T10:00:00Z"
  }
}
```

#### `DELETE /users/:id`

**Response 200**

```json
{
  "success": true,
  "message": "User deleted successfully"
}
```

---

### 3) Projects

| Method | Endpoint | Auth | Mô tả |
|---|---|---:|---|
| `POST` | `/projects` | Yes | Tạo project mới |
| `GET` | `/projects` | Yes | Danh sách project của user hiện tại |
| `GET` | `/projects/:id` | Yes | Xem chi tiết project |
| `PUT` | `/projects/:id` | Yes | Cập nhật project |
| `DELETE` | `/projects/:id` | Yes | Xóa project |
| `POST` | `/projects/:id/members` | Yes | Thêm member vào project |
| `GET` | `/projects/:id/members` | Yes | Danh sách member của project |
| `POST` | `/projects/:id/tasks` | Yes | Tạo task thuộc project |

#### `POST /projects`

**Request**

```json
{
  "name": "Website Revamp",
  "description": "Refactor landing page and auth flow"
}
```

**Response 201**

```json
{
  "success": true,
  "message": "Project created successfully",
  "data": {
    "id": 10,
    "name": "Website Revamp",
    "description": "Refactor landing page and auth flow",
    "owner_id": 1,
    "created_at": "2026-06-02T09:10:00Z",
    "updated_at": "2026-06-02T09:10:00Z"
  }
}
```

#### `GET /projects`

**Response 200**

```json
{
  "success": true,
  "message": "Projects retrieved successfully",
  "data": [
    {
      "id": 10,
      "name": "Website Revamp",
      "description": "Refactor landing page and auth flow",
      "owner_id": 1,
      "created_at": "2026-06-02T09:10:00Z",
      "updated_at": "2026-06-02T09:10:00Z"
    }
  ]
}
```

#### `GET /projects/:id`

**Response 200**

```json
{
  "success": true,
  "message": "Project retrieved successfully",
  "data": {
    "id": 10,
    "name": "Website Revamp",
    "description": "Refactor landing page and auth flow",
    "owner_id": 1,
    "created_at": "2026-06-02T09:10:00Z",
    "updated_at": "2026-06-02T09:10:00Z"
  }
}
```

#### `PUT /projects/:id`

**Request**

```json
{
  "name": "Website Revamp v2",
  "description": "Update product scope"
}
```

**Response 200**

```json
{
  "success": true,
  "message": "Project updated successfully",
  "data": {
    "id": 10,
    "name": "Website Revamp v2",
    "description": "Update product scope",
    "owner_id": 1,
    "created_at": "2026-06-02T09:10:00Z",
    "updated_at": "2026-06-02T10:00:00Z"
  }
}
```

#### `DELETE /projects/:id`

**Response 200**

```json
{
  "success": true,
  "message": "Project deleted successfully"
}
```

#### `POST /projects/:id/members`

**Request**

```json
{
  "email": "bob@example.com"
}
```

**Response 201**

```json
{
  "success": true,
  "message": "Project member added successfully",
  "data": {
    "id": 2,
    "email": "bob@example.com",
    "full_name": "Bob Tran",
    "role": "member"
  }
}
```

#### `GET /projects/:id/members`

**Response 200**

```json
{
  "success": true,
  "message": "Project members retrieved successfully",
  "data": [
    {
      "id": 1,
      "email": "alice@example.com",
      "full_name": "Alice Nguyen",
      "role": "owner"
    },
    {
      "id": 2,
      "email": "bob@example.com",
      "full_name": "Bob Tran",
      "role": "member"
    }
  ]
}
```

#### `POST /projects/:id/tasks`

> Hệ thống lấy `project_id` từ path; field `project_id` trong body là không bắt buộc.

**Request**

```json
{
  "project_id": 10,
  "title": "Design login screen",
  "description": "Prepare wireframe and copy",
  "status": "todo",
  "assignee_id": 1
}
```

**Response 201**

```json
{
  "success": true,
  "message": "Task created successfully",
  "data": {
    "id": 100,
    "project_id": 10,
    "title": "Design login screen",
    "description": "Prepare wireframe and copy",
    "status": "todo",
    "created_by": 1,
    "assignee_id": 1,
    "due_date": null,
    "created_at": "2026-06-02T09:20:00Z",
    "updated_at": "2026-06-02T09:20:00Z"
  }
}
```

---

### 4) Tasks

| Method | Endpoint | Auth | Mô tả |
|---|---|---:|---|
| `POST` | `/tasks` | Yes | Tạo task độc lập |
| `GET` | `/tasks` | Yes | Danh sách task của user hiện tại, lọc bằng `project_id` nếu có |
| `GET` | `/tasks/:id` | Yes | Lấy chi tiết task |
| `PUT` | `/tasks/:id` | Yes | Cập nhật task |
| `DELETE` | `/tasks/:id` | Yes | Xóa task |
| `GET` | `/tasks/:id/comments` | Yes | Danh sách comment của task |
| `POST` | `/tasks/:id/comments` | Yes | Tạo comment cho task |

#### `POST /tasks`

**Request**

```json
{
  "project_id": 10,
  "title": "Implement dark mode",
  "description": "Add theme switcher and persistence",
  "status": "in_progress",
  "assignee_id": 1,
  "due_date": "2026-06-10T00:00:00Z"
}
```

**Response 201**

```json
{
  "success": true,
  "message": "Task created successfully",
  "data": {
    "id": 100,
    "project_id": 10,
    "title": "Implement dark mode",
    "description": "Add theme switcher and persistence",
    "status": "in_progress",
    "created_by": 1,
    "assignee_id": 1,
    "due_date": "2026-06-10T00:00:00Z",
    "created_at": "2026-06-02T09:30:00Z",
    "updated_at": "2026-06-02T09:30:00Z"
  }
}
```

#### `GET /tasks`

**Ví dụ**: `GET /tasks?project_id=10`

**Response 200**

```json
{
  "success": true,
  "message": "Tasks retrieved successfully",
  "data": [
    {
      "id": 100,
      "project_id": 10,
      "title": "Implement dark mode",
      "description": "Add theme switcher and persistence",
      "status": "in_progress",
      "created_by": 1,
      "assignee_id": 1,
      "due_date": "2026-06-10T00:00:00Z",
      "created_at": "2026-06-02T09:30:00Z",
      "updated_at": "2026-06-02T09:30:00Z"
    }
  ]
}
```

#### `GET /tasks/:id`

**Response 200**

```json
{
  "success": true,
  "message": "Task retrieved successfully",
  "data": {
    "id": 100,
    "project_id": 10,
    "title": "Implement dark mode",
    "description": "Add theme switcher and persistence",
    "status": "in_progress",
    "created_by": 1,
    "assignee_id": 1,
    "due_date": "2026-06-10T00:00:00Z",
    "created_at": "2026-06-02T09:30:00Z",
    "updated_at": "2026-06-02T09:30:00Z"
  }
}
```

#### `PUT /tasks/:id`

**Request**

```json
{
  "status": "done",
  "description": "Completed and verified"
}
```

**Response 200**

```json
{
  "success": true,
  "message": "Task updated successfully",
  "data": {
    "id": 100,
    "project_id": 10,
    "title": "Implement dark mode",
    "description": "Completed and verified",
    "status": "done",
    "created_by": 1,
    "assignee_id": 1,
    "due_date": "2026-06-10T00:00:00Z",
    "created_at": "2026-06-02T09:30:00Z",
    "updated_at": "2026-06-02T10:05:00Z"
  }
}
```

#### `DELETE /tasks/:id`

**Response 200**

```json
{
  "success": true,
  "message": "Task deleted successfully"
}
```

#### `GET /tasks/:id/comments`

**Response 200**

```json
{
  "success": true,
  "message": "Comments retrieved successfully",
  "data": [
    {
      "id": 501,
      "task_id": 100,
      "user_id": 1,
      "username": "Alice Nguyen",
      "content": "Please review this with QA.",
      "created_at": "2026-06-02T10:10:00Z",
      "updated_at": "2026-06-02T10:10:00Z"
    }
  ]
}
```

#### `POST /tasks/:id/comments`

**Request**

```json
{
  "content": "Looks good, let's merge after final checks."
}
```

**Response 201**

```json
{
  "success": true,
  "message": "Comment created successfully",
  "data": {
    "id": 501,
    "task_id": 100,
    "user_id": 1,
    "username": "Alice Nguyen",
    "content": "Looks good, let's merge after final checks.",
    "created_at": "2026-06-02T10:12:00Z",
    "updated_at": "2026-06-02T10:12:00Z"
  }
}
```

---

### 5) Utilities & System Endpoints

| Method | Endpoint | Auth | Mô tả |
|---|---|---:|---|
| `GET` | `/ping` | No | Kiểm tra server sống |
| `GET` | `/health` | No | Kiểm tra DB + Redis |
| `GET` | `/ws` | No | WebSocket realtime |
| `GET` | `/` | No | Redirect sang UI tĩnh |
| `GET` | `/panic` | No | Endpoint test middleware recovery |

#### `GET /ping`

**Response 200**

```json
{
  "message": "pong"
}
```

#### `GET /health`

**Response 200**

```json
{
  "status": "UP",
  "services": {
    "database": {
      "status": "UP",
      "details": "database connection is healthy"
    },
    "redis": {
      "status": "UP",
      "details": "redis connection is healthy"
    }
  }
}
```

#### `GET /panic`

**Response 500**

```json
{
  "success": false,
  "message": "Internal server error",
  "error": "panic recovered"
}
```

---

### 6) WebSocket Events

Endpoint WebSocket:

```text
ws://localhost:8080/ws
```

Các event hiện có:

- `task_created`
- `task_updated`
- `comment_created`

Ví dụ payload event:

```json
{
  "event": "task_created",
  "task_id": 100,
  "project_id": 10,
  "title": "Implement dark mode",
  "status": "todo",
  "created_by": 1,
  "assignee_id": 1,
  "due_date": "2026-06-10T00:00:00Z"
}
```

---

## Git Workflow Cho Team

Quy trình chuẩn khi làm việc với team:

1. **Checkout nhánh `main`**

   ```bash
   git checkout main
   ```

2. **Pull code mới nhất**

   ```bash
   git pull origin main
   ```

3. **Tạo nhánh feature từ `main`**

   ```bash
   git checkout -b feature/<ten-tinh-nang>
   ```

4. **Phát triển và tự kiểm tra local**
   - Chạy `go test ./... -v`
   - Chạy `go vet ./...`
   - Nếu cần, xác nhận lại bằng Docker Compose

5. **Commit và push branch**

   ```bash
   git add .
   git commit -m "feat: add <ten-tinh-nang>"
   git push origin feature/<ten-tinh-nang>
   ```

6. **Mở Pull Request**
   - Mở PR vào `main`
   - Chờ robot CI chạy xong
   - Chỉ merge khi **status check xanh**
   - Lead review xong mới merge

### Nguyên Tắc Team

- Không push thẳng vào `main`
- Không merge khi CI đang đỏ
- Không bỏ qua review của Lead
- Ưu tiên PR nhỏ, rõ phạm vi, dễ review

---

## CI

Workflow GitHub Actions nằm tại:

```text
.github/workflows/ci.yml
```

Hiện tại CI được thiết kế để:

- Chạy khi có **push** hoặc **Pull Request** vào `main` / `master`
- Thực thi kiểm tra chất lượng code bằng:
  - `go vet ./...`
  - `go test -v -race ./...`

Mục tiêu của CI:
- Bắt lỗi sớm trước khi merge
- Giữ chất lượng code ổn định giữa các PR
- Giảm rủi ro khi tích hợp thay đổi mới

---

## Ghi Chú Nhanh

- API nghiệp vụ chính đều dùng JWT.
- Redis phục vụ cache và hàng đợi notification.
- MySQL là nguồn dữ liệu chính.
- `web/` cung cấp giao diện tĩnh cơ bản để demo / test nhanh.
