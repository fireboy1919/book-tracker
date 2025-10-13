import { useState, useEffect } from 'react'
import { XMarkIcon } from '@heroicons/react/24/outline'
import api from '../services/api'

export default function EditBookModal({ book, onClose, onBookUpdated }) {
  const [formData, setFormData] = useState({
    title: '',
    author: '',
    isbn: '',
    lexileLevel: '',
    dateRead: '',
    isPartial: false,
    partialComment: ''
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (book) {
      setFormData({
        title: book.title || '',
        author: book.author || '',
        isbn: book.isbn || '',
        lexileLevel: book.lexileLevel || '',
        dateRead: book.dateRead || '',
        isPartial: book.isPartial || false,
        partialComment: book.partialComment || ''
      })
    }
  }, [book])

  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }))
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const response = await api.put(`/books/${book.id}`, formData)
      onBookUpdated(response.data)
      onClose()
    } catch (error) {
      console.error('Failed to update book:', error)
      setError(error.response?.data?.message || 'Failed to update book. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  const isFormValid = formData.title.trim() && formData.author.trim() && formData.dateRead

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-[60]">
      <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-bold text-gray-900">Edit Book</h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <XMarkIcon className="h-6 w-6" />
          </button>
        </div>

        {error && (
          <div className="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative">
            <span className="block sm:inline">{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-1">
              Title *
            </label>
            <input
              type="text"
              id="title"
              name="title"
              value={formData.title}
              onChange={handleInputChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="Enter book title"
              required
              readOnly={book.isCustomBook === false} // Read-only for shared books
            />
          </div>

          <div className="mb-4">
            <label htmlFor="author" className="block text-sm font-medium text-gray-700 mb-1">
              Author *
            </label>
            <input
              type="text"
              id="author"
              name="author"
              value={formData.author}
              onChange={handleInputChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="Enter author name"
              required
              readOnly={book.isCustomBook === false} // Read-only for shared books
            />
          </div>

          {book.isCustomBook && (
            <div className="mb-4">
              <label htmlFor="isbn" className="block text-sm font-medium text-gray-700 mb-1">
                ISBN (optional)
              </label>
              <input
                type="text"
                id="isbn"
                name="isbn"
                value={formData.isbn}
                onChange={handleInputChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                placeholder="Enter ISBN (optional)"
              />
            </div>
          )}

          <div className="mb-4">
            <label htmlFor="lexileLevel" className="block text-sm font-medium text-gray-700 mb-1">
              Lexile Level (optional)
            </label>
            <input
              type="text"
              id="lexileLevel"
              name="lexileLevel"
              value={formData.lexileLevel}
              onChange={handleInputChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="e.g., 450L"
            />
          </div>

          <div className="mb-4">
            <label htmlFor="dateRead" className="block text-sm font-medium text-gray-700 mb-1">
              Date Read *
            </label>
            <input
              type="date"
              id="dateRead"
              name="dateRead"
              value={formData.dateRead}
              onChange={handleInputChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              required
            />
          </div>

          <div className="mb-4">
            <label className="flex items-center">
              <input
                type="checkbox"
                name="isPartial"
                checked={formData.isPartial}
                onChange={handleInputChange}
                className="rounded border-gray-300 text-indigo-600 shadow-sm focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50"
              />
              <span className="ml-2 text-sm text-gray-700">This is a partial reading</span>
            </label>
          </div>

          {formData.isPartial && (
            <div className="mb-4">
              <label htmlFor="partialComment" className="block text-sm font-medium text-gray-700 mb-1">
                What portion was read?
              </label>
              <textarea
                id="partialComment"
                name="partialComment"
                value={formData.partialComment}
                onChange={handleInputChange}
                rows={3}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                placeholder="e.g., Chapters 1-5, First half, etc."
              />
            </div>
          )}

          <div className="flex gap-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading || !isFormValid}
              className="flex-1 px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Updating...' : 'Update Book'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}