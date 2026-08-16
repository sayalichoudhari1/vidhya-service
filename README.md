# Vidhya Service

Vidhya Service is the backend for a multi-school ERP: student management,
attendance, and exams/marksheets, with role-based access control (RBAC) for
five roles — **Admin**, **Principal**, **Teacher**, **Student**, **Parent**.

Phase 1 scope (this repo, today):

- Go (Golang) REST API, PostgreSQL only. Redis and Kafka are wired in but
  **disabled by default** — nothing in phase 1 requires them.
- Local JWT auth (no external identity provider). Bcrypt password hashing.
- Multi-tenant: every academic record is scoped to a `school_id`; an `ADMIN`
  can see all schools, everyone else only their own.
- A separate UI team builds the frontend against this API; there is no
  bundled web/mobile client in this repo.

---

## 1. Architecture at a glance

```
src/
├── main.go              entrypoint, graceful shutdown
├── app/                 wiring: config → db → repos → services, seed admin
├── config/               env-var configuration (see .env.example)
├── db/                    Postgres connection (GORM)
├── cache/                 optional Redis client (REDIS_ENABLED)
├── messaging/             optional Kafka producer (KAFKA_ENABLED)
├── model/                 domain structs (School, User, Student, ...)
├── repository/            GORM data access, one file per aggregate
├── services/              business logic (auth, school, academic, student,
│                          teacher, attendance, exam)
├── web/                   HTTP layer: router, middleware (auth/RBAC/CORS),
│                          handlers per resource
├── util/                  JWT, password hashing, pagination, JSON helpers
├── logger/, apperrors/, requestid/   cross-cutting concerns
sql/
├── dbchangelog.xml        Liquibase root changelog (includeAll)
├── liquibase.properties   connection template (envsubst'd at container start)
└── changelogs/            001..010, one aggregate per file, ordered by FK deps
devops/
├── local/                docker-compose.yml + Postgres/Liquibase/Redis images
└── aws/                  CloudFormation for phase 2 (Aurora/ElastiCache/MSK)
```

Request flow: `router.go` → `AuthMiddleware` (parses JWT) → `RequireRoles`
(RBAC) → handler → service (validation + business rules) → repository (GORM)
→ Postgres. Multi-tenancy is enforced by `resolveSchoolID()` in
`src/web/scope.go`, which checks the `{school_id}` path param against the
caller's JWT claims before any handler logic runs.

---

## 2. Prerequisites

- Go 1.24+ (toolchain pinned via `go.mod`; `go build` will fetch a matching
  toolchain automatically if needed)
- Docker + Docker Compose (for local Postgres/Redis)
- `curl` (or Postman) to exercise the API

---

## 3. Running locally

### 3.1 Start Postgres (+ optional Redis)

```bash
cp devops/local/.env.example devops/local/.env   # first time only
make local-up          # starts postgres + redis containers
make migrate           # applies sql/changelogs via Liquibase (idempotent)
```

> Postgres publishes on **host port 55432** (not 5432) and Redis on
> **63790** (not 6379), so this stack won't collide with any other local
> Postgres/Redis you might already be running. Adjust in
> `devops/local/docker-compose.yml` if you'd rather use the standard ports.

Verify the schema landed:

```bash
docker exec vidhya_postgres psql -U appuser -d vidhyadb -c '\dt'
# schools, users, classes, subjects, class_subjects, students, teachers,
# attendance, exams, exam_results, databasechangelog(lock)
```

### 3.2 Run the Go service

```bash
cp .env.example .env      # first time only
set -a; source .env; set +a
go run ./src
```

or simply `make run` (after sourcing `.env`). You should see:

```
[INFO] Connected to Postgres at localhost:55432/vidhyadb
[INFO] Redis is disabled (REDIS_ENABLED=false) - skipping connection
[INFO] Kafka is disabled (KAFKA_ENABLED=false) - skipping connection
[INFO] Seeded default admin user: email=admin@vidhya.local password=Admin@123
[INFO] Vidhya Service listening on :8080/vidhyaservice
```

The service seeds a default `ADMIN` user (`admin@vidhya.local` /
`Admin@123`, overridable via `SEED_ADMIN_EMAIL`/`SEED_ADMIN_PASSWORD`) on a
fresh database, so you can log in immediately without touching SQL.

