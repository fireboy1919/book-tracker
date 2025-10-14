import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import api from '../services/api'

export default function GoogleCallback() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { login } = useAuth()
  const [user, setUser] = useState(null)

  useEffect(() => {
    const handleCallback = async () => {
      // The backend will handle the OAuth callback and redirect to the frontend with token info
      // Check if there's error from OAuth
      const error = searchParams.get('error')
      const token = searchParams.get('token')
      const userParam = searchParams.get('user')

      if (error) {
        navigate('/login', { 
          state: { 
            error: 'Google authentication failed. Please try again.' 
          }
        })
        return
      }

      if (token && userParam) {
        try {
          // Parse user data and store in localStorage
          const userData = JSON.parse(decodeURIComponent(userParam))
          
          localStorage.setItem('token', token)
          localStorage.setItem('user', JSON.stringify(userData))
          
          // Set authorization header for API calls
          api.defaults.headers.common['Authorization'] = `Bearer ${token}`
          
          // Update the user state to trigger re-render in AuthContext
          setUser(userData)
          
          // Navigate to dashboard
          navigate('/dashboard')
          
        } catch (error) {
          console.error('Failed to process Google login:', error)
          navigate('/login', { 
            state: { 
              error: 'Authentication processing failed. Please try again.' 
            }
          })
        }
      } else {
        // No token received, redirect to login
        navigate('/login', { 
          state: { 
            error: 'Authentication failed. Please try again.' 
          }
        })
      }
    }

    handleCallback()
  }, [navigate, searchParams])

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <div className="flex justify-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
        </div>
        <p className="mt-4 text-center text-gray-600">Completing Google authentication...</p>
      </div>
    </div>
  )
}