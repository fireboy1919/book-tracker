import { useState, useEffect } from 'react'
import { createPortal } from 'react-dom'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { PlusIcon, ChartBarIcon, ShareIcon, ChevronLeftIcon, ChevronRightIcon, PencilIcon } from '@heroicons/react/24/outline'
import api from '../services/api'
import { useAuth } from '../contexts/AuthContext'
import ChildCard from '../components/ChildCard'
import AddChildModal from '../components/AddChildModal'
import AddBookModal from '../components/AddBookModal'
import BulkShareModal from '../components/BulkShareModal'
import ChildManagementModal from '../components/ChildManagementModal'
import ReportModal from '../components/ReportModal'

export default function Dashboard() {
  const { user: currentUser } = useAuth()
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()
  const [children, setChildren] = useState([])
  const [loading, setLoading] = useState(true)
  const [showAddChild, setShowAddChild] = useState(false)
  const [showAddBook, setShowAddBook] = useState(false)
  const [showBulkShare, setShowBulkShare] = useState(false)
  const [showChildManagement, setShowChildManagement] = useState(false)
  const [showReport, setShowReport] = useState(false)
  const [selectedChild, setSelectedChild] = useState(null)
  const [refreshTrigger, setRefreshTrigger] = useState(0)

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

  const currentMonth = getCurrentMonth()

  useEffect(() => {
    fetchChildren()
  }, [searchParams])

  const fetchChildren = async () => {
    try {
      // Fetch children with book counts for the current month
      const year = currentMonth.getFullYear()
      const month = currentMonth.getMonth() + 1 // JavaScript months are 0-based
      const response = await api.get(`/children/with-counts?year=${year}&month=${month}`)
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

  const handleBookAdded = async () => {
    setShowAddBook(false)
    // Refresh data after book has been successfully added
    await fetchChildren()
    setRefreshTrigger(prev => prev + 1) // Force ChildCard components to refresh
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
    // Navigate to child detail page with current month context
    const monthParam = searchParams.get('month')
    const yearParam = searchParams.get('year')
    
    const params = new URLSearchParams()
    if (monthParam && yearParam) {
      params.set('month', monthParam)
      params.set('year', yearParam)
    }
    
    const queryString = params.toString()
    navigate(`/dashboard/child/${child.id}${queryString ? `?${queryString}` : ''}`)
  }

  const navigateMonth = (direction) => {
    const newDate = new Date(currentMonth)
    newDate.setMonth(currentMonth.getMonth() + direction)
    
    // Update URL parameters
    const newSearchParams = new URLSearchParams(searchParams)
    newSearchParams.set('year', newDate.getFullYear().toString())
    newSearchParams.set('month', (newDate.getMonth() + 1).toString()) // Convert back to 1-based
    setSearchParams(newSearchParams)
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
        <div>
          {/* Header with title and buttons aligned */}
          <div className="flex items-center justify-between">
            <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
              My Children
            </h2>
            <div className="flex space-x-1 ml-4">
            <button
              onClick={() => setShowReport(true)}
              className="p-1.5 md:px-3 md:py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 inline-flex items-center"
              title="Generate Report"
            >
              <ChartBarIcon className="h-6 w-6 md:h-7 md:w-7 md:mr-2" />
              <span className="hidden md:inline">Generate Report</span>
            </button>
            {children.length > 0 && (
              <button
                onClick={() => setShowBulkShare(true)}
                className="p-1.5 md:px-3 md:py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 inline-flex items-center"
                title="Share Children"
              >
                <ShareIcon className="h-6 w-6 md:h-7 md:w-7 md:mr-2" />
                <span className="hidden md:inline">Share Children</span>
              </button>
            )}
            <button
              onClick={() => setShowAddChild(true)}
              className="p-1.5 md:px-3 md:py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 inline-flex items-center"
              title="Add Child"
            >
              <PlusIcon className="h-6 w-6 md:h-7 md:w-7 md:mr-2" />
              <span className="hidden md:inline">Add Child</span>
            </button>
            </div>
          </div>
          
          {/* Month Navigation */}
          <div className="flex items-center mt-4 space-x-4">
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
                <ChildCard
                  key={`${child.id}-${refreshTrigger}`}
                  child={child}
                  currentMonth={currentMonth}
                  currentUser={currentUser}
                  refreshTrigger={refreshTrigger}
                  onAddBook={() => handleAddBook(child)}
                  onViewDetails={() => handleViewChild(child)}
                  onEditChild={() => handleManageChild(child)}
                  onChildUpdate={fetchChildren}
                />
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

      {showAddBook && selectedChild && createPortal(
        <AddBookModal
          child={selectedChild}
          onClose={() => setShowAddBook(false)}
          onBookAdded={handleBookAdded}
        />,
        document.body
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
          onChildDeleted={(childId) => {
            setChildren(prevChildren => prevChildren.filter(child => child.id !== childId))
            setShowChildManagement(false)
          }}
        />
      )}


      {showReport && (
        <ReportModal onClose={() => setShowReport(false)} currentMonth={currentMonth} />
      )}
    </div>
  )
}