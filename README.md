
<img src="source\tikeria-be.png" alt="App Screenshot" width="350">

### Tikeria Backend For Event Concert Made Using Golang 



![Golang](https://img.shields.io/badge/Golang-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Swagger](https://img.shields.io/badge/Documentation_ready-85EA2D?style=for-the-badge&logo=swagger&logoColor=173647&)
![Docker](https://img.shields.io/badge/Docker_Ready_To_Deploy-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)

---

## First Time Setup Project 

Follow these steps when setup project for the first time after `git clone`:

### 1. Clone Repository & Enter Directory
```bash
git clone <repository-url>
cd tikeria-be
```

### 2. Setup Environment Variables
Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```

### 3. Install Dependencies & Generate Swagger
```bash
# Download Go modules
go mod download

# Install swag tool (if not installed yet)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger Docs
swag init -g cmd/api/main.go
```

### 4. Run Application locally

**Normal Run:**
```bash
go run cmd/api/main.go
```

**Hot Reload Run (Recommended):**
```bash
# Install air tool for hot reload
go install github.com/air-verse/air@latest

# Run with hot reload
air
```

* Swagger UI Documentation: `http://localhost:8080/swagger/index.html`
* Root API Info: `http://localhost:8080/`

---
![Docker](https://img.shields.io/badge/Deploy_with_Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
---

### Option A: Using Docker Compose (Recommended)
Run the entire app with one command:
```bash
docker compose up --build
```

To stop containers:
```bash
docker compose down
```

### Option B: Using Docker CLI Manual
```bash
# Build Docker Image
docker build -t tikeria-be .

# Run Docker Container
docker run -p 8080:8080 tikeria-be
```

---

##  Project Roadmap & Feature Status

### Completed Features:
- [x] Clean Architecture Setup (Config, Models, DTO, Repositories, Services, Controllers, Middleware)
- [x] Database Migration (AutoMigrate via GORM - SQLite / PostgreSQL Supabase)
- [x] User Authentication & Authorization (Register & Login with JWT Bearer Token)
- [x] Password Security with Bcrypt Hashing
- [x] Event & Ticket Category Management (Create, List, Detail)
- [x] Interactive OpenAPI / Swagger UI Documentation (`swag`)
- [x] Hot Reload Support (`air`)
- [x] Production-ready Docker Setup (`Dockerfile` & `docker-compose.yml`)

### Unrealized / Planned Features:
- [ ] Order & Checkout Ticketing System
- [ ] Payment Gateway Integration (Midtrans Adapter / Mock Payment)
- [ ] Ticket Stock Reservation with Expiry Window
- [ ] Unique Signed QR Code Ticket Generation
- [ ] Email Notifications for E-Ticket Delivery
- [ ] Event Check-in & Scanner API Validation
- [ ] Organizer Application & Admin Approval Workflow
- [ ] Redis Caching & Rate Limiting
