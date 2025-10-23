import { useState, useEffect } from 'react'
import { PencilIcon, TrashIcon, UsersIcon, PlusIcon } from '@heroicons/react/24/outline'
import api from '../services/api'
import { useAuth } from '../contexts/AuthContext'

export default function AdminPanel() {
  const { user: currentUser } = useAuth()
  const [users, setUsers] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [classes, setClasses] = useState([])
  const [selectedUser, setSelectedUser] = useState(null)
  const [showClassModal, setShowClassModal] = useState(false)

  useEffect(() => {
    fetchUsers()
    fetchClasses()
  }, [])

  const fetchUsers = async () => {
    try {
      const response = await api.get('/users')
      setUsers(response.data)
    } catch (error) {
      setError('Failed to fetch users')
    } finally {
      setLoading(false)
    }
  }

  const fetchClasses = async () => {
    try {
      const response = await api.get('/classes/available')
      setClasses(response.data)
    } catch (error) {
      console.error('Failed to fetch classes:', error)
    }
  }

  const toggleAdmin = async (userId, currentIsAdmin) => {
    try {
      // Find the user to get their current data
      const user = users.find(u => u.id === userId)
      if (!user) {
        setError('User not found')
        return
      }

      // Prevent admins from disabling their own admin capability
      if (currentUser && currentUser.id === userId && currentIsAdmin) {
        setError('You cannot remove your own admin privileges')
        return
      }

      await api.put(`/users/${userId}`, {
        email: user.email,
        firstName: user.firstName,
        lastName: user.lastName,
        isAdmin: !currentIsAdmin,
        isTeacher: user.isTeacher
      })
      fetchUsers()
    } catch (error) {
      setError(error.response?.data?.message || 'Failed to update user')
    }
  }

  const makeTeacher = async (userId) => {
    try {
      await api.put(`/users/${userId}/make-teacher`)
      fetchUsers()
    } catch (error) {
      setError('Failed to make user a teacher')
    }
  }

  const removeTeacher = async (userId) => {
    try {
      await api.put(`/users/${userId}/remove-teacher`)
      fetchUsers()
    } catch (error) {
      setError('Failed to remove teacher role')
    }
  }

  const deleteUser = async (userId) => {
    if (!confirm('Are you sure you want to delete this user?')) return

    try {
      await api.delete(`/users/${userId}`)
      fetchUsers()
    } catch (error) {
      setError('Failed to delete user')
    }
  }

  const assignToClass = async (userId, classId) => {
    try {
      await api.post(`/classes/${classId}/members`, {
        userId: userId,
        role: users.find(u => u.id === userId)?.isTeacher ? 'TEACHER' : 'STUDENT'
      })
      setShowClassModal(false)
      setSelectedUser(null)
      setError('')
    } catch (error) {
      setError('Failed to assign user to class')
    }
  }

  const openClassAssignment = (user) => {
    setSelectedUser(user)
    setShowClassModal(true)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-indigo-600"></div>
      </div>
    )
  }

  return (
    <div className="py-6">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 md:px-8">
        <div className="md:flex md:items-center md:justify-between">
          <div className="flex-1 min-w-0">
            <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
              User Management
            </h2>
          </div>
        </div>

        {error && (
          <div className="mt-4 bg-red-50 border border-red-400 text-red-700 px-4 py-3 rounded">
            {error}
          </div>
        )}

        <div className="mt-8 flex flex-col">
          <div className="-my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
            <div className="py-2 align-middle inline-block min-w-full sm:px-6 lg:px-8">
              <div className="shadow overflow-hidden border-b border-gray-200 sm:rounded-lg">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        User
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Role
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Created
                      </th>
                      <th className="relative px-6 py-3">
                        <span className="sr-only">Actions</span>
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {users.map((user) => (
                      <tr key={user.id}>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="flex items-center">
                            <div>
                              <div className="text-sm font-medium text-gray-900">
                                {user.firstName} {user.lastName}
                              </div>
                              <div className="text-sm text-gray-500">
                                {user.email}
                              </div>
                            </div>
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="space-y-1">
                            <button
                              onClick={() => toggleAdmin(user.id, user.isAdmin)}
                              className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                                user.isAdmin
                                  ? 'bg-green-100 text-green-800'
                                  : 'bg-gray-100 text-gray-800'
                              }`}
                            >
                              {user.isAdmin ? 'Admin' : 'User'}
                            </button>
                            {user.isTeacher && (
                              <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800">
                                Teacher
                              </span>
                            )}
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          {new Date(user.createdAt).toLocaleDateString()}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                          <div className="flex justify-end space-x-2">
                            {!user.isTeacher ? (
                              <button
                                onClick={() => makeTeacher(user.id)}
                                className="text-blue-600 hover:text-blue-900 text-xs bg-blue-50 px-2 py-1 rounded"
                              >
                                Make Teacher
                              </button>
                            ) : (
                              <button
                                onClick={() => removeTeacher(user.id)}
                                className="text-orange-600 hover:text-orange-900 text-xs bg-orange-50 px-2 py-1 rounded"
                              >
                                Remove Teacher
                              </button>
                            )}
                            <button
                              onClick={() => openClassAssignment(user)}
                              className="text-indigo-600 hover:text-indigo-900 text-xs bg-indigo-50 px-2 py-1 rounded flex items-center"
                            >
                              <UsersIcon className="h-3 w-3 mr-1" />
                              Assign to Class
                            </button>
                            <button
                              onClick={() => deleteUser(user.id)}
                              className="text-red-600 hover:text-red-900"
                            >
                              <TrashIcon className="h-5 w-5" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Class Assignment Modal */}
      {showClassModal && selectedUser && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
          <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
            <div className="mt-3">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium text-gray-900">
                  Assign {selectedUser.firstName} {selectedUser.lastName} to Class
                </h3>
                <button
                  onClick={() => setShowClassModal(false)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <span className="sr-only">Close</span>
                  <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              
              <div className="mb-4">
                <p className="text-sm text-gray-500 mb-3">
                  Select a class to assign this {selectedUser.isTeacher ? 'teacher' : 'user'} to:
                </p>
                
                {classes.length === 0 ? (
                  <p className="text-sm text-gray-500">No classes available</p>
                ) : (
                  <div className="space-y-2 max-h-60 overflow-y-auto">
                    {classes.map((classItem) => (
                      <button
                        key={classItem.id}
                        onClick={() => assignToClass(selectedUser.id, classItem.id)}
                        className="w-full text-left p-3 border border-gray-200 rounded-lg hover:border-indigo-300 hover:bg-indigo-50 transition-colors"
                      >
                        <div className="font-medium text-gray-900">{classItem.name}</div>
                        {classItem.description && (
                          <div className="text-sm text-gray-500">{classItem.description}</div>
                        )}
                        <div className="text-xs text-gray-400 mt-1">
                          Student Goal: {classItem.studentBooksGoal} | Read-to Goal: {classItem.otherBooksGoal}
                        </div>
                      </button>
                    ))}
                  </div>
                )}
              </div>
              
              <div className="flex justify-end">
                <button
                  onClick={() => setShowClassModal(false)}
                  className="px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400 transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}