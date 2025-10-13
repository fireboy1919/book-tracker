import { useState, useEffect } from 'react'
import { BookOpenIcon, PlusIcon, EyeIcon } from '@heroicons/react/24/outline'
import api from '../services/api'

export default function ChildCard({ child, onAddBook, onViewDetails, currentMonth }) {
  const [books, setBooks] = useState([])
  const [loading, setLoading] = useState(true)
  const [currentMonthBooks, setCurrentMonthBooks] = useState([])

  useEffect(() => {
    fetchBooks()
  }, [child.id])

  useEffect(() => {
    filterBooksByMonth()
  }, [books, currentMonth])

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

  const filterBooksByMonth = () => {
    if (!currentMonth || !books.length) {
      setCurrentMonthBooks([])
      return
    }

    const currentYear = currentMonth.getFullYear()
    const currentMonthIndex = currentMonth.getMonth()

    const filtered = books.filter(book => {
      const bookDate = new Date(book.dateRead)
      return bookDate.getFullYear() === currentYear && bookDate.getMonth() === currentMonthIndex
    })

    setCurrentMonthBooks(filtered)
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
                {child.firstName} {child.lastName}
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
            {loading ? 'Loading...' : `${currentMonthBooks?.length || 0} books this month`}
          </div>
          {!loading && (
            <div className="text-gray-600">
              {currentMonth ? 
                currentMonth.toLocaleDateString('en-US', { month: 'long', year: 'numeric' }) :
                'Current Month'
              }
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
            onClick={() => onViewDetails(child)}
            className="inline-flex items-center px-3 py-1 border border-transparent text-xs font-medium rounded-full text-gray-700 bg-gray-100 hover:bg-gray-200"
          >
            <EyeIcon className="h-3 w-3 mr-1" />
            View All
          </button>
        </div>
      </div>
    </div>
  )
}