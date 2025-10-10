# Render.com Deployment Guide

This guide covers deploying the Book Tracker application to Render.com with the optimal architecture: static frontend + single executable backend.

## Architecture Overview

- **Frontend**: Static React site served via Render's static hosting
- **Backend**: Single GraalVM native executable with embedded SQLite database
- **Database**: SQLite file on persistent disk (auto-created via Flyway migrations)

## Prerequisites

1. Render.com account
2. GitHub repository connected to Render
3. Render Starter plan (required for persistent disk storage)

## Deployment Steps

### 1. Connect Repository
- Go to [Render Dashboard](https://dashboard.render.com)
- Click "New +" → "Blueprint"
- Connect your GitHub repository
- Render will automatically detect the `render.yaml` configuration

### 2. Configure Services

The `render.yaml` configures two services:

#### Backend API (`book-tracker-api`)
- **Type**: Web service with Docker
- **Plan**: Starter ($7/month - required for disk storage)
- **Features**:
  - Single 66MB native executable
  - 1GB persistent disk for SQLite database
  - Auto-generated JWT secret
  - Health checks enabled

#### Frontend (`book-tracker-frontend`)
- **Type**: Static site
- **Plan**: Free
- **Features**:
  - Pre-built React app served from CDN
  - Environment variable for API URL

### 3. Environment Variables

#### Backend (Auto-configured)
- `JWT_SECRET`: Auto-generated secure key
- `DATABASE_URL`: Points to persistent disk location
- `PORT`: Set to 10000 (Render requirement)

#### Frontend (Auto-configured)  
- `VITE_API_URL`: Points to your backend API URL

### 4. Custom Domain (Optional)
- Backend: `https://your-api-domain.com`
- Frontend: `https://your-app-domain.com`

## Database Management

### Automatic Setup
- Database is created automatically on first startup
- Flyway runs all migrations from `backend/src/main/resources/db/migration/`
- No manual database setup required

### Backups
- SQLite database persists on 1GB disk
- Render provides automatic disk snapshots
- For additional backups, use the admin test endpoint: `DELETE /api/test/reset-db`

### Migrations
- Add new migration files as `V2__Description.sql`, `V3__Description.sql`, etc.
- Migrations run automatically on deployment
- All migrations are reversible through versioning

## Monitoring

### Health Checks
- Endpoint: `GET /health`
- Returns: `OK` for healthy service
- Render automatically restarts unhealthy instances

### Logs
- View logs in Render dashboard
- GraalVM native executable provides fast startup (~100ms)
- Memory usage: ~50MB at runtime

## Cost Breakdown

| Service | Plan | Cost | Features |
|---------|------|------|----------|
| Frontend | Free | $0/month | 100GB bandwidth, global CDN |
| Backend | Starter | $7/month | 0.5 CPU, 512MB RAM, 1GB disk |
| **Total** | | **$7/month** | Production-ready setup |

## Security Features

- ✅ HTTPS termination by Render
- ✅ Auto-generated JWT secrets
- ✅ Minimal attack surface (distroless container)
- ✅ No exposed database ports
- ✅ CORS configured for frontend domain

## Performance

- **Backend**: 66MB native executable, ~100ms startup
- **Frontend**: Static files served from global CDN
- **Database**: SQLite with WAL mode for concurrent access
- **Scaling**: Vertical scaling available on higher plans

## Development vs Production

| Environment | Frontend | Backend | Database |
|-------------|----------|---------|----------|
| **Development** | `npm run dev` | `./gradlew run` | Local SQLite |
| **Production** | Static CDN | Native executable | Persistent SQLite |

## Troubleshooting

### Backend Won't Start
1. Check logs for database connection errors
2. Verify persistent disk is mounted at `/opt/render/project/src/data`
3. Ensure migrations can run (check SQL syntax)

### Frontend Can't Connect to Backend
1. Verify `VITE_API_URL` points to correct backend URL
2. Check CORS settings if using custom domain
3. Ensure backend health check passes

### Database Issues
1. Check disk space usage in Render dashboard
2. Verify Flyway migrations in backend logs
3. Use test reset endpoint for development: `DELETE /api/test/reset-db`

## Manual Deployment Commands

If needed, you can deploy manually:

```bash
# Build frontend
cd frontend && npm ci && npm run build

# Build backend native image  
cd backend && ./gradlew nativeCompile

# Test locally
./backend/build/native/nativeCompile/book-tracker
```

## Next Steps

1. Deploy using the `render.yaml` configuration
2. Set up custom domains (optional)
3. Configure monitoring alerts
4. Set up automated backups (optional)

Your Book Tracker application will be live with a single native executable backend and static frontend, exactly as requested!