**Change `JWT_SECRET` and the seed admin password before using this outside
your own machine.**

### 3.3 Stopping / resetting

```bash
make local-down    # stop containers, keep data volume
make local-reset   # stop + delete the Postgres volume (fresh DB next time)
```

---

## 4. API quick-start (curl)

Base URL: `http://localhost:8080/vidhyaservice`

```bash
BASE=http://localhost:8080/vidhyaservice

# 1. Log in as the seeded admin
TOKEN=$(curl -s -X POST $BASE/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@vidhya.local","password":"Admin@123"}' \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")

# 2. Create a school
SCHOOL_ID=$(curl -s -X POST $BASE/v1/schools \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Vidhya Public School","code":"VID-001","address":"MG Road, Pune"}' \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")

# 3. Create a class and a subject
CLASS_ID=$(curl -s -X POST $BASE/v1/schools/$SCHOOL_ID/classes \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"10","section":"A","academicYear":"2026-27"}' \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")

SUBJECT_ID=$(curl -s -X POST $BASE/v1/schools/$SCHOOL_ID/subjects \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Mathematics","code":"MATH"}' \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")

# 4. Create a student (this also provisions their login)
STUDENT_ID=$(curl -s -X POST $BASE/v1/schools/$SCHOOL_ID/students \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"firstName\":\"Riya\",\"lastName\":\"Sharma\",\"email\":\"riya@vidhya.local\",\"password\":\"Student@123\",\"admissionNumber\":\"ADM-2026-001\",\"rollNumber\":\"1\",\"classId\":\"$CLASS_ID\"}" \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['userId'])")

# 5. Mark attendance
curl -s -X POST $BASE/v1/schools/$SCHOOL_ID/attendance \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"studentId\":\"$STUDENT_ID\",\"classId\":\"$CLASS_ID\",\"date\":\"2026-08-16\",\"status\":\"PRESENT\"}"

# 6. Create an exam and record marks
EXAM_ID=$(curl -s -X POST $BASE/v1/schools/$SCHOOL_ID/exams \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"name\":\"Mid-Term Mathematics\",\"classId\":\"$CLASS_ID\",\"subjectId\":\"$SUBJECT_ID\",\"examType\":\"MID_TERM\",\"examDate\":\"2026-09-10\",\"maxMarks\":100}" \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")

curl -s -X POST $BASE/v1/schools/$SCHOOL_ID/exams/$EXAM_ID/results \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"studentId\":\"$STUDENT_ID\",\"marksObtained\":87}"

# 7. Consolidated marksheet
curl -s $BASE/v1/schools/$SCHOOL_ID/students/$STUDENT_ID/marksheet \
  -H "Authorization: Bearer $TOKEN"
```

This exact sequence was run against a fresh local database while building
this service; every step above returns `200`/`201` with the JSON shown.

---

## 5. Full endpoint reference

Role groups used below: `staff` = ADMIN, PRINCIPAL · `staffAndTeachers` =
ADMIN, PRINCIPAL, TEACHER · `anyRole` = all five roles (with row-level
self/parent checks inside the handler where noted).

### Auth — `/v1/auth`
| Method | Path | Roles | Notes |
|---|---|---|---|
| POST | `/v1/auth/login` | public | email + password → JWT + user profile |
| GET | `/v1/auth/me` | anyRole | current authenticated user |

### Schools — `/v1/schools`
| Method | Path | Roles |
|---|---|---|
| POST | `/v1/schools` | ADMIN only |
| GET | `/v1/schools` | ADMIN only |
| GET | `/v1/schools/{school_id}` | anyRole |
| PUT | `/v1/schools/{school_id}` | staff |
| DELETE | `/v1/schools/{school_id}` | ADMIN only |

### Classes & Subjects — `/v1/schools/{school_id}/...`
| Method | Path | Roles |
|---|---|---|
| POST | `/classes` | staff |
| GET | `/classes` | staffAndTeachers |
| GET | `/classes/{class_id}` | anyRole |
| PUT | `/classes/{class_id}` | staff |
| DELETE | `/classes/{class_id}` | staff |
| POST | `/subjects` | staff |
| GET | `/subjects` | staffAndTeachers |
| PUT | `/subjects/{subject_id}` | staff |
| DELETE | `/subjects/{subject_id}` | staff |
| POST | `/classes/{class_id}/subjects` | staff — assign subject+teacher to class |
| GET | `/classes/{class_id}/subjects` | staffAndTeachers |
| DELETE | `/classes/{class_id}/subjects/{class_subject_id}` | staff |

