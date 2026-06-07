import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { toast } from 'react-toastify'
import authService from '../../services/authService'
import { useAuth } from '../../contexts/AuthContext'
import { getGoogleClientId, renderGoogleIdentityButton } from '../../utils/googleIdentity'

const Login = () => {
  const [form, setForm] = useState({ email: '', password: '' })
  const [loading, setLoading] = useState(false)
  const [registerEnabled, setRegisterEnabled] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const googleButtonRef = useRef(null)
  const { login, loginWithGoogle } = useAuth()
  const navigate = useNavigate()
  const googleClientId = getGoogleClientId()

  useEffect(() => {
    authService.registerStatus()
      .then((response) => setRegisterEnabled(Boolean(response.data.data?.enabled)))
      .catch(() => setRegisterEnabled(false))
  }, [])

  const redirectAfterLogin = useCallback((result) => {
    const canViewDashboard = result.user?.role === 'superadmin' || (result.permissions || []).some(
      (permission) => permission.resource === 'dashboard' && permission.action === 'view',
    )
    navigate(canViewDashboard ? '/dashboard' : '/profile', { replace: true })
  }, [navigate])

  const handleGoogleCredential = useCallback(async (credential) => {
    if (!credential) {
      toast.error('Google credential is missing')
      return
    }
    setLoading(true)
    const result = await loginWithGoogle(credential)
    setLoading(false)
    if (!result.success) {
      toast.error(result.error)
      return
    }
    toast.success('Login successful')
    redirectAfterLogin(result)
  }, [loginWithGoogle, redirectAfterLogin])

  useEffect(() => {
    if (!googleClientId || !googleButtonRef.current) return undefined
    return renderGoogleIdentityButton({
      element: googleButtonRef.current,
      clientId: googleClientId,
      onCredential: handleGoogleCredential,
      text: 'signin_with',
    })
  }, [googleClientId, handleGoogleCredential])

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
    redirectAfterLogin(result)
  }

  return (
    <div className="auth-screen auth-modern">
      <section className="auth-visual auth-visual-modern">
        <div className="auth-visual-inner">
          <div className="brand-mark auth-brand-mark">3ZA</div>
          <div className="marketplace-tabs">
            <span><i className="bi bi-stars"></i> SMM Services</span>
            <span><i className="bi bi-receipt"></i> Orders</span>
            <span><i className="bi bi-wallet2"></i> Wallet</span>
          </div>
          <h1>Operate digital orders with clearer control.</h1>
          <p>Manage SMM services, deposits, wallet balances, and provider status from one focused workspace.</p>
          <div className="auth-signal-grid">
            <div><strong>24/7</strong><span>Order monitoring</span></div>
            <div><strong>QRIS</strong><span>Deposit support</span></div>
            <div><strong>RBAC</strong><span>Permission based</span></div>
          </div>
        </div>
      </section>

      <section className="auth-card auth-panel">
        <div className="auth-mobile-brand">
          <div className="brand-mark">3ZA</div>
          <div>
            <strong>3ZA Digital</strong>
            <span>Orders, wallet, and deposits</span>
          </div>
        </div>
        <div className="auth-heading">
          <div>
            <h2>Welcome back</h2>
            <p>Sign in to continue managing your workspace.</p>
          </div>
        </div>

        {googleClientId && (
          <>
            <div className="google-auth-button" ref={googleButtonRef}></div>
            <div className="auth-divider"><span>or use email</span></div>
          </>
        )}

        <form className="auth-form" onSubmit={submit}>
          <label className="auth-field">
            <span>Email</span>
            <div className="auth-input">
              <i className="bi bi-envelope"></i>
              <input
                type="email"
                value={form.email}
                onChange={(event) => setForm({ ...form, email: event.target.value })}
                placeholder="name@example.com"
                autoComplete="email"
                required
              />
            </div>
          </label>

          <label className="auth-field">
            <span>Password</span>
            <div className="auth-input">
              <i className="bi bi-lock"></i>
              <input
                type={showPassword ? 'text' : 'password'}
                value={form.password}
                onChange={(event) => setForm({ ...form, password: event.target.value })}
                placeholder="Enter password"
                autoComplete="current-password"
                required
              />
              <button type="button" className="auth-input-action" onClick={() => setShowPassword((value) => !value)} aria-label={showPassword ? 'Hide password' : 'Show password'}>
                <i className={`bi ${showPassword ? 'bi-eye-slash' : 'bi-eye'}`}></i>
              </button>
            </div>
          </label>

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
