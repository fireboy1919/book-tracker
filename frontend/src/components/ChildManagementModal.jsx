import { useState, useEffect } from 'react'
import { XMarkIcon, TrashIcon, EyeIcon, PencilIcon } from '@heroicons/react/24/outline'
import api from '../services/api'

export default function ChildManagementModal({ child, onClose, onChildUpdated }) {
  const [activeTab, setActiveTab] = useState('details')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  
  // Child details form
  const [childData, setChildData] = useState({
    firstName: child.firstName,
    lastName: child.lastName,
    grade: child.grade
  })
  
  // Permissions data
  const [permissions, setPermissions] = useState([])
  const [permissionsLoading, setPermissionsLoading] = useState(false)
  
  // New invitation form
  const [inviteData, setInviteData] = useState({
    email: '',
    permissionType: 'VIEW'
  })

  useEffect(() => {
    if (activeTab === 'sharing') {
      fetchPermissions()
    }
  }, [activeTab])

  const fetchPermissions = async () => {
    setPermissionsLoading(true)
    try {
      const response = await api.get(`/children/${child.id}/permissions`)
      setPermissions(response.data || [])
    } catch (error) {
      console.error('Failed to fetch permissions:', error)
      setPermissions([])
    } finally {
      setPermissionsLoading(false)
    }
  }

  const handleChildUpdate = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    setSuccess('')

    try {
      await api.put(`/children/${child.id}`, childData)
      setSuccess('Child details updated successfully!')
      onChildUpdated()
      setTimeout(() => {
        onClose()
      }, 2000)
    } catch (error) {
      setError(error.response?.data?.message || 'Failed to update child')
    } finally {
      setLoading(false)
    }
  }

  const handleInvite = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    setSuccess('')

    try {
      await api.post(`/children/${child.id}/invite`, inviteData)
      setSuccess('Invitation sent successfully!')
      setInviteData({ email: '', permissionType: 'VIEW' })
      fetchPermissions()
    } catch (error) {
      setError(error.response?.data?.message || 'Failed to send invitation')
    } finally {
      setLoading(false)
    }
  }

  const handleRevokePermission = async (permissionId) => {
    if (!confirm('Are you sure you want to revoke this permission?')) {
      return
    }

    try {
      await api.delete(`/permissions/${permissionId}`)
      setSuccess('Permission revoked successfully!')
      fetchPermissions()
    } catch (error) {
      setError(error.response?.data?.message || 'Failed to revoke permission')
    }
  }

  const getPermissionIcon = (type) => {
    return type === 'EDIT' ? 
      <PencilIcon className="h-4 w-4 text-green-600" /> : 
      <EyeIcon className="h-4 w-4 text-blue-600" />
  }

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-10 mx-auto p-5 border w-full max-w-3xl shadow-lg rounded-md bg-white">
        <div className="flex justify-between items-center mb-6">
          <h3 className="text-xl font-medium">Manage {child.firstName} {child.lastName}</h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <XMarkIcon className="h-6 w-6" />
          </button>
        </div>

        {/* Tab Navigation */}
        <div className="border-b border-gray-200 mb-6">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setActiveTab('details')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'details'
                  ? 'border-indigo-500 text-indigo-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Child Details
            </button>
            <button
              onClick={() => setActiveTab('sharing')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'sharing'
                  ? 'border-indigo-500 text-indigo-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Sharing & Permissions
            </button>
          </nav>
        </div>

        {/* Child Details Tab */}
        {activeTab === 'details' && (
          <form onSubmit={handleChildUpdate} className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-700">
                First Name
              </label>
              <input
                type="text"
                required
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                value={childData.firstName}
                onChange={(e) => setChildData({...childData, firstName: e.target.value})}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">
                Last Name
              </label>
              <input
                type="text"
                required
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                value={childData.lastName}
                onChange={(e) => setChildData({...childData, lastName: e.target.value})}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">
                Grade
              </label>
              <input
                type="text"
                required
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                value={childData.grade}
                onChange={(e) => setChildData({...childData, grade: e.target.value})}
                placeholder="e.g., 3rd Grade, Kindergarten"
              />
            </div>

            {error && (
              <div className="text-red-600 text-sm">{error}</div>
            )}

            {success && (
              <div className="text-green-600 text-sm">{success}</div>
            )}

            <div className="flex justify-end space-x-3 pt-4 border-t border-gray-200">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={loading}
                className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50"
              >
                {loading ? 'Updating...' : 'Update Child'}
              </button>
            </div>
          </form>
        )}

        {/* Sharing Tab */}
        {activeTab === 'sharing' && (
          <div className="space-y-6">
            {/* Send New Invitation */}
            <div className="bg-gray-50 p-4 rounded-lg">
              <h4 className="text-lg font-medium mb-4">Send New Invitation</h4>
              <form onSubmit={handleInvite} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Email Address
                  </label>
                  <input
                    type="email"
                    required
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                    value={inviteData.email}
                    onChange={(e) => setInviteData({...inviteData, email: e.target.value})}
                    placeholder="Enter email address..."
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Permission Level
                  </label>
                  <select
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                    value={inviteData.permissionType}
                    onChange={(e) => setInviteData({...inviteData, permissionType: e.target.value})}
                  >
                    <option value="VIEW">View Only</option>
                    <option value="EDIT">View & Edit</option>
                  </select>
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50"
                >
                  {loading ? 'Sending...' : 'Send Invitation'}
                </button>
              </form>
            </div>

            {/* Current Permissions */}
            <div>
              <h4 className="text-lg font-medium mb-4">Current Permissions</h4>
              {permissionsLoading ? (
                <div className="text-center py-4">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mx-auto"></div>
                </div>
              ) : permissions.length === 0 ? (
                <p className="text-gray-500 text-center py-4">No permissions granted yet</p>
              ) : (
                <div className="space-y-3">
                  {permissions.map((permission) => (
                    <div key={permission.id} className="flex items-center justify-between p-3 border border-gray-200 rounded-md">
                      <div className="flex items-center space-x-3">
                        {getPermissionIcon(permission.permissionType)}
                        <div>
                          <div className="font-medium">{permission.user?.email}</div>
                          <div className="text-sm text-gray-500">
                            {permission.permissionType === 'EDIT' ? 'Can view and edit' : 'Can view only'}
                          </div>
                        </div>
                      </div>
                      <button
                        onClick={() => handleRevokePermission(permission.id)}
                        className="p-2 text-red-600 hover:bg-red-50 rounded-md"
                        title="Revoke permission"
                      >
                        <TrashIcon className="h-4 w-4" />
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {error && (
              <div className="text-red-600 text-sm">{error}</div>
            )}

            {success && (
              <div className="text-green-600 text-sm">{success}</div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}