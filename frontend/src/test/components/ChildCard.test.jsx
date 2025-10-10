import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import ChildCard from '../../components/ChildCard'

// Mock the API
vi.mock('../../services/api', () => ({
  default: {
    get: vi.fn(() => Promise.resolve({
      data: [
        {
          id: 1,
          title: 'The Cat in the Hat',
          author: 'Dr. Seuss',
          dateRead: '2024-01-01T00:00:00Z',
          childId: 1
        },
        {
          id: 2,
          title: 'Green Eggs and Ham',
          author: 'Dr. Seuss',
          dateRead: '2024-01-02T00:00:00Z',
          childId: 1
        }
      ]
    }))
  }
}))

describe('ChildCard', () => {
  const mockChild = {
    id: 1,
    name: 'Alice',
    age: 8,
    ownerId: 1,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z'
  }

  const mockProps = {
    child: mockChild,
    onAddBook: vi.fn(),
    onInviteUser: vi.fn()
  }

  it('renders child information', async () => {
    render(<ChildCard {...mockProps} />)

    expect(screen.getByText('Alice')).toBeInTheDocument()
    expect(screen.getByText('Age 8')).toBeInTheDocument()
    
    // Wait for the loading to complete to avoid act() warnings
    await waitFor(() => {
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument()
    })
  })

  it('loads and displays book count', async () => {
    render(<ChildCard {...mockProps} />)

    expect(screen.getByText('Loading...')).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByText('2 books read')).toBeInTheDocument()
    })
  })

  it('shows recent book when books are loaded', async () => {
    render(<ChildCard {...mockProps} />)

    await waitFor(() => {
      expect(screen.getByText('Recent: Green Eggs and Ham')).toBeInTheDocument()
    })
  })

  it('calls onAddBook when Add Book button is clicked', async () => {
    const user = userEvent.setup()
    
    render(<ChildCard {...mockProps} />)

    const addBookButton = screen.getByText('Add Book')
    await user.click(addBookButton)

    expect(mockProps.onAddBook).toHaveBeenCalledWith(mockChild)
  })

  it('calls onInviteUser when Share button is clicked', async () => {
    const user = userEvent.setup()
    
    render(<ChildCard {...mockProps} />)

    const shareButton = screen.getByText('Share')
    await user.click(shareButton)

    expect(mockProps.onInviteUser).toHaveBeenCalledWith(mockChild)
  })

  it('handles API error gracefully', async () => {
    // Mock API error
    const api = await import('../../services/api')
    api.default.get.mockRejectedValueOnce(new Error('API Error'))

    render(<ChildCard {...mockProps} />)

    await waitFor(() => {
      expect(screen.getByText('0 books read')).toBeInTheDocument()
    })
  })
})