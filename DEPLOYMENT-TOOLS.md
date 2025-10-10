# Deployment Automation Tools for Freeshell.org

## 🎯 1. Ansible (Recommended)

**Best for: Traditional SSH-based hosting like freeshell.org**

### Installation:
```bash
pip install ansible
```

### Usage:
```bash
# Edit inventory.yml with your freeshell username
# Deploy:
ansible-playbook -i ansible/inventory.yml ansible/deploy.yml

# Check status:
ansible freeshell -i ansible/inventory.yml -m shell -a "~/book-tracker/manage.sh status"
```

### Features:
- ✅ **Idempotent** - safe to run multiple times
- ✅ **Systemd integration** - proper service management
- ✅ **Rollback support** with `--check` and `--diff`
- ✅ **Secrets management** - auto-generates JWT secrets
- ✅ **Built-in logging** and health checks
- ✅ **Template-based configuration**

---

## 🚀 2. Fabric (Python-based)

**Best for: Python developers who want programmatic control**

### Setup:
```bash
pip install fabric
```

### Create `fabfile.py`:
```python
from fabric import task, Connection
import os

@task
def deploy(c, host="freeshell.org", user="your-username"):
    """Deploy Book Tracker to freeshell.org"""
    
    # Build locally
    c.local("cd frontend && npm ci && npm run build")
    c.local("cd backend && ./gradlew clean nativeCompile")
    
    # Connect to remote
    conn = Connection(f"{user}@{host}")
    
    # Create directories
    conn.run("mkdir -p ~/book-tracker/backend ~/book-tracker/data ~/public_html/book-tracker")
    
    # Upload files
    conn.put("backend/build/native/nativeCompile/book-tracker", "~/book-tracker/backend/")
    conn.put("frontend/dist/", "~/public_html/book-tracker/", recursive=True)
    
    # Set permissions and restart
    conn.run("chmod +x ~/book-tracker/backend/book-tracker")
    conn.run("~/book-tracker/manage.sh restart")

@task
def status(c, host="freeshell.org", user="your-username"):
    """Check application status"""
    conn = Connection(f"{user}@{host}")
    conn.run("~/book-tracker/manage.sh status")
```

### Usage:
```bash
fab deploy
fab status
```

---

## 🔧 3. GitHub Actions + SSH

**Best for: Automated CI/CD on git push**

### Create `.github/workflows/deploy.yml`:
```yaml
name: Deploy to Freeshell.org

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '18'
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json
    
    - name: Setup GraalVM
      uses: graalvm/setup-graalvm@v1
      with:
        java-version: '21'
        distribution: 'graalvm-community'
    
    - name: Build Frontend
      run: |
        cd frontend
        npm ci
        npm run build
    
    - name: Build Backend
      run: |
        cd backend
        ./gradlew clean nativeCompile
    
    - name: Deploy to Freeshell
      uses: appleboy/ssh-action@v1.0.0
      with:
        host: freeshell.org
        username: ${{ secrets.FREESHELL_USERNAME }}
        key: ${{ secrets.FREESHELL_SSH_KEY }}
        script: |
          mkdir -p ~/book-tracker/backend ~/book-tracker/data ~/public_html/book-tracker
    
    - name: Upload Files
      uses: appleboy/scp-action@v0.1.4
      with:
        host: freeshell.org
        username: ${{ secrets.FREESHELL_USERNAME }}
        key: ${{ secrets.FREESHELL_SSH_KEY }}
        source: "backend/build/native/nativeCompile/book-tracker,frontend/dist/*"
        target: "~/book-tracker/"
        strip_components: 2
    
    - name: Restart Service
      uses: appleboy/ssh-action@v1.0.0
      with:
        host: freeshell.org
        username: ${{ secrets.FREESHELL_USERNAME }}
        key: ${{ secrets.FREESHELL_SSH_KEY }}
        script: |
          chmod +x ~/book-tracker/backend/book-tracker
          ~/book-tracker/manage.sh restart
```

---

## 🐳 4. Docker Compose + SSH (Hybrid)

**Best for: Consistent local/remote environments**

### Create `docker-compose.prod.yml`:
```yaml
version: '3.8'
services:
  book-tracker:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    volumes:
      - ~/book-tracker/data:/app/data
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - DATABASE_URL=file:/app/data/booktracker.db
    restart: unless-stopped
```

### Deploy script:
```bash
#!/bin/bash
# Build and export Docker image
docker build -t book-tracker:latest backend/
docker save book-tracker:latest | gzip > book-tracker.tar.gz

# Upload and deploy
scp book-tracker.tar.gz your-username@freeshell.org:~/
ssh your-username@freeshell.org << 'EOF'
  docker load < book-tracker.tar.gz
  docker-compose -f docker-compose.prod.yml up -d
EOF
```

---

## 📊 Comparison Matrix

| Tool | Complexity | Features | Best For |
|------|------------|----------|----------|
| **Ansible** | Medium | ⭐⭐⭐⭐⭐ | Production deployments |
| **Fabric** | Low | ⭐⭐⭐ | Python developers |
| **GitHub Actions** | Medium | ⭐⭐⭐⭐ | Automated CI/CD |
| **Shell Script** | Low | ⭐⭐ | Quick deployments |
| **Docker Compose** | High | ⭐⭐⭐⭐ | Container environments |

## 🎯 Recommendation

**Use Ansible** for freeshell.org deployment because:
- Perfect for SSH-based traditional hosting
- Handles systemd service management
- Idempotent and safe
- Great for both development and production
- Built-in rollback and health checking

## 🚀 Quick Start with Ansible

1. **Install Ansible:**
   ```bash
   pip install ansible
   ```

2. **Edit inventory:**
   ```bash
   # Edit ansible/inventory.yml
   # Replace "your-username" with your freeshell username
   ```

3. **Deploy:**
   ```bash
   ansible-playbook -i ansible/inventory.yml ansible/deploy.yml
   ```

4. **Manage:**
   ```bash
   # SSH to freeshell and use management script
   ssh your-username@freeshell.org
   ~/book-tracker/manage.sh status
   ~/book-tracker/manage.sh logs
   ~/book-tracker/manage.sh backup-db
   ```

Your single executable + SQLite architecture is perfect for freeshell.org hosting!