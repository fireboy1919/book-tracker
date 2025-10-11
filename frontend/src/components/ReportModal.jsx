import { useState, useEffect } from 'react'
import { XMarkIcon, DocumentArrowDownIcon } from '@heroicons/react/24/outline'
import api from '../services/api'

export default function ReportModal({ onClose }) {
  const [report, setReport] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    fetchReport()
  }, [])

  const fetchReport = async () => {
    try {
      const response = await api.get('/reports/my-books')
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
        csvContent += `"${childReport.child.name}","${childReport.child.grade}","${book.title}","${book.author}","${dateRead}"\n`
      })
    })

    const blob = new Blob([csvContent], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `book-report-${new Date().toISOString().split('T')[0]}.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  }

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-10 mx-auto p-5 border w-4/5 max-w-4xl shadow-lg rounded-md bg-white">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-medium">Reading Report</h3>
          <div className="flex items-center space-x-2">
            {report && (
              <button
                onClick={downloadReport}
                className="inline-flex items-center px-3 py-1 border border-transparent text-sm font-medium rounded-md text-indigo-700 bg-indigo-100 hover:bg-indigo-200"
              >
                <DocumentArrowDownIcon className="h-4 w-4 mr-1" />
                Download CSV
              </button>
            )}
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600"
            >
              <XMarkIcon className="h-6 w-6" />
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
                <div key={childReport.child.id} className="border rounded-lg p-4">
                  <div className="flex justify-between items-center mb-3">
                    <h4 className="text-lg font-medium">
                      {childReport.child.name} ({childReport.child.grade})
                    </h4>
                    <span className="text-sm text-gray-500">
                      {childReport.totalBooks} books read
                    </span>
                  </div>
                  
                  {childReport.books.length === 0 ? (
                    <div className="text-gray-500 text-sm">No books recorded</div>
                  ) : (
                    <div className="overflow-x-auto">
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