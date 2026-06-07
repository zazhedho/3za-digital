import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { toast } from 'react-toastify'
import authService from '../../services/authService'
import { getErrorMessage } from '../../services/api'
import { useAuth } from '../../contexts/AuthContext'
import { getGoogleClientId, renderGoogleIdentityButton } from '../../utils/googleIdentity'
import { isPasswordValid, passwordRequirements, passwordStrength, passwordStrengthLabel, validatePassword } from '../../utils/passwordValidation'

const Register = () => {
  const [form, setForm] = useState({ name: '', email: '', password: '', confirm_password: '', phone: '', otp_code: '' })
  const [status, setStatus] = useState({ enabled: true, otp_enabled: false })
  const [statusLoading, setStatusLoading] = useState(true)
  const [loading, setLoading] = useState(false)
  const [otpLoading, setOtpLoading] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const googleButtonRef = useRef(null)
  const { register, loginWithGoogle } = useAuth()
  const navigate = useNavigate()
  const googleClientId = getGoogleClientId()
  const validation = validatePassword(form.password)
  const strength = passwordStrength(validation)
  const passwordMatches = form.confirm_password && form.password === form.confirm_password

  useEffect(() => {
    let mounted = true
    authService.registerStatus()
      .then((response) => {
        if (mounted) setStatus({ enabled: true, otp_enabled: false, ...(response.data.data || {}) })
      })
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load registration settings')))
      .finally(() => {
        if (mounted) setStatusLoading(false)
      })
    return () => {
      mounted = false
    }
  }, [])

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
    toast.success('Signed in with Google')
    const canViewDashboard = result.user?.role === 'superadmin' || (result.permissions || []).some(
      (permission) => permission.resource === 'dashboard' && permission.action === 'view',
    )
    navigate(canViewDashboard ? '/dashboard' : '/profile', { replace: true })
  }, [loginWithGoogle, navigate])

  useEffect(() => {
    if (!status.enabled || !googleClientId || !googleButtonRef.current) return undefined
    return renderGoogleIdentityButton({
      element: googleButtonRef.current,
      clientId: googleClientId,
      onCredential: handleGoogleCredential,
      text: 'signup_with',
    })
  }, [googleClientId, handleGoogleCredential, status.enabled])

  const submit = async (event) => {
    event.preventDefault()
    if (!status.enabled) {
      toast.error('Registration is disabled')
      return
    }
    if (status.otp_enabled && !form.otp_code.trim()) {
      toast.error('OTP code is required')
      return
    }
    if (!isPasswordValid(validation)) {
      toast.error('Password does not meet all requirements')
      return
    }
    if (form.password !== form.confirm_password) {
      toast.error('Passwords do not match')
      return
    }
    setLoading(true)
    const payload = { ...form }
    delete payload.confirm_password
    if (!status.otp_enabled || !payload.otp_code) delete payload.otp_code
    const result = await register(payload)
    setLoading(false)
    if (!result.success) {
      toast.error(result.error)
      return
    }
    toast.success('Account created')
    navigate('/login')
  }

  const sendOTP = async () => {
    if (!form.email) {
      toast.error('Email is required')
      return
    }
    if (!status.otp_enabled) {
      toast.error('Registration OTP is disabled')
      return
    }
    setOtpLoading(true)
    try {
      await authService.sendRegisterOTP(form.email)
      toast.success('OTP sent')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to send OTP'))
    } finally {
      setOtpLoading(false)
    }
  }

  return (
    <div className="auth-screen auth-modern auth-register-screen">
      <section className="auth-visual auth-visual-modern">
        <div className="auth-visual-inner">
          <div className="brand-mark auth-brand-mark">3ZA</div>
          <div className="marketplace-tabs">
            <span><i className="bi bi-shield-check"></i> Secure access</span>
            <span><i className="bi bi-phone"></i> Mobile ready</span>
            <span><i className="bi bi-key"></i> OTP support</span>
          </div>
          <h1>Start with a verified account.</h1>
          <p>Create your workspace access with strong credentials and optional email verification.</p>
          <div className="auth-signal-grid">
            <div><strong>OTP</strong><span>Config based</span></div>
            <div><strong>Google</strong><span>One tap signup</span></div>
            <div><strong>Secure</strong><span>Strong password</span></div>
          </div>
        </div>
      </section>

      <section className="auth-card auth-panel auth-register-panel">
        <div className="auth-mobile-brand">
          <div className="brand-mark">3ZA</div>
          <div>
            <strong>3ZA Digital</strong>
            <span>Secure account creation</span>
          </div>
        </div>
        <div className="auth-heading">
          <div>
            <h2>Create account</h2>
            <p>Use Google or create a secured email account.</p>
          </div>
        </div>
        {googleClientId && status.enabled && (
          <>
            <div className="google-auth-button" ref={googleButtonRef}></div>
            <div className="auth-divider"><span>or use email</span></div>
          </>
        )}
        {!statusLoading && !status.enabled && (
          <div className="auth-alert">Public registration is currently disabled.</div>
        )}
        <form className="auth-form" onSubmit={submit}>
          <div className="auth-two-col">
            <label className="auth-field">
              <span>Name</span>
              <div className="auth-input">
                <i className="bi bi-person"></i>
                <input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} placeholder="Full name" autoComplete="name" required />
              </div>
            </label>
            <label className="auth-field">
              <span>Phone</span>
              <div className="auth-input">
                <i className="bi bi-phone"></i>
                <input value={form.phone} onChange={(event) => setForm({ ...form, phone: event.target.value })} placeholder="628..." autoComplete="tel" required />
              </div>
            </label>
          </div>

          <label className="auth-field">
            <span>Email</span>
            <div className="auth-input">
              <i className="bi bi-envelope"></i>
              <input type="email" value={form.email} onChange={(event) => setForm({ ...form, email: event.target.value })} placeholder="name@example.com" autoComplete="email" required />
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
                placeholder="Create password"
                autoComplete="new-password"
                required
              />
              <button type="button" className="auth-input-action" onClick={() => setShowPassword((value) => !value)} aria-label={showPassword ? 'Hide password' : 'Show password'}>
                <i className={`bi ${showPassword ? 'bi-eye-slash' : 'bi-eye'}`}></i>
              </button>
            </div>
          </label>

          {form.password && (
            <div className="password-validation-card">
              <div className="password-meter-row">
                <div className="password-meter">
                  <span style={{ width: `${(strength / 5) * 100}%` }}></span>
                </div>
                <strong>{passwordStrengthLabel(strength)}</strong>
              </div>
              <div className="password-requirements">
                {passwordRequirements.map(([key, label]) => (
                  <span className={validation[key] ? 'valid' : ''} key={key}>
                    <i className={`bi ${validation[key] ? 'bi-check-circle-fill' : 'bi-circle'}`}></i>{label}
                  </span>
                ))}
              </div>
            </div>
          )}

          <label className="auth-field">
            <span>Confirm password</span>
            <div className="auth-input">
              <i className="bi bi-shield-lock"></i>
              <input
                type={showConfirmPassword ? 'text' : 'password'}
                value={form.confirm_password}
                onChange={(event) => setForm({ ...form, confirm_password: event.target.value })}
                placeholder="Repeat password"
                autoComplete="new-password"
                required
              />
              <button type="button" className="auth-input-action" onClick={() => setShowConfirmPassword((value) => !value)} aria-label={showConfirmPassword ? 'Hide confirm password' : 'Show confirm password'}>
                <i className={`bi ${showConfirmPassword ? 'bi-eye-slash' : 'bi-eye'}`}></i>
              </button>
            </div>
          </label>
          {form.confirm_password && (
            <div className={`password-match-note ${passwordMatches ? 'valid' : ''}`}>
              <i className={`bi ${passwordMatches ? 'bi-check-circle-fill' : 'bi-exclamation-circle-fill'}`}></i>
              {passwordMatches ? 'Passwords match' : 'Passwords do not match'}
            </div>
          )}

          {status.otp_enabled && (
            <>
              <label className="auth-field">
                <span>OTP code</span>
                <div className="auth-input auth-input-with-button">
                  <i className="bi bi-key"></i>
                  <input
                    value={form.otp_code}
                    onChange={(event) => setForm({ ...form, otp_code: event.target.value })}
                    maxLength="6"
                    placeholder="Enter OTP"
                    required
                  />
                </div>
              </label>
              <div className="auth-inline-action">
                <button className="btn btn-outline-dark" type="button" onClick={sendOTP} disabled={otpLoading || !form.email}>
                  {otpLoading ? 'Sending...' : 'Send OTP'}
                </button>
              </div>
            </>
          )}
          <button className="btn btn-primary w-100 mt-4" disabled={loading || statusLoading || !status.enabled}>
            {loading ? 'Creating...' : 'Create account'}
          </button>
        </form>
        <Link className="auth-return" to="/login">Back to login</Link>
      </section>
    </div>
  )
}

export default Register
