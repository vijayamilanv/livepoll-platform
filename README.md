# LivePoll — Real-Time Audience Polling Platform

> **GUVI × HCL Developer Internship Task**

A production-quality live polling platform: **Create poll → Share link → Audience votes → Live results (no refresh needed).**

---

## 🚀 Live Demo

| Service | URL |
|---------|-----|
| Frontend | *(deploy to Vercel — see Deployment section)* |
| Backend | *(deploy to Render — see Deployment section)* |

---

## 🏗 Architecture

```
Browser (React/Vite)
  ├── REST (axios)  → auth, create poll, cast vote, fetch poll
  └── WebSocket     → /ws/poll/:shareCode → live count stream

Go (Gin) Backend
  ├── REST handlers:  /auth/*, /polls, /polls/:code/vote
  ├── WebSocket hub:  goroutine-safe room per poll
  ├── On vote:        validate → Mongo write → Redis HINCRBY → Redis PUBLISH
  └── Redis sub goroutine → fans out to all WS clients in that room

MongoDB: users, polls, votes (source of truth, audit log)
Redis:   hash counters + pub/sub channels (real-time, ephemeral)
```

---

## 🔴 How Redis Is Used (Graded Criteria)

### 1. Live Vote Counters — O(1) per vote
```
HINCRBY poll:{pollId}:counts {optionId} 1
```
Every vote increments an in-memory Redis hash field atomically. Reading
live counts is a single `HGETALL` — no Mongo aggregation on every request.

### 2. Redis Pub/Sub — instant fan-out
```
PUBLISH poll:{pollId}:events '{"pollId":"...","counts":[...]}'
```
When a vote lands, the backend publishes the new counts JSON to a
channel. A subscriber goroutine (one per active poll room) receives it
instantly and pushes it to every connected WebSocket client — regardless
of which backend instance handled the vote request.

### 3. Cold-Cache Reconciliation
On first fetch of a poll, if Redis has no counters (`HEXISTS` == 0),
the backend runs a MongoDB aggregation (`$group` by optionId) and warms
Redis with `HSET`. Thereafter Redis is authoritative for live counts.

### 4. Horizontal Scale Reasoning
Because all backend instances share the same Redis pub/sub channel, a
vote received on Instance A is instantly fanned out to WebSocket clients
connected to Instance B. This is the standard Redis fan-out pattern for
horizontally-scaled real-time services.

---

## 📦 Tech Stack

| Layer | Tech |
|-------|------|
| Frontend | React 18, Vite, TypeScript, React Router v6, Axios |
| Backend | Go 1.22, Gin, Gorilla WebSocket |
| Database | MongoDB 7 (Atlas in prod, Docker in dev) |
| Realtime | Redis 7 — counters (HINCRBY) + pub/sub |
| Auth | JWT (HS256), bcrypt (cost 10) |
| Dev infra | Docker Compose (all 4 services) |

---

## ⚙️ Local Development (Docker — No installs needed)

### Prerequisites
- Docker Desktop (running)
- Node.js 18+ (only for local frontend dev without Docker)

### 1. Start everything with Docker Compose

```bash
# From the project root:
docker compose up --build
```

This starts:
- **MongoDB** on port 27017
- **Redis** on port 6379
- **Go backend** on port 8080
- **React frontend** on port 3000

Open **http://localhost:3000**

### 2. Verify the backend health check
```bash
curl http://localhost:8080/health
# → {"status":"ok"}
```

### 3. Local frontend dev (hot reload)
```bash
cd frontend
npm install
npm run dev   # → http://localhost:5173
```

Make sure `frontend/.env` points to `http://localhost:8080`.

---

## 🌐 Deployment Guide

### Step 1 — MongoDB Atlas (free tier)
1. Sign up at https://cloud.mongodb.com
2. Create a free cluster (M0)
3. Create a DB user with read/write access
4. Get the connection string: `mongodb+srv://user:pass@cluster.mongodb.net/livepoll`
5. Add your backend server IP to the IP allowlist (or `0.0.0.0/0` for open)

### Step 2 — Upstash Redis (free tier, supports pub/sub)
1. Sign up at https://upstash.com
2. Create a Redis database (Global or regional)
3. Copy **Endpoint** and **Password**
4. ✅ Upstash free tier supports Redis Pub/Sub

### Step 3 — Deploy Backend (Render)
1. Push this repo to GitHub
2. Go to https://render.com → New → Web Service
3. Connect your GitHub repo, root dir: `backend`
4. Build command: `go build -o server .`
5. Start command: `./server`
6. Set environment variables:
   ```
   MONGO_URI=<atlas connection string>
   MONGO_DB=livepoll
   REDIS_ADDR=<upstash endpoint>:<port>
   REDIS_PASSWORD=<upstash password>
   JWT_SECRET=<generate a 64-char random string>
   PORT=8080
   FRONTEND_URL=https://your-app.vercel.app
   ```
