import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { toast } from 'react-toastify'
import authService from '../../services/authService'
import { getErrorMessage } from '../../services/api'
import { useAuth } from '../../contexts/AuthContext'
import { getGoogleClientId, renderGoogleIdentityButton } from '../../utils/googleIdentity'
import { isPasswordValid, passwordRequirements, passwordStrength, passwordStrengthLabel, validatePassword } from '../../utils/passwordValidation'

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const maxNameLength = 100
const normalizeEmailInput = (value) => value.trim().toLowerCase()
const normalizeNameInput = (value) => value.trim().replace(/\s+/g, ' ')
const sanitizePhoneInput = (value) => {
  const trimmed = value.trim()
  const prefix = trimmed.startsWith('+') ? '+' : ''
  return prefix + trimmed.replace(/[^\d]/g, '')
}
const isValidPhone = (value) => /^\+?\d{9,15}$/.test(value)

const Register = () => {
  const [form, setForm] = useState({ name: '', email: '', password: '', confirm_password: '', phone: '', otp_code: '' })
  const [status, setStatus] = useState({ enabled: true, otp_enabled: false, otp_cooldown: 60 })
  const [statusLoading, setStatusLoading] = useState(true)
  const [loading, setLoading] = useState(false)
  const [otpLoading, setOtpLoading] = useState(false)
  const [otpStep, setOtpStep] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const googleButtonRef = useRef(null)
  const otpInputRefs = useRef([])
  const { register, loginWithGoogle } = useAuth()
  const navigate = useNavigate()
  const googleClientId = getGoogleClientId()
  const validation = validatePassword(form.password)
  const strength = passwordStrength(validation)
  const passwordMatches = form.confirm_password && form.password === form.confirm_password

  useEffect(() => {
    let timer
    if (countdown > 0) {
      timer = setInterval(() => {
        setCountdown((current) => {
          if (current <= 1) {
            localStorage.removeItem('3za_otp_expiry')
            return 0
          }
          return current - 1
        })
      }, 1000)
    }
    return () => {
      if (timer) clearInterval(timer)
    }
  }, [countdown])

  useEffect(() => {
    let mounted = true
    
    // Restore countdown from localStorage
    const expiry = localStorage.getItem('3za_otp_expiry')
    if (expiry) {
      const remaining = Math.ceil((parseInt(expiry, 10) - Date.now()) / 1000)
      if (remaining > 0) {
        setCountdown(remaining)
        setOtpStep(true) // If there's an active timer, user should be in OTP step
      } else {
        localStorage.removeItem('3za_otp_expiry')
      }
    }

    authService.registerStatus()
      .then((response) => {
        if (mounted) setStatus({ enabled: true, otp_enabled: false, otp_cooldown: 60, ...(response.data.data || {}) })
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

  const validateRegistrationDetails = () => {
    if (!status.enabled) {
      toast.error('Registration is disabled')
      return false
    }
    const name = normalizeNameInput(form.name)
    if (name.length < 3 || name.length > maxNameLength) {
      toast.error('Name must contain 3 to 100 characters')
      return false
    }
    if (!emailPattern.test(form.email)) {
      toast.error('Enter a valid email address')
      return false
    }
    if (!isValidPhone(form.phone)) {
      toast.error('Phone must contain 9 to 15 digits')
      return false
    }
    if (!isPasswordValid(validation)) {
      toast.error('Password does not meet all requirements')
      return false
    }
    if (form.password !== form.confirm_password) {
      toast.error('Passwords do not match')
      return false
    }
    return true
  }

  const requestOTP = async () => {
    const email = normalizeEmailInput(form.email)
    const phone = sanitizePhoneInput(form.phone)
    
    if (!emailPattern.test(email)) {
      toast.error('Enter a valid email address')
      return false
    }
    if (!status.otp_enabled) {
      toast.error('Registration OTP is disabled')
      return false
    }
    setOtpLoading(true)
    try {
      await authService.sendRegisterOTP(email, phone)
      setForm((current) => ({ ...current, email, phone }))
      const cooldownSeconds = status.otp_cooldown || 60
      setCountdown(cooldownSeconds)
      localStorage.setItem('3za_otp_expiry', (Date.now() + cooldownSeconds * 1000).toString())
      toast.success('OTP sent')
      return true
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to send OTP'))
      return false
    } finally {
      setOtpLoading(false)
    }
  }

  const submit = async (event) => {
    if (event) event.preventDefault()
    if (!validateRegistrationDetails()) return

    if (status.otp_enabled && !otpStep) {
      const sent = await requestOTP()
      if (sent) setOtpStep(true)
      return
    }

    if (status.otp_enabled && !form.otp_code.trim()) {
      toast.error('OTP code is required')
      return
    }

    setLoading(true)
    const payload = {
      ...form,
      name: normalizeNameInput(form.name),
      email: normalizeEmailInput(form.email),
      phone: sanitizePhoneInput(form.phone),
    }
    delete payload.confirm_password
    if (!status.otp_enabled || !payload.otp_code) delete payload.otp_code
    const result = await register(payload)
    setLoading(false)
    if (!result.success) {
      toast.error(result.error)
      return
    }
    localStorage.removeItem('3za_otp_expiry')
    toast.success('Account created')
    navigate('/login')
  }

  const handleOtpChange = (index, value) => {
    const cleanValue = value.replace(/\D/g, '').slice(-1)
    if (!cleanValue && value !== '') return

    const newOtpArray = form.otp_code.split('')
    // Pad with empty strings if length is less than index
    while (newOtpArray.length <= index) newOtpArray.push('')
    newOtpArray[index] = cleanValue
    const newOtp = newOtpArray.join('').slice(0, 6)
    
    setForm({ ...form, otp_code: newOtp })

    // Move to next input
    if (cleanValue && index < 5) {
      otpInputRefs.current[index + 1]?.focus()
    }
  }

  const handleOtpKeyDown = (index, event) => {
    if (event.key === 'Backspace' && !form.otp_code[index] && index > 0) {
      otpInputRefs.current[index - 1]?.focus()
    }
  }

  const handleOtpPaste = (event) => {
    event.preventDefault()
    const pastedData = event.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6)
    if (pastedData) {
      setForm({ ...form, otp_code: pastedData })
      // Focus the last filled input or the next one
      const nextFocus = Math.min(pastedData.length, 5)
      otpInputRefs.current[nextFocus]?.focus()
    }
  }

  return (
    <div className="auth-screen auth-modern auth-register-screen">
      <section className="auth-visual auth-visual-modern">
        <div className="auth-visual-inner">
          <div className="brand-mark auth-brand-mark">3ZA</div>
          <div className="marketplace-tabs">
            <span><i className="bi bi-rocket-takeoff"></i> Quick Setup</span>
            <span><i className="bi bi-phone"></i> Mobile Friendly</span>
            <span><i className="bi bi-shield-check"></i> Safe & Secure</span>
          </div>
          <h1>Join us and get started.</h1>
          <p>Create your account in seconds and get instant access to all our premium digital services.</p>
          <div className="auth-signal-grid">
            <div><strong>Fast</strong><span>Quick verification</span></div>
            <div><strong>Easy</strong><span>One-tap login</span></div>
            <div><strong>Protected</strong><span>Secure data</span></div>
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
            <div className="auth-kicker"><i className="bi bi-person-check"></i> New account</div>
            <h2>{otpStep ? 'Verify Email' : 'Create account'}</h2>
            <p>{otpStep ? 'Enter the code we just sent to you.' : 'Use Google or create a secured email account.'}</p>
          </div>
        </div>

        {!statusLoading && !status.enabled && (
          <div className="auth-alert">Public registration is currently disabled.</div>
        )}
        <form className="auth-form" onSubmit={submit}>
          {!otpStep ? (
            <>
              <div className="auth-two-col">
                <label className="auth-field">
                  <span>Name</span>
                  <div className="auth-input">
                    <i className="bi bi-person"></i>
                    <input
                      value={form.name}
                      onBlur={() => setForm({ ...form, name: normalizeNameInput(form.name) })}
                      onChange={(event) => setForm({ ...form, name: event.target.value.slice(0, maxNameLength) })}
                      placeholder="Full name"
                      autoComplete="name"
                      maxLength={maxNameLength}
                      minLength="3"
                      required
                    />
                  </div>
                </label>
                <label className="auth-field">
                  <span>Phone</span>
                  <div className="auth-input">
                    <i className="bi bi-phone"></i>
                    <input
                      value={form.phone}
                      onChange={(event) => setForm({ ...form, phone: sanitizePhoneInput(event.target.value) })}
                      placeholder="628..."
                      autoComplete="tel"
                      inputMode="tel"
                      maxLength="16"
                      pattern="^\+?\d{9,15}$"
                      title="Phone must contain 9 to 15 digits"
                      required
                    />
                  </div>
                </label>
              </div>

              <label className="auth-field">
                <span>Email</span>
                <div className="auth-input">
                  <i className="bi bi-envelope"></i>
                  <input
                    type="email"
                    value={form.email}
                    onBlur={() => setForm({ ...form, email: normalizeEmailInput(form.email) })}
                    onChange={(event) => setForm({ ...form, email: event.target.value.replace(/\s/g, '') })}
                    placeholder="name@example.com"
                    autoComplete="email"
                    inputMode="email"
                    pattern="^[^\s@]+@[^\s@]+\.[^\s@]+$"
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
            </>
          ) : (
            <div className="auth-otp-modern-container">
              <div className="auth-otp-visual-info">
                <div className="otp-icon-luxe"><i className="bi bi-shield-lock"></i></div>
                <p>We've sent a 6-digit code to <strong>{form.email}</strong></p>
              </div>
              
              <div className="otp-segmented-input" onPaste={handleOtpPaste}>
                {[...Array(6)].map((_, i) => (
                  <input
                    key={i}
                    ref={(el) => (otpInputRefs.current[i] = el)}
                    type="text"
                    inputMode="numeric"
                    maxLength="1"
                    value={form.otp_code[i] || ''}
                    onChange={(e) => handleOtpChange(i, e.target.value)}
                    onKeyDown={(e) => handleOtpKeyDown(i, e)}
                    className={form.otp_code[i] ? 'has-value' : ''}
                  />
                ))}
              </div>

              <div className="otp-countdown-luxe">
                {countdown > 0 ? (
                  <div className="otp-timer-row">
                    <span className="timer-text">Resend code in</span>
                    <span className="timer-value">{countdown}s</span>
                  </div>
                ) : (
                  <button 
                    type="button" 
                    className="otp-resend-link" 
                    onClick={requestOTP} 
                    disabled={otpLoading}
                  >
                    {otpLoading ? 'Sending...' : "Didn't receive code? Resend"}
                  </button>
                )}
              </div>

              <div className="otp-change-email-row">
                <button 
                  type="button" 
                  className="btn btn-link btn-sm" 
                  onClick={() => {
                    setOtpStep(false)
                    setForm({ ...form, otp_code: '' })
                  }}
                >
                  Change email address
                </button>
              </div>
            </div>
          )}

          <div className="auth-submit-area">
            <button className="btn btn-primary w-100 d-flex align-items-center justify-content-center gap-2" disabled={loading || otpLoading || statusLoading || !status.enabled}>
              {loading || otpLoading ? (
                <>
                  <span className="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
                  <span>{otpLoading ? 'Processing...' : 'Loading...'}</span>
                </>
              ) : (
                otpStep ? 'Verify and Create Account' : 'Create account'
              )}
            </button>
          </div>
        </form>
        
        {!otpStep && googleClientId && status.enabled && (
          <>
            <div className="auth-divider"><span>or continue with</span></div>
            <div className="google-auth-button" ref={googleButtonRef}></div>
          </>
        )}
        <Link className="auth-return" to="/login">Back to login</Link>
      </section>
    </div>
  )
}

export default Register
