import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import authService from '../../services/authService'
import { getErrorMessage } from '../../services/api'
import { passwordRequirements, passwordStrength, passwordStrengthLabel, validatePassword } from '../../utils/passwordValidation'

const ResetPassword = () => {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const [token, setToken] = useState(params.get('token') || '')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [showPassword, setShowPassword] = useState(false)

  const validation = validatePassword(password)
  const strength = passwordStrength(validation)
  const passwordMatches = confirmPassword && password === confirmPassword

  const submit = async (event) => {
    event.preventDefault()
    if (password !== confirmPassword) {
      toast.error('Passwords do not match')
      return
    }
    setLoading(true)
    try {
      await authService.resetPassword(token, password)
      toast.success('Password reset successful')
      navigate('/login')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to reset password'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-screen auth-modern">
      <section className="auth-visual auth-visual-modern">
        <div className="auth-visual-inner">
          <div className="brand-mark auth-brand-mark">3ZA</div>
          <div className="marketplace-tabs">
            <span><i className="bi bi-shield-lock-fill"></i> Secure Update</span>
            <span><i className="bi bi-check-all"></i> Fast Reset</span>
          </div>
          <h1>Stronger security.</h1>
          <p>Create a new strong password to protect your account and personal digital assets.</p>
          <div className="auth-signal-grid">
            <div><strong>Verified</strong><span>Secure token check</span></div>
            <div><strong>Protected</strong><span>Enhanced safety</span></div>
          </div>
        </div>
      </section>

      <section className="auth-card auth-panel">
        <div className="auth-mobile-brand">
          <div className="brand-mark">3ZA</div>
          <div>
            <strong>3ZA Digital</strong>
            <span>Secure account access</span>
          </div>
        </div>
        <div className="auth-heading">
          <div>
            <div className="auth-kicker"><i className="bi bi-shield-plus"></i> Password Reset</div>
            <h2>Reset password</h2>
            <p>Please enter and confirm your new strong password.</p>
          </div>
        </div>

        <form className="auth-form" onSubmit={submit}>
          <label className="auth-field">
            <span>Reset Token</span>
            <div className="auth-input">
              <i className="bi bi-ticket-perforated"></i>
              <input
                value={token}
                onChange={(event) => setToken(event.target.value)}
                placeholder="Enter your reset token"
                required
              />
            </div>
          </label>

          <label className="auth-field">
            <span>New Password</span>
            <div className="auth-input">
              <i className="bi bi-lock"></i>
              <input
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                placeholder="Create new password"
                autoComplete="new-password"
                required
              />
              <button type="button" className="auth-input-action" onClick={() => setShowPassword((value) => !value)}>
                <i className={`bi ${showPassword ? 'bi-eye-slash' : 'bi-eye'}`}></i>
              </button>
            </div>
          </label>

          {password && (
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
            <span>Confirm New Password</span>
            <div className="auth-input">
              <i className="bi bi-shield-lock"></i>
              <input
                type={showPassword ? 'text' : 'password'}
                value={confirmPassword}
                onChange={(event) => setConfirmPassword(event.target.value)}
                placeholder="Confirm new password"
                autoComplete="new-password"
                required
              />
            </div>
          </label>

          {confirmPassword && (
            <div className={`password-match-note ${passwordMatches ? 'valid' : ''}`}>
              <i className={`bi ${passwordMatches ? 'bi-check-circle-fill' : 'bi-exclamation-circle-fill'}`}></i>
              {passwordMatches ? 'Passwords match' : 'Passwords do not match'}
            </div>
          )}

          <button className="btn btn-primary w-100 d-flex align-items-center justify-content-center gap-2 mt-2" disabled={loading}>
            {loading ? (
              <>
                <span className="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
                <span>Saving...</span>
              </>
            ) : (
              'Reset Password'
            )}
          </button>
        </form>

        <Link className="auth-return" to="/login">
          <i className="bi bi-arrow-left me-2"></i> Back to login
        </Link>
      </section>
    </div>
  )
}

export default ResetPassword
