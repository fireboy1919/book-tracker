import { useState, useEffect } from 'react'
import { XMarkIcon, ChevronLeftIcon, ChevronRightIcon, PlusIcon } from '@heroicons/react/24/outline'
import api from '../services/api'

export default function FullScreenChildView({ child, onClose, onAddBook }) {
  const [books, setBooks] = useState([])
  const [loading, setLoading] = useState(true)
  const [currentDate, setCurrentDate] = useState(new Date())
  const [filteredBooks, setFilteredBooks] = useState([])

  useEffect(() => {
    fetchBooks()
  }, [child.id])

  useEffect(() => {
    filterBooksByMonth()
  }, [books, currentDate])

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
    const currentYear = currentDate.getFullYear()
    const currentMonth = currentDate.getMonth()

    const filtered = books.filter(book => {
      const bookDate = new Date(book.dateRead)
      return bookDate.getFullYear() === currentYear && bookDate.getMonth() === currentMonth
    })

    setFilteredBooks(filtered)
  }

  const navigateMonth = (direction) => {
    const newDate = new Date(currentDate)
    newDate.setMonth(currentDate.getMonth() + direction)
    setCurrentDate(newDate)
  }

  const formatMonthYear = (date) => {
    return date.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
  }

  const formatDateRead = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', { 
      month: 'short', 
      day: 'numeric',
      year: 'numeric'
    })
  }

  const handleAddBook = () => {
    onAddBook(child)
  }

  if (loading) {
    return (
      <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
        <div className="relative top-20 mx-auto p-5 border w-full max-w-6xl shadow-lg rounded-md bg-white">
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-indigo-600"></div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-10 mx-auto p-5 border w-full max-w-6xl shadow-lg rounded-md bg-white min-h-[80vh]">
        {/* Header */}
        <div className="flex justify-between items-center mb-6">
          <div>
            <h3 className="text-2xl font-bold text-gray-900">{child.name}'s Books</h3>
            <p className="text-sm text-gray-500">{child.grade}</p>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <XMarkIcon className="h-6 w-6" />
          </button>
        </div>

        {/* Month Navigation */}
        <div className="flex items-center justify-between mb-6 bg-gray-50 p-4 rounded-lg">
          <button
            onClick={() => navigateMonth(-1)}
            className="p-2 rounded-full hover:bg-gray-200 transition-colors"
          >
            <ChevronLeftIcon className="h-5 w-5" />
          </button>
          
          <div className="text-center">
            <h4 className="text-lg font-semibold text-gray-900">
              {formatMonthYear(currentDate)}
            </h4>
            <p className="text-sm text-gray-600">
              {filteredBooks.length} book{filteredBooks.length !== 1 ? 's' : ''} read
            </p>
          </div>
          
          <button
            onClick={() => navigateMonth(1)}
            className="p-2 rounded-full hover:bg-gray-200 transition-colors"
          >
            <ChevronRightIcon className="h-5 w-5" />
          </button>
        </div>

        {/* Add Book Button */}
        <div className="mb-6">
          <button
            onClick={handleAddBook}
            className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700"
          >
            <PlusIcon className="h-4 w-4 mr-2" />
            Add Book for {formatMonthYear(currentDate)}
          </button>
        </div>

        {/* Books List */}
        <div className="space-y-4">
          {filteredBooks.length === 0 ? (
            <div className="text-center py-12">
              <div className="text-gray-400 text-lg mb-4">
                No books recorded for {formatMonthYear(currentDate)}
              </div>
              <button
                onClick={handleAddBook}
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700"
              >
                <PlusIcon className="h-4 w-4 mr-2" />
                Add First Book
              </button>
            </div>
          ) : (
            <div className="grid gap-4">
              {filteredBooks.map((book, index) => (
                <div key={book.id} className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm hover:shadow-md transition-shadow">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center mb-2">
                        <span className="bg-indigo-100 text-indigo-800 text-xs font-medium px-2 py-1 rounded-full mr-3">
                          Book #{index + 1}
                        </span>
                        <span className="text-sm text-gray-500">
                          Read on {formatDateRead(book.dateRead)}
                        </span>
                      </div>
                      <h5 className="text-lg font-semibold text-gray-900 mb-1">
                        {book.title}
                      </h5>
                      <p className="text-gray-600">
                        by {book.author}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Summary Footer */}
        {filteredBooks.length > 0 && (
          <div className="mt-8 pt-6 border-t border-gray-200">
            <div className="text-center">
              <p className="text-lg font-medium text-gray-900">
                Total books read in {formatMonthYear(currentDate)}: {filteredBooks.length}
              </p>
              <p className="text-sm text-gray-500 mt-1">
                All time total: {books.length} books
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}