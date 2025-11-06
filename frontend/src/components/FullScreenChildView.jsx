import { useState, useEffect } from 'react'
import { XMarkIcon, ChevronLeftIcon, ChevronRightIcon, PlusIcon, PencilIcon, TrashIcon, DocumentArrowDownIcon, BookOpenIcon } from '@heroicons/react/24/outline'
import api from '../services/api'
import EditBookModal from './EditBookModal'
import EditChildModal from './EditChildModal'

export default function FullScreenChildView({ child, onClose, onAddBook, onSetBookUpdateCallback, onChildDeleted, currentMonth }) {
  const [books, setBooks] = useState([])
  const [loading, setLoading] = useState(true)
  const [currentDate, setCurrentDate] = useState(currentMonth || new Date())
  const [totalBookCount, setTotalBookCount] = useState(0)
  const [showEditModal, setShowEditModal] = useState(false)
  const [selectedBook, setSelectedBook] = useState(null)
  const [canEdit, setCanEdit] = useState(false)
  const [checkingPermissions, setCheckingPermissions] = useState(true)
  const [showEditChildModal, setShowEditChildModal] = useState(false)
  const [childData, setChildData] = useState(child)

  useEffect(() => {
    fetchBooksForCurrentMonth()
    fetchTotalBookCount()
    checkEditPermission()
    // Set up the callback for when books are added
    if (onSetBookUpdateCallback) {
      onSetBookUpdateCallback(async () => {
        await Promise.all([
          fetchBooksForCurrentMonth(),
          fetchTotalBookCount()
        ])
      })
    }
  }, [child.id, onSetBookUpdateCallback])

  useEffect(() => {
    fetchBooksForCurrentMonth()
  }, [currentDate])

  // Update currentDate when currentMonth prop changes
  useEffect(() => {
    if (currentMonth) {
      setCurrentDate(new Date(currentMonth))
    }
  }, [currentMonth])

  const fetchBooksForCurrentMonth = async () => {
    setLoading(true)
    try {
      const year = currentDate.getFullYear()
      const month = currentDate.getMonth() + 1 // JavaScript months are 0-based
      const response = await api.get(`/books/child/${child.id}?year=${year}&month=${month}`)
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
      const response = await api.get(`/books/child/${child.id}`)
      setTotalBookCount(response.data ? response.data.length : 0)
    } catch (error) {
      console.error('Failed to fetch total book count:', error)
      setTotalBookCount(0)
    }
  }

  const checkEditPermission = async () => {
    try {
      // Try to create a book request - this will check EDIT permission
      // We're not actually creating, just checking if we have permission
      const currentUser = JSON.parse(localStorage.getItem('user'))
      if (currentUser && child.ownerId === currentUser.id) {
        // User owns this child, so they have EDIT permission
        setCanEdit(true)
      } else {
        // For non-owners, we'll try a test request to see if we have EDIT permission
        // We can use a HEAD request or check permissions via another endpoint
        // For now, assume VIEW only for non-owners (can be enhanced later)
        setCanEdit(false)
      }
    } catch (error) {
      console.error('Failed to check permissions:', error)
      setCanEdit(false)
    } finally {
      setCheckingPermissions(false)
    }
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
    // Open the AddBookModal above this modal (don't close this one)
    onAddBook(child)
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

  const handleBookUpdated = (updatedBook) => {
    // Refresh the books list
    fetchBooksForCurrentMonth()
    fetchTotalBookCount()
  }

  const handleChildUpdated = (updatedChild) => {
    setChildData(updatedChild)
    setShowEditChildModal(false)
  }

  const handleChildDeleted = (childId) => {
    if (onChildDeleted) {
      onChildDeleted(childId)
    }
    onClose() // Close the full screen view since the child no longer exists
  }

  const handleDownloadPDF = async () => {
    const currentYear = currentDate.getFullYear()
    const currentMonth = currentDate.getMonth() + 1
    
    try {
      const response = await api.get(`/reports/child/${child.id}/monthly-pdf`, {
        params: { year: currentYear, month: currentMonth },
        responseType: 'blob'
      })
      
      // Create blob URL and download
      const blob = new Blob([response.data], { type: 'application/pdf' })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `${childData.firstName}_${childData.lastName}_books_${currentMonth}_${currentYear}.pdf`
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Failed to download PDF:', error)
      alert('Failed to download PDF report. Please try again.')
    }
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
          <div className="flex items-center space-x-4">
            <div>
              <h3 className="text-2xl font-bold text-gray-900">{childData.firstName} {childData.lastName}'s Books</h3>
              <p className="text-sm text-gray-500">{childData.grade}</p>
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
                  onClick={handleAddBook}
                  className="inline-flex items-center px-3 py-1 border border-transparent rounded-md shadow-sm text-xs font-medium text-indigo-600 bg-indigo-100 hover:bg-indigo-200 transition-colors"
                  data-testid="fullscreen-add-book-button"
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
                  onClick={handleAddBook}
                  className="inline-flex items-center px-3 py-1 border border-transparent rounded-md shadow-sm text-xs font-medium text-indigo-600 bg-indigo-100 hover:bg-indigo-200 transition-colors"
                  data-testid="fullscreen-add-book-button"
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
      </div>

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
          child={childData}
          onClose={() => setShowEditChildModal(false)}
          onChildUpdated={handleChildUpdated}
          onChildDeleted={handleChildDeleted}
        />
      )}
    </div>
  )
}