# Book Tracker

A full-stack application for tracking children's reading progress with sharing capabilities.

## Features

### User Roles
- **Administrator**: Manage users, promote/demote admins
- **Normal User**: Manage children and their book lists, invite others to view/edit

### Core Functionality
- Add and manage children (name, age)
- Track books read (title, author, date read)
- Share child data with others (viewer or editor permissions)
- Generate reading reports (exportable to CSV)
- Responsive design for all devices

## Tech Stack

### Frontend
- React 18 with Vite
- Tailwind CSS for styling
- React Router for navigation
- Axios for API calls
- Heroicons for icons

### Backend
- Node.js with Express
- Drizzle ORM with SQLite/Turso
- JWT authentication
- bcryptjs for password hashing

## Development Setup

### Prerequisites
- Node.js 20+ 
- npm

### Frontend Setup
```bash
npm install
npm run dev
```

### Backend Setup
```bash
cd backend
npm install
cp .env.example .env
# Edit .env with your configuration
npm run dev
```

### Database Setup
```bash
cd backend
npm run db:generate
npm run db:migrate
```

## Environment Variables

### Backend (.env)
```
DATABASE_URL=file:./booktracker.db
JWT_SECRET=your-secret-key-change-this-in-production
PORT=8080
```

For Turso (recommended for production):
```
DATABASE_URL=libsql://your-database-url
DATABASE_AUTH_TOKEN=your-auth-token
```

## Deployment

### Render.com (Recommended)

1. Connect your GitHub repository to Render
2. The `render.yaml` blueprint will automatically configure:
   - Backend service (Node.js)
   - Frontend service (Static site)
   - PostgreSQL database (free tier)

3. Set environment variables in Render dashboard:
   - `JWT_SECRET` (auto-generated)
   - `DATABASE_URL` (auto-configured)

### Manual Deployment

#### Frontend
```bash
npm run build
# Deploy dist/ folder to any static hosting service
```

#### Backend
```bash
cd backend
npm install
npm start
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user

### Users (Admin only)
- `GET /api/users` - List all users
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user

### Children
- `GET /api/children` - List user's children
- `POST /api/children` - Create child
- `GET /api/children/:id` - Get child details
- `PUT /api/children/:id` - Update child
- `DELETE /api/children/:id` - Delete child

### Books
- `GET /api/books/child/:childId` - List child's books
- `POST /api/books/child/:childId` - Add book to child
- `GET /api/books/:id` - Get book details
- `PUT /api/books/:id` - Update book
- `DELETE /api/books/:id` - Delete book

### Permissions
- `POST /api/permissions/invite` - Invite user to access child
- `GET /api/permissions/child/:childId` - List child permissions
- `DELETE /api/permissions/:userId/:childId` - Remove permission

### Reports
- `GET /api/reports/my-books` - Generate reading report

## Database Schema

### Users
- id, email, passwordHash, firstName, lastName, isAdmin
- timestamps: createdAt, updatedAt

### Children
- id, name, age, ownerId (references users)
- timestamps: createdAt, updatedAt

### Books
- id, title, author, dateRead, childId (references children)
- timestamps: createdAt, updatedAt

### Permissions
- id, userId (references users), childId (references children)
- permissionType: 'VIEWER' | 'EDITOR'
- timestamp: createdAt

## License

MIT
