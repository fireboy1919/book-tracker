import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ChildCard from '../../components/ChildCard'
import api from '../../services/api'

// Mock the API module
vi.mock('../../services/api')

describe('ChildCard Refresh Functionality', () => {
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
    onChildUpdate: vi.fn(),
    currentMonth: new Date('2025-10-01'),
    currentUser: { id: 1, isTeacher: false },
    refreshTrigger: 0
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches books when component mounts', async () => {
    api.get.mockResolvedValue({
      data: [
        {
          id: 1,
          title: 'Test Book',
          author: 'Test Author',
          dateRead: '2025-10-01',
          childId: 1
        }
      ]
    })

    render(<ChildCard {...mockProps} />)

    await waitFor(() => {
      expect(api.get).toHaveBeenCalledWith('/books/child/1')
    })

    await waitFor(() => {
      expect(screen.getByText('1 books this month')).toBeInTheDocument()
    })
  })

  it('refetches books when refreshTrigger changes', async () => {
    // Initial render
    api.get.mockResolvedValueOnce({
      data: [
        {
          id: 1,
          title: 'Initial Book',
          author: 'Initial Author',
          dateRead: '2025-10-01',
          childId: 1
        }
      ]
    })

    const { rerender } = render(<ChildCard {...mockProps} />)

    // Wait for initial fetch
    await waitFor(() => {
      expect(api.get).toHaveBeenCalledTimes(3) // books, classes, current class
    })

    // Mock new API response with additional book
    api.get.mockResolvedValueOnce({
      data: [
        {
          id: 1,
          title: 'Initial Book',
          author: 'Initial Author',
          dateRead: '2025-10-01',
          childId: 1
        },
        {
          id: 2,
          title: 'New Book',
          author: 'New Author',
          dateRead: '2025-10-02',
          childId: 1
        }
      ]
    })

    // Re-render with updated refreshTrigger
    rerender(<ChildCard {...mockProps} refreshTrigger={1} />)

    // Should fetch books again
    await waitFor(() => {
      expect(api.get).toHaveBeenCalledWith('/books/child/1')
    })

    await waitFor(() => {
      expect(screen.getByText('2 books this month')).toBeInTheDocument()
    })
  })

  it('filters books correctly by current month', async () => {
    const currentMonth = new Date('2025-10-01')
    
    api.get.mockResolvedValue({
      data: [
        {
          id: 1,
          title: 'October Book',
          author: 'Author A',
          dateRead: '2025-10-15', // October - should be included
          childId: 1
        },
        {
          id: 2,
          title: 'September Book',
          author: 'Author B',
          dateRead: '2025-09-15', // September - should be excluded
          childId: 1
        },
        {
          id: 3,
          title: 'Another October Book',
          author: 'Author C',
          dateRead: '2025-10-25', // October - should be included
          childId: 1
        }
      ]
    })

    render(<ChildCard {...mockProps} currentMonth={currentMonth} />)

    await waitFor(() => {
      // Should only count books from October (2 books)
      expect(screen.getByText('2 books this month')).toBeInTheDocument()
    })
  })

  it('shows loading state during refresh', async () => {
    // Make API call hang
    const hangingPromise = new Promise(() => {}) // Never resolves
    api.get.mockReturnValue(hangingPromise)

    render(<ChildCard {...mockProps} />)

    // Should show loading state
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('handles API error gracefully during refresh', async () => {
    api.get.mockRejectedValue(new Error('Network error'))

    render(<ChildCard {...mockProps} />)

    await waitFor(() => {
      expect(screen.getByText('0 books this month')).toBeInTheDocument()
    })
  })
})