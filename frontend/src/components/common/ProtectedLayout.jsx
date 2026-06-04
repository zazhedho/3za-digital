import { Navigate } from 'react-router-dom'
import DashboardLayout from './DashboardLayout'
import Loading from './Loading'
import { useAuth } from '../../contexts/AuthContext'

const ProtectedLayout = () => {
  const { loading, isAuthenticated } = useAuth()

  if (loading) return <Loading />
  if (!isAuthenticated) return <Navigate to="/login" replace />

  return <DashboardLayout />
}

export default ProtectedLayout