### Students — `/v1/schools/{school_id}/students`
| Method | Path | Roles | Notes |
|---|---|---|---|
| POST | `/` | staff | creates login + enrollment |
| GET | `/` | staffAndTeachers | optional `?classId=` filter |
| GET | `/{student_id}` | anyRole | student/parent restricted to own record |
| PUT | `/{student_id}` | staff | |
| DELETE | `/{student_id}` | staff | |
| GET | `/{student_id}/attendance` | anyRole | `?from=&to=` (default: current month) |
| GET | `/{student_id}/attendance/summary` | anyRole | `?from=&to=` |
| GET | `/{student_id}/marksheet` | anyRole | consolidated results |

### Teachers — `/v1/schools/{school_id}/teachers`
| Method | Path | Roles |
|---|---|---|
| POST | `/` | staff |
| GET | `/` | staff |
| GET | `/{teacher_id}` | staff |
| DELETE | `/{teacher_id}` | staff |

### Attendance — `/v1/schools/{school_id}/...`
| Method | Path | Roles | Notes |
|---|---|---|---|
| POST | `/attendance` | staffAndTeachers | single entry |
| POST | `/attendance/bulk` | staffAndTeachers | whole-class register: `{classId, date, entries:[{studentId,status}]}` |
| DELETE | `/attendance/{attendance_id}` | staffAndTeachers | |
| GET | `/classes/{class_id}/attendance?date=YYYY-MM-DD` | staffAndTeachers | |

### Exams & Marksheets — `/v1/schools/{school_id}/...`
| Method | Path | Roles | Notes |
|---|---|---|---|
| POST | `/exams` | staffAndTeachers | `examType`: `UNIT_TEST`\|`MID_TERM`\|`FINAL`\|`QUIZ` |
| GET | `/classes/{class_id}/exams` | anyRole | list exams for a class |
| GET | `/exams/{exam_id}` | staffAndTeachers | |
| DELETE | `/exams/{exam_id}` | staffAndTeachers | |
| POST | `/exams/{exam_id}/results` | staffAndTeachers | single result; auto-computes grade |
| POST | `/exams/{exam_id}/results/bulk` | staffAndTeachers | `{results:[{studentId,marksObtained}]}` |
| GET | `/exams/{exam_id}/results` | staffAndTeachers | |
| GET | `/students/{student_id}/marksheet` | anyRole | consolidated results across exams |

### Health
| Method | Path |
|---|---|
| GET | `/health` | pings Postgres, returns `{"status":"UP","database":"UP"}` |
| GET | `/version` | service name/version |

Grading (`src/services/exam`): `A+` ≥90%, `A` ≥80%, `B` ≥70%, `C` ≥60%, `D`
≥50%, `E` ≥40%, `F` <40% — adjust in `gradeFor()` if your school uses a
different scale.

---

## 6. Configuration reference

All configuration is via environment variables (see `.env.example` for the
full annotated list and defaults) — `config.Read()` in `src/config`.

