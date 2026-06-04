/* eslint-disable react-refresh/only-export-components */
import { createContext, useCallback, useContext, useEffect, useState } from 'react'
import api, { getErrorMessage } from '../services/api'
import authService from '../services/authService'
import permissionService from '../services/permissionService'

const AuthContext = createContext()

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuth must be used within AuthProvider')
  return context
}

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null)
  const [permissions, setPermissions] = useState([])
  const [loading, setLoading] = useState(true)
  const [token, setToken] = useState(localStorage.getItem('token'))

  const clearSession = useCallback(() => {
    localStorage.removeItem('token')
    localStorage.removeItem('refresh_token')
    setToken(null)
    setUser(null)
    setPermissions([])
    delete api.defaults.headers.common.Authorization
  }, [])

  const fetchUser = useCallback(async () => {
    try {
      const response = await authService.me()
      const fetchedUser = response.data.data
      setUser(fetchedUser)

      let fetchedPermissions = []
      try {
        const permissionResponse = await permissionService.getUserPermissions()
        fetchedPermissions = permissionResponse.data.data || []
        setPermissions(fetchedPermissions)
      } catch {
        setPermissions([])
      }
      return { user: fetchedUser, permissions: fetchedPermissions }
    } catch (error) {
      if ([401, 403].includes(error.response?.status)) {
        clearSession()
      }
      return null
    } finally {
      setLoading(false)
    }
  }, [clearSession])

  useEffect(() => {
    if (!token) {
      setLoading(false)
      return
    }
    api.defaults.headers.common.Authorization = `Bearer ${token}`
    fetchUser()
  }, [fetchUser, token])

  const login = async (email, password) => {
    try {
      const response = await authService.login({ email, password })
      const authData = response.data.data || {}
      const nextToken = authData.access_token || authData.token
      const refreshToken = authData.refresh_token

      if (!nextToken) {
        return { success: false, error: 'Login response missing access token' }
      }

      localStorage.setItem('token', nextToken)
      if (refreshToken) localStorage.setItem('refresh_token', refreshToken)
      setToken(nextToken)
      api.defaults.headers.common.Authorization = `Bearer ${nextToken}`
      const fetchedAuth = await fetchUser()
      if (!fetchedAuth?.user) {
        clearSession()
        return { success: false, error: 'Login succeeded but user session could not be loaded' }
      }
      return {
        success: true,
        user: fetchedAuth?.user,
        permissions: fetchedAuth?.permissions || [],
      }
    } catch (error) {
      return { success: false, error: getErrorMessage(error, 'Login failed') }
    }
  }

  const register = async (payload) => {
    try {
      await authService.register(payload)
      return { success: true }
    } catch (error) {
      return { success: false, error: getErrorMessage(error, 'Registration failed') }
    }
  }

  const logout = async () => {
    try {
      if (token) await authService.logout()
    } catch {
      // Backend logout can fail if token expired; local session still must clear.
    } finally {
      clearSession()
    }
  }

  const hasPermission = (resource, action) => {
    if (user?.role === 'superadmin') return true
    if (!action) return permissions.some((item) => item.name === resource)
    return permissions.some((item) => item.resource === resource && item.action === action)
  }

  return (
    <AuthContext.Provider value={{
      user,
      permissions,
      loading,
      isAuthenticated: Boolean(user),
      login,
      register,
      logout,
      fetchUser,
      hasPermission,
    }}>
      {children}
    </AuthContext.Provider>
  )
}
