import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { AuthProvider, useAuth } from '../../contexts/AuthContext'

// Test component to access auth context
const TestComponent = () => {
  const { user, login, register, logout, loading } = useAuth()
  
  return (
    <div>
      <div data-testid="loading">{loading ? 'loading' : 'not loading'}</div>
      <div data-testid="user">{user ? user.email : 'no user'}</div>
      <button 
        onClick={() => login('test@example.com', 'password')}
        data-testid="login-btn"
      >
        Login
      </button>
      <button 
        onClick={() => register({ email: 'test@example.com', password: 'password', firstName: 'Test', lastName: 'User' })}
        data-testid="register-btn"
      >
        Register
      </button>
      <button onClick={logout} data-testid="logout-btn">
        Logout
      </button>
    </div>
  )
}

describe('AuthContext', () => {
  afterEach(() => {
    // Clear localStorage between tests
    localStorage.clear()
  })

  it('provides initial loading state', () => {
    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    expect(screen.getByTestId('loading')).toHaveTextContent('not loading')
    expect(screen.getByTestId('user')).toHaveTextContent('no user')
  })

  it('handles successful login', async () => {
    const user = userEvent.setup()
    
    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    const loginBtn = screen.getByTestId('login-btn')
    await user.click(loginBtn)

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('test@example.com')
    })
  })

  it('handles successful registration', async () => {
    const user = userEvent.setup()
    
    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    const registerBtn = screen.getByTestId('register-btn')
    await user.click(registerBtn)

    // Registration doesn't automatically log in
    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('no user')
    })
  })

  it('handles logout', async () => {
    const user = userEvent.setup()
    
    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    // First login
    const loginBtn = screen.getByTestId('login-btn')
    await user.click(loginBtn)

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('test@example.com')
    })

    // Then logout
    const logoutBtn = screen.getByTestId('logout-btn')
    await user.click(logoutBtn)

    expect(screen.getByTestId('user')).toHaveTextContent('no user')
  })

  it('restores auth state from localStorage', () => {
    // Mock localStorage
    const mockUser = { id: 1, email: 'test@example.com' }
    const mockToken = 'mock-token'
    
    vi.spyOn(Storage.prototype, 'getItem')
      .mockImplementation((key) => {
        if (key === 'user') return JSON.stringify(mockUser)
        if (key === 'token') return mockToken
        return null
      })

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    expect(screen.getByTestId('user')).toHaveTextContent('test@example.com')
  })
})