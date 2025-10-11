import { useState, useEffect, useRef } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { CheckCircleIcon, ExclamationTriangleIcon, BookOpenIcon } from '@heroicons/react/24/outline'
import api from '../services/api'

export default function VerifyEmail() {
  const [status, setStatus] = useState('verifying') // 'verifying', 'success', 'error'
  const [message, setMessage] = useState('')
  const [user, setUser] = useState(null)
  const hasAttempted = useRef(false)
  const location = useLocation()
  const navigate = useNavigate()

  useEffect(() => {
    const verifyToken = async () => {
      // Prevent multiple API calls
      if (hasAttempted.current) return
      hasAttempted.current = true

      try {
        const params = new URLSearchParams(location.search)
        const token = params.get('token')

        if (!token) {
          setStatus('error')
          setMessage('No verification token provided')
          return
        }

        const response = await api.get(`/auth/verify-email?token=${token}`)
        setStatus('success')
        setMessage(response.data.message)
        setUser(response.data.user)

        // Auto-redirect to login after 3 seconds
        setTimeout(() => {
          navigate('/login', { 
            state: { 
              message: 'Email verified successfully! You can now log in.',
              type: 'success'
            }
          })
        }, 3000)

      } catch (error) {
        setStatus('error')
        setMessage(error.response?.data?.message || 'Verification failed')
      }
    }

    verifyToken()
  }, [location, navigate])

  const handleGoToLogin = () => {
    navigate('/login', { 
      state: { 
        message: status === 'success' ? 'Email verified successfully! You can now log in.' : null,
        type: 'success'
      }
    })
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <div className="flex justify-center">
            <BookOpenIcon className="h-12 w-12 text-indigo-600" />
          </div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Email Verification
          </h2>
        </div>

        <div className="bg-white py-8 px-6 shadow rounded-lg">
          {status === 'verifying' && (
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
              <p className="mt-4 text-sm text-gray-600">Verifying your email address...</p>
            </div>
          )}

          {status === 'success' && (
            <div className="text-center">
              <CheckCircleIcon className="h-12 w-12 text-green-500 mx-auto" />
              <h3 className="mt-4 text-lg font-medium text-gray-900">Verification Successful!</h3>
              <p className="mt-2 text-sm text-gray-600">{message}</p>
              {user && (
                <p className="mt-2 text-sm text-gray-500">
                  Welcome, {user.firstName}! You will be redirected to login in a few seconds.
                </p>
              )}
              <button
                onClick={handleGoToLogin}
                className="mt-4 w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                Go to Login
              </button>
            </div>
          )}

          {status === 'error' && (
            <div className="text-center">
              <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto" />
              <h3 className="mt-4 text-lg font-medium text-gray-900">Verification Failed</h3>
              <p className="mt-2 text-sm text-gray-600">{message}</p>
              <div className="mt-6 flex flex-col space-y-3">
                <button
                  onClick={handleGoToLogin}
                  className="w-full flex justify-center py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  Go to Login
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}