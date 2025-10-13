import { useState, useEffect, useRef } from 'react'
import { XMarkIcon, CameraIcon } from '@heroicons/react/24/outline'
import { Html5QrcodeScanner, Html5QrcodeScanType } from 'html5-qrcode'
import api from '../services/api'

export default function AddBookModal({ child, onClose, onBookAdded }) {
  const [formData, setFormData] = useState({
    isbn: '',
    title: '',
    author: '',
    lexileLevel: '',
    dateRead: new Date().toISOString().split('T')[0]
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [existingBooks, setExistingBooks] = useState([])
  const [isDuplicate, setIsDuplicate] = useState(false)
  const [isbnLookupLoading, setIsbnLookupLoading] = useState(false)
  const [isbnLookupError, setIsbnLookupError] = useState('')
  const [showScanner, setShowScanner] = useState(false)
  const [scannerError, setScannerError] = useState('')
  const scannerRef = useRef(null)
  const html5QrcodeScannerRef = useRef(null)

  useEffect(() => {
    fetchExistingBooks()
  }, [child.id])

  useEffect(() => {
    // Cleanup scanner when component unmounts or scanner is hidden
    return () => {
      if (html5QrcodeScannerRef.current) {
        html5QrcodeScannerRef.current.clear()
      }
    }
  }, [])

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
      setError(`${child.firstName} ${child.lastName} has already read "${formData.title}" by ${formData.author}`)
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

  const lookupISBN = async () => {
    if (!formData.isbn.trim()) {
      setIsbnLookupError('Please enter an ISBN first')
      return
    }

    setIsbnLookupLoading(true)
    setIsbnLookupError('')

    try {
      const response = await api.post('/books/lookup-isbn', {
        isbn: formData.isbn
      })

      if (response.data.found) {
        setFormData({
          ...formData,
          title: response.data.title || formData.title,
          author: response.data.author || formData.author,
          lexileLevel: response.data.lexileLevel || formData.lexileLevel
        })
        setIsbnLookupError('')
      } else {
        setIsbnLookupError('Book not found. Please fill in manually.')
      }
    } catch (error) {
      setIsbnLookupError(error.response?.data?.message || 'Failed to lookup ISBN')
    } finally {
      setIsbnLookupLoading(false)
    }
  }

  const startScanner = () => {
    setShowScanner(true)
    setScannerError('')
    
    setTimeout(() => {
      if (scannerRef.current) {
        const html5QrcodeScanner = new Html5QrcodeScanner(
          "barcode-scanner",
          {
            fps: 10,
            qrbox: { width: 250, height: 100 },
            aspectRatio: 1.777778,
            supportedScanTypes: [Html5QrcodeScanType.SCAN_TYPE_CAMERA]
          },
          false
        )

        html5QrcodeScannerRef.current = html5QrcodeScanner

        html5QrcodeScanner.render(
          (decodedText) => {
            // Success callback - ISBN found
            const cleanedISBN = decodedText.replace(/[^\d]/g, '') // Remove non-digits
            if (cleanedISBN.length === 10 || cleanedISBN.length === 13) {
              setFormData(prev => ({ ...prev, isbn: cleanedISBN }))
              stopScanner()
              // Auto-lookup the book info after scanning
              setTimeout(() => {
                // Trigger lookup with the scanned ISBN
                setIsbnLookupLoading(true)
                setIsbnLookupError('')
                
                api.post('/books/lookup-isbn', { isbn: cleanedISBN })
                  .then(response => {
                    if (response.data.found) {
                      setFormData(prev => ({
                        ...prev,
                        isbn: cleanedISBN,
                        title: response.data.title || prev.title,
                        author: response.data.author || prev.author,
                        lexileLevel: response.data.lexileLevel || prev.lexileLevel
                      }))
                      setIsbnLookupError('')
                    } else {
                      setIsbnLookupError('Book not found. Please fill in manually.')
                    }
                  })
                  .catch(error => {
                    setIsbnLookupError(error.response?.data?.message || 'Failed to lookup ISBN')
                  })
                  .finally(() => {
                    setIsbnLookupLoading(false)
                  })
              }, 500)
            } else {
              setScannerError('Invalid ISBN format. Please try again.')
            }
          },
          (error) => {
            // Error callback - scanning failed
            console.log('Scanner error:', error)
            // Don't show every scanning attempt error, only real failures
          }
        )
      }
    }, 100)
  }

  const stopScanner = () => {
    if (html5QrcodeScannerRef.current) {
      html5QrcodeScannerRef.current.clear()
      html5QrcodeScannerRef.current = null
    }
    setShowScanner(false)
    setScannerError('')
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    
    if (isDuplicate) {
      return
    }

    setLoading(true)

    try {
      await api.post(`/books/child/${child.id}`, {
        isbn: formData.isbn,
        title: formData.title,
        author: formData.author,
        lexileLevel: formData.lexileLevel,
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
    if (!formData.isbn || !formData.title || !formData.author || isDuplicate) return
    
    setLoading(true)

    try {
      await api.post(`/books/child/${child.id}`, {
        isbn: formData.isbn,
        title: formData.title,
        author: formData.author,
        lexileLevel: formData.lexileLevel,
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

  const isFormValid = formData.isbn.trim() && formData.title.trim() && formData.author.trim() && !isDuplicate

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border w-[520px] shadow-lg rounded-md bg-white">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-medium">Add Book for {child.firstName} {child.lastName}</h3>
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
              ISBN
            </label>
            <div className="flex space-x-2">
              <input
                type="text"
                name="isbn"
                required
                placeholder="978-0-123456-78-9"
                className="mt-1 block flex-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                value={formData.isbn}
                onChange={handleChange}
              />
              <button
                type="button"
                onClick={startScanner}
                disabled={showScanner}
                className="mt-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                title="Scan barcode"
              >
                <CameraIcon className="h-4 w-4" />
              </button>
              <button
                type="button"
                onClick={lookupISBN}
                disabled={isbnLookupLoading || !formData.isbn.trim()}
                className="mt-1 px-3 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isbnLookupLoading ? 'Looking up...' : 'Lookup'}
              </button>
            </div>
            {isbnLookupError && (
              <div className="text-red-600 text-sm mt-1">{isbnLookupError}</div>
            )}
            {scannerError && (
              <div className="text-red-600 text-sm mt-1">{scannerError}</div>
            )}
          </div>

          {/* Barcode Scanner */}
          {showScanner && (
            <div className="border rounded-lg p-4 bg-gray-50">
              <div className="flex justify-between items-center mb-3">
                <h4 className="text-sm font-medium text-gray-700">Scan ISBN Barcode</h4>
                <button
                  type="button"
                  onClick={stopScanner}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <XMarkIcon className="h-5 w-5" />
                </button>
              </div>
              <div 
                id="barcode-scanner" 
                ref={scannerRef}
                className="w-full"
              ></div>
              <p className="text-xs text-gray-500 mt-2">
                Position the ISBN barcode within the scanner frame. The camera will automatically detect and scan the code.
              </p>
            </div>
          )}

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
              Lexile Level (optional)
            </label>
            <div className="flex space-x-2">
              <input
                type="text"
                name="lexileLevel"
                placeholder="e.g., 650L"
                className="mt-1 block flex-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                value={formData.lexileLevel}
                onChange={handleChange}
              />
              {formData.isbn && (
                <a
                  href={`https://hub.lexile.com/find-a-book/details/${formData.isbn}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="mt-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50"
                  title={`View Lexile level for ISBN ${formData.isbn}`}
                >
                  Find Lexile
                </a>
              )}
            </div>
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
                {child.firstName} {child.lastName} has already read this book. Each child can only read a book once.
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
                title={isDuplicate ? "Cannot add duplicate book" : !isFormValid ? "Please fill in ISBN, title and author first" : "Mark as finished today"}
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