import { useState, useEffect } from 'react'
import { BookOpenIcon, PlusIcon, EyeIcon, PencilIcon, DocumentArrowDownIcon } from '@heroicons/react/24/outline'
import api from '../services/api'

export default function ChildCard({ child, onAddBook, onViewDetails, onEditChild, currentMonth }) {
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

  const handleDownloadPDF = async () => {
    try {
      const month = currentMonth.getMonth() + 1
      const year = currentMonth.getFullYear()
      
      const response = await api.get(`/reports/child/${child.id}/monthly-pdf`, {
        params: { month, year },
        responseType: 'blob'
      })
      
      const blob = new Blob([response.data], { type: 'application/pdf' })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `${child.firstName}_${child.lastName}_books_${currentMonth.toLocaleDateString('en-US', { month: 'long', year: 'numeric' }).replace(' ', '_')}.pdf`
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Failed to download PDF:', error)
      alert('Failed to download PDF report')
    }
  }


  return (
    <div className="bg-white overflow-hidden shadow rounded-lg">
      <div className="p-3 sm:p-5">
        <div className="flex items-center justify-between">
          <div className="flex items-center flex-1 min-w-0">
            <div className="flex-shrink-0">
              <BookOpenIcon className="h-6 w-6 sm:h-8 sm:w-8 text-indigo-600" />
            </div>
            <div className="ml-3 sm:ml-5 flex-1 min-w-0">
              <dl>
                <dt className="text-sm font-medium text-gray-500 truncate">
                  {child.firstName} {child.lastName}
                </dt>
                <dd className="text-base sm:text-lg font-medium text-gray-900">
                  {child.grade}
                </dd>
              </dl>
            </div>
          </div>
          
          {/* Action buttons on the right */}
          <div className="flex space-x-1 flex-shrink-0 ml-3">
            <button
              onClick={handleDownloadPDF}
              className="p-1.5 rounded-full bg-white shadow-md hover:bg-blue-50 hover:text-blue-600 transition-colors border border-gray-200"
              title="Download monthly PDF report"
            >
              <DocumentArrowDownIcon className="h-3 w-3 sm:h-4 sm:w-4 text-gray-600" />
            </button>
            <button
              onClick={() => onEditChild(child)}
              className="p-1.5 rounded-full bg-white shadow-md hover:bg-indigo-50 hover:text-indigo-600 transition-colors border border-gray-200"
              title="Edit child information"
            >
              <PencilIcon className="h-3 w-3 sm:h-4 sm:w-4 text-gray-600" />
            </button>
          </div>
        </div>
      </div>
      <div className="bg-gray-50 px-3 sm:px-5 py-3">
        <div className="text-sm">
          <div className="font-medium text-gray-900 mb-2">
            {loading ? 'Loading...' : `${currentMonthBooks?.length || 0} books this month`}
          </div>
          {!loading && (
            <div className="text-gray-600 text-xs sm:text-sm">
              {currentMonth ? 
                currentMonth.toLocaleDateString('en-US', { month: 'long', year: 'numeric' }) :
                'Current Month'
              }
            </div>
          )}
        </div>
        <div className="mt-3 flex flex-wrap gap-2">
          <button
            onClick={() => onAddBook(child)}
            className="inline-flex items-center px-2 sm:px-3 py-1 border border-transparent text-xs font-medium rounded-full text-indigo-700 bg-indigo-100 hover:bg-indigo-200"
          >
            <PlusIcon className="h-3 w-3 mr-1" />
            Add Book
          </button>
          <button
            onClick={() => onViewDetails(child)}
            className="inline-flex items-center px-2 sm:px-3 py-1 border border-transparent text-xs font-medium rounded-full text-gray-700 bg-gray-100 hover:bg-gray-200"
          >
            <EyeIcon className="h-3 w-3 mr-1" />
            View All
          </button>
        </div>
      </div>
    </div>
  )
}