import { useState } from 'react'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import authService from '../../services/authService'
import { getErrorMessage } from '../../services/api'

const Profile = () => {
  const { user, fetchUser } = useAuth()
  const [form, setForm] = useState({
    name: user?.name || '',
    phone: user?.phone || '',
  })
  const [password, setPassword] = useState({ current_password: '', new_password: '' })

  const saveProfile = async (event) => {
    event.preventDefault()
    try {
      await authService.updateProfile(form)
      await fetchUser()
      toast.success('Profile updated')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to update profile'))
    }
  }

  const savePassword = async (event) => {
    event.preventDefault()
    try {
      await authService.changePassword(password)
      setPassword({ current_password: '', new_password: '' })
      toast.success('Password updated')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to update password'))
    }
  }

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>Profile</h1>
          <p>{user?.email}</p>
        </div>
      </div>

      <div className="content-grid two">
        <section className="form-panel">
          <h2>Account</h2>
          <form onSubmit={saveProfile}>
            <label className="form-label">Name</label>
            <input className="form-control" value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} />
            <label className="form-label mt-3">Phone</label>
            <input className="form-control" value={form.phone} onChange={(event) => setForm({ ...form, phone: event.target.value })} />
            <button className="btn btn-primary mt-4">Save profile</button>
          </form>
        </section>

        <section className="form-panel">
          <h2>Password</h2>
          <form onSubmit={savePassword}>
            <label className="form-label">Current password</label>
            <input className="form-control" type="password" value={password.current_password} onChange={(event) => setPassword({ ...password, current_password: event.target.value })} required />
            <label className="form-label mt-3">New password</label>
            <input className="form-control" type="password" value={password.new_password} onChange={(event) => setPassword({ ...password, new_password: event.target.value })} required />
            <button className="btn btn-primary mt-4">Change password</button>
          </form>
        </section>
      </div>
    </div>
  )
}

export default Profile
