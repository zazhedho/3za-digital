import { useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import authService from '../../services/authService'
import { getErrorMessage } from '../../services/api'

const ResetPassword = () => {
  const [params] = useSearchParams()
  const [token, setToken] = useState(params.get('token') || '')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)

  const submit = async (event) => {
    event.preventDefault()
    setLoading(true)
    try {
      await authService.resetPassword(token, password)
      toast.success('Password reset successful')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to reset password'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-screen single">
      <section className="auth-card">
        <h2>Reset password</h2>
        <form onSubmit={submit}>
          <label className="form-label">Token</label>
          <input className="form-control" value={token} onChange={(event) => setToken(event.target.value)} required />
          <label className="form-label mt-3">New password</label>
          <input className="form-control" type="password" value={password} onChange={(event) => setPassword(event.target.value)} required />
          <button className="btn btn-primary w-100 mt-4" disabled={loading}>{loading ? 'Saving...' : 'Reset password'}</button>
        </form>
        <Link className="auth-return" to="/login">Back to login</Link>
      </section>
    </div>
  )
}

export default ResetPassword