7. Enable "WebSockets" in Render settings (required for WS support)

### Step 4 — Deploy Frontend (Vercel)
1. Go to https://vercel.com → New Project → import your repo
2. Set root directory to `frontend`
3. Set environment variables:
   ```
   VITE_API_URL=https://your-backend.onrender.com
   VITE_WS_URL=wss://your-backend.onrender.com
   ```
4. Deploy → get your URL → update `FRONTEND_URL` on Render

---

## 🔑 Key Design Decisions

### Vote Deduplication
**Decision:** One vote per (IP + User-Agent) hash per poll.

A `SHA-256(IP | UserAgent)` fingerprint is stored per vote. A unique
compound index `{pollId: 1, voterFingerprint: 1}` on the votes collection
enforces this at the database level — no application-level race conditions.

**Tradeoff:** A determined user with VPN + private browsing can bypass this.
Full deduplication requires user accounts for voting, which conflicts with the
spec requirement that voting be open to unauthenticated users. This approach
provides best-effort fairness without being onerous.

### JWT Storage
**Decision:** `localStorage` with a 24-hour expiry.

**Security note:** `localStorage` is susceptible to XSS attacks. The
production-hardened alternative is an httpOnly cookie via a `/auth/refresh`
endpoint. localStorage was chosen here for simplicity and developer UX (no
cross-origin cookie configuration needed). For a real production app, use
httpOnly cookies and implement CSRF protection.

### CORS
The backend only accepts requests from the `FRONTEND_URL` env var origin.
`AllowWildcard` is explicitly set to `false`. Update `FRONTEND_URL` to your
deployed Vercel URL before production use.

### Rate Limiting
The vote endpoint is protected by the voter fingerprint dedup (prevents
one user from spamming votes). Full IP-based rate limiting (e.g., 5 votes/min
per IP) is out of scope but documented here as the next production step.
Recommended library: `github.com/ulule/limiter` with a Redis backend.

---

## 📁 Project Structure

```
/backend
  ├── main.go              → Server entry: config, DB, routes
  ├── config/config.go     → Env var loading
  ├── database/mongo.go    → MongoDB connection + index setup
  ├── redisdb/redis.go     → Redis client, HINCRBY, pub/sub helpers
  ├── middleware/auth.go   → JWT middleware
  ├── models/models.go     → User, Poll, Vote structs
  ├── services/
  │   ├── auth.go          → Signup, login, JWT generation
  │   └── poll.go          → CreatePoll, GetPoll, Vote + Redis logic
  ├── handlers/
  │   ├── auth.go          → POST /auth/signup, POST /auth/login
  │   ├── poll.go          → Poll CRUD + vote handlers
  │   └── websocket.go     → WebSocket upgrade + initial snapshot
  ├── hub/hub.go           → WebSocket hub, rooms, Redis subscriber goroutine
  └── Dockerfile

/frontend
  ├── src/
  │   ├── api/             → Axios instance, auth.ts, polls.ts
  │   ├── contexts/        → AuthContext (JWT + user state)
  │   ├── hooks/
  │   │   └── useLivePoll.ts → WebSocket hook with auto-reconnect
  │   ├── components/      → Navbar, AuthGuard
  │   └── pages/           → Login, Signup, CreatePoll, Vote, Results, MyPolls
  └── Dockerfile + nginx.conf
```

---

## 📋 API Reference

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /health | — | Health check |
| POST | /auth/signup | — | Register, returns JWT |
| POST | /auth/login | — | Login, returns JWT |
| POST | /polls | JWT | Create poll |
| GET | /polls/mine | JWT | List my polls |
| GET | /polls/:shareCode | — | Get poll + live counts |
| POST | /polls/:shareCode/vote | — | Cast vote |
| GET | /ws/poll/:shareCode | — | WebSocket upgrade |

---

## 🎬 Video Walkthrough

*(Record a 3–5 min video covering: the hardest challenge solved and AI tool usage disclosure. Upload to YouTube/Drive and link here.)*

---

## 🧪 Testing the Live Feature

1. Open two browser tabs: `http://localhost:3000/poll/<shareCode>/results`
2. In a third tab: `http://localhost:3000/poll/<shareCode>` → cast a vote
3. Watch both result tabs update **instantly** without any refresh

To verify Redis is doing real work:
```bash
docker exec -it livepoll_redis redis-cli
> HGETALL poll:<pollId>:counts
> SUBSCRIBE poll:<pollId>:events
```
