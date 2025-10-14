import { useState, useEffect } from 'react'
import { XMarkIcon, DocumentArrowDownIcon } from '@heroicons/react/24/outline'
import api from '../services/api'

export default function ReportModal({ onClose, currentMonth }) {
  const [report, setReport] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    fetchReport()
  }, [currentMonth])

  const fetchReport = async () => {
    try {
      const year = currentMonth.getFullYear()
      const month = currentMonth.getMonth() + 1 // JavaScript months are 0-indexed
      const response = await api.get(`/reports/my-books?year=${year}&month=${month}`)
      setReport(response.data)
    } catch (error) {
      setError('Failed to generate report')
    } finally {
      setLoading(false)
    }
  }

  const downloadReport = () => {
    if (!report) return

    let csvContent = "Child Name,Grade,Book Title,Author,Date Read\n"
    
    report.children.forEach(childReport => {
      childReport.books.forEach(book => {
        const dateRead = new Date(book.dateRead).toLocaleDateString()
        csvContent += `"${childReport.child.firstName} ${childReport.child.lastName}","${childReport.child.grade}","${book.title}","${book.author}","${dateRead}"\n`
      })
    })

    const blob = new Blob([csvContent], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    const monthYear = `${currentMonth.getFullYear()}-${String(currentMonth.getMonth() + 1).padStart(2, '0')}`
    a.download = `book-report-${monthYear}.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  }

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-2 sm:top-10 mx-auto p-3 sm:p-5 border w-full sm:w-4/5 max-w-4xl shadow-lg rounded-md bg-white m-2 sm:m-0">
        <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center mb-4 space-y-2 sm:space-y-0">
          <h3 className="text-base sm:text-lg font-medium pr-8 sm:pr-0">
            Reading Report - {currentMonth.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}
          </h3>
          <div className="flex items-center space-x-2 absolute top-3 right-3 sm:relative sm:top-auto sm:right-auto">
            {report && (
              <button
                onClick={downloadReport}
                className="inline-flex items-center px-2 py-1 sm:px-3 sm:py-1 border border-transparent text-xs sm:text-sm font-medium rounded-md text-indigo-700 bg-indigo-100 hover:bg-indigo-200"
              >
                <DocumentArrowDownIcon className="h-3 w-3 sm:h-4 sm:w-4 mr-1" />
                <span className="hidden sm:inline">Download CSV</span>
                <span className="sm:hidden">CSV</span>
              </button>
            )}
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600"
            >
              <XMarkIcon className="h-5 w-5 sm:h-6 sm:w-6" />
            </button>
          </div>
        </div>

        {loading && (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
          </div>
        )}

        {error && (
          <div className="text-red-600 text-center py-4">{error}</div>
        )}

        {report && (
          <div className="space-y-6">
            {report.children.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                No children or books found
              </div>
            ) : (
              report.children.map((childReport) => (
                <div key={childReport.child.id} className="border rounded-lg p-3 sm:p-4">
                  <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center mb-3 space-y-1 sm:space-y-0">
                    <h4 className="text-base sm:text-lg font-medium">
                      {childReport.child.firstName} {childReport.child.lastName} ({childReport.child.grade})
                    </h4>
                    <span className="text-xs sm:text-sm text-gray-500">
                      {childReport.totalBooks} books read
                    </span>
                  </div>
                  
                  {childReport.books.length === 0 ? (
                    <div className="text-gray-500 text-sm">No books recorded</div>
                  ) : (
                    <div className="space-y-2 sm:space-y-0">
                      {/* Mobile: Card View */}
                      <div className="sm:hidden space-y-3">
                        {childReport.books.map((book) => (
                          <div key={book.id} className="bg-gray-50 rounded-lg p-3">
                            <h5 className="font-medium text-sm text-gray-900 mb-1">{book.title}</h5>
                            <p className="text-xs text-gray-600 mb-1">by {book.author}</p>
                            <p className="text-xs text-gray-500">Read: {new Date(book.dateRead).toLocaleDateString()}</p>
                          </div>
                        ))}
                      </div>
                      
                      {/* Desktop: Table View */}
                      <div className="hidden sm:block overflow-x-auto">
                        <table className="min-w-full divide-y divide-gray-200">
                          <thead className="bg-gray-50">
                            <tr>
                              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                                Title
                              </th>
                              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                                Author
                              </th>
                              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                                Date Read
                              </th>
                            </tr>
                          </thead>
                          <tbody className="bg-white divide-y divide-gray-200">
                            {childReport.books.map((book) => (
                              <tr key={book.id}>
                                <td className="px-4 py-2 text-sm text-gray-900">
                                  {book.title}
                                </td>
                                <td className="px-4 py-2 text-sm text-gray-900">
                                  {book.author}
                                </td>
                                <td className="px-4 py-2 text-sm text-gray-900">
                                  {new Date(book.dateRead).toLocaleDateString()}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}
                </div>
              ))
            )}
          </div>
        )}
      </div>
    </div>
  )
}