import { useState } from 'react'
import { useAuth } from '../contexts/AuthContext'
import api from '../services/api'
import { ExclamationTriangleIcon, EnvelopeIcon } from '@heroicons/react/24/outline'

export default function EmailVerificationRequired() {
  const [resendLoading, setResendLoading] = useState(false)
  const [resendMessage, setResendMessage] = useState('')
  const [messageType, setMessageType] = useState('') // 'success' or 'error'
  const { user, logout } = useAuth()

  const handleResendVerification = async () => {
    setResendLoading(true)
    setResendMessage('')
    setMessageType('')

    try {
      await api.post('/auth/resend-verification', {
        email: user?.email
      })
      setResendMessage('Verification email sent successfully! Please check your inbox.')
      setMessageType('success')
    } catch (error) {
      setResendMessage(error.response?.data?.message || 'Failed to send verification email')
      setMessageType('error')
    }

    setResendLoading(false)
  }

  const handleLogout = () => {
    logout()
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <img
            className="mx-auto h-12 w-auto"
            src="/book-icon.svg"
            alt="Book Tracker"
          />
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Email Verification Required
          </h2>
        </div>

        <div className="bg-white py-8 px-6 shadow rounded-lg">
          <div className="text-center">
            <ExclamationTriangleIcon className="h-12 w-12 text-yellow-500 mx-auto" />
            <h3 className="mt-4 text-lg font-medium text-gray-900">Please Verify Your Email</h3>
            <p className="mt-2 text-sm text-gray-600">
              Hi {user?.firstName}! You need to verify your email address before accessing the application.
            </p>
            <p className="mt-1 text-sm font-medium text-gray-900">
              {user?.email}
            </p>

            <div className="mt-4 p-4 bg-blue-50 rounded-md">
              <div className="flex">
                <div className="flex-shrink-0">
                  <EnvelopeIcon className="h-5 w-5 text-blue-400" />
                </div>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-blue-800">
                    Check your email
                  </h3>
                  <div className="mt-2 text-sm text-blue-700">
                    <p>
                      We've sent a verification link to your email address. 
                      Click the link to verify your account and gain access.
                    </p>
                  </div>
                </div>
              </div>
            </div>

            {resendMessage && (
              <div className={`mt-4 p-3 rounded-md ${
                messageType === 'success' 
                  ? 'bg-green-50 border border-green-200' 
                  : 'bg-red-50 border border-red-200'
              }`}>
                <p className={`text-sm ${
                  messageType === 'success' ? 'text-green-700' : 'text-red-700'
                }`}>
                  {resendMessage}
                </p>
              </div>
            )}

            <div className="mt-6 flex flex-col space-y-3">
              <button
                onClick={handleResendVerification}
                disabled={resendLoading}
                className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {resendLoading ? 'Sending...' : 'Resend Verification Email'}
              </button>
              
              <button
                onClick={handleLogout}
                className="w-full flex justify-center py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                Sign Out
              </button>
            </div>

            <div className="mt-4 text-xs text-gray-500">
              <p>Didn't receive the email? Check your spam folder or click "Resend"</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}