import { useState } from 'react'
import { XMarkIcon } from '@heroicons/react/24/outline'
import api from '../services/api'

export default function CreateClassModal({ isOpen, onClose, onSuccess }) {
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    studentBooksGoal: 0,
    otherBooksGoal: 0
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      await api.post('/classes', formData)
      onSuccess()
      onClose()
      setFormData({
        name: '',
        description: '',
        studentBooksGoal: 0,
        otherBooksGoal: 0
      })
    } catch (error) {
      setError(error.response?.data?.message || 'Failed to create class')
    } finally {
      setLoading(false)
    }
  }

  const handleChange = (e) => {
    const { name, value, type } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: type === 'number' ? parseInt(value) || 0 : value
    }))
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-medium text-gray-900">Create New Class</h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <XMarkIcon className="h-6 w-6" />
          </button>
        </div>

        {error && (
          <div className="mb-4 bg-red-50 border border-red-400 text-red-700 px-4 py-3 rounded">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-700">
              Class Name *
            </label>
            <input
              type="text"
              id="name"
              name="name"
              required
              value={formData.name}
              onChange={handleChange}
              className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="e.g., Mrs. Smith's 3rd Grade"
            />
          </div>

          <div>
            <label htmlFor="description" className="block text-sm font-medium text-gray-700">
              Description
            </label>
            <textarea
              id="description"
              name="description"
              value={formData.description}
              onChange={handleChange}
              rows={3}
              className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="Optional description"
            />
          </div>

          <div>
            <label htmlFor="studentBooksGoal" className="block text-sm font-medium text-gray-700">
              Student Books Goal
            </label>
            <input
              type="number"
              id="studentBooksGoal"
              name="studentBooksGoal"
              min="0"
              value={formData.studentBooksGoal}
              onChange={handleChange}
              className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="Number of books students should read"
            />
          </div>

          <div>
            <label htmlFor="otherBooksGoal" className="block text-sm font-medium text-gray-700">
              Books Read to Student Goal
            </label>
            <input
              type="number"
              id="otherBooksGoal"
              name="otherBooksGoal"
              min="0"
              value={formData.otherBooksGoal}
              onChange={handleChange}
              className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="Number of books others should read to/with student"
            />
          </div>

          <div className="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-md disabled:opacity-50"
            >
              {loading ? 'Creating...' : 'Create Class'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}