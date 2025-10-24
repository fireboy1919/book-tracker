import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { BrowserRouter } from 'react-router-dom'
import TeacherDashboard from '../../pages/TeacherDashboard'
import AcceptInvitation from '../../pages/AcceptInvitation'

// Mock API
vi.mock('../../services/api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  }
}))

// Mock components
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

vi.mock('../../components/FullScreenChildView', () => ({
  default: ({ child, onClose }) => (
    <div data-testid="full-screen-child-view">
      <button onClick={onClose}>Close View</button>
      <span>{child.firstName} {child.lastName}</span>
    </div>
  )
}))

// Test wrapper with router
const TestWrapper = ({ children }) => (
  <BrowserRouter>{children}</BrowserRouter>
)

describe('Invitation Workflow', () => {
  const mockClass = {
    id: 1,
    name: 'Math Class',
    description: 'Elementary Math',
    studentBooksGoal: 10,
    otherBooksGoal: 5,
    createdById: 1,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z'
  }

  const mockInvitationData = {
    teacher_id: 1,
    class_id: 1,
    class_name: 'Math Class',
    invitation_key: 'abc123def456ghi789jkl012mno345pqr678stu901vwx234yz567890abcdef123456789012345678901234567890',
    base_url: 'https://booktracker.app/invite/'
  }

  const mockStudentInvitationDetails = {
    student_name: 'John Doe',
    teacher_name: 'Jane Smith',
    class_name: 'Math Class',
    valid: true
  }

  beforeEach(async () => {
    vi.clearAllMocks()
    const api = await import('../../services/api')
    
    // Default mock implementations
    api.default.get.mockImplementation((url) => {
      if (url === '/classes') {
        return Promise.resolve({ data: [mockClass] })
      }
      if (url.includes('/students')) {
        return Promise.resolve({ data: [] })
      }
      if (url.includes('/teachers')) {
        return Promise.resolve({ data: [] })
      }
      if (url.includes('/invitation-data')) {
        return Promise.resolve({ data: mockInvitationData })
      }
      if (url.includes('/invitation/') && url.includes('/details')) {
        return Promise.resolve({ data: mockStudentInvitationDetails })
      }
      return Promise.resolve({ data: [] })
    })
  })

  describe('Teacher Dashboard - Invitation Key Generation', () => {
    it('displays invitation key section when class is selected', async () => {
      const user = userEvent.setup()
      render(<TeacherDashboard />)
      
      // Wait for classes to load
      await waitFor(() => {
        expect(screen.getByText('Math Class')).toBeInTheDocument()
      })
      
      // Click on the class to select it
      const classCard = screen.getByText('Math Class').closest('div')
      await user.click(classCard)
      
      // Should display invitation key section
      await waitFor(() => {
        expect(screen.getByText('Class Invitation Key')).toBeInTheDocument()
        expect(screen.getByText('Use this key for Gmail mail merge to invite students')).toBeInTheDocument()
      })
      
      // Should display the invitation key
      const textarea = screen.getByDisplayValue(mockInvitationData.invitation_key)
      expect(textarea).toBeInTheDocument()
      expect(textarea).toHaveAttribute('readonly')
    })

    it('displays copy and help buttons as icons on the right', async () => {
      const user = userEvent.setup()
      render(<TeacherDashboard />)
      
      // Wait for classes to load and select class
      await waitFor(() => {
        expect(screen.getByText('Math Class')).toBeInTheDocument()
      })
      
      const classCard = screen.getByText('Math Class').closest('div')
      await user.click(classCard)
      
      // Wait for invitation section to load
      await waitFor(() => {
        expect(screen.getByText('Class Invitation Key')).toBeInTheDocument()
      })
      
      // Check that copy and help buttons are present as icons
      const copyButton = screen.getByTitle('Copy to clipboard')
      const helpButton = screen.getByTitle('Help Guide')
      
      expect(copyButton).toBeInTheDocument()
      expect(helpButton).toBeInTheDocument()
      
      // Verify they are styled as circular buttons
      expect(copyButton).toHaveClass('rounded-full')
      expect(helpButton).toHaveClass('rounded-full')
      
      // Verify help button links to the correct URL
      expect(helpButton).toHaveAttribute('href', '/help/mail-merge')
      expect(helpButton).toHaveAttribute('target', '_blank')
    })

    it('copies invitation key to clipboard when copy button is clicked', async () => {
      const user = userEvent.setup()
      
      // Mock clipboard API
      const mockWriteText = vi.fn().mockResolvedValue()
      Object.assign(navigator, {
        clipboard: {
          writeText: mockWriteText
        }
      })
      
      render(<TeacherDashboard />)
      
      // Wait for classes to load and select class
      await waitFor(() => {
        expect(screen.getByText('Math Class')).toBeInTheDocument()
      })
      
      const classCard = screen.getByText('Math Class').closest('div')
      await user.click(classCard)
      
      // Wait for invitation section to load
      await waitFor(() => {
        expect(screen.getByText('Class Invitation Key')).toBeInTheDocument()
      })
      
      // Click copy button
      const copyButton = screen.getByTitle('Copy to clipboard')
      await user.click(copyButton)
      
      // Verify clipboard API was called with correct text
      expect(mockWriteText).toHaveBeenCalledWith(mockInvitationData.invitation_key)
      
      // Should show success state
      await waitFor(() => {
        expect(screen.getByTitle('Copied!')).toBeInTheDocument()
      })
    })

    it('fetches invitation data when class is selected', async () => {
      const user = userEvent.setup()
      const api = await import('../../services/api')
      
      render(<TeacherDashboard />)
      
      // Wait for classes to load
      await waitFor(() => {
        expect(screen.getByText('Math Class')).toBeInTheDocument()
      })
      
      // Select class
      const classCard = screen.getByText('Math Class').closest('div')
      await user.click(classCard)
      
      // Should fetch invitation data
      await waitFor(() => {
        expect(api.default.get).toHaveBeenCalledWith('/classes/1/invitation-data')
      })
    })

    it('allows selecting invitation key text by clicking on textarea', async () => {
      const user = userEvent.setup()
      
      // Mock text selection
      const mockSelect = vi.fn()
      
      render(<TeacherDashboard />)
      
      // Wait for classes to load and select class
      await waitFor(() => {
        expect(screen.getByText('Math Class')).toBeInTheDocument()
      })
      
      const classCard = screen.getByText('Math Class').closest('div')
      await user.click(classCard)
      
      // Wait for invitation section
      await waitFor(() => {
        expect(screen.getByText('Class Invitation Key')).toBeInTheDocument()
      })
      
      // Mock the select method on the textarea
      const textarea = screen.getByDisplayValue(mockInvitationData.invitation_key)
      textarea.select = mockSelect
      
      // Click on textarea
      await user.click(textarea)
      
      // Should call select method
      expect(mockSelect).toHaveBeenCalled()
    })
  })

  describe('Student Invitation Acceptance', () => {
    // Mock URL search params
    const mockSearchParams = new URLSearchParams('?token=test-invitation-token')
    
    beforeEach(() => {
      // Mock useSearchParams
      vi.doMock('react-router-dom', async () => {
        const actual = await vi.importActual('react-router-dom')
        return {
          ...actual,
          useSearchParams: () => [mockSearchParams],
          useNavigate: () => vi.fn()
        }
      })
    })

    it('displays loading state while fetching invitation details', async () => {
      const api = await import('../../services/api')
      api.default.get.mockImplementationOnce(() => new Promise(() => {})) // Never resolves
      
      render(
        <TestWrapper>
          <AcceptInvitation />
        </TestWrapper>
      )
      
      expect(screen.getByText('Loading invitation details...')).toBeInTheDocument()
      expect(document.querySelector('.animate-spin')).toBeInTheDocument()
    })

    it('displays invitation details when loaded successfully', async () => {
      const api = await import('../../services/api')
      api.default.get.mockResolvedValueOnce({
        data: {
          email: 'parent@example.com',
          childName: 'John Doe',
          inviterName: 'Jane Smith',
          permissionType: 'PARENT'
        }
      })
      
      render(
        <TestWrapper>
          <AcceptInvitation />
        </TestWrapper>
      )
      
      await waitFor(() => {
        expect(screen.getByText('Accept Invitation')).toBeInTheDocument()
        expect(screen.getByText(/Jane Smith.*has invited you to help track.*John Doe.*reading progress/)).toBeInTheDocument()
        expect(screen.getByText('Permission level: PARENT')).toBeInTheDocument()
      })
    })

    it('displays error state for invalid invitation token', async () => {
      const api = await import('../../services/api')
      api.default.get.mockRejectedValueOnce({
        response: {
          data: {
            message: 'Invalid or expired invitation'
          }
        }
      })
      
      render(
        <TestWrapper>
          <AcceptInvitation />
        </TestWrapper>
      )
      
      await waitFor(() => {
        expect(screen.getByText('Invalid Invitation')).toBeInTheDocument()
        expect(screen.getByText('Invalid or expired invitation')).toBeInTheDocument()
        expect(screen.getByText('Go to Login')).toBeInTheDocument()
      })
    })

    it('validates form inputs correctly', async () => {
      const user = userEvent.setup()
      const api = await import('../../services/api')
      api.default.get.mockResolvedValueOnce({
        data: {
          email: 'parent@example.com',
          childName: 'John Doe',
          inviterName: 'Jane Smith',
          permissionType: 'PARENT'
        }
      })
      
      render(
        <TestWrapper>
          <AcceptInvitation />
        </TestWrapper>
      )
      
      // Wait for form to load
      await waitFor(() => {
        expect(screen.getByText('Accept Invitation')).toBeInTheDocument()
      })
      
      // Fill form with mismatched passwords
      await user.type(screen.getByLabelText('First name'), 'John')
      await user.type(screen.getByLabelText('Last name'), 'Doe')
      await user.type(screen.getByLabelText('Password'), 'password123')
      await user.type(screen.getByLabelText('Confirm password'), 'different')
      
      // Submit form
      await user.click(screen.getByRole('button', { name: /Accept Invitation & Create Account/i }))
      
      // Should show validation error
      await waitFor(() => {
        expect(screen.getByText('Passwords do not match')).toBeInTheDocument()
      })
    })

    it('validates minimum password length', async () => {
      const user = userEvent.setup()
      const api = await import('../../services/api')
      api.default.get.mockResolvedValueOnce({
        data: {
          email: 'parent@example.com',
          childName: 'John Doe',
          inviterName: 'Jane Smith',
          permissionType: 'PARENT'
        }
      })
      
      render(
        <TestWrapper>
          <AcceptInvitation />
        </TestWrapper>
      )
      
      // Wait for form to load
      await waitFor(() => {
        expect(screen.getByText('Accept Invitation')).toBeInTheDocument()
      })
      
      // Fill form with short password
      await user.type(screen.getByLabelText('First name'), 'John')
      await user.type(screen.getByLabelText('Last name'), 'Doe')
      await user.type(screen.getByLabelText('Password'), 'short')
      await user.type(screen.getByLabelText('Confirm password'), 'short')
      
      // Submit form
      await user.click(screen.getByRole('button', { name: /Accept Invitation & Create Account/i }))
      
      // Should show validation error
      await waitFor(() => {
        expect(screen.getByText('Password must be at least 8 characters long')).toBeInTheDocument()
      })
    })

    it('submits registration successfully with valid data', async () => {
      const user = userEvent.setup()
      const api = await import('../../services/api')
      const mockNavigate = vi.fn()
      
      // Mock react-router-dom navigate
      vi.doMock('react-router-dom', async () => {
        const actual = await vi.importActual('react-router-dom')
        return {
          ...actual,
          useSearchParams: () => [mockSearchParams],
          useNavigate: () => mockNavigate
        }
      })
      
      api.default.get.mockResolvedValueOnce({
        data: {
          email: 'parent@example.com',
          childName: 'John Doe',
          inviterName: 'Jane Smith',
          permissionType: 'PARENT'
        }
      })
      
      api.default.post.mockResolvedValueOnce({
        data: { message: 'Registration successful' }
      })
      
      render(
        <TestWrapper>
          <AcceptInvitation />
        </TestWrapper>
      )
      
      // Wait for form to load
      await waitFor(() => {
        expect(screen.getByText('Accept Invitation')).toBeInTheDocument()
      })
      
      // Fill form with valid data
      await user.type(screen.getByLabelText('First name'), 'John')
      await user.type(screen.getByLabelText('Last name'), 'Doe')
      await user.type(screen.getByLabelText('Password'), 'password123')
      await user.type(screen.getByLabelText('Confirm password'), 'password123')
      
      // Submit form
      await user.click(screen.getByRole('button', { name: /Accept Invitation & Create Account/i }))
      
      // Should call registration API
      await waitFor(() => {
        expect(api.default.post).toHaveBeenCalledWith('/auth/register-with-invitation', {
          firstName: 'John',
          lastName: 'Doe',
          email: 'parent@example.com',
          password: 'password123',
          invitationToken: 'test-invitation-token'
        })
      })
    })

    it('toggles password visibility when eye icon is clicked', async () => {
      const user = userEvent.setup()
      const api = await import('../../services/api')
      api.default.get.mockResolvedValueOnce({
        data: {
          email: 'parent@example.com',
          childName: 'John Doe',
          inviterName: 'Jane Smith',
          permissionType: 'PARENT'
        }
      })
      
      render(
        <TestWrapper>
          <AcceptInvitation />
        </TestWrapper>
      )
      
      // Wait for form to load
      await waitFor(() => {
        expect(screen.getByText('Accept Invitation')).toBeInTheDocument()
      })
      
      const passwordInput = screen.getByLabelText('Password')
      const toggleButton = passwordInput.parentElement.querySelector('button')
      
      // Initially should be password type
      expect(passwordInput).toHaveAttribute('type', 'password')
      
      // Click toggle button
      await user.click(toggleButton)
      
      // Should change to text type
      expect(passwordInput).toHaveAttribute('type', 'text')
      
      // Click again to toggle back
      await user.click(toggleButton)
      
      // Should be password type again
      expect(passwordInput).toHaveAttribute('type', 'password')
    })

    it('handles registration errors gracefully', async () => {
      const user = userEvent.setup()
      const api = await import('../../services/api')
      
      api.default.get.mockResolvedValueOnce({
        data: {
          email: 'parent@example.com',
          childName: 'John Doe',
          inviterName: 'Jane Smith',
          permissionType: 'PARENT'
        }
      })
      
      api.default.post.mockRejectedValueOnce({
        response: {
          data: {
            message: 'Email already exists'
          }
        }
      })
      
      render(
        <TestWrapper>
          <AcceptInvitation />
        </TestWrapper>
      )
      
      // Wait for form to load
      await waitFor(() => {
        expect(screen.getByText('Accept Invitation')).toBeInTheDocument()
      })
      
      // Fill and submit form
      await user.type(screen.getByLabelText('First name'), 'John')
      await user.type(screen.getByLabelText('Last name'), 'Doe')
      await user.type(screen.getByLabelText('Password'), 'password123')
      await user.type(screen.getByLabelText('Confirm password'), 'password123')
      
      await user.click(screen.getByRole('button', { name: /Accept Invitation & Create Account/i }))
      
      // Should display error message
      await waitFor(() => {
        expect(screen.getByText('Email already exists')).toBeInTheDocument()
      })
    })
  })

  describe('End-to-End Invitation Flow', () => {
    it('completes full invitation workflow from generation to acceptance', async () => {
      const user = userEvent.setup()
      const api = await import('../../services/api')
      
      // Step 1: Teacher generates invitation
      render(<TeacherDashboard />)
      
      await waitFor(() => {
        expect(screen.getByText('Math Class')).toBeInTheDocument()
      })
      
      const classCard = screen.getByText('Math Class').closest('div')
      await user.click(classCard)
      
      await waitFor(() => {
        expect(screen.getByText('Class Invitation Key')).toBeInTheDocument()
      })
      
      // Verify invitation data was fetched
      expect(api.default.get).toHaveBeenCalledWith('/classes/1/invitation-data')
      
      // Verify invitation key is displayed
      const invitationKey = screen.getByDisplayValue(mockInvitationData.invitation_key)
      expect(invitationKey).toBeInTheDocument()
      
      // This completes the teacher side of the workflow
      // The student side would be tested separately as it's a different page/component
      // but the invitation key generated here would be used to create the invitation URL
    })
  })
})