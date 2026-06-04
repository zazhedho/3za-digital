import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { toast } from 'react-toastify'
import authService from '../../services/authService'
import { getErrorMessage } from '../../services/api'
import { useAuth } from '../../contexts/AuthContext'

const Register = () => {
  const [form, setForm] = useState({ name: '', email: '', password: '', phone: '', otp_code: '' })
  const [loading, setLoading] = useState(false)
  const { register } = useAuth()
  const navigate = useNavigate()

  const submit = async (event) => {
    event.preventDefault()
    setLoading(true)
    const payload = { ...form }
    if (!payload.otp_code) delete payload.otp_code
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
    try {
      await authService.sendRegisterOTP(form.email)
      toast.success('OTP sent')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to send OTP'))
    }
  }

  return (
    <div className="auth-screen single">
      <section className="auth-card">
        <h2>Create account</h2>
        <form onSubmit={submit}>
          {['name', 'email', 'phone'].map((field) => (
            <div key={field} className="mt-3">
              <label className="form-label text-capitalize">{field}</label>
              <input
                className="form-control"
                type={field === 'email' ? 'email' : 'text'}
                value={form[field]}
                onChange={(event) => setForm({ ...form, [field]: event.target.value })}
                required
              />
            </div>
          ))}
          <label className="form-label mt-3">Password</label>
          <input
            className="form-control"
            type="password"
            value={form.password}
            onChange={(event) => setForm({ ...form, password: event.target.value })}
            required
          />
          <label className="form-label mt-3">OTP code</label>
          <div className="input-group">
            <input
              className="form-control"
              value={form.otp_code}
              onChange={(event) => setForm({ ...form, otp_code: event.target.value })}
              maxLength="6"
              placeholder="Optional"
            />
            <button className="btn btn-outline-dark" type="button" onClick={sendOTP}>Send OTP</button>
          </div>
          <button className="btn btn-primary w-100 mt-4" disabled={loading}>
            {loading ? 'Creating...' : 'Create account'}
          </button>
        </form>
        <Link className="auth-return" to="/login">Back to login</Link>
      </section>
    </div>
  )
}

export default Register
