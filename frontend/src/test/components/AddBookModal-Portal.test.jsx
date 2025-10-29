import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createPortal } from 'react-dom'
import AddBookModal from '../../components/AddBookModal'

// Mock the API
vi.mock('../../services/api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn()
  }
}))

// Mock Html5QrcodeScanner to avoid camera access issues in tests
vi.mock('html5-qrcode', () => ({
  Html5QrcodeScanner: vi.fn().mockImplementation(() => ({
    render: vi.fn(),
    clear: vi.fn()
  })),
  Html5QrcodeScanType: {
    SCAN_TYPE_CAMERA: 'camera'
  }
}))

describe('AddBookModal Portal Test', () => {
  const mockChild = {
    id: 1,
    firstName: 'John',
    lastName: 'Doe'
  }

  const mockOnClose = vi.fn()
  const mockOnBookAdded = vi.fn()

  beforeEach(async () => {
    vi.clearAllMocks()
    
    // Clear document.body for clean portal testing
    document.body.innerHTML = ''
    
    const api = await import('../../services/api')
    api.default.get.mockResolvedValue({ data: [] })
    api.default.post.mockResolvedValue({ data: { success: true } })
  })

  it('should render AddBookModal with correct z-index via portal', async () => {
    // Create a mock container to simulate the parent component
    const TestParentComponent = () => {
      return (
        <div>
          <div className="some-parent-content z-50">Parent Content</div>
          {createPortal(
            <AddBookModal
              child={mockChild}
              onClose={mockOnClose}
              onBookAdded={mockOnBookAdded}
            />,
            document.body
          )}
        </div>
      )
    }

    render(<TestParentComponent />)

    // Wait for the modal to be rendered
    await waitFor(() => {
      expect(screen.getByText('Add Book for John Doe')).toBeInTheDocument()
    })

    // Verify the modal is rendered in document.body (portal)
    const modalInBody = document.body.querySelector('[data-testid="add-book-modal"]')
    expect(modalInBody).toBeTruthy()

    // Verify it has the correct z-index class
    expect(modalInBody).toHaveClass('z-50')

    // Verify it's a direct child of document.body (not nested)
    expect(modalInBody.parentElement).toBe(document.body)
  })

  it('should render above other high z-index elements when using portal', async () => {
    // Create a scenario with multiple layered elements
    const TestWithMultipleLayers = () => {
      return (
        <div>
          {/* Simulate FullScreenChildView */}
          <div className="fixed inset-0 z-50 bg-gray-500" data-testid="background-modal">
            Background Modal Content
          </div>
          
          {/* Simulate other modals */}
          <div className="fixed inset-0 z-50 bg-blue-500" data-testid="other-modal">
            Other Modal
          </div>

          {/* AddBookModal via portal */}
          {createPortal(
            <AddBookModal
              child={mockChild}
              onClose={mockOnClose}
              onBookAdded={mockOnBookAdded}
            />,
            document.body
          )}
        </div>
      )
    }

    render(<TestWithMultipleLayers />)

    // Wait for all elements to render
    await waitFor(() => {
      expect(screen.getByText('Add Book for John Doe')).toBeInTheDocument()
    })

    // Get all the elements
    const backgroundModal = document.querySelector('[data-testid="background-modal"]')
    const otherModal = document.querySelector('[data-testid="other-modal"]')
    const addBookModal = document.querySelector('[data-testid="add-book-modal"]')

    // Verify they all exist
    expect(backgroundModal).toBeTruthy()
    expect(otherModal).toBeTruthy()
    expect(addBookModal).toBeTruthy()

    // Verify z-index hierarchy - All use z-50, but portal ensures AddBookModal appears last (on top)
    expect(backgroundModal).toHaveClass('z-50')
    expect(otherModal).toHaveClass('z-50')
    expect(addBookModal).toHaveClass('z-50')

    // Verify AddBookModal is rendered as direct child of body (portal behavior)
    expect(addBookModal.parentElement).toBe(document.body)
  })

  it('should handle closing the modal properly when using portal', async () => {
    const TestWithPortal = () => {
      return (
        <div>
          {createPortal(
            <AddBookModal
              child={mockChild}
              onClose={mockOnClose}
              onBookAdded={mockOnBookAdded}
            />,
            document.body
          )}
        </div>
      )
    }

    render(<TestWithPortal />)

    // Wait for modal to render
    await waitFor(() => {
      expect(screen.getByText('Add Book for John Doe')).toBeInTheDocument()
    })

    // Find and click the close button (the X button in the header)
    const closeButton = screen.getByRole('button', { name: 'Close modal' })
    fireEvent.click(closeButton)

    // Verify the onClose callback was called
    expect(mockOnClose).toHaveBeenCalledTimes(1)
  })

  it('should prevent stacking context issues by using portal', async () => {
    // This test verifies that the portal approach bypasses stacking context issues
    const TestWithStackingContext = () => {
      return (
        <div>
          {/* Parent with transform creates stacking context */}
          <div className="transform translate-x-0 z-50" data-testid="stacking-context-parent">
            <div className="fixed inset-0 z-999 bg-red-500" data-testid="trapped-modal">
              This would be trapped by stacking context
            </div>
          </div>

          {/* Portal bypasses stacking context */}
          {createPortal(
            <AddBookModal
              child={mockChild}
              onClose={mockOnClose}
              onBookAdded={mockOnBookAdded}
            />,
            document.body
          )}
        </div>
      )
    }

    render(<TestWithStackingContext />)

    await waitFor(() => {
      expect(screen.getByText('Add Book for John Doe')).toBeInTheDocument()
    })

    const trappedModal = document.querySelector('[data-testid="trapped-modal"]')
    const addBookModal = document.querySelector('[data-testid="add-book-modal"]')
    const stackingParent = document.querySelector('[data-testid="stacking-context-parent"]')

    // Verify the trapped modal is within the stacking context
    expect(trappedModal.closest('[data-testid="stacking-context-parent"]')).toBe(stackingParent)

    // Verify AddBookModal is NOT within the stacking context (escaped via portal)
    expect(addBookModal.closest('[data-testid="stacking-context-parent"]')).toBeNull()
    expect(addBookModal.parentElement).toBe(document.body)
  })
})