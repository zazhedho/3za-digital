import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import roleService from '../../services/roleService'
import userService from '../../services/userService'
import { getErrorMessage, getListPayload } from '../../services/api'
import SearchableSelect from '../../components/common/SearchableSelect'

const UserForm = () => {
  const { id } = useParams()
  const isEdit = Boolean(id)
  const navigate = useNavigate()
  const [roles, setRoles] = useState([])
  const [form, setForm] = useState({ name: '', email: '', phone: '', password: '', role: '' })

  useEffect(() => {
    roleService.getAll({ limit: 100 })
      .then((response) => setRoles(getListPayload(response).rows))
      .catch(() => setRoles([]))

    if (isEdit) {
      userService.getById(id)
        .then((response) => {
          const user = response.data.data
          setForm({ name: user.name || '', email: user.email || '', phone: user.phone || '', password: '', role: user.role || '' })
        })
        .catch((error) => toast.error(getErrorMessage(error, 'Failed to load user')))
    }
  }, [id, isEdit])

  const submit = async (event) => {
    event.preventDefault()
    if (!form.role) {
      toast.error('Role is required')
      return
    }
    try {
      const payload = { ...form }
      if (isEdit) delete payload.password
      if (isEdit) await userService.update(id, payload)
      else await userService.create(payload)
      toast.success(isEdit ? 'User updated' : 'User created')
      navigate(isEdit ? `/users/${id}` : '/users')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to save user'))
    }
  }

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>{isEdit ? 'Edit User' : 'Create User'}</h1>
          <p>User account and role.</p>
        </div>
      </div>

      <section className="form-panel">
        <form onSubmit={submit}>
          <label className="form-label">Name</label>
          <input className="form-control" value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} required />
          <label className="form-label mt-3">Email</label>
          <input className="form-control" type="email" value={form.email} onChange={(event) => setForm({ ...form, email: event.target.value })} required />
          <label className="form-label mt-3">Phone</label>
          <input className="form-control" value={form.phone} onChange={(event) => setForm({ ...form, phone: event.target.value })} />
          {!isEdit && (
            <>
              <label className="form-label mt-3">Password</label>
              <input className="form-control" type="password" value={form.password} onChange={(event) => setForm({ ...form, password: event.target.value })} required />
            </>
          )}
          <label className="form-label mt-3">Role</label>
          <SearchableSelect
            value={form.role}
            onChange={(roleName) => setForm({ ...form, role: roleName })}
            placeholder="Select role"
            searchPlaceholder="Search role"
            options={roles.map((role) => ({
              value: role.name,
              label: role.display_name || role.name,
              description: role.name,
            }))}
          />
          <div className="d-flex gap-2 mt-4">
            <button className="btn btn-primary">Save</button>
            <button className="btn btn-outline-secondary" type="button" onClick={() => navigate(-1)}>Cancel</button>
          </div>
        </form>
      </section>
    </div>
  )
}

export default UserForm
