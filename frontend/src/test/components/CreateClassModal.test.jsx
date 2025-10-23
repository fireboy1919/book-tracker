import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import CreateClassModal from '../../components/CreateClassModal'

// Mock the API
vi.mock('../../services/api', () => ({
  default: {
    post: vi.fn()
  }
}))

describe('CreateClassModal', () => {
  const mockProps = {
    isOpen: true,
    onClose: vi.fn(),
    onSuccess: vi.fn()
  }

  beforeEach(async () => {
    vi.clearAllMocks()
    const api = await import('../../services/api')
    api.default.post.mockClear()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('renders when open', () => {
    render(<CreateClassModal {...mockProps} />)
    
    expect(screen.getByText('Create New Class')).toBeInTheDocument()
    expect(screen.getByLabelText(/class name/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/description/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/student books goal/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/books read to student goal/i)).toBeInTheDocument()
  })

  it('does not render when closed', () => {
    render(<CreateClassModal {...mockProps} isOpen={false} />)
    
    expect(screen.queryByText('Create New Class')).not.toBeInTheDocument()
  })

  it('calls onClose when cancel button is clicked', async () => {
    const user = userEvent.setup()
    render(<CreateClassModal {...mockProps} />)
    
    const cancelButton = screen.getByText('Cancel')
    await user.click(cancelButton)
    
    expect(mockProps.onClose).toHaveBeenCalledTimes(1)
  })

  it('calls onClose when X button is clicked', async () => {
    const user = userEvent.setup()
    render(<CreateClassModal {...mockProps} />)
    
    const closeButton = screen.getByRole('button', { name: '' }) // X button
    await user.click(closeButton)
    
    expect(mockProps.onClose).toHaveBeenCalledTimes(1)
  })

  it('submits form with correct data', async () => {
    const user = userEvent.setup()
    const api = await import('../../services/api')
    api.default.post.mockResolvedValueOnce({ data: { id: 1, name: 'Test Class' } })
    
    render(<CreateClassModal {...mockProps} />)
    
    // Fill out the form
    await user.type(screen.getByLabelText(/class name/i), 'Test Class')
    await user.type(screen.getByLabelText(/description/i), 'A test class description')
    await user.clear(screen.getByLabelText(/student books goal/i))
    await user.type(screen.getByLabelText(/student books goal/i), '10')
    await user.clear(screen.getByLabelText(/books read to student goal/i))
    await user.type(screen.getByLabelText(/books read to student goal/i), '5')
    
    // Submit the form
    const submitButton = screen.getByText('Create Class')
    await user.click(submitButton)
    
    // Wait for API call
    await waitFor(() => {
      expect(api.default.post).toHaveBeenCalledWith('/classes', {
        name: 'Test Class',
        description: 'A test class description',
        studentBooksGoal: 10,
        otherBooksGoal: 5
      })
    })
    
    expect(mockProps.onSuccess).toHaveBeenCalledTimes(1)
    expect(mockProps.onClose).toHaveBeenCalledTimes(1)
  })

  it('displays error message when API call fails', async () => {
    const user = userEvent.setup()
    const api = await import('../../services/api')
    const errorMessage = 'Failed to create class'
    api.default.post.mockRejectedValueOnce({
      response: { data: { message: errorMessage } }
    })
    
    render(<CreateClassModal {...mockProps} />)
    
    // Fill out required field
    await user.type(screen.getByLabelText(/class name/i), 'Test Class')
    
    // Submit the form
    const submitButton = screen.getByText('Create Class')
    await user.click(submitButton)
    
    // Wait for error message
    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument()
    })
    
    expect(mockProps.onSuccess).not.toHaveBeenCalled()
    expect(mockProps.onClose).not.toHaveBeenCalled()
  })

  it('prevents submission with empty name', async () => {
    const user = userEvent.setup()
    const api = await import('../../services/api')
    render(<CreateClassModal {...mockProps} />)
    
    // Try to submit without filling required field
    const submitButton = screen.getByText('Create Class')
    await user.click(submitButton)
    
    // API should not be called
    expect(api.default.post).not.toHaveBeenCalled()
  })

  it('shows loading state when submitting', async () => {
    const user = userEvent.setup()
    const api = await import('../../services/api')
    // Mock a delayed response
    api.default.post.mockImplementationOnce(() => new Promise(() => {}))
    
    render(<CreateClassModal {...mockProps} />)
    
    // Fill out required field
    await user.type(screen.getByLabelText(/class name/i), 'Test Class')
    
    // Submit the form
    const submitButton = screen.getByText('Create Class')
    await user.click(submitButton)
    
    // Check loading state
    expect(screen.getByText('Creating...')).toBeInTheDocument()
    expect(submitButton).toBeDisabled()
  })

  it('resets form after successful submission', async () => {
    const user = userEvent.setup()
    const api = await import('../../services/api')
    api.default.post.mockResolvedValueOnce({ data: { id: 1, name: 'Test Class' } })
    
    render(<CreateClassModal {...mockProps} />)
    
    const nameInput = screen.getByLabelText(/class name/i)
    const descriptionInput = screen.getByLabelText(/description/i)
    
    // Fill out the form
    await user.type(nameInput, 'Test Class')
    await user.type(descriptionInput, 'A test description')
    
    // Submit the form
    const submitButton = screen.getByText('Create Class')
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(mockProps.onSuccess).toHaveBeenCalledTimes(1)
    })
    
    // Form should be reset (checked by re-opening modal)
    render(<CreateClassModal {...mockProps} />)
    expect(screen.getByLabelText(/class name/i)).toHaveValue('')
    expect(screen.getByLabelText(/description/i)).toHaveValue('')
  })
})