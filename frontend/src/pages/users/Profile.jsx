import { useEffect, useState } from 'react'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import authService from '../../services/authService'
import { getErrorMessage } from '../../services/api'

const Profile = () => {
  const { user, fetchUser } = useAuth()
  const [form, setForm] = useState({ name: '', phone: '' })
  const [passwordForm, setPasswordForm] = useState({ current_password: '', new_password: '' })
  const [loading, setLoading] = useState(false)
  const [passLoading, setPasswordLoading] = useState(false)

  useEffect(() => {
    if (user) setForm({ name: user.name || '', phone: user.phone || '' })
  }, [user])

  const saveProfile = async (event) => {
    event.preventDefault()
    setLoading(true)
    try {
      await authService.updateProfile(form)
      await fetchUser()
      toast.success('Profile updated')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Update failed'))
    } finally {
      setLoading(false)
    }
  }

  const changePassword = async (event) => {
    event.preventDefault()
    setPasswordLoading(true)
    try {
      await authService.changePassword(passwordForm)
      setPasswordForm({ current_password: '', new_password: '' })
      toast.success('Password changed')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to change password'))
    } finally {
      setPasswordLoading(false)
    }
  }

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>My Profile</h1>
          <p>Manage your account settings and security</p>
        </div>
      </div>

      <div className="luxe-detail-hero">
        <div className="luxe-hero-content d-flex align-items-center gap-4">
          <div className="user-avatar-luxe">
            {user?.avatar_url ? (
              <img src={user.avatar_url} alt={user.name} />
            ) : (
              <div className="avatar-placeholder">{user?.name?.charAt(0) || '?'}</div>
            )}
          </div>
          <div>
            <div className="luxe-hero-kicker">
              <i className="bi bi-person-badge"></i> {user?.role || 'Member'}
            </div>
            <h2 className="luxe-hero-title">{user?.name || 'My Account'}</h2>
            <p className="luxe-hero-subtitle">{user?.email}</p>
          </div>
        </div>
      </div>

      <div className="content-grid two">
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-person-gear"></i> Account Settings</h3>
          </div>
          <div className="luxe-card-body">
            <form onSubmit={saveProfile} className="deposit-form-modern">
              <div className="deposit-input-group">
                <label>Full Name</label>
                <div className="auth-input m-0">
                   <i className="bi bi-person"></i>
                   <input 
                      value={form.name} 
                      onChange={(event) => setForm({ ...form, name: event.target.value })} 
                      placeholder="Your name"
                      style={{ background: 'white' }}
                      required
                   />
                </div>
              </div>

              <div className="deposit-input-group">
                <label>Phone Number</label>
                <div className="auth-input m-0">
                   <i className="bi bi-phone"></i>
                   <input 
                      value={form.phone} 
                      onChange={(event) => setForm({ ...form, phone: event.target.value })} 
                      placeholder="628..."
                      style={{ background: 'white' }}
                      required
                   />
                </div>
              </div>

              <div className="toolbar-actions justify-content-end mt-4">
                <button className="btn btn-primary px-4" disabled={loading}>
                  {loading ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </form>
          </div>
        </section>

        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-shield-lock"></i> Security</h3>
          </div>
          <div className="luxe-card-body">
            <form onSubmit={changePassword} className="deposit-form-modern">
              <div className="deposit-input-group">
                <label>Current Password</label>
                <div className="auth-input m-0">
                   <i className="bi bi-lock"></i>
                   <input 
                      type="password"
                      value={passwordForm.current_password} 
                      onChange={(event) => setPasswordForm({ ...passwordForm, current_password: event.target.value })} 
                      placeholder="••••••••"
                      style={{ background: 'white' }}
                      required
                   />
                </div>
              </div>

              <div className="deposit-input-group">
                <label>New Password</label>
                <div className="auth-input m-0">
                   <i className="bi bi-key"></i>
                   <input 
                      type="password"
                      value={passwordForm.new_password} 
                      onChange={(event) => setPasswordForm({ ...passwordForm, new_password: event.target.value })} 
                      placeholder="New strong password"
                      style={{ background: 'white' }}
                      required
                   />
                </div>
              </div>

              <div className="toolbar-actions justify-content-end mt-4">
                <button className="btn btn-outline-danger px-4" disabled={passLoading}>
                  {passLoading ? 'Updating...' : 'Change Password'}
                </button>
              </div>
            </form>
          </div>
        </section>
      </div>
    </div>
  )
}

export default Profile
