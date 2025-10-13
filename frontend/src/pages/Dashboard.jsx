import { useState, useEffect } from 'react'
import { PlusIcon, ChartBarIcon, ShareIcon, ChevronLeftIcon, ChevronRightIcon, PencilIcon } from '@heroicons/react/24/outline'
import api from '../services/api'
import ChildCard from '../components/ChildCard'
import AddChildModal from '../components/AddChildModal'
import AddBookModal from '../components/AddBookModal'
import BulkShareModal from '../components/BulkShareModal'
import ChildManagementModal from '../components/ChildManagementModal'
import FullScreenChildView from '../components/FullScreenChildView'
import ReportModal from '../components/ReportModal'

export default function Dashboard() {
  const [children, setChildren] = useState([])
  const [loading, setLoading] = useState(true)
  const [showAddChild, setShowAddChild] = useState(false)
  const [showAddBook, setShowAddBook] = useState(false)
  const [showBulkShare, setShowBulkShare] = useState(false)
  const [showChildManagement, setShowChildManagement] = useState(false)
  const [showFullScreenView, setShowFullScreenView] = useState(false)
  const [showReport, setShowReport] = useState(false)
  const [selectedChild, setSelectedChild] = useState(null)
  const [refreshTrigger, setRefreshTrigger] = useState(0)
  const [currentMonth, setCurrentMonth] = useState(new Date())

  useEffect(() => {
    fetchChildren()
  }, [])

  const fetchChildren = async () => {
    try {
      const response = await api.get('/children')
      setChildren(response.data || [])
    } catch (error) {
      console.error('Failed to fetch children:', error)
      setChildren([])
    } finally {
      setLoading(false)
    }
  }

  const handleChildAdded = () => {
    fetchChildren()
    setShowAddChild(false)
  }

  const handleBookAdded = () => {
    fetchChildren()
    setRefreshTrigger(prev => prev + 1) // Force ChildCard components to refresh
    setShowAddBook(false)
  }

  const handleAddBook = (child) => {
    setSelectedChild(child)
    setShowAddBook(true)
  }

  const handleManageChild = (child) => {
    setSelectedChild(child)
    setShowChildManagement(true)
  }

  const handleViewChild = (child) => {
    setSelectedChild(child)
    setShowFullScreenView(true)
  }

  const navigateMonth = (direction) => {
    const newDate = new Date(currentMonth)
    newDate.setMonth(currentMonth.getMonth() + direction)
    setCurrentMonth(newDate)
  }

  const formatMonthYear = (date) => {
    return date.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
  }


  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-indigo-600"></div>
      </div>
    )
  }

  return (
    <div className="py-6">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 md:px-8">
        <div className="md:flex md:items-center md:justify-between">
          <div className="flex-1 min-w-0">
            <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
              My Children
            </h2>
            {/* Month Navigation */}
            <div className="flex items-center mt-2 space-x-4">
              <button
                onClick={() => navigateMonth(-1)}
                className="p-1 rounded-full hover:bg-gray-200 transition-colors"
              >
                <ChevronLeftIcon className="h-5 w-5 text-gray-600" />
              </button>
              
              <div className="text-lg font-medium text-gray-700">
                {formatMonthYear(currentMonth)}
              </div>
              
              <button
                onClick={() => navigateMonth(1)}
                className="p-1 rounded-full hover:bg-gray-200 transition-colors"
              >
                <ChevronRightIcon className="h-5 w-5 text-gray-600" />
              </button>
            </div>
          </div>
          <div className="mt-4 flex md:mt-0 md:ml-4 space-x-3">
            <button
              onClick={() => setShowReport(true)}
              className="inline-flex items-center px-3 py-2 md:px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              title="Generate Report"
            >
              <ChartBarIcon className="h-5 w-5 md:-ml-1 md:mr-2" />
              <span className="hidden md:inline">Generate Report</span>
            </button>
            {children.length > 0 && (
              <button
                onClick={() => setShowBulkShare(true)}
                className="inline-flex items-center px-3 py-2 md:px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
                title="Share Children"
              >
                <ShareIcon className="h-5 w-5 md:-ml-1 md:mr-2" />
                <span className="hidden md:inline">Share Children</span>
              </button>
            )}
            <button
              onClick={() => setShowAddChild(true)}
              className="inline-flex items-center px-3 py-2 md:px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700"
              title="Add Child"
            >
              <PlusIcon className="h-5 w-5 md:-ml-1 md:mr-2" />
              <span className="hidden md:inline">Add Child</span>
            </button>
          </div>
        </div>

        <div className="mt-8">
          {!children || children.length === 0 ? (
            <div className="text-center py-12">
              <div className="text-gray-400 text-lg">No children added yet</div>
              <button
                onClick={() => setShowAddChild(true)}
                className="mt-4 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700"
              >
                <PlusIcon className="-ml-1 mr-2 h-5 w-5" />
                Add Your First Child
              </button>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {children.map((child) => (
                <div key={`${child.id}-${refreshTrigger}`} className="relative">
                  <ChildCard
                    child={child}
                    currentMonth={currentMonth}
                    onAddBook={() => handleAddBook(child)}
                    onViewDetails={() => handleViewChild(child)}
                  />
                  {/* Child Edit Button */}
                  <div className="absolute top-2 right-2">
                    <button
                      onClick={() => handleManageChild(child)}
                      className="p-1.5 rounded-full bg-white shadow-md hover:bg-indigo-50 hover:text-indigo-600 transition-colors border border-gray-200"
                      title="Edit child information"
                    >
                      <PencilIcon className="h-4 w-4 text-gray-600" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {showAddChild && (
        <AddChildModal
          onClose={() => setShowAddChild(false)}
          onChildAdded={handleChildAdded}
        />
      )}

      {showAddBook && selectedChild && (
        <AddBookModal
          child={selectedChild}
          onClose={() => setShowAddBook(false)}
          onBookAdded={handleBookAdded}
        />
      )}

      {showBulkShare && (
        <BulkShareModal
          children={children}
          onClose={() => setShowBulkShare(false)}
        />
      )}

      {showChildManagement && selectedChild && (
        <ChildManagementModal
          child={selectedChild}
          onClose={() => setShowChildManagement(false)}
          onChildUpdated={() => {
            fetchChildren()
            setRefreshTrigger(prev => prev + 1)
          }}
        />
      )}

      {showFullScreenView && selectedChild && (
        <FullScreenChildView
          child={selectedChild}
          onClose={() => setShowFullScreenView(false)}
          onAddBook={handleAddBook}
        />
      )}

      {showReport && (
        <ReportModal onClose={() => setShowReport(false)} currentMonth={currentMonth} />
      )}
    </div>
  )
}