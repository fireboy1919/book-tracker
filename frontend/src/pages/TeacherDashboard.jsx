import { useState, useEffect } from 'react'
import { PlusIcon, UsersIcon, BookOpenIcon, UserPlusIcon, TrashIcon, ClipboardDocumentIcon } from '@heroicons/react/24/outline'
import { CheckIcon } from '@heroicons/react/24/solid'
import api from '../services/api'
import CreateClassModal from '../components/CreateClassModal'
import FullScreenChildView from '../components/FullScreenChildView'

export default function TeacherDashboard() {
  const [classes, setClasses] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [selectedClass, setSelectedClass] = useState(null)
  const [classStudents, setClassStudents] = useState([])
  const [classTeachers, setClassTeachers] = useState([])
  const [showAssignModal, setShowAssignModal] = useState(false)
  const [assignModalType, setAssignModalType] = useState('') // 'teacher' or 'student'
  const [availableUsers, setAvailableUsers] = useState([])
  const [invitationData, setInvitationData] = useState(null)
  const [isEditingClassName, setIsEditingClassName] = useState(false)
  const [editingClassName, setEditingClassName] = useState('')
  const [copySuccess, setCopySuccess] = useState(false)
  const [assigningUser, setAssigningUser] = useState(false)
  const [successMessage, setSuccessMessage] = useState('')
  const [showFullScreenView, setShowFullScreenView] = useState(false)
  const [selectedStudent, setSelectedStudent] = useState(null)

  useEffect(() => {
    fetchClasses()
  }, [])

  const fetchClasses = async () => {
    try {
      const response = await api.get('/classes')
      setClasses(response.data)
    } catch (error) {
      setError('Failed to fetch classes')
    } finally {
      setLoading(false)
    }
  }

  const fetchClassStudents = async (classId) => {
    try {
      const response = await api.get(`/classes/${classId}/students`)
      setClassStudents(response.data)
    } catch (error) {
      setError('Failed to fetch class students')
    }
  }

  const fetchClassTeachers = async (classId) => {
    try {
      const response = await api.get(`/classes/${classId}/teachers`)
      setClassTeachers(response.data)
    } catch (error) {
      setError('Failed to fetch class teachers')
    }
  }

  const fetchInvitationData = async (classId) => {
    try {
      const response = await api.get(`/classes/${classId}/invitation-data`)
      setInvitationData(response.data)
    } catch (error) {
      console.error('Failed to fetch invitation data:', error)
    }
  }

  const handleClassClick = async (classItem) => {
    setSelectedClass(classItem)
    await fetchClassStudents(classItem.id)
    await fetchClassTeachers(classItem.id)
    await fetchInvitationData(classItem.id)
  }

  const handleCreateSuccess = () => {
    fetchClasses()
  }

  const fetchAvailableUsers = async (userType) => {
    try {
      console.log(`Fetching available ${userType}s`)
      if (userType === 'teacher') {
        const response = await api.get('/users')
        // Filter teachers and exclude already assigned ones
        const availableTeachers = response.data.filter(user => 
          user.isTeacher && !(classTeachers || []).some(teacher => teacher.id === user.id)
        )
        console.log('Available teachers:', availableTeachers)
        setAvailableUsers(availableTeachers)
      } else {
        // For students, fetch children instead of users
        const response = await api.get('/children')
        // Filter children not already assigned to this class
        const availableChildren = response.data.filter(child => 
          !(classStudents || []).some(student => student.id === child.id)
        )
        console.log('Available children:', availableChildren)
        setAvailableUsers(availableChildren)
      }
    } catch (error) {
      console.error(`Error fetching available ${userType}s:`, error)
      setError(`Failed to fetch available ${userType === 'teacher' ? 'teachers' : 'children'}`)
    }
  }

  const openAssignModal = (type) => {
    if (!selectedClass) {
      setError('Please select a class first')
      return
    }
    setAssignModalType(type)
    setShowAssignModal(true)
    fetchAvailableUsers(type)
  }

  const assignUserToClass = async (userId) => {
    if (assigningUser) return // Prevent double-clicks
    
    setAssigningUser(true)
    try {
      console.log(`Assigning ${assignModalType} with ID ${userId} to class ${selectedClass.id}`)
      
      if (assignModalType === 'teacher') {
        await api.post(`/classes/${selectedClass.id}/members`, {
          userId: userId,
          role: 'TEACHER'
        })
      } else {
        // For students (children), use the assign-child endpoint
        const response = await api.post('/classes/assign-child', {
          childId: userId,
          classId: selectedClass.id
        })
        console.log('Assign child response:', response)
      }
      setShowAssignModal(false)
      setError('')
      setSuccessMessage(`Successfully added ${assignModalType === 'teacher' ? 'teacher' : 'student'} to class!`)
      setTimeout(() => setSuccessMessage(''), 3000)
      // Refresh students and teachers list if we're looking at this class
      if (selectedClass) {
        await fetchClassStudents(selectedClass.id)
        await fetchClassTeachers(selectedClass.id)
      }
      console.log(`Successfully assigned ${assignModalType} to class`)
    } catch (error) {
      console.error('Error assigning user to class:', error)
      const errorMessage = error.response?.data?.error || error.message || 'Unknown error'
      setError(`Failed to assign ${assignModalType === 'teacher' ? 'teacher' : 'child'} to class: ${errorMessage}`)
    } finally {
      setAssigningUser(false)
    }
  }

  const removeUserFromClass = async (userId, userType) => {
    if (!selectedClass) return
    
    try {
      if (userType === 'teacher') {
        await api.delete(`/classes/${selectedClass.id}/members/${userId}`)
        fetchClassTeachers(selectedClass.id)
      } else {
        // For students (children), use the new endpoint to remove them from the class
        await api.delete(`/classes/${selectedClass.id}/children/${userId}`)
        fetchClassStudents(selectedClass.id)
      }
      setError('')
    } catch (error) {
      setError('Failed to remove user from class')
    }
  }

  const copyInvitationKey = async () => {
    try {
      await navigator.clipboard.writeText(invitationData.invitation_key)
      setCopySuccess(true)
      setTimeout(() => setCopySuccess(false), 2000)
    } catch (error) {
      setError('Failed to copy invitation key')
    }
  }

  const startEditingClassName = () => {
    setEditingClassName(selectedClass.name)
    setIsEditingClassName(true)
  }

  const cancelEditingClassName = () => {
    setIsEditingClassName(false)
    setEditingClassName('')
  }

  const saveClassName = async () => {
    if (!selectedClass || !editingClassName.trim()) return
    
    try {
      await api.put(`/classes/${selectedClass.id}`, {
        ...selectedClass,
        name: editingClassName.trim()
      })
      
      // Update the selected class locally
      setSelectedClass(prev => ({ ...prev, name: editingClassName.trim() }))
      
      // Update in the classes list
      setClasses(prev => 
        prev.map(cls => 
          cls.id === selectedClass.id 
            ? { ...cls, name: editingClassName.trim() }
            : cls
        )
      )
      
      setIsEditingClassName(false)
      setEditingClassName('')
      setError('')
    } catch (error) {
      setError('Failed to update class name')
    }
  }

  const handleViewStudent = (student) => {
    setSelectedStudent(student)
    setShowFullScreenView(true)
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
              Teacher Dashboard
            </h2>
          </div>
          <div className="mt-4 flex md:mt-0 md:ml-4">
            <button
              onClick={() => setShowCreateModal(true)}
              className="ml-3 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              <PlusIcon className="-ml-1 mr-2 h-5 w-5" />
              Create Class
            </button>
          </div>
        </div>

        {error && (
          <div className="mt-4 bg-red-50 border border-red-400 text-red-700 px-4 py-3 rounded">
            {error}
          </div>
        )}

        {successMessage && (
          <div className="mt-4 bg-green-50 border border-green-400 text-green-700 px-4 py-3 rounded">
            {successMessage}
          </div>
        )}

        <div className="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-3">
          {/* Classes List */}
          <div className="lg:col-span-1 order-1 lg:order-none">
            <div className="bg-white shadow rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  My Classes
                </h3>
                {(classes?.length || 0) === 0 ? (
                  <p className="text-gray-500 text-sm">No classes yet. Create your first class!</p>
                ) : (
                  <div className="space-y-3">
                    {(classes || []).map((classItem) => (
                      <div
                        key={classItem.id}
                        onClick={() => handleClassClick(classItem)}
                        className={`p-3 rounded-lg border cursor-pointer transition-colors ${
                          selectedClass?.id === classItem.id
                            ? 'border-indigo-500 bg-indigo-50'
                            : 'border-gray-200 hover:border-gray-300'
                        }`}
                      >
                        <div className="flex items-center justify-between">
                          <h4 className="text-sm font-medium text-gray-900">
                            {classItem.name}
                          </h4>
                          <UsersIcon className="h-5 w-5 text-gray-400" />
                        </div>
                        {classItem.description && (
                          <p className="text-xs text-gray-500 mt-1">
                            {classItem.description}
                          </p>
                        )}
                        <div className="flex justify-between text-xs text-gray-500 mt-2">
                          <span>Student Goal: {classItem.studentBooksGoal}</span>
                          <span>Read-to Goal: {classItem.otherBooksGoal}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Class Details */}
          <div className="lg:col-span-2 order-2 lg:order-none">
            {selectedClass ? (
              <div className="bg-white shadow rounded-lg">
                <div className="px-4 py-5 sm:p-6">
                  <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-4 space-y-3 sm:space-y-0">
                    {isEditingClassName ? (
                      <div className="flex items-center space-x-2">
                        <input
                          type="text"
                          value={editingClassName}
                          onChange={(e) => setEditingClassName(e.target.value)}
                          onKeyPress={(e) => {
                            if (e.key === 'Enter') saveClassName()
                            if (e.key === 'Escape') cancelEditingClassName()
                          }}
                          className="text-lg font-medium border border-gray-300 rounded px-2 py-1 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                          autoFocus
                        />
                        <button
                          onClick={saveClassName}
                          className="text-green-600 hover:text-green-800"
                        >
                          ✓
                        </button>
                        <button
                          onClick={cancelEditingClassName}
                          className="text-red-600 hover:text-red-800"
                        >
                          ✕
                        </button>
                      </div>
                    ) : (
                      <h3 
                        className="text-lg leading-6 font-medium text-gray-900 cursor-pointer hover:text-indigo-600 flex items-center"
                        onClick={startEditingClassName}
                        title="Click to edit class name"
                      >
                        {selectedClass.name}
                        <span className="ml-2 text-sm text-gray-400">✏️</span>
                      </h3>
                    )}
                    <div className="flex flex-col sm:flex-row sm:items-center space-y-2 sm:space-y-0 sm:space-x-3">
                      <div className="flex items-center text-sm text-gray-500">
                        <UsersIcon className="h-4 w-4 mr-1" />
                        {classStudents?.length || 0} students, {classTeachers?.length || 0} teachers
                      </div>
                      <div className="flex space-x-2">
                        <button
                          onClick={() => openAssignModal('teacher')}
                          className="inline-flex items-center px-3 py-1 border border-transparent text-xs font-medium rounded-full text-blue-700 bg-blue-100 hover:bg-blue-200"
                        >
                          <UserPlusIcon className="h-3 w-3 mr-1" />
                          <span className="hidden sm:inline">Add </span>Teacher
                        </button>
                        <button
                          onClick={() => openAssignModal('student')}
                          className="inline-flex items-center px-3 py-1 border border-transparent text-xs font-medium rounded-full text-green-700 bg-green-100 hover:bg-green-200"
                        >
                          <UserPlusIcon className="h-3 w-3 mr-1" />
                          <span className="hidden sm:inline">Add </span>Student
                        </button>
                      </div>
                    </div>
                  </div>

                  {/* Invitation Key Section */}
                  {invitationData && (
                    <div className="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
                      <div className="space-y-3">
                        <div>
                          <h4 className="text-sm font-medium text-blue-900 mb-1">Class Invitation Key</h4>
                          <p className="text-xs text-blue-700 mb-2">Use this key for Gmail mail merge to invite students</p>
                        </div>
                        
                        <div className="relative">
                          <textarea
                            readOnly
                            value={invitationData.invitation_key}
                            className="w-full text-xs bg-white px-3 py-2 rounded border text-gray-800 resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            rows="3"
                            style={{ wordBreak: 'break-all' }}
                            onClick={(e) => e.target.select()}
                          />
                          <button
                            onClick={copyInvitationKey}
                            className={`absolute top-2 right-2 p-1 rounded transition-colors shadow-sm ${
                              copySuccess 
                                ? 'bg-green-600 text-white' 
                                : 'bg-blue-600 text-white hover:bg-blue-700'
                            }`}
                            title={copySuccess ? 'Copied!' : 'Copy to clipboard'}
                          >
                            <ClipboardDocumentIcon className="h-4 w-4" />
                          </button>
                        </div>
                        
                        <div className="flex flex-col sm:flex-row space-y-2 sm:space-y-0 sm:space-x-2">
                          <button
                            onClick={copyInvitationKey}
                            className={`text-xs px-3 py-1 rounded transition-colors flex items-center justify-center space-x-1 ${
                              copySuccess 
                                ? 'bg-green-600 text-white' 
                                : 'bg-blue-600 text-white hover:bg-blue-700'
                            }`}
                          >
                            <ClipboardDocumentIcon className="h-3 w-3" />
                            <span>{copySuccess ? 'Copied!' : 'Copy Key'}</span>
                          </button>
                          <a
                            href="/help/mail-merge"
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-xs px-3 py-1 bg-green-600 text-white rounded hover:bg-green-700 text-center"
                          >
                            Help Guide
                          </a>
                        </div>
                      </div>
                    </div>
                  )}

                  <div className="border-b border-gray-200">
                    <nav className="-mb-px flex space-x-8">
                      <button
                        className="border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm"
                      >
                        Students ({classStudents?.length || 0})
                      </button>
                      <button
                        className="border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm"
                      >
                        Teachers ({classTeachers?.length || 0})
                      </button>
                    </nav>
                  </div>
                  
                  {/* Students Section */}
                  <div className="mt-4">
                    <h4 className="text-md font-medium text-gray-900 mb-3">Students</h4>
                    {(classStudents?.length || 0) === 0 ? (
                      <div className="text-center py-8">
                        <BookOpenIcon className="mx-auto h-8 w-8 text-gray-400" />
                        <h3 className="mt-2 text-sm font-medium text-gray-900">No students assigned</h3>
                        <p className="mt-1 text-sm text-gray-500">
                          Students can be assigned to this class by parents or admins.
                        </p>
                      </div>
                    ) : (
                      <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
                        <table className="min-w-full divide-y divide-gray-200">
                          <thead className="bg-gray-50">
                            <tr>
                              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Student
                              </th>
                              <th className="px-3 py-2 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Read
                              </th>
                              <th className="px-3 py-2 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Read-to
                              </th>
                              <th className="px-3 py-2 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Status
                              </th>
                              <th className="px-3 py-2 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Actions
                              </th>
                            </tr>
                          </thead>
                          <tbody className="bg-white divide-y divide-gray-200">
                            {classStudents.map((student) => (
                              <tr 
                                key={student.id} 
                                className="hover:bg-gray-50 cursor-pointer transition-colors"
                                onClick={() => handleViewStudent(student)}
                              >
                                <td className="px-3 py-2 whitespace-nowrap">
                                  <div className="flex items-center">
                                    <div className="flex-shrink-0 h-6 w-6">
                                      <div className="h-6 w-6 rounded-full bg-indigo-100 flex items-center justify-center">
                                        <span className="text-xs font-medium text-indigo-700">
                                          {student.firstName[0]}{student.lastName[0]}
                                        </span>
                                      </div>
                                    </div>
                                    <div className="ml-2">
                                      <div className="text-sm font-medium text-gray-900">
                                        {student.firstName} {student.lastName}
                                      </div>
                                    </div>
                                  </div>
                                </td>
                                <td className="px-3 py-2 whitespace-nowrap text-center">
                                  <div className="text-sm font-medium text-gray-900">
                                    {student.studentBooksRead}/{student.studentGoal}
                                  </div>
                                </td>
                                <td className="px-3 py-2 whitespace-nowrap text-center">
                                  <div className="text-sm font-medium text-gray-900">
                                    {student.readToBooksRead}/{student.readToGoal}
                                  </div>
                                </td>
                                <td className="px-3 py-2 whitespace-nowrap text-center">
                                  {student.goalsReached ? (
                                    <CheckIcon className="h-5 w-5 text-green-500 mx-auto" title="Goals completed!" />
                                  ) : (
                                    <span className="text-gray-400">-</span>
                                  )}
                                </td>
                                <td className="px-3 py-2 whitespace-nowrap text-center">
                                  <button
                                    onClick={(e) => {
                                      e.stopPropagation() // Prevent row click
                                      if (window.confirm(`Remove ${student.firstName} ${student.lastName} from this class?`)) {
                                        removeUserFromClass(student.id, 'student')
                                      }
                                    }}
                                    className="text-red-500 hover:text-red-700 p-1 rounded-md hover:bg-red-50 transition-colors"
                                    title={`Remove ${student.firstName} ${student.lastName} from class`}
                                  >
                                    <TrashIcon className="h-4 w-4" />
                                  </button>
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </div>

                  {/* Teachers Section */}
                  <div className="mt-6">
                    <h4 className="text-md font-medium text-gray-900 mb-3">Teachers</h4>
                    {(classTeachers?.length || 0) === 0 ? (
                      <div className="text-center py-8">
                        <UsersIcon className="mx-auto h-8 w-8 text-gray-400" />
                        <h3 className="mt-2 text-sm font-medium text-gray-900">No other teachers assigned</h3>
                        <p className="mt-1 text-sm text-gray-500">
                          Add other teachers to help manage this class.
                        </p>
                      </div>
                    ) : (
                      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-1 xl:grid-cols-2">
                        {classTeachers.map((teacher) => (
                          <div key={teacher.id} className="border border-blue-200 rounded-lg p-4 bg-blue-50">
                            <div className="flex items-center justify-between">
                              <div className="flex items-center">
                                <div className="flex-shrink-0">
                                  <div className="h-10 w-10 rounded-full bg-blue-100 flex items-center justify-center">
                                    <span className="text-sm font-medium text-blue-700">
                                      {teacher.firstName[0]}{teacher.lastName[0]}
                                    </span>
                                  </div>
                                </div>
                                <div className="ml-4">
                                  <div className="text-sm font-medium text-gray-900">
                                    {teacher.firstName} {teacher.lastName}
                                  </div>
                                  <div className="text-sm text-gray-500">
                                    {teacher.email}
                                  </div>
                                  <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800">
                                    Teacher
                                  </span>
                                </div>
                              </div>
                              <button
                                onClick={() => {
                                  if (window.confirm(`Remove ${teacher.firstName} ${teacher.lastName} from this class?`)) {
                                    removeUserFromClass(teacher.id, 'teacher')
                                  }
                                }}
                                className="text-red-500 hover:text-red-700 p-1 rounded-md hover:bg-red-50 transition-colors"
                                title={`Remove ${teacher.firstName} ${teacher.lastName} from class`}
                              >
                                <TrashIcon className="h-4 w-4" />
                              </button>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ) : (
              <div className="bg-white shadow rounded-lg">
                <div className="px-4 py-5 sm:p-6">
                  <div className="text-center py-12">
                    <UsersIcon className="mx-auto h-12 w-12 text-gray-400" />
                    <h3 className="mt-2 text-sm font-medium text-gray-900">Select a class</h3>
                    <p className="mt-1 text-sm text-gray-500">
                      Choose a class from the left to view its students and manage assignments.
                    </p>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      <CreateClassModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSuccess={handleCreateSuccess}
      />

      {/* User Assignment Modal */}
      {showAssignModal && selectedClass && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
          <div className="relative top-20 mx-auto p-5 border max-w-md w-full mx-4 shadow-lg rounded-md bg-white">
            <div className="mt-3">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium text-gray-900">
                  Add {assignModalType === 'teacher' ? 'Teacher' : 'Student'} to {selectedClass.name}
                </h3>
                <button
                  onClick={() => setShowAssignModal(false)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <span className="sr-only">Close</span>
                  <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              
              <div className="mb-4">
                <p className="text-sm text-gray-500 mb-3">
                  Select a {assignModalType === 'teacher' ? 'teacher' : 'child'} to assign to this class:
                </p>
                
                {(availableUsers?.length || 0) === 0 ? (
                  <p className="text-sm text-gray-500">
                    No available {assignModalType === 'teacher' ? 'teachers' : 'children'} found
                  </p>
                ) : (
                  <div className="space-y-2 max-h-60 overflow-y-auto">
                    {(availableUsers || []).map((user) => (
                      <button
                        key={user.id}
                        onClick={() => assignUserToClass(user.id)}
                        disabled={assigningUser}
                        className={`w-full text-left p-3 border border-gray-200 rounded-lg transition-colors ${
                          assigningUser 
                            ? 'opacity-50 cursor-not-allowed' 
                            : 'hover:border-indigo-300 hover:bg-indigo-50'
                        }`}
                      >
                        <div className="flex items-center justify-between">
                          <div className="font-medium text-gray-900">
                            {user.firstName} {user.lastName}
                          </div>
                          {assigningUser && (
                            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-indigo-600"></div>
                          )}
                        </div>
                        {assignModalType === 'teacher' ? (
                          <>
                            <div className="text-sm text-gray-500">{user.email}</div>
                            {user.isTeacher && (
                              <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800 mt-1">
                                Teacher
                              </span>
                            )}
                            {user.isAdmin && (
                              <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800 mt-1 ml-1">
                                Admin
                              </span>
                            )}
                          </>
                        ) : (
                          <div className="text-sm text-gray-500">Grade: {user.grade}</div>
                        )}
                      </button>
                    ))}
                  </div>
                )}
              </div>
              
              <div className="flex justify-end">
                <button
                  onClick={() => setShowAssignModal(false)}
                  className="px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400 transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Full Screen Child View */}
      {showFullScreenView && selectedStudent && (
        <FullScreenChildView
          child={selectedStudent}
          onClose={() => setShowFullScreenView(false)}
          onAddBook={() => {
            // Teachers shouldn't be able to add books
            // This is just for viewing
          }}
        />
      )}
    </div>
  )
}