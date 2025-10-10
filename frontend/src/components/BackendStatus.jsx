import { useState, useEffect } from 'react'
import { ExclamationTriangleIcon, CheckCircleIcon } from '@heroicons/react/24/outline'

export default function BackendStatus() {
  const [status, setStatus] = useState('checking') // checking, online, waking, offline
  const [retryCount, setRetryCount] = useState(0)

  const checkBackendHealth = async () => {
    try {
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 5000) // 5s timeout

      const response = await fetch(`${import.meta.env.VITE_API_URL}/health`, {
        signal: controller.signal
      })
      
      clearTimeout(timeoutId)
      
      if (response.ok) {
        setStatus('online')
        setRetryCount(0)
        return true
      }
    } catch (error) {
      if (error.name === 'AbortError') {
        // Timeout - backend is probably waking up
        setStatus('waking')
      } else {
        setStatus('offline')
      }
    }
    return false
  }

  useEffect(() => {
    let intervalId

    const startHealthCheck = async () => {
      const isOnline = await checkBackendHealth()
      
      if (!isOnline && retryCount < 10) {
        setRetryCount(prev => prev + 1)
        // Retry with exponential backoff, max 5 seconds
        const delay = Math.min(1000 * Math.pow(1.5, retryCount), 5000)
        intervalId = setTimeout(startHealthCheck, delay)
      }
    }

    startHealthCheck()

    return () => {
      if (intervalId) clearTimeout(intervalId)
    }
  }, [retryCount])

  if (status === 'online') {
    return null // Don't show anything when backend is working
  }

  return (
    <div className="fixed top-0 left-0 right-0 z-50 bg-yellow-50 border-b border-yellow-200">
      <div className="max-w-7xl mx-auto py-2 px-3 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between flex-wrap">
          <div className="w-0 flex-1 flex items-center">
            {status === 'waking' ? (
              <>
                <span className="flex p-2 rounded-lg bg-yellow-400">
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                </span>
                <p className="ml-3 font-medium text-yellow-800 truncate">
                  Backend is waking up... Please wait a moment.
                </p>
              </>
            ) : status === 'offline' ? (
              <>
                <span className="flex p-2 rounded-lg bg-red-400">
                  <ExclamationTriangleIcon className="h-4 w-4 text-white" />
                </span>
                <p className="ml-3 font-medium text-red-800 truncate">
                  Backend is temporarily unavailable. Retrying...
                </p>
              </>
            ) : (
              <>
                <span className="flex p-2 rounded-lg bg-blue-400">
                  <div className="animate-pulse rounded-full h-4 w-4 bg-white"></div>
                </span>
                <p className="ml-3 font-medium text-blue-800 truncate">
                  Checking backend status...
                </p>
              </>
            )}
          </div>
          {status === 'waking' && (
            <div className="text-xs text-yellow-600">
              Attempt {retryCount}/10
            </div>
          )}
        </div>
      </div>
    </div>
  )
}