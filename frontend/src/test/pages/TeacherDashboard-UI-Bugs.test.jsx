import { render, screen, waitFor } from '@testing-library/react'
import { vi } from 'vitest'
import TeacherDashboard from '../../pages/TeacherDashboard'
import api from '../../services/api'

// Mock the API
vi.mock('../../services/api')

// Mock the icons
vi.mock('@heroicons/react/24/outline', () => ({
  PlusIcon: () => <div data-testid="plus-icon" />,
  UsersIcon: () => <div data-testid="users-icon" />,
  BookOpenIcon: () => <div data-testid="book-open-icon" />,
  UserPlusIcon: () => <div data-testid="user-plus-icon" />,
  TrashIcon: () => <div data-testid="trash-icon" />,
  ClipboardDocumentIcon: () => <div data-testid="clipboard-icon" />,
  QuestionMarkCircleIcon: () => <div data-testid="question-icon" />
}))

vi.mock('@heroicons/react/24/solid', () => ({
  CheckIcon: () => <div data-testid="check-icon" />
}))

// Mock other components
vi.mock('../../components/CreateClassModal', () => ({
  default: () => <div data-testid="create-class-modal" />
}))

vi.mock('../../components/FullScreenChildView', () => ({
  default: () => <div data-testid="full-screen-view" />
}))

describe('TeacherDashboard UI Bugs', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    
    // Mock API responses
    api.get.mockImplementation((url) => {
      if (url === '/classes') {
        return Promise.resolve({ 
          data: [{
            id: 1,
            name: 'Test Class',
            description: 'Test Description',
            studentBooksGoal: 10,
            otherBooksGoal: 5
          }]
        })
      }
      if (url.includes('/students')) {
        return Promise.resolve({
          data: [{
            id: 1,
            firstName: 'John',
            lastName: 'Doe',
            studentBooksRead: 5,
            studentGoal: 10,
            readToBooksRead: 3,
            readToGoal: 5,
            goalsReached: false
          }]
        })
      }
      if (url.includes('/teachers')) {
        return Promise.resolve({ data: [] })
      }
      if (url.includes('/invitation-data')) {
        return Promise.resolve({
          data: { invitation_key: 'test-invitation-key-123' }
        })
      }
      return Promise.resolve({ data: [] })
    })
  })

  test('should show hand cursor when hovering over student row (except on trash button)', async () => {
    render(<TeacherDashboard />)
    
    // Wait for data to load
    await waitFor(() => {
      expect(screen.getByText('Test Class')).toBeInTheDocument()
    })
    
    // Click on class to show students
    screen.getByText('Test Class').click()
    
    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })
    
    // Find the student row
    const studentRow = screen.getByText('John Doe').closest('tr')
    
    // Check that the row has cursor-pointer class
    expect(studentRow).toHaveClass('cursor-pointer')
    
    // The row should be clickable (has onClick handler)
    expect(studentRow).toHaveAttribute('class')
    const classes = studentRow.getAttribute('class')
    expect(classes).toContain('cursor-pointer')
  })

  test('should show visible copy button with proper background color', async () => {
    render(<TeacherDashboard />)
    
    // Wait for data to load and click on class
    await waitFor(() => {
      expect(screen.getByText('Test Class')).toBeInTheDocument()
    })
    
    screen.getByText('Test Class').click()
    
    // Wait for invitation data to load
    await waitFor(() => {
      expect(screen.getByDisplayValue('test-invitation-key-123')).toBeInTheDocument()
    })
    
    // Find the copy button by its icon
    const copyButton = screen.getByTestId('clipboard-icon').closest('button')
    
    // Check that the button has proper background color classes
    expect(copyButton).toHaveClass('bg-blue-600')
    expect(copyButton).toHaveClass('text-white')
    
    // The button should be visible (not white text on white background)
    const buttonClasses = copyButton.getAttribute('class')
    expect(buttonClasses).toContain('bg-blue-600')
    expect(buttonClasses).toContain('text-white')
    
    // Should not have white background that would make white text invisible
    expect(buttonClasses).not.toContain('bg-white')
  })

  test('should show 100% emoji for completed goals instead of checkmark', async () => {
    // Mock a student with completed goals  
    api.get.mockImplementation((url) => {
      if (url === '/classes') {
        return Promise.resolve({ 
          data: [{
            id: 1,
            name: 'Test Class',
            description: 'Test Description',
            studentBooksGoal: 10,
            otherBooksGoal: 5
          }]
        })
      }
      if (url.includes('/students')) {
        return Promise.resolve({
          data: [{
            id: 1,
            firstName: 'Jane',
            lastName: 'Smith',
            studentBooksRead: 10,
            studentGoal: 10,
            readToBooksRead: 5,
            readToGoal: 5,
            goalsReached: true // This student completed goals
          }]
        })
      }
      if (url.includes('/teachers')) {
        return Promise.resolve({ data: [] })
      }
      if (url.includes('/invitation-data')) {
        return Promise.resolve({
          data: { invitation_key: 'test-invitation-key-123' }
        })
      }
      return Promise.resolve({ data: [] })
    })

    render(<TeacherDashboard />)
    
    // Wait for data to load and click on class
    await waitFor(() => {
      expect(screen.getByText('Test Class')).toBeInTheDocument()
    })
    
    screen.getByText('Test Class').click()
    
    await waitFor(() => {
      expect(screen.getByText('Jane Smith')).toBeInTheDocument()
    })
    
    // Should show 100% emoji, not checkmark icon
    expect(screen.getByText('💯')).toBeInTheDocument()
    expect(screen.queryByTestId('check-icon')).not.toBeInTheDocument()
  })

  test('should not show dash for incomplete status', async () => {
    render(<TeacherDashboard />)
    
    // Wait for data to load and click on class
    await waitFor(() => {
      expect(screen.getByText('Test Class')).toBeInTheDocument()
    })
    
    screen.getByText('Test Class').click()
    
    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })
    
    // Should not show dash for incomplete status
    expect(screen.queryByText('-')).not.toBeInTheDocument()
  })
})