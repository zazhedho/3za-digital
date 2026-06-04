import { Navigate } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import Loading from './Loading'

const PermissionRoute = ({ resource, action, children }) => {
  const { loading, hasPermission } = useAuth()

  if (loading) return <Loading />
  if (!hasPermission(resource, action)) return <Navigate to="/unauthorized" replace />

  return children
}

export default PermissionRoute
