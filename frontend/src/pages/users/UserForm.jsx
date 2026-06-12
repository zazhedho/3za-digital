import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import roleService from '../../services/roleService'
import userService from '../../services/userService'
import { getErrorMessage, getListPayload } from '../../services/api'
import SearchableSelect from '../../components/common/SearchableSelect'
import BackButton from '../../components/common/BackButton'

const UserForm = () => {
  const { id } = useParams()
  const isEdit = Boolean(id)
  const navigate = useNavigate()
  const [roles, setRoles] = useState([])
  const [form, setForm] = useState({ name: '', email: '', phone: '', password: '', role: '' })
  const [loading, setLoading] = useState(isEdit)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    roleService.getAll({ limit: 100 })
      .then((response) => setRoles(getListPayload(response).rows))
      .catch(() => setRoles([]))

    if (isEdit) {
      userService.getById(id)
        .then((response) => {
          const user = response.data.data
          setForm({ 
            name: user.name || '', 
            email: user.email || '', 
            phone: user.phone || '', 
            password: '', 
            role: user.role || '' 
          })
        })
        .catch((error) => toast.error(getErrorMessage(error, 'Failed to load user')))
        .finally(() => setLoading(false))
    }
  }, [id, isEdit])

  const submit = async (event) => {
    event.preventDefault()
    if (!form.role) {
      toast.error('User role is required')
      return
    }
    setSaving(true)
    try {
      const payload = { ...form }
      if (isEdit) delete payload.password
      
      if (isEdit) await userService.update(id, payload)
      else await userService.create(payload)
      
      toast.success(isEdit ? 'User account updated' : 'User account created')
      navigate(isEdit ? `/users/${id}` : '/users', { replace: isEdit })
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to save user'))
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <div className="loading-fade">Loading member data...</div>

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>{isEdit ? 'Edit User' : 'Create New User'}</h1>
          <p>{isEdit ? 'Modify existing member profile and access level.' : 'Add a new member to the platform with specific role.'}</p>
        </div>
        <div className="toolbar-actions">
           <BackButton fallback="/users" />
        </div>
      </div>

      <div className="content-grid single max-w-lg mx-auto">
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className={`bi ${isEdit ? 'bi-person-gear' : 'bi-person-plus'}`}></i> Account Details</h3>
          </div>
          <div className="luxe-card-body">
            <form onSubmit={submit} className="deposit-form-modern">
              <div className="deposit-input-group">
                <label>Full Name</label>
                <div className="auth-input m-0">
                   <i className="bi bi-person"></i>
                   <input 
                     value={form.name} 
                     onChange={(event) => setForm({ ...form, name: event.target.value })} 
                     placeholder="Member's full name"
                     required 
                     style={{ background: 'white' }}
                   />
                </div>
              </div>

              <div className="deposit-input-group mt-3">
                <label>Email Address</label>
                <div className="auth-input m-0">
                   <i className="bi bi-envelope"></i>
                   <input 
                     type="email" 
                     value={form.email} 
                     onChange={(event) => setForm({ ...form, email: event.target.value })} 
                     placeholder="email@example.com"
                     required 
                     style={{ background: 'white' }}
                   />
                </div>
              </div>

              <div className="deposit-input-group mt-3">
                <label>Phone Number</label>
                <div className="auth-input m-0">
                   <i className="bi bi-phone"></i>
                   <input 
                     value={form.phone} 
                     onChange={(event) => setForm({ ...form, phone: event.target.value })} 
                     placeholder="628..."
                     style={{ background: 'white' }}
                   />
                </div>
              </div>

              {!isEdit && (
                <div className="deposit-input-group mt-3">
                  <label>Initial Password</label>
                  <div className="auth-input m-0">
                     <i className="bi bi-lock"></i>
                     <input 
                       type="password" 
                       value={form.password} 
                       onChange={(event) => setForm({ ...form, password: event.target.value })} 
                       placeholder="Assign secure password"
                       required 
                       style={{ background: 'white' }}
                     />
                  </div>
                </div>
              )}

              <div className="deposit-input-group mt-3">
                <label>System Role</label>
                <SearchableSelect
                  value={form.role}
                  onChange={(roleName) => setForm({ ...form, role: roleName })}
                  placeholder="Assign a role..."
                  searchPlaceholder="Search available roles"
                  options={roles.map((role) => ({
                    value: role.name,
                    label: role.display_name || role.name,
                    description: role.description || `Access as ${role.name}`,
                  }))}
                />
              </div>

              <div className="toolbar-actions justify-content-end mt-5 pt-3 border-top">
                <button className="btn btn-outline-dark px-4" type="button" onClick={() => navigate(-1)} disabled={saving}>
                  Cancel
                </button>
                <button className="btn btn-primary px-5" disabled={saving}>
                   {saving ? 'Saving...' : isEdit ? 'Update Member' : 'Create Member'}
                </button>
              </div>
            </form>
          </div>
        </section>
      </div>
    </div>
  )
}

export default UserForm
