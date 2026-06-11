import { useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import authService from '../../services/authService'
import { getErrorMessage } from '../../services/api'

const ForgotPassword = () => {
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [sent, setSent] = useState(false)

  const submit = async (event) => {
    event.preventDefault()
    setLoading(true)
    try {
      await authService.forgotPassword(email)
      setSent(true)
      toast.success('Reset link sent')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to send reset link'))
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
            <span><i className="bi bi-shield-lock"></i> Account Recovery</span>
            <span><i className="bi bi-envelope-check"></i> Email Verification</span>
          </div>
          <h1>Don't worry, we got you.</h1>
          <p>Locked out? Enter your registered email address and we'll send you instructions to reset your password safely.</p>
          <div className="auth-signal-grid">
            <div><strong>Secure</strong><span>End-to-end encryption</span></div>
            <div><strong>Fast</strong><span>Instant delivery</span></div>
          </div>
        </div>
      </section>

      <section className="auth-card auth-panel">
        <div className="auth-mobile-brand">
          <div className="brand-mark">3ZA</div>
          <div>
            <strong>3ZA Digital</strong>
            <span>Secure account recovery</span>
          </div>
        </div>
        <div className="auth-heading">
          <div>
            <div className="auth-kicker"><i className="bi bi-key"></i> Password Recovery</div>
            <h2>{sent ? 'Check your inbox' : 'Forgot password?'}</h2>
            <p>{sent ? "We've sent a recovery link to your email." : 'Enter your email to receive a password reset link.'}</p>
          </div>
        </div>

        {sent ? (
          <div className="auth-success-state py-4 text-center">
            <div className="otp-icon-luxe mb-4 mx-auto" style={{ background: '#e8f9ee', color: '#198754' }}>
              <i className="bi bi-check2-circle"></i>
            </div>
            <p className="text-muted">A reset link has been sent to <strong>{email}</strong>. Please follow the instructions in the email to regain access.</p>
            <button className="btn btn-outline-dark w-100 mt-4" onClick={() => setSent(false)}>Try another email</button>
          </div>
        ) : (
          <form className="auth-form" onSubmit={submit}>
            <label className="auth-field">
              <span>Email Address</span>
              <div className="auth-input">
                <i className="bi bi-envelope"></i>
                <input
                  type="email"
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  placeholder="name@example.com"
                  autoComplete="email"
                  required
                />
              </div>
            </label>

            <button className="btn btn-primary w-100 d-flex align-items-center justify-content-center gap-2 mt-2" disabled={loading}>
              {loading ? (
                <>
                  <span className="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
                  <span>Sending Link...</span>
                </>
              ) : (
                'Send Reset Link'
              )}
            </button>
          </form>
        )}

        <Link className="auth-return" to="/login">
          <i className="bi bi-arrow-left me-2"></i> Back to login
        </Link>
      </section>
    </div>
  )
}

export default ForgotPassword
