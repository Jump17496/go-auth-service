# 🔐 Go Auth Service

RESTful authentication service built with Go, PostgreSQL, and JWT.

---

## 📋 สารบัญ

- [ภาพรวม](#ภาพรวม)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [การติดตั้งและรัน](#การติดตั้งและรัน)
- [API Endpoints](#api-endpoints)
- [Configuration](#configuration)
- [โครงสร้างโปรเจกต์](#โครงสร้างโปรเจกต์)
- [Docker](#docker)

---

## 🎯 ภาพรวม

Go Auth Service เป็น Backend API สำหรับระบบ Authentication ที่รองรับ:
- User Registration และ Login
- JWT Token Authentication
- Token Refresh Mechanism
- Protected Routes
- PostgreSQL Database Integration

---

## ✨ Features

- ✅ User registration with password hashing (bcrypt)
- ✅ User login with JWT token generation
- ✅ Protected user data endpoint with JWT verification
- ✅ Token refresh mechanism
- ✅ PostgreSQL database for user storage
- ✅ CORS support for frontend integration
- ✅ Configuration management with Viper
- ✅ Docker support

---

## 📦 Prerequisites

- **Go** 1.21 หรือสูงกว่า
- **PostgreSQL** database (หรือใช้ Docker)
- **Docker & Docker Compose** (แนะนำ, optional)

---

## 🚀 การติดตั้งและรัน

### วิธีที่ 1: ใช้ Docker (แนะนำ) ⭐

#### ขั้นตอนที่ 1: เริ่ม PostgreSQL

```bash
docker compose up -d postgres
```

รอประมาณ 10 วินาทีให้ PostgreSQL พร้อมใช้งาน

#### ขั้นตอนที่ 2: ตั้งค่า Configuration

```bash
# คัดลอกไฟล์ config example
cp config.yaml.example config.yaml

# แก้ไข config.yaml ตามต้องการ (หรือใช้ค่า default ได้)
```

#### ขั้นตอนที่ 3: ติดตั้ง Dependencies

```bash
go mod download
```

#### ขั้นตอนที่ 4: รัน Service

```bash
go run main.go
```

Server จะรันที่ `http://localhost:8080`

#### ขั้นตอนที่ 5: หยุด PostgreSQL (เมื่อเสร็จแล้ว)

```bash
docker compose down
```

---

### วิธีที่ 2: รันทั้งหมดด้วย Docker 🐳

```bash
docker compose -f docker-compose.full.yml up --build
```

วิธีนี้จะรันทั้ง PostgreSQL และ Go Service ใน Docker containers

---

### วิธีที่ 3: ใช้ PostgreSQL แบบ Local 💻

#### ขั้นตอนที่ 1: ติดตั้ง Dependencies

```bash
go mod download
```

#### ขั้นตอนที่ 2: สร้าง Database

```bash
createdb authdb
```

#### ขั้นตอนที่ 3: ตั้งค่า Configuration

**Option A: ใช้ Environment Variables**

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/authdb?sslmode=disable"
export JWT_SECRET="your-secret-key-change-in-production"
export PORT="8080"
```

**Option B: ใช้ Config File**

```bash
# คัดลอกไฟล์ config example
cp config.yaml.example config.yaml

# แก้ไข config.yaml
```

**config.yaml:**
```yaml
database_url: "postgres://postgres:postgres@localhost:5432/authdb?sslmode=disable"
jwt_secret: "your-secret-key-change-in-production"
port: "8080"
```

#### ขั้นตอนที่ 4: รัน Service

```bash
go run main.go
```

---

## 🔌 API Endpoints

### Base URL
```
http://localhost:8080/api/auth
```

### 1. Register User

**POST** `/api/auth/register`

**Request Body:**
```json
{
  "username": "testuser",
  "password": "password123",
  "confirmPassword": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "User registered successfully",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "testuser",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 2. Login

**POST** `/api/auth/login`

**Request Body:**
```json
{
  "username": "testuser",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "testuser"
  }
}
```

---

### 3. Refresh Token

**POST** `/api/auth/refresh`

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

### 4. Get User Data (Protected) 🔒

**GET** `/api/auth/user`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "success": true,
  "message": "User data retrieved successfully",
  "user": {
    "id": 1,
    "username": "testuser",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 5. Health Check

**GET** `/health`

**Response:**
```
OK
```

---

## ⚙️ Configuration

### Environment Variables

Service รองรับการตั้งค่าผ่าน Environment Variables:

- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret key สำหรับ JWT signing
- `PORT` - Server port (default: 8080)

### Config File

สร้างไฟล์ `config.yaml` จาก `config.yaml.example`:

```yaml
database_url: "postgres://postgres:postgres@localhost:5432/authdb?sslmode=disable"
jwt_secret: "your-secret-key-change-in-production"
port: "8080"
```

**Priority:** Environment Variables > Config File > Default Values

---

## 📁 โครงสร้างโปรเจกต์

```
go-auth-service/
├── main.go                 # Application entry point
├── config.yaml             # Configuration file (create from example)
├── config.yaml.example     # Configuration template
├── go.mod                  # Go module file
├── go.sum                  # Go dependencies checksum
├── Dockerfile              # Docker image definition
├── docker-compose.yml      # Docker Compose for PostgreSQL only
├── docker-compose.full.yml # Docker Compose for full stack
└── internal/
    ├── config/             # Configuration management
    │   └── config.go
    ├── database/           # Database connection & migrations
    │   └── database.go
    ├── handlers/           # HTTP handlers
    │   └── auth_handler.go
    ├── middleware/         # Middleware (auth, CORS)
    │   └── auth.go
    ├── models/             # Data models
    │   └── user.go
    └── utils/              # Utility functions
        └── response.go
```

---

## 🐳 Docker

### Docker Compose Files

#### 1. `docker-compose.yml` - PostgreSQL Only

รันเฉพาะ PostgreSQL database:

```bash
docker compose up -d postgres
```

#### 2. `docker-compose.full.yml` - Full Stack

รันทั้ง PostgreSQL และ Go Service:

```bash
docker compose -f docker-compose.full.yml up --build
```

### Docker Commands

```bash
# Build image
docker build -t go-auth-service .

# Run container
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://..." \
  -e JWT_SECRET="your-secret" \
  go-auth-service

# View logs
docker compose logs -f

# Stop services
docker compose down

# Stop and remove volumes
docker compose down -v
```

---

## 🔒 Security Features

- ✅ Password Hashing ด้วย bcrypt
- ✅ JWT Token Authentication
- ✅ Token Refresh Mechanism
- ✅ CORS Protection
- ✅ Input Validation
- ✅ SQL Injection Protection (Prepared Statements)

---

## 🧪 Testing

### Manual Testing with cURL

```bash
# Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123","confirmPassword":"password123"}'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'

# Get User (replace TOKEN with actual token)
curl -X GET http://localhost:8080/api/auth/user \
  -H "Authorization: Bearer TOKEN"
```

---

## 🐛 Troubleshooting

### Database Connection Error

```bash
# ตรวจสอบว่า PostgreSQL กำลังรันอยู่
docker ps

# ตรวจสอบ logs
docker compose logs postgres

# ทดสอบ connection
psql -h localhost -U postgres -d authdb
```

### Port Already in Use

```bash
# เปลี่ยน port ใน config.yaml
port: "8081"

# หรือใช้ environment variable
export PORT="8081"
```

### Migration Errors

```bash
# ลบ database และสร้างใหม่
dropdb authdb
createdb authdb

# รัน service อีกครั้ง (จะ migrate อัตโนมัติ)
go run main.go
```

---

## 📚 Additional Commands

```bash
# Build executable
go build -o auth-service main.go

# Run executable
./auth-service

# Run tests
go test ./...

# Format code
go fmt ./...

# Check for issues
go vet ./...

# Download dependencies
go mod download

# Tidy dependencies
go mod tidy
```

---

## 📄 License

MIT License

---

## 👨‍💻 Author

Created with ❤️

---

## 🤝 Contributing

Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.
