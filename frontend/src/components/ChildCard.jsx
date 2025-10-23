import { useState, useEffect } from 'react'
import { BookOpenIcon, PlusIcon, EyeIcon, PencilIcon } from '@heroicons/react/24/outline'
import api from '../services/api'

export default function ChildCard({ child, onAddBook, onViewDetails, onEditChild, currentMonth, onChildUpdate, currentUser }) {
  const [books, setBooks] = useState([])
  const [loading, setLoading] = useState(true)
  const [currentMonthBooks, setCurrentMonthBooks] = useState([])
  const [availableClasses, setAvailableClasses] = useState([])
  const [assigningClass, setAssigningClass] = useState(false)
  const [currentClass, setCurrentClass] = useState(null)

  useEffect(() => {
    fetchBooks()
    fetchAvailableClasses()
    fetchCurrentClass()
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

  const fetchAvailableClasses = async () => {
    try {
      const response = await api.get('/classes/available')
      setAvailableClasses(response.data || [])
    } catch (error) {
      console.error('Failed to fetch available classes:', error)
    }
  }

  const fetchCurrentClass = async () => {
    if (child.classId) {
      try {
        const response = await api.get(`/classes/${child.classId}`)
        setCurrentClass(response.data)
      } catch (error) {
        console.error('Failed to fetch current class:', error)
      }
    } else {
      setCurrentClass(null)
    }
  }

  const assignToClass = async (classId) => {
    setAssigningClass(true)
    try {
      if (classId === '') {
        // Remove from class (set to null)
        await api.post('/classes/assign-child', {
          childId: child.id,
          classId: null
        })
      } else {
        // Assign to new class
        await api.post('/classes/assign-child', {
          childId: child.id,
          classId: parseInt(classId)
        })
      }
      
      // Refresh data
      fetchCurrentClass()
      if (onChildUpdate) {
        onChildUpdate()
      }
    } catch (error) {
      console.error('Failed to assign child to class:', error)
      alert(error.response?.data?.message || 'Failed to assign to class. Please try again.')
    } finally {
      setAssigningClass(false)
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
                {/* Class assignment section */}
                <dd className="mt-2">
                  {child.classId && currentClass ? (
                    <div className="text-xs">
                      <span className="text-gray-600">Class: </span>
                      <span className="font-medium text-indigo-600">{currentClass.name}</span>
                      {currentUser && (currentUser.isTeacher || currentUser.isAdmin) && (
                        <button
                          onClick={() => {
                            if (window.confirm('Remove child from this class?')) {
                              assignToClass('')
                            }
                          }}
                          className="ml-2 text-red-500 hover:text-red-700 text-xs"
                        >
                          ✕
                        </button>
                      )}
                    </div>
                  ) : (
                    availableClasses.length > 0 && (
                      <select
                        onChange={(e) => assignToClass(e.target.value)}
                        disabled={assigningClass}
                        className="text-xs border border-gray-300 rounded px-2 py-1 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                        defaultValue=""
                      >
                        <option value="">Assign to class...</option>
                        {availableClasses.map(classItem => (
                          <option key={classItem.id} value={classItem.id}>
                            {classItem.name}
                          </option>
                        ))}
                      </select>
                    )
                  )}
                </dd>
              </dl>
            </div>
          </div>
          
          {/* Action buttons on the right */}
          <div className="flex-shrink-0 ml-2 sm:ml-3">
            <button
              onClick={() => onEditChild(child)}
              className="p-1 sm:p-2 rounded hover:bg-indigo-50 hover:text-indigo-600 transition-colors"
              title="Edit child information"
            >
              <PencilIcon className="h-5 w-5 sm:h-6 sm:w-6 text-gray-600" />
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