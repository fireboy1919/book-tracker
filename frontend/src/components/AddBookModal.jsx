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
    dateRead: new Date().toISOString().split('T')[0],
    isPartial: false,
    partialComment: '',
    coverUrl: ''
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
      // Clear ISBN lookup timeout
      if (window.isbnLookupTimeout) {
        clearTimeout(window.isbnLookupTimeout)
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
    const { name, value, type, checked } = e.target
    const newFormData = {
      ...formData,
      [name]: type === 'checkbox' ? checked : value
    }
    
    setFormData(newFormData)
    
    // Auto-lookup ISBN when a complete ISBN is entered
    if (name === 'isbn' && value) {
      const cleanedISBN = value.replace(/[^\d]/g, '')
      if ((cleanedISBN.length === 10 || cleanedISBN.length === 13) && 
          (cleanedISBN.startsWith('978') || cleanedISBN.startsWith('979') || cleanedISBN.length === 10)) {
        // Debounce the lookup to avoid excessive API calls
        if (window.isbnLookupTimeout) {
          clearTimeout(window.isbnLookupTimeout)
        }
        
        window.isbnLookupTimeout = setTimeout(() => {
          performISBNLookup(cleanedISBN)
        }, 500) // Wait 500ms after user stops typing
      }
    }
  }

  const performISBNLookup = async (isbn) => {
    if (!isbn) return

    setIsbnLookupLoading(true)
    setIsbnLookupError('')

    try {
      const response = await api.post('/books/lookup-isbn', { isbn: isbn })
      if (response.data.found) {
        setFormData(prev => ({
          ...prev,
          isbn: isbn,
          title: response.data.title || prev.title,
          author: response.data.author || prev.author,
          lexileLevel: response.data.lexileLevel || prev.lexileLevel,
          coverUrl: response.data.coverUrl || prev.coverUrl
        }))
        setIsbnLookupError('')
      } else {
        setIsbnLookupError('Book not found. Please fill in manually.')
      }
    } catch (error) {
      console.error('ISBN lookup error:', error)
      setIsbnLookupError(error.response?.data?.message || 'Failed to lookup ISBN')
    } finally {
      setIsbnLookupLoading(false)
    }
  }

  const lookupISBN = async () => {
    if (!formData.isbn.trim()) {
      setIsbnLookupError('Please enter an ISBN first')
      return
    }

    const cleanedISBN = formData.isbn.replace(/[^\d]/g, '')
    await performISBNLookup(cleanedISBN)
  }

  const extractISBN = (decodedText) => {
    // Remove all non-digit characters
    const cleanedText = decodedText.replace(/[^\d]/g, '')
    
    // Check for EAN-13 ISBN (most common book barcode format)
    if (cleanedText.length === 13) {
      // Valid ISBN-13 must start with 978 or 979
      if (cleanedText.startsWith('978') || cleanedText.startsWith('979')) {
        return cleanedText
      }
    }
    
    // Check for ISBN-10 format
    if (cleanedText.length === 10) {
      return cleanedText
    }
    
    // Check if the scanned text contains multiple numbers (common with price barcodes nearby)
    // Look for patterns like "9781234567890" within longer strings
    const isbn13Match = decodedText.match(/\b(978|979)\d{10}\b/)
    if (isbn13Match) {
      return isbn13Match[0]
    }
    
    // Look for 10-digit ISBN patterns
    const isbn10Match = decodedText.match(/\b\d{10}\b/)
    if (isbn10Match) {
      return isbn10Match[0]
    }
    
    return null // No valid ISBN found
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
            qrbox: { width: 300, height: 120 },
            aspectRatio: 1.777778,
            supportedScanTypes: [Html5QrcodeScanType.SCAN_TYPE_CAMERA],
            // Improve barcode detection
            experimentalFeatures: {
              useBarCodeDetectorIfSupported: true
            },
            rememberLastUsedCamera: true
          },
          false
        )

        html5QrcodeScannerRef.current = html5QrcodeScanner

        html5QrcodeScanner.render(
          (decodedText) => {
            // Success callback - attempt to extract valid ISBN
            const validISBN = extractISBN(decodedText)
            if (validISBN) {
              setFormData(prev => ({ ...prev, isbn: validISBN }))
              stopScanner()
              // Auto-lookup the book info after scanning
              setTimeout(() => {
                // Trigger lookup with the scanned ISBN
                setIsbnLookupLoading(true)
                setIsbnLookupError('')
                
                api.post('/books/lookup-isbn', { isbn: validISBN })
                  .then(response => {
                    if (response.data.found) {
                      setFormData(prev => ({
                        ...prev,
                        isbn: validISBN,
                        title: response.data.title || prev.title,
                        author: response.data.author || prev.author,
                        lexileLevel: response.data.lexileLevel || prev.lexileLevel,
                        coverUrl: response.data.coverUrl || prev.coverUrl
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
              setScannerError('No valid ISBN found. Try scanning the larger barcode on the back cover.')
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
        childId: child.id,
        isCustomBook: !formData.isbn || formData.isbn.trim() === '',
        isPartial: formData.isPartial,
        partialComment: formData.partialComment
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
        childId: child.id,
        isCustomBook: !formData.isbn || formData.isbn.trim() === '',
        isPartial: formData.isPartial,
        partialComment: formData.partialComment
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
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-[60]">
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
                title="Scan the main book barcode (usually on the back cover)"
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

          {/* Book Display Section */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Book Information
            </label>
            <div 
              className={`border-2 rounded-lg p-4 transition-colors ${
                formData.title || formData.author ? 'border-indigo-200 bg-indigo-50' : 'border-gray-300 bg-gray-50'
              }`}
              onClick={() => {
                // Allow editing if no ISBN lookup data or if it's a custom book
                if (!formData.isbn || (!formData.title && !formData.author)) {
                  document.querySelector('input[name="title"]')?.focus()
                }
              }}
              style={{ cursor: (!formData.isbn || (!formData.title && !formData.author)) ? 'pointer' : 'default' }}
            >
              <div className="flex items-start space-x-4">
                {/* Book Cover Placeholder */}
                <div className="flex-shrink-0 w-16 h-20 bg-gray-200 rounded border border-gray-300 flex items-center justify-center">
                  {formData.coverUrl ? (
                    <img
                      src={formData.coverUrl}
                      alt="Book cover"
                      className="w-full h-full object-cover rounded"
                      onError={(e) => {
                        e.target.style.display = 'none'
                      }}
                    />
                  ) : (
                    <div className="text-xs text-gray-400 text-center p-1">No Cover</div>
                  )}
                </div>
                
                {/* Book Details */}
                <div className="flex-1 min-w-0">
                  {formData.title || formData.author ? (
                    <>
                      <h4 className="text-base font-semibold text-gray-900 truncate">
                        {formData.title || 'Untitled'}
                      </h4>
                      <p className="text-sm text-gray-600 truncate">
                        by {formData.author || 'Unknown Author'}
                      </p>
                      {formData.isbn && (
                        <p className="text-xs text-gray-500 mt-1">
                          {formData.isbn.startsWith('978') || formData.isbn.startsWith('979') ? 'ISBN-13: ' : 'ISBN-10: '}{formData.isbn}
                        </p>
                      )}
                    </>
                  ) : (
                    <div className="text-gray-500">
                      <p className="text-sm font-medium">Click to add book details</p>
                      <p className="text-xs">Or use ISBN lookup above</p>
                    </div>
                  )}
                </div>
              </div>
              
              {/* Editable fields (hidden by default, shown when clicked or no data) */}
              {(!formData.isbn || !formData.title || !formData.author) && (
                <div className="mt-3 space-y-2 border-t pt-3">
                  <input
                    type="text"
                    name="title"
                    placeholder="Enter book title"
                    className="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm text-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                    value={formData.title}
                    onChange={handleChange}
                  />
                  <input
                    type="text"
                    name="author"
                    placeholder="Enter author name"
                    className="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm text-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                    value={formData.author}
                    onChange={handleChange}
                  />
                </div>
              )}
            </div>
            
            {formData.isbn && formData.title && formData.author && (
              <p className="text-xs text-gray-500 mt-1">
                Book details populated from ISBN lookup. This will be a shared book.
              </p>
            )}
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

          {/* Partial Book Checkbox */}
          <div className="flex items-center">
            <input
              type="checkbox"
              id="isPartial"
              name="isPartial"
              checked={formData.isPartial}
              onChange={handleChange}
              className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
            />
            <label htmlFor="isPartial" className="ml-2 block text-sm text-gray-900">
              This is a partial reading (e.g., only read some chapters)
            </label>
          </div>

          {/* Partial Comment Field */}
          {formData.isPartial && (
            <div>
              <label className="block text-sm font-medium text-gray-700">
                What portion was read?
              </label>
              <textarea
                name="partialComment"
                rows={3}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                placeholder="e.g., Chapters 1-5, First half, Pages 1-100, etc."
                value={formData.partialComment}
                onChange={handleChange}
              />
              <p className="mt-1 text-sm text-gray-500">
                Describe which part of the book was read (optional)
              </p>
            </div>
          )}

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