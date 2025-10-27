import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import api from '../services/api'

export default function StudentInvitation() {
  const { classId, token } = useParams()
  const navigate = useNavigate()
  const { user, loading: authLoading } = useAuth()
  
  const [loading, setLoading] = useState(true)
  const [invitationDetails, setInvitationDetails] = useState(null)
  const [error, setError] = useState('')
  const [redeeming, setRedeeming] = useState(false)
  const [success, setSuccess] = useState(false)

  useEffect(() => {
    if (!classId || !token) {
      setError('Invalid invitation link - missing class ID or token')
      setLoading(false)
      return
    }
    
    fetchInvitationDetails()
  }, [classId, token])

  const fetchInvitationDetails = async () => {
    try {
      const response = await api.get(`/api/invite/${classId}/${token}`)
      setInvitationDetails(response.data)
    } catch (error) {
      setError(error.response?.data?.message || 'Invalid or expired invitation link')
    } finally {
      setLoading(false)
    }
  }

  const handleRedeem = async () => {
    if (!user) {
      // Store invitation and redirect to login
      const inviteUrl = `/invite/${classId}/${token}`
      navigate(`/login?redirect=${encodeURIComponent(inviteUrl)}`)
      return
    }

    setRedeeming(true)
    setError('')

    try {
      const response = await api.post(`/api/invite/${classId}/${token}/redeem`)
      setSuccess(true)
      setInvitationDetails(prev => ({
        ...prev,
        child: response.data.child
      }))
    } catch (error) {
      setError(error.response?.data?.message || 'Failed to redeem invitation')
    } finally {
      setRedeeming(false)
    }
  }

  if (authLoading || loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
        <div className="sm:mx-auto sm:w-full sm:max-w-md">
          <div className="flex justify-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
          </div>
          <p className="mt-4 text-center text-gray-600">Loading invitation details...</p>
        </div>
      </div>
    )
  }

  if (error && !invitationDetails) {
    return (
      <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
        <div className="sm:mx-auto sm:w-full sm:max-w-md">
          <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
            <div className="text-center">
              <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100">
                <svg className="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </div>
              <h2 className="mt-4 text-lg font-medium text-gray-900">Invalid Invitation</h2>
              <p className="mt-2 text-sm text-gray-600">{error}</p>
              <div className="mt-6">
                <Link 
                  to="/login"
                  className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  Go to Login
                </Link>
              </div>
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (success) {
    return (
      <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
        <div className="sm:mx-auto sm:w-full sm:max-w-md">
          <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
            <div className="text-center">
              <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100">
                <svg className="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7" />
                </svg>
              </div>
              <h2 className="mt-4 text-lg font-medium text-gray-900">Invitation Accepted!</h2>
              <p className="mt-2 text-sm text-gray-600">
                <span className="font-medium">{invitationDetails?.student_name}</span> has been added to your account
                and enrolled in <span className="font-medium">{invitationDetails?.class_name}</span>.
              </p>
              <div className="mt-6">
                <Link 
                  to="/dashboard"
                  className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  Go to Dashboard
                </Link>
              </div>
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
          Student Invitation
        </h2>
        {invitationDetails && (
          <div className="mt-4 text-center">
            <p className="text-sm text-gray-600">
              You've been invited to track reading progress for{' '}
              <span className="font-medium">{invitationDetails.student_name}</span> in{' '}
              <span className="font-medium">{invitationDetails.class_name}</span>
            </p>
            <p className="text-xs text-gray-500 mt-1">
              Teacher: {invitationDetails.teacher_name}
            </p>
          </div>
        )}
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
          {error && (
            <div className="mb-4 rounded-md bg-red-50 p-4">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {!user ? (
            <div className="space-y-4">
              <p className="text-sm text-gray-600">
                Please log in or create an account to accept this invitation.
              </p>
              <div className="grid grid-cols-2 gap-3">
                <Link
                  to={`/login?redirect=${encodeURIComponent(`/invite/${classId}/${token}`)}`}
                  className="w-full flex justify-center py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  Log In
                </Link>
                <Link
                  to={`/register?redirect=${encodeURIComponent(`/invite/${classId}/${token}`)}`}
                  className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  Sign Up
                </Link>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              <p className="text-sm text-gray-600">
                Logged in as <span className="font-medium">{user.firstName} {user.lastName}</span> ({user.email})
              </p>
              <button
                onClick={handleRedeem}
                disabled={redeeming}
                className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {redeeming ? 'Accepting Invitation...' : 'Accept Invitation'}
              </button>
              <p className="text-xs text-gray-500 text-center">
                This will add {invitationDetails?.student_name} to your account
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}