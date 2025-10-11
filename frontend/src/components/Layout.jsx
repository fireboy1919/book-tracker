import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { 
  HomeIcon, 
  UserGroupIcon, 
  BookOpenIcon,
  ChartBarIcon,
  UserIcon,
  ArrowRightOnRectangleIcon,
  Bars3Icon,
  XMarkIcon
} from '@heroicons/react/24/outline'
import { Link, useLocation, useNavigate } from 'react-router-dom'

export default function Layout({ children }) {
  const { user, logout } = useAuth()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const location = useLocation()
  const navigate = useNavigate()

  const handleLogout = () => {
    console.log('handleLogout called')
    logout()
    console.log('logout() called, navigating to /login')
    navigate('/login')
  }

  const navigation = [
    { name: 'Dashboard', href: '/dashboard', icon: HomeIcon, color: 'bg-blue-500 hover:bg-blue-600' },
    ...(user?.isAdmin ? [{ name: 'Admin Panel', href: '/admin', icon: UserGroupIcon, color: 'bg-purple-500 hover:bg-purple-600' }] : [])
  ]

  return (
    <div className="min-h-screen bg-gradient-to-br from-yellow-100 via-pink-100 to-blue-100">
      {/* Top Navigation Bar - Always Visible */}
      <nav className="bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 shadow-lg">
        <div className="max-w-7xl mx-auto px-4">
          <div className="flex justify-between items-center h-16">
            {/* Logo */}
            <div className="flex items-center">
              <BookOpenIcon className="h-8 w-8 text-white" />
              <span className="ml-2 text-xl font-bold text-white">📚 Book Tracker</span>
            </div>

            {/* Desktop Navigation */}
            <div className="hidden md:flex items-center space-x-4">
              {navigation.map((item) => {
                const current = location.pathname === item.href
                return (
                  <Link
                    key={item.name}
                    to={item.href}
                    className={`flex items-center px-4 py-2 rounded-full text-white font-medium transition-all duration-200 ${
                      current
                        ? 'bg-white/20 shadow-md'
                        : 'hover:bg-white/10'
                    }`}
                  >
                    <item.icon className="h-5 w-5 mr-2" />
                    {item.name}
                  </Link>
                )
              })}
              
              {/* User Info & Logout */}
              <div className="flex items-center space-x-3 ml-6 pl-6 border-l border-white/20">
                <div className="flex items-center">
                  <div className="w-8 h-8 bg-white/20 rounded-full flex items-center justify-center">
                    <UserIcon className="h-5 w-5 text-white" />
                  </div>
                  <div className="ml-2 text-white">
                    <p className="text-sm font-medium">{user?.firstName} {user?.lastName}</p>
                    <p className="text-xs opacity-75">{user?.email}</p>
                  </div>
                </div>
                <button
                  onClick={handleLogout}
                  className="p-2 rounded-full text-white hover:bg-white/10 transition-colors"
                  title="Logout"
                >
                  <ArrowRightOnRectangleIcon className="h-5 w-5" />
                </button>
              </div>
            </div>

            {/* Mobile menu button */}
            <div className="md:hidden">
              <button
                onClick={() => setSidebarOpen(!sidebarOpen)}
                className="p-2 rounded-md text-white hover:bg-white/10"
              >
                {sidebarOpen ? (
                  <XMarkIcon className="h-6 w-6" />
                ) : (
                  <Bars3Icon className="h-6 w-6" />
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Mobile Navigation Dropdown */}
        {sidebarOpen && (
          <div className="md:hidden bg-white/10 backdrop-blur-sm border-t border-white/20">
            <div className="px-4 py-3 space-y-2">
              {navigation.map((item) => {
                const current = location.pathname === item.href
                return (
                  <Link
                    key={item.name}
                    to={item.href}
                    onClick={() => setSidebarOpen(false)}
                    className={`flex items-center px-3 py-2 rounded-lg text-white font-medium ${
                      current
                        ? 'bg-white/20'
                        : 'hover:bg-white/10'
                    }`}
                  >
                    <item.icon className="h-5 w-5 mr-3" />
                    {item.name}
                  </Link>
                )
              })}
              
              {/* Mobile User Info */}
              <div className="pt-3 mt-3 border-t border-white/20">
                <div className="flex items-center px-3 py-2">
                  <div className="w-8 h-8 bg-white/20 rounded-full flex items-center justify-center">
                    <UserIcon className="h-5 w-5 text-white" />
                  </div>
                  <div className="ml-3 text-white">
                    <p className="text-sm font-medium">{user?.firstName} {user?.lastName}</p>
                    <p className="text-xs opacity-75">{user?.email}</p>
                  </div>
                  <button
                    onClick={handleLogout}
                    className="ml-auto p-2 rounded-full text-white hover:bg-white/10"
                    title="Logout"
                  >
                    <ArrowRightOnRectangleIcon className="h-5 w-5" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </nav>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 py-6">
        <div className="bg-white rounded-2xl shadow-xl min-h-[calc(100vh-8rem)] p-6">
          {children}
        </div>
      </main>
    </div>
  )
}