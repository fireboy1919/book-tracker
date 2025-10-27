#!/bin/bash

echo '🛑 Stopping running instances...'

# Kill backend processes
pkill -f 'book-tracker-go' 2>/dev/null && echo 'Killed backend processes' || echo 'No backend processes found'

# Kill frontend processes  
pkill -f 'npm run dev' 2>/dev/null && echo 'Killed npm dev processes' || echo 'No npm dev processes found'
pkill -f 'vite' 2>/dev/null && echo 'Killed vite processes' || echo 'No vite processes found'

# Kill Vercel processes
pkill -f 'vercel dev' 2>/dev/null && echo 'Killed Vercel dev processes' || echo 'No Vercel dev processes found'

# Kill processes on common development ports
for port in 3000 5173 8080; do
    if lsof -ti:$port >/dev/null 2>&1; then
        lsof -ti:$port | xargs kill -9 2>/dev/null && echo "Freed port $port"
    else
        echo "Port $port already free"
    fi
done

echo 'Waiting 2 seconds for cleanup...'
sleep 2

# Check if we should use Vercel or regular dev
if [ "$1" = "vercel" ]; then
    echo '⚡ Starting Vercel development server...'
    mise run vercel:dev
else
    echo '🚀 Starting development servers...'
    mise run dev
fi