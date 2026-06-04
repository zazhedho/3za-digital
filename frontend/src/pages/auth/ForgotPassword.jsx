import { useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import authService from '../../services/authService'
import { getErrorMessage } from '../../services/api'

const ForgotPassword = () => {
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)

  const submit = async (event) => {
    event.preventDefault()
    setLoading(true)
    try {
      await authService.forgotPassword(email)
      toast.success('Reset link sent')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to send reset link'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-screen single">
      <section className="auth-card">
        <h2>Forgot password</h2>
        <form onSubmit={submit}>
          <label className="form-label">Email</label>
          <input className="form-control" type="email" value={email} onChange={(event) => setEmail(event.target.value)} required />
          <button className="btn btn-primary w-100 mt-4" disabled={loading}>{loading ? 'Sending...' : 'Send reset link'}</button>
        </form>
        <Link className="auth-return" to="/login">Back to login</Link>
      </section>
    </div>
  )
}

export default ForgotPassword
