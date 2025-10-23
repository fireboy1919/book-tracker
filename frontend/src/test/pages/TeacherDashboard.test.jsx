import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import TeacherDashboard from '../../pages/TeacherDashboard'

// Mock the CreateClassModal component
vi.mock('../../components/CreateClassModal', () => ({
  default: ({ isOpen, onClose, onSuccess }) => {
    if (!isOpen) return null
    return (
      <div data-testid="create-class-modal">
        <button onClick={onClose}>Close Modal</button>
        <button onClick={() => { onSuccess(); onClose(); }}>Create Class</button>
      </div>
    )
  }
}))

// Mock the API
vi.mock('../../services/api', () => ({
  default: {
    get: vi.fn()
  }
}))

describe('TeacherDashboard', () => {
  const mockClasses = [
    {
      id: 1,
      name: 'Math Class',
      description: 'Elementary Math',
      studentBooksGoal: 10,
      otherBooksGoal: 5,
      createdById: 1,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z'
    },
    {
      id: 2,
      name: 'Reading Class',
      description: 'Advanced Reading',
      studentBooksGoal: 15,
      otherBooksGoal: 8,
      createdById: 1,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z'
    }
  ]

  const mockStudents = [
    {
      id: 1,
      firstName: 'Alice',
      lastName: 'Johnson',
      email: 'alice@example.com',
      isAdmin: false,
      isTeacher: false,
      emailVerified: true,
      createdAt: '2024-01-01T00:00:00Z'
    },
    {
      id: 2,
      firstName: 'Bob',
      lastName: 'Smith',
      email: 'bob@example.com',
      isAdmin: false,
      isTeacher: false,
      emailVerified: true,
      createdAt: '2024-01-01T00:00:00Z'
    }
  ]

  beforeEach(async () => {
    vi.clearAllMocks()
    const api = await import('../../services/api')
    // Default mock for getting classes
    api.default.get.mockImplementation((url) => {
      if (url === '/classes') {
        return Promise.resolve({ data: mockClasses })
      }
      if (url.includes('/students')) {
        return Promise.resolve({ data: mockStudents })
      }
      return Promise.resolve({ data: [] })
    })
  })

  it('renders teacher dashboard with title', async () => {
    render(<TeacherDashboard />)
    
    // Wait for loading to complete and content to render
    await waitFor(() => {
      expect(screen.getByText('Teacher Dashboard')).toBeInTheDocument()
    })
    expect(screen.getByText('Create Class')).toBeInTheDocument()
  })

  it('displays loading state initially', async () => {
    const api = await import('../../services/api')
    api.default.get.mockImplementationOnce(() => new Promise(() => {}))
    render(<TeacherDashboard />)
    
    // Check for the loading spinner element by its CSS class
    expect(document.querySelector('.animate-spin')).toBeInTheDocument()
  })

  it('fetches and displays classes', async () => {
    render(<TeacherDashboard />)
    
    await waitFor(() => {
      expect(screen.getByText('Math Class')).toBeInTheDocument()
      expect(screen.getByText('Reading Class')).toBeInTheDocument()
    })
    
    const api = await import('../../services/api')
    expect(api.default.get).toHaveBeenCalledWith('/classes')
  })

  it('displays empty state when no classes', async () => {
    const api = await import('../../services/api')
    api.default.get.mockImplementationOnce(() => Promise.resolve({ data: [] }))
    
    render(<TeacherDashboard />)
    
    await waitFor(() => {
      expect(screen.getByText('No classes yet. Create your first class!')).toBeInTheDocument()
    })
  })

  it('opens create class modal when button is clicked', async () => {
    const user = userEvent.setup()
    render(<TeacherDashboard />)
    
    // Wait for component to load first
    await waitFor(() => {
      expect(screen.getByText('Teacher Dashboard')).toBeInTheDocument()
    })
    
    const createButton = screen.getByRole('button', { name: /create class/i })
    await user.click(createButton)
    
    expect(screen.getByTestId('create-class-modal')).toBeInTheDocument()
  })

  it('closes create class modal', async () => {
    const user = userEvent.setup()
    render(<TeacherDashboard />)
    
    // Wait for component to load first
    await waitFor(() => {
      expect(screen.getByText('Teacher Dashboard')).toBeInTheDocument()
    })
    
    // Open modal
    const createButton = screen.getByRole('button', { name: /create class/i })
    await user.click(createButton)
    
    // Close modal
    const closeButton = screen.getByText('Close Modal')
    await user.click(closeButton)
    
    expect(screen.queryByTestId('create-class-modal')).not.toBeInTheDocument()
  })

  it('refetches classes after successful creation', async () => {
    const user = userEvent.setup()
    const api = await import('../../services/api')
    render(<TeacherDashboard />)
    
    // Wait for initial load
    await waitFor(() => {
      expect(api.default.get).toHaveBeenCalledWith('/classes')
    })
    
    api.default.get.mockClear()
    
    // Open modal and create class
    const createButton = screen.getByRole('button', { name: /create class/i })
    await user.click(createButton)
    
    // Click the "Create Class" button inside the modal
    const modalCreateButton = screen.getByTestId('create-class-modal').querySelector('button:last-child')
    await user.click(modalCreateButton)
    
    // Should refetch classes
    await waitFor(() => {
      expect(api.default.get).toHaveBeenCalledWith('/classes')
    })
  })

  it('selects class and fetches students', async () => {
    const user = userEvent.setup()
    const api = await import('../../services/api')
    render(<TeacherDashboard />)
    
    // Wait for classes to load
    await waitFor(() => {
      expect(screen.getByText('Math Class')).toBeInTheDocument()
    })
    
    // Click on a class
    const classCard = screen.getByText('Math Class').closest('div')
    await user.click(classCard)
    
    // Should fetch students for the selected class
    await waitFor(() => {
      expect(api.default.get).toHaveBeenCalledWith('/classes/1/students')
    })
    
    // Should display students
    expect(screen.getByText('Alice Johnson')).toBeInTheDocument()
    expect(screen.getByText('Bob Smith')).toBeInTheDocument()
  })

  it('displays class details when selected', async () => {
    const user = userEvent.setup()
    const api = await import('../../services/api')
    render(<TeacherDashboard />)
    
    // Wait for classes to load
    await waitFor(() => {
      expect(screen.getByText('Math Class')).toBeInTheDocument()
    })
    
    // Click on a class
    const classCard = screen.getByText('Math Class').closest('div')
    await user.click(classCard)
    
    // Should display class name and student count
    await waitFor(() => {
      expect(screen.getByText('Math Class - Students')).toBeInTheDocument()
      expect(screen.getByText('2 students')).toBeInTheDocument()
    })
  })

  it('displays empty state when class has no students', async () => {
    const user = userEvent.setup()
    const api = await import('../../services/api')
    api.default.get.mockImplementation((url) => {
      if (url === '/classes') {
        return Promise.resolve({ data: mockClasses })
      }
      if (url.includes('/students')) {
        return Promise.resolve({ data: [] })
      }
      return Promise.resolve({ data: [] })
    })
    
    render(<TeacherDashboard />)
    
    // Wait for classes to load and click on one
    await waitFor(() => {
      expect(screen.getByText('Math Class')).toBeInTheDocument()
    })
    
    const classCard = screen.getByText('Math Class').closest('div')
    await user.click(classCard)
    
    // Should display empty state
    await waitFor(() => {
      expect(screen.getByText('No students assigned')).toBeInTheDocument()
      expect(screen.getByText('Students can be assigned to this class by parents or admins.')).toBeInTheDocument()
    })
  })

  it('displays error message when API fails', async () => {
    const api = await import('../../services/api')
    api.default.get.mockRejectedValueOnce(new Error('API Error'))
    
    render(<TeacherDashboard />)
    
    await waitFor(() => {
      expect(screen.getByText('Failed to fetch classes')).toBeInTheDocument()
    })
  })

  it('highlights selected class', async () => {
    const user = userEvent.setup()
    const api = await import('../../services/api')
    render(<TeacherDashboard />)
    
    // Wait for classes to load
    await waitFor(() => {
      expect(screen.getByText('Math Class')).toBeInTheDocument()
    })
    
    // Find the class card by looking for the clickable element with correct classes
    const classCard = screen.getByText('Math Class').closest('.cursor-pointer')
    await user.click(classCard)
    
    // Wait for selection to take effect and check for selected styling
    await waitFor(() => {
      expect(classCard).toHaveClass('border-indigo-500', 'bg-indigo-50')
    })
  })

  it('shows class goals in class cards', async () => {
    render(<TeacherDashboard />)
    
    await waitFor(() => {
      expect(screen.getByText('Student Goal: 10')).toBeInTheDocument()
      expect(screen.getByText('Read-to Goal: 5')).toBeInTheDocument()
      expect(screen.getByText('Student Goal: 15')).toBeInTheDocument()
      expect(screen.getByText('Read-to Goal: 8')).toBeInTheDocument()
    })
  })
})