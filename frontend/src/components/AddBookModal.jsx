import { useState, useEffect } from 'react'
import { XMarkIcon } from '@heroicons/react/24/outline'
import api from '../services/api'

export default function AddBookModal({ child, onClose, onBookAdded }) {
  const [formData, setFormData] = useState({
    title: '',
    author: '',
    dateRead: new Date().toISOString().split('T')[0]
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [existingBooks, setExistingBooks] = useState([])
  const [isDuplicate, setIsDuplicate] = useState(false)

  useEffect(() => {
    fetchExistingBooks()
  }, [child.id])

  useEffect(() => {
    checkForDuplicate()
  }, [formData.title, formData.author, existingBooks])

  const fetchExistingBooks = async () => {
    try {
      const response = await api.get(`/books/child/${child.id}`)
      setExistingBooks(response.data || [])
    } catch (error) {
      console.error('Failed to fetch existing books:', error)
      setExistingBooks([])
    }
  }

  const checkForDuplicate = () => {
    if (!formData.title.trim() || !formData.author.trim()) {
      setIsDuplicate(false)
      return
    }

    const normalizeString = (str) => str.toLowerCase().trim()
    const normalizedTitle = normalizeString(formData.title)
    const normalizedAuthor = normalizeString(formData.author)

    const duplicate = existingBooks.some(book => 
      normalizeString(book.title) === normalizedTitle && 
      normalizeString(book.author) === normalizedAuthor
    )

    setIsDuplicate(duplicate)
    if (duplicate) {
      setError(`${child.name} has already read "${formData.title}" by ${formData.author}`)
    } else {
      setError('')
    }
  }

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    })
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    
    if (isDuplicate) {
      return
    }

    setLoading(true)

    try {
      await api.post(`/books/child/${child.id}`, {
        title: formData.title,
        author: formData.author,
        dateRead: formData.dateRead,
        childId: child.id
      })
      onBookAdded()
    } catch (error) {
      setError(error.response?.data?.message || 'Failed to add book')
    } finally {
      setLoading(false)
    }
  }

  const handleFinishedToday = async () => {
    if (!formData.title || !formData.author || isDuplicate) return
    
    setLoading(true)

    try {
      await api.post(`/books/child/${child.id}`, {
        title: formData.title,
        author: formData.author,
        dateRead: new Date().toISOString().split('T')[0], // Use current date as string
        childId: child.id
      })
      onBookAdded()
    } catch (error) {
      setError(error.response?.data?.message || 'Failed to add book')
    } finally {
      setLoading(false)
    }
  }

  const isFormValid = formData.title.trim() && formData.author.trim() && !isDuplicate

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-medium">Add Book for {child.name}</h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <XMarkIcon className="h-6 w-6" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700">
              Book Title
            </label>
            <input
              type="text"
              name="title"
              required
              className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              value={formData.title}
              onChange={handleChange}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">
              Author
            </label>
            <input
              type="text"
              name="author"
              required
              className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              value={formData.author}
              onChange={handleChange}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">
              Date Read
            </label>
            <input
              type="date"
              name="dateRead"
              required
              className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              value={formData.dateRead}
              onChange={handleChange}
            />
          </div>

          {error && (
            <div className="text-red-600 text-sm">{error}</div>
          )}

          {isDuplicate && (
            <div className="bg-red-50 border border-red-200 rounded-md p-3">
              <div className="text-red-800 text-sm font-medium">
                Duplicate Book Detected
              </div>
              <div className="text-red-700 text-sm mt-1">
                {child.name} has already read this book. Each child can only read a book once.
              </div>
            </div>
          )}

          <div className="flex justify-between">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              Cancel
            </button>
            <div className="space-x-3">
              <button
                type="button"
                onClick={handleFinishedToday}
                disabled={loading || !isFormValid}
                className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-green-600 hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
                title={isDuplicate ? "Cannot add duplicate book" : !isFormValid ? "Please fill in title and author first" : "Mark as finished today"}
              >
                {loading ? 'Adding...' : 'Finished Today'}
              </button>
              <button
                type="submit"
                disabled={loading}
                className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50"
              >
                {loading ? 'Adding...' : 'Add Book'}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  )
}