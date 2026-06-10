import { Navigate } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import Loading from './Loading'

const GuestRoute = ({ children }) => {
  const { loading, isAuthenticated, hasPermission } = useAuth()

  if (loading) return <Loading />
  if (isAuthenticated) {
    const canViewDashboard = hasPermission('dashboard', 'view')
    return <Navigate to={canViewDashboard ? '/dashboard' : '/profile'} replace />
  }

  return children
}

export default GuestRoute
