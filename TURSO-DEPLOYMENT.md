# Turso + Render.com Free Deployment Guide

Deploy your Book Tracker application completely FREE using Turso (SQLite-compatible database) + Render.com hosting.

## 🎯 Why This Setup?

- **FREE**: Both Turso and Render.com have generous free tiers
- **Zero code changes**: Turso is SQLite-compatible  
- **Global performance**: Turso's edge database network
- **Simple deployment**: Single Docker container + static frontend

## 📋 Prerequisites

1. [Turso account](https://turso.tech) (free)
2. [Render.com account](https://render.com) (free)
3. GitHub repository with your code

## 🚀 Step 1: Set up Turso Database

### 1.1 Install Turso CLI
```bash
curl -sSfL https://get.tur.so/install.sh | bash
```

### 1.2 Login and Create Database
```bash
# Login to Turso
turso auth login

# Create your database
turso db create book-tracker

# Get the database URL
turso db show book-tracker

# Create an auth token
turso db tokens create book-tracker
```

### 1.3 Note Your Connection Details
You'll get something like:
- **Database URL**: `libsql://book-tracker-[username].turso.io`
- **Auth Token**: `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9...`

## 🔧 Step 2: Test Locally with Turso

### 2.1 Set Environment Variable
```bash
export DATABASE_URL="libsql://book-tracker-username.turso.io?authToken=your-token-here"
```

### 2.2 Test Local Connection
```bash
cd backend
./gradlew run
```

You should see:
```
Running database migrations for: jdbc:libsql://book-tracker-username.turso.io
Database migrations completed successfully
Starting server on port 8080
```

### 2.3 Verify Database Setup
```bash
# Check health endpoint
curl http://localhost:8080/health

# Check if tables were created
turso db shell book-tracker
> .tables
> SELECT * FROM flyway_schema_history;
```

## 🌐 Step 3: Deploy to Render.com

### 3.1 Connect Repository
1. Go to [Render Dashboard](https://dashboard.render.com)
2. Click "New +" → "Blueprint"
3. Connect your GitHub repository
4. Select `render-external-db.yaml`

### 3.2 Update Configuration
Before deploying, update `render-external-db.yaml`:

```yaml
# Replace these values with your actual Turso credentials
- key: DATABASE_URL
  value: libsql://your-database-name.turso.io?authToken=your-auth-token
```

### 3.3 Deploy Services
Render will automatically deploy:
- **Backend API**: https://book-tracker-api.onrender.com
- **Frontend**: https://book-tracker-frontend.onrender.com

## ✅ Step 4: Verify Deployment

### 4.1 Check Backend Health
```bash
curl https://book-tracker-api.onrender.com/health
# Should return: OK
```

### 4.2 Test Registration
```bash
curl -X POST https://book-tracker-api.onrender.com/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"firstName":"Test","lastName":"User","email":"test@example.com","password":"password123","isAdmin":false}'
```

### 4.3 Access Frontend
Visit: https://book-tracker-frontend.onrender.com

## 📊 Free Tier Limits

| Service | Free Tier | Limits |
|---------|-----------|--------|
| **Turso** | 500MB storage | 1M row reads/month |
| **Render.com Backend** | 750 hours/month | Sleeps after 15min inactivity |
| **Render.com Frontend** | Unlimited | Global CDN |

## 🔄 Database Management

### Backup Database
```bash
# Export current data
turso db shell book-tracker ".dump" > backup.sql

# Or query specific data
turso db shell book-tracker "SELECT * FROM users;"
```

### View Database in Browser
```bash
# Open Turso web console
turso db shell book-tracker --web
```

### Reset Database (Development)
```bash
# Use the test endpoint
curl -X DELETE https://book-tracker-api.onrender.com/api/test/reset-db
```

## 🚀 Production Considerations

### Environment Variables
For production, set these in Render.com dashboard:
- `JWT_SECRET`: Auto-generated secure key (✓ already configured)
- `DATABASE_URL`: Your Turso connection string (✓ already configured)

### Monitoring
- **Backend logs**: Render.com dashboard → Service → Logs
- **Database metrics**: Turso dashboard → Usage
- **Health checks**: Automated by Render.com

### Scaling
- **Turso**: Upgrade to $29/month for unlimited usage
- **Render.com**: Upgrade to Starter $7/month to prevent sleep

## 🐛 Troubleshooting

### Backend Won't Start
1. Check Render.com logs for database connection errors
2. Verify DATABASE_URL format: `libsql://db.turso.io?authToken=token`
3. Regenerate auth token if expired

### Migration Failures
```bash
# Check migration status in Turso
turso db shell book-tracker
> SELECT * FROM flyway_schema_history;

# Manually run migrations if needed
turso db shell book-tracker < backend/src/main/resources/db/migration/V1__Create_initial_schema.sql
```

### Frontend Can't Connect
1. Verify backend URL in render configuration
2. Check CORS settings for custom domains
3. Ensure backend health check passes

## 🎉 Success!

Your Book Tracker is now running:
- **Backend**: https://book-tracker-api.onrender.com
- **Frontend**: https://book-tracker-frontend.onrender.com
- **Database**: Global Turso SQLite edge network
- **Cost**: FREE (with generous limits)

Perfect for a production book tracking application!