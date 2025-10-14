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
          dateRead: '2025-10-01T00:00:00Z',
          childId: 1
        },
        {
          id: 2,
          title: 'Green Eggs and Ham',
          author: 'Dr. Seuss',
          dateRead: '2025-10-02T00:00:00Z',
          childId: 1
        }
      ]
    }))
  }
}))

describe('ChildCard', () => {
  const mockChild = {
    id: 1,
    firstName: 'Alice',
    lastName: 'Smith',
    grade: '3rd Grade',
    ownerId: 1,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z'
  }

  const mockProps = {
    child: mockChild,
    onAddBook: vi.fn(),
    onViewDetails: vi.fn(),
    onEditChild: vi.fn(),
    currentMonth: new Date('2025-10-01')
  }

  it('renders child information', async () => {
    render(<ChildCard {...mockProps} />)

    expect(screen.getByText('Alice Smith')).toBeInTheDocument()
    expect(screen.getByText('3rd Grade')).toBeInTheDocument()
    
    // Wait for the loading to complete to avoid act() warnings
    await waitFor(() => {
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument()
    })
  })

  it('loads and displays book count', async () => {
    render(<ChildCard {...mockProps} />)

    expect(screen.getByText('Loading...')).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByText('2 books this month')).toBeInTheDocument()
    })
  })

  it('calls onAddBook when Add Book button is clicked', async () => {
    const user = userEvent.setup()
    
    render(<ChildCard {...mockProps} />)

    const addBookButton = screen.getByText('Add Book')
    await user.click(addBookButton)

    expect(mockProps.onAddBook).toHaveBeenCalledWith(mockChild)
  })

  it('calls onViewDetails when View All button is clicked', async () => {
    const user = userEvent.setup()
    
    render(<ChildCard {...mockProps} />)

    const viewButton = screen.getByText('View All')
    await user.click(viewButton)

    expect(mockProps.onViewDetails).toHaveBeenCalledWith(mockChild)
  })

  it('calls onEditChild when pencil icon is clicked', async () => {
    const user = userEvent.setup()
    
    render(<ChildCard {...mockProps} />)

    const editButton = screen.getByTitle('Edit child information')
    await user.click(editButton)

    expect(mockProps.onEditChild).toHaveBeenCalledWith(mockChild)
  })

  it('handles API error gracefully', async () => {
    // Mock API error
    const api = await import('../../services/api')
    api.default.get.mockRejectedValueOnce(new Error('API Error'))

    render(<ChildCard {...mockProps} />)

    await waitFor(() => {
      expect(screen.getByText('0 books this month')).toBeInTheDocument()
    })
  })
})