import { http, HttpResponse } from 'msw'

// Use the same base URL configuration as the actual app
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

export const handlers = [
  // Auth endpoints
  http.post(`${API_BASE_URL}/auth/login`, async ({ request }) => {
    const body = await request.json()
    
    // Add small delay to allow loading state to be visible
    await new Promise(resolve => setTimeout(resolve, 100))
    
    // Check for invalid credentials
    if (body.email === 'wrong@example.com' || body.password === 'wrongpassword') {
      return HttpResponse.json(
        { error: 'Invalid credentials' },
        { status: 401 }
      )
    }
    
    return HttpResponse.json({
      token: 'mock-jwt-token',
      user: {
        id: 1,
        email: 'test@example.com',
        firstName: 'Test',
        lastName: 'User',
        isAdmin: false,
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z'
      }
    })
  }),

  http.post(`${API_BASE_URL}/auth/register`, ({ request }) => {
    return HttpResponse.json({
      id: 1,
      email: 'test@example.com',
      firstName: 'Test',
      lastName: 'User',
      isAdmin: false,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z'
    }, { status: 201 })
  }),

  // Children endpoints
  http.get(`${API_BASE_URL}/children`, ({ request }) => {
    return HttpResponse.json([
      {
        id: 1,
        name: 'Alice',
        age: 8,
        ownerId: 1,
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z'
      },
      {
        id: 2,
        name: 'Bob',
        age: 10,
        ownerId: 1,
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z'
      }
    ])
  }),

  http.post(`${API_BASE_URL}/children`, ({ request }) => {
    return HttpResponse.json({
      id: 3,
      name: 'New Child',
      age: 6,
      ownerId: 1,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z'
    }, { status: 201 })
  }),

  // Books endpoints
  http.get(`${API_BASE_URL}/books/child/:childId`, ({ params }) => {
    return HttpResponse.json([
      {
        id: 1,
        title: 'The Cat in the Hat',
        author: 'Dr. Seuss',
        dateRead: '2024-01-01T00:00:00Z',
        childId: parseInt(params.childId),
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z'
      }
    ])
  }),

  http.post(`${API_BASE_URL}/books/child/:childId`, ({ params, request }) => {
    return HttpResponse.json({
      id: 2,
      title: 'Green Eggs and Ham',
      author: 'Dr. Seuss',
      dateRead: '2024-01-01T00:00:00Z',
      childId: parseInt(params.childId),
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z'
    }, { status: 201 })
  }),

  // Reports endpoints
  http.get(`${API_BASE_URL}/reports/my-books`, ({ request }) => {
    return HttpResponse.json({
      children: [
        {
          child: {
            id: 1,
            name: 'Alice',
            age: 8,
            ownerId: 1,
            createdAt: '2024-01-01T00:00:00Z',
            updatedAt: '2024-01-01T00:00:00Z'
          },
          books: [
            {
              id: 1,
              title: 'The Cat in the Hat',
              author: 'Dr. Seuss',
              dateRead: '2024-01-01T00:00:00Z',
              childId: 1,
              createdAt: '2024-01-01T00:00:00Z',
              updatedAt: '2024-01-01T00:00:00Z'
            }
          ],
          totalBooks: 1
        }
      ]
    })
  }),

  // Error cases
  http.post(`${API_BASE_URL}/auth/login`, ({ request }) => {
    const url = new URL(request.url)
    if (url.searchParams.get('error') === 'unauthorized') {
      return HttpResponse.json({ error: 'Invalid credentials' }, { status: 401 })
    }
  })
]