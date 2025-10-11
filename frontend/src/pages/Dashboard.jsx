import { useState, useEffect } from 'react'
import { PlusIcon, UserPlusIcon, ChartBarIcon } from '@heroicons/react/24/outline'
import api from '../services/api'
import ChildCard from '../components/ChildCard'
import AddChildModal from '../components/AddChildModal'
import AddBookModal from '../components/AddBookModal'
import InviteUserModal from '../components/InviteUserModal'
import ReportModal from '../components/ReportModal'

export default function Dashboard() {
  const [children, setChildren] = useState([])
  const [loading, setLoading] = useState(true)
  const [showAddChild, setShowAddChild] = useState(false)
  const [showAddBook, setShowAddBook] = useState(false)
  const [showInviteUser, setShowInviteUser] = useState(false)
  const [showReport, setShowReport] = useState(false)
  const [selectedChild, setSelectedChild] = useState(null)
  const [refreshTrigger, setRefreshTrigger] = useState(0)

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

  const handleInviteUser = (child) => {
    setSelectedChild(child)
    setShowInviteUser(true)
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
          </div>
          <div className="mt-4 flex md:mt-0 md:ml-4 space-x-3">
            <button
              onClick={() => setShowReport(true)}
              className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
            >
              <ChartBarIcon className="-ml-1 mr-2 h-5 w-5" />
              Generate Report
            </button>
            <button
              onClick={() => setShowAddChild(true)}
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700"
            >
              <PlusIcon className="-ml-1 mr-2 h-5 w-5" />
              Add Child
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
                  onAddBook={() => handleAddBook(child)}
                  onInviteUser={() => handleInviteUser(child)}
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

      {showAddBook && selectedChild && (
        <AddBookModal
          child={selectedChild}
          onClose={() => setShowAddBook(false)}
          onBookAdded={handleBookAdded}
        />
      )}

      {showInviteUser && selectedChild && (
        <InviteUserModal
          child={selectedChild}
          onClose={() => setShowInviteUser(false)}
        />
      )}

      {showReport && (
        <ReportModal onClose={() => setShowReport(false)} />
      )}
    </div>
  )
}