import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { toast } from 'react-toastify'
import authService from '../../services/authService'
import { useAuth } from '../../contexts/AuthContext'

const Login = () => {
  const [form, setForm] = useState({ email: '', password: '' })
  const [loading, setLoading] = useState(false)
  const [registerEnabled, setRegisterEnabled] = useState(false)
  const { login } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    authService.registerStatus()
      .then((response) => setRegisterEnabled(Boolean(response.data.data?.enabled)))
      .catch(() => setRegisterEnabled(false))
  }, [])

  const submit = async (event) => {
    event.preventDefault()
    setLoading(true)
    const result = await login(form.email, form.password)
    setLoading(false)

    if (!result.success) {
      toast.error(result.error)
      return
    }
    toast.success('Login successful')
    const canViewDashboard = result.user?.role === 'superadmin' || (result.permissions || []).some(
      (permission) => permission.resource === 'dashboard' && permission.action === 'view',
    )
    navigate(canViewDashboard ? '/dashboard' : '/profile', { replace: true })
  }

  return (
    <div className="auth-screen">
      <section className="auth-visual">
        <div className="marketplace-tabs">
          <span><i className="bi bi-house-heart"></i> Services</span>
          <span><i className="bi bi-stars"></i> Orders</span>
          <span><i className="bi bi-wallet2"></i> Wallet</span>
        </div>
        <h1>3ZA Digital</h1>
        <p>SMM transactions, wallet, deposits, and provider operations in one workspace.</p>
      </section>

      <section className="auth-card">
        <div className="auth-heading">
          <div className="brand-mark">3ZA</div>
          <div>
            <h2>Welcome back</h2>
            <p>Sign in to the 3ZA Digital dashboard.</p>
          </div>
        </div>

        <form onSubmit={submit}>
          <label className="form-label">Email</label>
          <input
            className="form-control"
            type="email"
            value={form.email}
            onChange={(event) => setForm({ ...form, email: event.target.value })}
            required
          />

          <label className="form-label mt-3">Password</label>
          <input
            className="form-control"
            type="password"
            value={form.password}
            onChange={(event) => setForm({ ...form, password: event.target.value })}
            required
          />

          <button className="btn btn-primary w-100 mt-4" disabled={loading}>
            {loading ? 'Signing in...' : 'Sign in'}
          </button>
        </form>

        <div className="auth-links">
          <Link to="/forgot-password">Forgot password?</Link>
          {registerEnabled && <Link to="/register">Create account</Link>}
        </div>
      </section>
    </div>
  )
}

export default Login
