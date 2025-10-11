import { useState, useEffect } from 'react'
import { BookOpenIcon, PlusIcon, UserPlusIcon } from '@heroicons/react/24/outline'
import api from '../services/api'

export default function ChildCard({ child, onAddBook, onInviteUser }) {
  const [books, setBooks] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchBooks()
  }, [child.id])

  const fetchBooks = async () => {
    try {
      const response = await api.get(`/books/child/${child.id}`)
      setBooks(response.data || [])
    } catch (error) {
      console.error('Failed to fetch books:', error)
      setBooks([])
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="bg-white overflow-hidden shadow rounded-lg">
      <div className="p-5">
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <BookOpenIcon className="h-8 w-8 text-indigo-600" />
          </div>
          <div className="ml-5 w-0 flex-1">
            <dl>
              <dt className="text-sm font-medium text-gray-500 truncate">
                {child.name}
              </dt>
              <dd className="text-lg font-medium text-gray-900">
                {child.grade}
              </dd>
            </dl>
          </div>
        </div>
      </div>
      <div className="bg-gray-50 px-5 py-3">
        <div className="text-sm">
          <div className="font-medium text-gray-900 mb-2">
            {loading ? 'Loading...' : `${books?.length || 0} books read`}
          </div>
          {!loading && books?.length > 0 && (
            <div className="text-gray-600">
              Recent: {books[books.length - 1]?.title}
            </div>
          )}
        </div>
        <div className="mt-3 flex space-x-2">
          <button
            onClick={() => onAddBook(child)}
            className="inline-flex items-center px-3 py-1 border border-transparent text-xs font-medium rounded-full text-indigo-700 bg-indigo-100 hover:bg-indigo-200"
          >
            <PlusIcon className="h-3 w-3 mr-1" />
            Add Book
          </button>
          <button
            onClick={() => onInviteUser(child)}
            className="inline-flex items-center px-3 py-1 border border-transparent text-xs font-medium rounded-full text-gray-700 bg-gray-100 hover:bg-gray-200"
          >
            <UserPlusIcon className="h-3 w-3 mr-1" />
            Share
          </button>
        </div>
      </div>
    </div>
  )
}