| Variable | Default | Purpose |
|---|---|---|
| `APP_SERVER_ADDRESS` | `:8080` | HTTP bind address |
| `APP_SERVER_URL_PREFIX` | `/vidhyaservice` | Path prefix for all routes |
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASS` / `DB_NAME` | `localhost` / `55432` (local compose) / `appuser` / `appuser` / `vidhyadb` | Postgres connection |
| `JWT_SECRET` | *(dev placeholder — change it!)* | HMAC signing key |
| `JWT_ACCESS_TOKEN_DURATION` | `2h` | Access token lifetime |
| `REDIS_ENABLED` | `false` | Set `true` to connect Redis (currently no code path uses it) |
| `KAFKA_ENABLED` | `false` | Set `true` to connect Kafka (currently no code path uses it) |
| `SEED_ADMIN_ENABLED` | `true` | Auto-create the default admin user |

When Redis/Kafka are enabled but unreachable, startup logs a warning and
continues — **the service always runs on Postgres alone**.

---

## 7. Local troubleshooting

**Port already in use (5432/6379/8080)**
Something else (maybe another project's containers) is bound to that port.
Either stop it, or remap the host side of the `ports:` entry in
`devops/local/docker-compose.yml` (and the matching `DB_PORT`/`REDIS_ADDRESS`
in your `.env`). This repo already uses 55432/63790 by default for exactly
this reason.

**`liquibase` container exits immediately with "Permission denied"**
The scripts in `devops/local/liquibase/init-scripts/*.sh` lost their execute
bit (common after a fresh `git clone` on some systems/editors). Fix with:
```bash
chmod +x devops/local/liquibase/init-scripts/*.sh
```

**Postgres container logs repeated `Error: in 18+, these Docker images...`**
The named volume was created under an older Postgres major version layout.
Run `make local-reset` (deletes the local volume) and `make local-up` again.

**`compile: version "go1.x" does not match go tool version "go1.y"`**
Warnings only, seen on some environments where the `go` launcher and its
bundled compiler tools disagree — the build still succeeds (exit code 0).
If it actually fails to produce a binary, pin `GOTOOLCHAIN=local` and point
`GOROOT` at a consistent Go install (e.g. the one under
`~/go/pkg/mod/golang.org/toolchain@<version>`).

**API returns `401 Unauthorized`**
Missing/expired/garbled `Authorization: Bearer <token>` header. Re-run
`/v1/auth/login`; tokens expire after `JWT_ACCESS_TOKEN_DURATION` (default
2h).

**API returns `403 Forbidden`**
The route enforces RBAC (`RequireRoles` in `src/web/router.go`) and/or
tenant isolation (`resolveSchoolID` in `src/web/scope.go`). Check the
caller's `role` and `schoolId` claims against the route's allowed roles and
the `{school_id}` in the path — non-admins can only act within their own
school, and students/parents can only read their own/child's records.

**`migrate` re-runs cleanly every time**
That's expected — Liquibase tracks applied changesets in
`databasechangelog`; re-running `make migrate` after adding new files in
`sql/changelogs/` only applies the new ones.

**Fresh schema changes not reflected**
Add a new numbered file under `sql/changelogs/` (never edit an already-
applied one — Liquibase checksums the content) and run `make migrate`
again.

---

## 8. Phase 2 (AWS)

Two self-contained CloudFormation templates provision the AWS footprint;
neither is deployed or required for local development.

- **`devops/aws/aws.cfn.core.yml`** — one-time per environment: IAM roles
  (deploy/execution/task), ECR repo, CloudWatch log group, artifacts S3
  bucket. Only supports `Environment: dev` or `prod` for now.
- **`aws.cfn.app.yml`** — everything else: its own VPC (2 AZs, public +
  private subnets, NAT), an ALB + ECS Fargate service, Aurora PostgreSQL
  (Serverless v2, credentials auto-managed in Secrets Manager), and
  **skeleton, OFF-by-default** ElastiCache Redis and MSK Kafka resources
  (`EnableRedis`/`EnableKafka` parameters) — flip them on later with a stack
  update once phase 2 code actually uses `src/cache`/`src/messaging`, no
  template rewrite needed.

Deploy order: `aws.cfn.core.yml` first, then `aws.cfn.app.yml` with the same
`ServiceName`/`Environment` parameters (the app stack imports the core
stack's role ARNs/ECR URI/log group via CloudFormation exports).

Both templates also carry two placeholder parameters, `MatiServiceBaseUrl`
and `GatiServiceBaseUrl`, reserved for the first external microservice
integrations Vidhya is expected to grow into (e.g. notifications, transport/
logistics). They're wired through as env vars (`APP_MATI_SERVICE_URL`,
`APP_GATI_SERVICE_URL`) and SSM parameters but unused by the app today -
fill them in when those services exist instead of editing the template.

These templates create real, billable AWS resources (VPC/NAT/ALB/Aurora at
minimum). Get phase 1 fully working locally first, and review
`ContainerCpu`/`ContainerMemory`/`DbMinCapacityAcu`/instance types before
deploying to keep dev-environment cost low.
