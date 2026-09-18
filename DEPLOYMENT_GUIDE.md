# LivePoll Platform — Step-by-Step Production Deployment Guide

This guide walks you through deploying the **LivePoll Platform** to production using free-tier cloud services:
- **Database**: MongoDB Atlas (Free M0 Cluster)
- **Cache & Pub/Sub**: Upstash Redis (Free Tier with WebSocket/PubSub support)
- **Backend Service**: Render (Go Web Service)
- **Frontend App**: Vercel (React + Vite SPA)

---

## 🛠 Prerequisites Checklist

- [ ] GitHub account (repo pushed to GitHub)
- [ ] MongoDB Atlas account ([https://cloud.mongodb.com](https://cloud.mongodb.com))
- [ ] Upstash account ([https://upstash.com](https://upstash.com))
- [ ] Render account ([https://render.com](https://render.com))
- [ ] Vercel account ([https://vercel.com](https://vercel.com))

---

## Step 1: MongoDB Atlas Setup (Database)

1. Log in to [MongoDB Atlas](https://cloud.mongodb.com).
2. Click **Create a Cluster** $\rightarrow$ select the **M0 Shared (Free)** tier.
3. Choose a region close to your target users (e.g., `aws / us-east-1` or `aws / ap-south-1`).
4. **Create Database User**:
   - Go to **Security** $\rightarrow$ **Database Access** $\rightarrow$ **Add New Database User**.
   - Auth Method: **Password**.
   - Save your username (e.g., `livepoll_admin`) and password.
   - User Privileges: `Read and write to any database`.
5. **Network Access**:
   - Go to **Security** $\rightarrow$ **Network Access** $\rightarrow$ **Add IP Address**.
   - Select **Allow Access from Anywhere** (`0.0.0.0/0`) so Render instances can connect.
6. **Get Connection String**:
   - Go to **Database** $\rightarrow$ **Connect** $\rightarrow$ **Drivers**.
   - Copy connection URI. It looks like:
     ```text
     mongodb+srv://<username>:<password>@cluster0.abcde.mongodb.net/livepoll?retryWrites=true&w=majority
     ```

---

## Step 2: Upstash Redis Setup (Real-Time Pub/Sub)

1. Log in to [Upstash Console](https://console.upstash.com).
2. Click **Create Database**.
3. Name: `livepoll-redis`
4. Type: **Primary**, select your region.
5. Click **Create**.
6. Under **Details / Connection Details**:
   - Note down **Endpoint** (e.g., `epic-bird-12345.upstash.io`)
   - Note down **Port** (e.g., `6379` or `31234`)
   - Note down **Password** (click eye icon to view)
7. Final `REDIS_ADDR` format: `epic-bird-12345.upstash.io:6379` (or the specific port shown).

---

## Step 3: Deploy Backend on Render

1. Log in to [Render](https://dashboard.render.com).
2. Click **New +** $\rightarrow$ **Web Service**.
3. Connect your GitHub repository.
4. Fill out service settings:
   - **Name**: `livepoll-backend`
   - **Root Directory**: `backend`
   - **Runtime**: `Go`
   - **Build Command**: `go build -o server .`
   - **Start Command**: `./server`
5. **Environment Variables**:
   Add the following variables under **Environment**:

   | Variable Name | Example / Value |
   |---|---|
   | `MONGO_URI` | `mongodb+srv://user:pass@cluster.mongodb.net/livepoll?retryWrites=true&w=majority` |
   | `MONGO_DB` | `livepoll` |
   | `REDIS_ADDR` | `your-upstash-endpoint.upstash.io:6379` |
   | `REDIS_PASSWORD` | `your_upstash_password` |
   | `JWT_SECRET` | *(Generate a 64-character random string)* |
   | `PORT` | `8080` |
   | `FRONTEND_URL` | `https://your-app.vercel.app` *(update after Step 4)* |

6. Click **Create Web Service**.
7. Once deployed, note down your backend URL (e.g., `https://livepoll-backend.onrender.com`).

---

## Step 4: Deploy Frontend on Vercel

1. Log in to [Vercel](https://vercel.com).
2. Click **Add New...** $\rightarrow$ **Project**.
3. Import your GitHub repository.
4. Configure Project:
   - **Framework Preset**: `Vite`
   - **Root Directory**: Select `frontend`
5. **Environment Variables**:
   Add the following:

   | Variable Name | Value |
   |---|---|
   | `VITE_API_URL` | `https://livepoll-backend.onrender.com` |
   | `VITE_WS_URL` | `wss://livepoll-backend.onrender.com` |

   *(Note: Use `wss://` for secure WebSockets on HTTPS!)*

6. Click **Deploy**.
7. Copy your production frontend URL (e.g., `https://livepoll-platform.vercel.app`).

---

## Step 5: Final Linkage & Verification

1. Return to **Render** $\rightarrow$ `livepoll-backend` $\rightarrow$ **Environment Variables**.
2. Update `FRONTEND_URL` to your exact Vercel URL (e.g., `https://livepoll-platform.vercel.app`).
3. Save changes (Render will automatically re-deploy backend).
4. **Smoke Test**:
   - Open Vercel URL in browser.
   - Register a user account & log in.
   - Create a poll.
   - Open the poll results page in Tab 1 and Tab 2.
   - Submit a vote in Tab 3 $\rightarrow$ Confirm live update appears instantaneously in Tab 1 and Tab 2 without refreshing!
