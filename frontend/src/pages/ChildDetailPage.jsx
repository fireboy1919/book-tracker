import { useState, useEffect } from 'react'
import { useParams, useSearchParams, useNavigate, useLocation } from 'react-router-dom'
import { ChevronLeftIcon, ChevronRightIcon, PlusIcon, PencilIcon, TrashIcon, DocumentArrowDownIcon, BookOpenIcon, ArrowLeftIcon } from '@heroicons/react/24/outline'
import api from '../services/api'
import EditBookModal from '../components/EditBookModal'
import EditChildModal from '../components/EditChildModal'
import AddBookModal from '../components/AddBookModal'
import { createPortal } from 'react-dom'

export default function ChildDetailPage() {
  const { childId } = useParams()
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()
  const location = useLocation()
  
  const [child, setChild] = useState(null)
  const [books, setBooks] = useState([])
  const [loading, setLoading] = useState(true)
  const [totalBookCount, setTotalBookCount] = useState(0)
  const [showEditModal, setShowEditModal] = useState(false)
  const [selectedBook, setSelectedBook] = useState(null)
  const [canEdit, setCanEdit] = useState(false)
  const [checkingPermissions, setCheckingPermissions] = useState(true)
  const [showEditChildModal, setShowEditChildModal] = useState(false)
  const [showAddBook, setShowAddBook] = useState(false)

  // Get current month from URL params or default to current date
  const getCurrentMonth = () => {
    const monthParam = searchParams.get('month')
    const yearParam = searchParams.get('year')
    
    if (monthParam && yearParam) {
      const month = parseInt(monthParam, 10) - 1 // JavaScript months are 0-based
      const year = parseInt(yearParam, 10)
      if (!isNaN(month) && !isNaN(year) && month >= 0 && month <= 11) {
        return new Date(year, month, 1)
      }
    }
    
    // Default to current month
    const now = new Date()
    return new Date(now.getFullYear(), now.getMonth(), 1)
  }

  const currentDate = getCurrentMonth()

  useEffect(() => {
    if (childId) {
      fetchChild()
      fetchBooksForCurrentMonth()
      fetchTotalBookCount()
    }
  }, [childId, searchParams])

  // Separate useEffect for permission checking that runs when child data is available
  useEffect(() => {
    if (child) {
      checkEditPermission()
    }
  }, [child])

  const fetchChild = async () => {
    try {
      const response = await api.get(`/children/${childId}`)
      setChild(response.data)
    } catch (error) {
      console.error('Failed to fetch child:', error)
      if (error.response?.status === 404) {
        navigate('/dashboard')
      }
    }
  }

  const fetchBooksForCurrentMonth = async () => {
    setLoading(true)
    try {
      const year = currentDate.getFullYear()
      const month = currentDate.getMonth() + 1 // JavaScript months are 0-based
      const response = await api.get(`/books/child/${childId}?year=${year}&month=${month}`)
      setBooks(response.data || [])
    } catch (error) {
      console.error('Failed to fetch books for month:', error)
      setBooks([])
    } finally {
      setLoading(false)
    }
  }

  const fetchTotalBookCount = async () => {
    try {
      const response = await api.get(`/books/child/${childId}`)
      setTotalBookCount(response.data ? response.data.length : 0)
    } catch (error) {
      console.error('Failed to fetch total book count:', error)
      setTotalBookCount(0)
    }
  }

  const checkEditPermission = async () => {
    try {
      // Check if user has EDIT permission by calling the permissions endpoint
      // This endpoint requires EDIT permission, so if it succeeds, user can edit
      await api.get(`/children/${childId}/permissions`)
      setCanEdit(true)
    } catch (error) {
      console.error('Failed to check edit permissions:', error)
      if (error.response?.status === 403) {
        setCanEdit(false)
      } else {
        setCanEdit(false)
      }
    } finally {
      setCheckingPermissions(false)
    }
  }

  const navigateMonth = (direction) => {
    const newDate = new Date(currentDate)
    newDate.setMonth(currentDate.getMonth() + direction)
    
    // Update URL parameters
    const newSearchParams = new URLSearchParams(searchParams)
    newSearchParams.set('year', newDate.getFullYear().toString())
    newSearchParams.set('month', (newDate.getMonth() + 1).toString()) // Convert back to 1-based
    setSearchParams(newSearchParams)
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

  const handleEditBook = (book) => {
    setSelectedBook(book)
    setShowEditModal(true)
  }

  const handleDeleteBook = async (book) => {
    if (window.confirm(`Are you sure you want to delete "${book.title}"?`)) {
      try {
        await api.delete(`/books/${book.id}`)
        // Refresh the books list
        fetchBooksForCurrentMonth()
        fetchTotalBookCount()
      } catch (error) {
        console.error('Failed to delete book:', error)
        alert('Failed to delete book. Please try again.')
      }
    }
  }

  const handleBookUpdated = () => {
    // Refresh the books list
    fetchBooksForCurrentMonth()
    fetchTotalBookCount()
  }

  const handleBookAdded = () => {
    setShowAddBook(false)
    // Refresh the books list
    fetchBooksForCurrentMonth()
    fetchTotalBookCount()
  }

  const handleChildUpdated = (updatedChild) => {
    setChild(updatedChild)
    setShowEditChildModal(false)
  }

  const handleChildDeleted = () => {
    // Navigate back to where the user came from, or default to dashboard
    navigate(-1)
  }

  const handleDownloadPDF = async () => {
    const currentYear = currentDate.getFullYear()
    const currentMonth = currentDate.getMonth() + 1
    
    try {
      const response = await api.get(`/reports/child/${childId}/monthly-pdf`, {
        params: { year: currentYear, month: currentMonth },
        responseType: 'blob'
      })
      
      // Create blob URL and download
      const blob = new Blob([response.data], { type: 'application/pdf' })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `${child?.firstName}_${child?.lastName}_books_${currentMonth}_${currentYear}.pdf`
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Failed to download PDF:', error)
      alert('Failed to download PDF report. Please try again.')
    }
  }

  if (loading && !child) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-indigo-600"></div>
      </div>
    )
  }

  if (!child) {
    return (
      <div className="text-center py-12">
        <div className="text-gray-400 text-lg">Child not found</div>
        <button
          onClick={() => navigate(-1)}
          className="mt-4 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700"
        >
          <ArrowLeftIcon className="mr-2 h-4 w-4" />
          Go Back
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div className="flex items-center space-x-4">
          <button
            onClick={() => navigate(-1)}
            className="p-2 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-full transition-colors"
            title="Go Back"
          >
            <ArrowLeftIcon className="h-6 w-6" />
          </button>
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{child.firstName} {child.lastName}'s Books</h1>
            <p className="text-lg text-gray-500">{child.grade}</p>
          </div>
          {canEdit && (
            <button
              onClick={() => setShowEditChildModal(true)}
              className="p-2 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-full transition-colors"
              title="Edit child information"
            >
              <PencilIcon className="h-5 w-5" />
            </button>
          )}
        </div>
      </div>

      {/* Month Navigation */}
      <div className="flex items-center justify-between bg-gray-50 p-6 rounded-lg">
        <button
          onClick={() => navigateMonth(-1)}
          className="p-2 rounded-full hover:bg-gray-200 transition-colors"
        >
          <ChevronLeftIcon className="h-5 w-5" />
        </button>
        
        <div className="text-center">
          <h2 className="text-xl font-semibold text-gray-900">
            {formatMonthYear(currentDate)}
          </h2>
          <p className="text-sm text-gray-600">
            {books.length} book{books.length !== 1 ? 's' : ''} read
          </p>
          <div className="mt-2 flex justify-center gap-2">
            {books.length > 0 && (
              <button
                onClick={handleDownloadPDF}
                className="inline-flex items-center px-3 py-1 border border-transparent rounded-md shadow-sm text-xs font-medium text-indigo-600 bg-indigo-100 hover:bg-indigo-200 transition-colors"
              >
                <DocumentArrowDownIcon className="h-4 w-4 mr-1" />
                PDF
              </button>
            )}
            {canEdit && (
              <button
                onClick={() => setShowAddBook(true)}
                className="inline-flex items-center px-3 py-1 border border-transparent rounded-md shadow-sm text-xs font-medium text-indigo-600 bg-indigo-100 hover:bg-indigo-200 transition-colors"
                data-testid="add-book-button"
              >
                <PlusIcon className="h-4 w-4 mr-1" />
                Add
              </button>
            )}
          </div>
        </div>
        
        <button
          onClick={() => navigateMonth(1)}
          className="p-2 rounded-full hover:bg-gray-200 transition-colors"
        >
          <ChevronRightIcon className="h-5 w-5" />
        </button>
      </div>

      {/* Books List */}
      <div className="space-y-4">
        {books.length === 0 ? (
          <div className="text-center py-12">
            <div className="text-gray-400 text-lg mb-4">
              No books recorded for {formatMonthYear(currentDate)}
            </div>
            {canEdit && (
              <button
                onClick={() => setShowAddBook(true)}
                className="inline-flex items-center px-3 py-1 border border-transparent rounded-md shadow-sm text-xs font-medium text-indigo-600 bg-indigo-100 hover:bg-indigo-200 transition-colors"
                data-testid="add-book-button"
              >
                <PlusIcon className="h-4 w-4 mr-1" />
                Add
              </button>
            )}
          </div>
        ) : (
          <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {books.map((book, index) => (
              <div key={book.id} className="bg-white border border-gray-200 rounded-lg p-3 sm:p-4 shadow-sm hover:shadow-md transition-shadow">
                {/* Book Cover - Top positioned */}
                <div className="flex justify-center mb-3">
                  <div className="w-16 h-20 sm:w-20 sm:h-24">
                    {book.coverUrl ? (
                      <img
                        src={book.coverUrl}
                        alt={`Cover of ${book.title}`}
                        className="w-full h-full object-cover rounded-md border border-gray-200"
                        onError={(e) => {
                          e.target.style.display = 'none'
                        }}
                      />
                    ) : (
                      <div className="w-full h-full bg-gray-100 rounded-md border border-gray-200 flex items-center justify-center">
                        <BookOpenIcon className="h-6 w-6 text-gray-400" />
                      </div>
                    )}
                  </div>
                </div>
                
                {/* Book Details */}
                <div className="text-center">
                  <div className="flex flex-wrap justify-center items-center gap-2 mb-2">
                    {book.isPartial && (
                      <span className="bg-yellow-100 text-yellow-800 text-xs font-medium px-2 py-1 rounded-full">
                        Partial
                      </span>
                    )}
                    <span className="text-xs text-gray-500 flex items-center gap-1">
                      {formatDateRead(book.dateRead)}
                      {!book.readByParent && (
                        <span className="text-yellow-500">⭐</span>
                      )}
                    </span>
                  </div>
                  <h5 className="text-sm font-semibold text-gray-900 mb-1 truncate">
                    {book.title}
                  </h5>
                  <p className="text-xs text-gray-600 mb-2 truncate">
                    by {book.author}
                  </p>
                  {book.lexileLevel && (
                    <p className="text-xs text-gray-500 mb-2">
                      Lexile: {book.lexileLevel}
                    </p>
                  )}
                  {book.isPartial && book.partialComment && (
                    <div className="bg-yellow-50 border border-yellow-200 rounded p-2 mt-2">
                      <p className="text-xs text-yellow-800">
                        <span className="font-medium">Portion:</span> {book.partialComment}
                      </p>
                    </div>
                  )}
                </div>

                {/* Action Buttons - Bottom positioned */}
                {canEdit && (
                  <div className="flex justify-center gap-2 mt-3 pt-3 border-t border-gray-100">
                    <button
                      onClick={() => handleEditBook(book)}
                      className="p-2 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-full transition-colors border border-gray-200 bg-white shadow-sm"
                      title="Edit book"
                    >
                      <PencilIcon className="h-4 w-4" />
                    </button>
                    <button
                      onClick={() => handleDeleteBook(book)}
                      className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-full transition-colors border border-gray-200 bg-white shadow-sm"
                      title="Delete book"
                    >
                      <TrashIcon className="h-4 w-4" />
                    </button>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Summary Footer */}
      {books.length > 0 && (
        <div className="mt-8 pt-6 border-t border-gray-200">
          <div className="text-center">
            <p className="text-lg font-medium text-gray-900">
              Total books read in {formatMonthYear(currentDate)}: {books.length}
            </p>
            <p className="text-sm text-gray-500 mt-1">
              All time total: {totalBookCount} books
            </p>
          </div>
        </div>
      )}

      {/* Edit Book Modal */}
      {showEditModal && selectedBook && (
        <EditBookModal
          book={selectedBook}
          onClose={() => {
            setShowEditModal(false)
            setSelectedBook(null)
          }}
          onBookUpdated={handleBookUpdated}
        />
      )}

      {/* Edit Child Modal */}
      {showEditChildModal && (
        <EditChildModal
          child={child}
          onClose={() => setShowEditChildModal(false)}
          onChildUpdated={handleChildUpdated}
          onChildDeleted={handleChildDeleted}
        />
      )}

      {/* Add Book Modal */}
      {showAddBook && child && createPortal(
        <AddBookModal
          child={child}
          onClose={() => setShowAddBook(false)}
          onBookAdded={handleBookAdded}
        />,
        document.body
      )}
    </div>
  )
}