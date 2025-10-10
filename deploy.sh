#!/bin/bash
set -e

# Configuration
FREESHELL_USER="your-username"
FREESHELL_HOST="freeshell.org"
REMOTE_DIR="~/book-tracker"
FRONTEND_REMOTE_DIR="~/public_html/book-tracker"

echo "🚀 Deploying Book Tracker to freeshell.org..."

# Build frontend
echo "📦 Building frontend..."
cd frontend
npm ci
npm run build
cd ..

# Build backend native image
echo "🔨 Building backend native executable..."
cd backend
./gradlew clean nativeCompile
cd ..

# Create deployment package
echo "📋 Creating deployment package..."
mkdir -p dist/backend
mkdir -p dist/frontend

# Copy backend executable and resources
cp backend/build/native/nativeCompile/book-tracker dist/backend/
cp -r backend/src/main/resources dist/backend/

# Copy frontend build
cp -r frontend/dist/* dist/frontend/

# Create startup script
cat > dist/backend/start.sh << 'EOF'
#!/bin/bash
export PORT=8080
export JWT_SECRET=${JWT_SECRET:-"$(openssl rand -hex 32)"}
export DATABASE_URL="file:$HOME/book-tracker/data/booktracker.db"

# Create data directory
mkdir -p "$HOME/book-tracker/data"

# Start the application
cd "$HOME/book-tracker/backend"
exec ./book-tracker
EOF

chmod +x dist/backend/start.sh

# Create systemd-style service script for process management
cat > dist/backend/book-tracker.service << 'EOF'
#!/bin/bash
# Simple process manager for book-tracker

PIDFILE="$HOME/book-tracker/book-tracker.pid"
LOGFILE="$HOME/book-tracker/book-tracker.log"

case "$1" in
    start)
        if [ -f "$PIDFILE" ] && kill -0 $(cat "$PIDFILE") 2>/dev/null; then
            echo "book-tracker is already running"
            exit 1
        fi
        echo "Starting book-tracker..."
        cd "$HOME/book-tracker/backend"
        nohup ./start.sh > "$LOGFILE" 2>&1 &
        echo $! > "$PIDFILE"
        echo "book-tracker started with PID $(cat $PIDFILE)"
        ;;
    stop)
        if [ -f "$PIDFILE" ]; then
            PID=$(cat "$PIDFILE")
            if kill -0 "$PID" 2>/dev/null; then
                kill "$PID"
                rm -f "$PIDFILE"
                echo "book-tracker stopped"
            else
                echo "book-tracker was not running"
                rm -f "$PIDFILE"
            fi
        else
            echo "book-tracker is not running"
        fi
        ;;
    restart)
        $0 stop
        sleep 2
        $0 start
        ;;
    status)
        if [ -f "$PIDFILE" ] && kill -0 $(cat "$PIDFILE") 2>/dev/null; then
            echo "book-tracker is running with PID $(cat $PIDFILE)"
        else
            echo "book-tracker is not running"
        fi
        ;;
    logs)
        tail -f "$LOGFILE"
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|logs}"
        exit 1
        ;;
esac
EOF

chmod +x dist/backend/book-tracker.service

# Deploy to freeshell.org
echo "🌐 Deploying to freeshell.org..."

# Create directories on remote
ssh "${FREESHELL_USER}@${FREESHELL_HOST}" "mkdir -p ${REMOTE_DIR}/backend ${REMOTE_DIR}/data ${FRONTEND_REMOTE_DIR}"

# Upload backend
echo "📤 Uploading backend..."
scp -r dist/backend/* "${FREESHELL_USER}@${FREESHELL_HOST}:${REMOTE_DIR}/backend/"

# Upload frontend to public_html
echo "📤 Uploading frontend..."
scp -r dist/frontend/* "${FREESHELL_USER}@${FREESHELL_HOST}:${FRONTEND_REMOTE_DIR}/"

# Set executable permissions
ssh "${FREESHELL_USER}@${FREESHELL_HOST}" "chmod +x ${REMOTE_DIR}/backend/book-tracker ${REMOTE_DIR}/backend/start.sh ${REMOTE_DIR}/backend/book-tracker.service"

# Create .htaccess for React Router (SPA support)
cat > dist/.htaccess << 'EOF'
Options -MultiViews
RewriteEngine On
RewriteCond %{REQUEST_FILENAME} !-f
RewriteRule ^ index.html [QSA,L]
EOF

scp dist/.htaccess "${FREESHELL_USER}@${FREESHELL_HOST}:${FRONTEND_REMOTE_DIR}/"

echo "✅ Deployment complete!"
echo ""
echo "🔧 To start the backend service:"
echo "   ssh ${FREESHELL_USER}@${FREESHELL_HOST}"
echo "   ~/book-tracker/backend/book-tracker.service start"
echo ""
echo "🌐 Your application will be available at:"
echo "   Frontend: https://${FREESHELL_USER}.freeshell.org/book-tracker/"
echo "   Backend:  https://${FREESHELL_USER}.freeshell.org:8080/health"
echo ""
echo "📋 Service management commands:"
echo "   ~/book-tracker/backend/book-tracker.service {start|stop|restart|status|logs}"

# Cleanup
rm -rf dist
echo "🧹 Cleanup complete"