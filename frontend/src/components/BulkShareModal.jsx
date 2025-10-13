import { useState } from 'react'
import { XMarkIcon } from '@heroicons/react/24/outline'
import api from '../services/api'

export default function BulkShareModal({ children, onClose }) {
  const [formData, setFormData] = useState({
    email: '',
    shareAll: false,
    allPermissionType: 'VIEW',
    individualPermissions: {}
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }))
  }

  const handleIndividualPermissionChange = (childId, permissionType) => {
    setFormData(prev => ({
      ...prev,
      individualPermissions: {
        ...prev.individualPermissions,
        [childId]: permissionType
      }
    }))
  }

  const toggleChildSelection = (childId) => {
    setFormData(prev => {
      const newPermissions = { ...prev.individualPermissions }
      if (newPermissions[childId]) {
        delete newPermissions[childId]
      } else {
        newPermissions[childId] = 'VIEW'
      }
      return {
        ...prev,
        individualPermissions: newPermissions
      }
    })
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    setSuccess('')

    try {
      if (formData.shareAll) {
        // Share all children with the same permission
        const promises = children.map(child =>
          api.post(`/children/${child.id}/invite`, {
            email: formData.email,
            permissionType: formData.allPermissionType
          })
        )
        await Promise.all(promises)
      } else {
        // Share selected children with individual permissions
        const selectedChildren = Object.keys(formData.individualPermissions)
        if (selectedChildren.length === 0) {
          setError('Please select at least one child to share or use "Share All" option.')
          setLoading(false)
          return
        }

        const promises = selectedChildren.map(childId =>
          api.post(`/children/${childId}/invite`, {
            email: formData.email,
            permissionType: formData.individualPermissions[childId]
          })
        )
        await Promise.all(promises)
      }

      setSuccess('All invitations sent successfully!')
      setTimeout(() => {
        onClose()
      }, 2000)
    } catch (error) {
      setError(error.response?.data?.message || 'Failed to send invitations')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-10 mx-auto p-5 border w-full max-w-2xl shadow-lg rounded-md bg-white">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-medium">Share Children</h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <XMarkIcon className="h-6 w-6" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-gray-700">
              Email Address
            </label>
            <input
              type="email"
              name="email"
              required
              className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              value={formData.email}
              onChange={handleChange}
              placeholder="Enter email address..."
            />
          </div>

          <div className="border-t border-gray-200 pt-4">
            <div className="flex items-center mb-4">
              <input
                type="checkbox"
                name="shareAll"
                id="shareAll"
                className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                checked={formData.shareAll}
                onChange={handleChange}
              />
              <label htmlFor="shareAll" className="ml-2 block text-sm font-medium text-gray-700">
                Share all children with the same permission level
              </label>
            </div>

            {formData.shareAll && (
              <div className="ml-6 mb-4">
                <label className="block text-sm font-medium text-gray-700">
                  Permission Level for All
                </label>
                <select
                  name="allPermissionType"
                  className="mt-1 block w-48 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                  value={formData.allPermissionType}
                  onChange={handleChange}
                >
                  <option value="VIEW">View Only</option>
                  <option value="EDIT">View & Edit</option>
                </select>
              </div>
            )}

            {!formData.shareAll && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-3">
                  Select children and set individual permission levels:
                </label>
                <div className="space-y-3 max-h-60 overflow-y-auto">
                  {children.map(child => (
                    <div key={child.id} className="flex items-center justify-between p-3 border border-gray-200 rounded-md">
                      <div className="flex items-center">
                        <input
                          type="checkbox"
                          id={`child-${child.id}`}
                          className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                          checked={!!formData.individualPermissions[child.id]}
                          onChange={() => toggleChildSelection(child.id)}
                        />
                        <label htmlFor={`child-${child.id}`} className="ml-3 text-sm font-medium text-gray-900">
                          {child.name} ({child.grade})
                        </label>
                      </div>
                      {formData.individualPermissions[child.id] && (
                        <select
                          className="ml-4 px-3 py-1 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                          value={formData.individualPermissions[child.id]}
                          onChange={(e) => handleIndividualPermissionChange(child.id, e.target.value)}
                        >
                          <option value="VIEW">View Only</option>
                          <option value="EDIT">View & Edit</option>
                        </select>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}
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
              {loading ? 'Sending Invitations...' : 'Send Invitations'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}