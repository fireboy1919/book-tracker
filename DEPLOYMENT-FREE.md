# Free Deployment Options

Since Render.com requires a paid plan for persistent disk storage, here are your best free alternatives:

## 🎯 Recommended: Railway.app

**Best option for your single executable + SQLite setup:**

### Why Railway?
- ✅ **Free tier includes 512MB persistent storage**
- ✅ **$5/month for unlimited** (cheaper than Render's $7)
- ✅ **Perfect for SQLite + native executable**
- ✅ **GitHub integration**
- ✅ **No 30-day expiration**

### Deployment:
1. Create `railway.toml`:
```toml
[build]
builder = "dockerfile"
dockerfilePath = "backend/Dockerfile"

[deploy]
startCommand = "./book-tracker"
healthcheckPath = "/health"
healthcheckTimeout = 300

[environment]
JWT_SECRET = { generate = true }
DATABASE_URL = "file:/app/data/booktracker.db"
PORT = 3000
```

2. Deploy frontend to **Vercel** (free static hosting)

**Total cost: FREE** (with limitations) or **$5/month** (unlimited)

---

## 🏗️ Alternative: Fly.io

### Why Fly.io?
- ✅ **Generous free tier** (3 shared-cpu-1x, 3GB persistent storage)
- ✅ **Perfect for single executable**
- ✅ **Global edge deployment**
- ✅ **No time limits**

### Deployment:
1. Install Fly CLI: `curl -L https://fly.io/install.sh | sh`
2. Create `fly.toml`:
```toml
app = "book-tracker"
primary_region = "ord"

[build]
  dockerfile = "backend/Dockerfile"

[env]
  PORT = "8080"
  DATABASE_URL = "file:/data/booktracker.db"

[mounts]
  source = "data"
  destination = "/data"

[[services]]
  http_checks = []
  internal_port = 8080
  processes = ["app"]
  protocol = "tcp"
  script_checks = []

  [services.concurrency]
    hard_limit = 25
    soft_limit = 20
    type = "connections"

  [[services.ports]]
    force_https = true
    handlers = ["http"]
    port = 80

  [[services.ports]]
    handlers = ["tls", "http"]
    port = 443

  [[services.tcp_checks]]
    grace_period = "1s"
    interval = "15s"
    restart_limit = 0
    timeout = "2s"
```

3. Deploy: `fly deploy`
4. Deploy frontend to **Netlify** (free static hosting)

**Total cost: FREE**

---

## 🔧 Option 3: Modify for Render Free + PostgreSQL

If you want to stick with Render.com's free tier:

### Required Changes:

1. **Switch from SQLite to PostgreSQL**:
```kotlin
// Update build.gradle.kts
implementation("org.postgresql:postgresql:42.7.2")
```

2. **Update DatabaseConfig.kt**:
```kotlin
config.driverClassName = "org.postgresql.Driver"
config.jdbcUrl = System.getenv("DATABASE_URL")
```

3. **Convert Flyway migrations to PostgreSQL syntax**
4. **Use the `render-free.yaml` configuration**

### Limitations:
- ❌ **Database expires after 30 days**
- ❌ **Must recreate every month**
- ❌ **Only 1GB storage**
- ❌ **Not suitable for production**

---

## 🎯 Recommendation Matrix

| Platform | Monthly Cost | Persistent Storage | Best For |
|----------|--------------|-------------------|----------|
| **Railway.app** | FREE (limited) or $5 | 512MB free, unlimited paid | **Single executable + SQLite** |
| **Fly.io** | FREE | 3GB free | **Production apps** |
| **Render.com** | $7 | 1GB | **Enterprise features** |
| **Render.com Free** | FREE | None (PostgreSQL expires 30d) | **Prototyping only** |

## 🚀 Quick Start: Railway.app (Recommended)

1. **Sign up**: [railway.app](https://railway.app)
2. **Connect GitHub repo**
3. **Use existing Dockerfile**
4. **Deploy frontend to Vercel**
5. **Total setup time**: ~10 minutes
6. **Total cost**: FREE (with limits) or $5/month

Your single executable + SQLite architecture is perfect for Railway.app!

---

## 🔄 Migration Path

If you want to start free and upgrade later:

1. **Start**: Railway.app free tier
2. **Upgrade**: Railway.app $5/month when you hit limits
3. **Scale**: Fly.io or Render.com for global deployment

This gives you a clear progression without architectural changes.