import { useAuth } from '../contexts/AuthContext'

export const usePermissions = () => {
  const { hasPermission, user } = useAuth()
  return {
    user,
    can: hasPermission,
    isSuperAdmin: user?.role === 'superadmin',
  }
}

export default usePermissions
