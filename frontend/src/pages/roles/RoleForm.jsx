import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import roleService from '../../services/roleService'
import { getErrorMessage } from '../../services/api'

const RoleForm = () => {
  const { id } = useParams()
  const isEdit = Boolean(id)
  const navigate = useNavigate()
  const [form, setForm] = useState({ name: '', display_name: '', description: '' })

  useEffect(() => {
    if (isEdit) {
      roleService.getById(id)
        .then((response) => {
          const role = response.data.data
          setForm({ name: role.name || '', display_name: role.display_name || '', description: role.description || '' })
        })
        .catch((error) => toast.error(getErrorMessage(error, 'Failed to load role')))
    }
  }, [id, isEdit])

  const submit = async (event) => {
    event.preventDefault()
    try {
      if (isEdit) await roleService.update(id, { display_name: form.display_name, description: form.description })
      else await roleService.create(form)
      toast.success(isEdit ? 'Role updated' : 'Role created')
      navigate(isEdit ? `/roles/${id}` : '/roles')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to save role'))
    }
  }

  return (
    <div>
      <div className="page-toolbar">
        <div><h1>{isEdit ? 'Edit Role' : 'Create Role'}</h1><p>Role metadata.</p></div>
      </div>
      <section className="form-panel">
        <form onSubmit={submit}>
          <label className="form-label">Name</label>
          <input className="form-control" value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} required disabled={isEdit} />
          <label className="form-label mt-3">Display name</label>
          <input className="form-control" value={form.display_name} onChange={(event) => setForm({ ...form, display_name: event.target.value })} required />
          <label className="form-label mt-3">Description</label>
          <textarea className="form-control" value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} />
          <div className="d-flex gap-2 mt-4">
            <button className="btn btn-primary">Save</button>
            <button className="btn btn-outline-secondary" type="button" onClick={() => navigate(-1)}>Cancel</button>
          </div>
        </form>
      </section>
    </div>
  )
}

export default RoleForm
