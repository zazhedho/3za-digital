import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import roleService from '../../services/roleService'
import permissionService from '../../services/permissionService'
import { getErrorMessage, getListPayload } from '../../services/api'
import { useAuth } from '../../contexts/AuthContext'

const formatResource = (value) => (value || 'other').replace(/_/g, ' ')

const RoleForm = () => {
  const { id } = useParams()
  const isEdit = Boolean(id)
  const navigate = useNavigate()
  const { hasPermission } = useAuth()
  const [form, setForm] = useState({ name: '', display_name: '', description: '' })
  const [permissions, setPermissions] = useState([])
  const [selectedPermissions, setSelectedPermissions] = useState([])
  const [permissionSearch, setPermissionSearch] = useState('')
  const [isSystem, setIsSystem] = useState(false)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const canAssignPermissions = hasPermission('roles', 'assign_permissions')

  useEffect(() => {
    let mounted = true
    const load = async () => {
      setLoading(true)
      try {
        const requests = [permissionService.getAll({ limit: 1000 })]
        if (isEdit) requests.push(roleService.getById(id))
        const [permissionResponse, roleResponse] = await Promise.all(requests)
        if (!mounted) return

        setPermissions(getListPayload(permissionResponse).rows)
        if (roleResponse) {
          const role = roleResponse.data.data
          setForm({ name: role.name || '', display_name: role.display_name || '', description: role.description || '' })
          setSelectedPermissions(role.permission_ids || [])
          setIsSystem(Boolean(role.is_system))
        }
      } catch (error) {
        toast.error(getErrorMessage(error, isEdit ? 'Failed to load role' : 'Failed to load permissions'))
      } finally {
        if (mounted) setLoading(false)
      }
    }
    load()
    return () => {
      mounted = false
    }
  }, [id, isEdit])

  useEffect(() => {
    if (isEdit) {
      return
    }
    setForm((current) => ({
      ...current,
      name: current.name.toLowerCase().replace(/\s+/g, ''),
    }))
  }, [form.name, isEdit])

  const togglePermission = (permissionId) => {
    setSelectedPermissions((current) => (
      current.includes(permissionId)
        ? current.filter((id) => id !== permissionId)
        : [...current, permissionId]
    ))
  }

  const selectAllPermissions = () => {
    setSelectedPermissions(permissions.map((permission) => permission.id))
  }

  const clearPermissions = () => {
    setSelectedPermissions([])
  }

  const toggleResourcePermissions = (resourcePermissions) => {
    const resourcePermissionIds = resourcePermissions.map((permission) => permission.id)
    const hasAllResourcePermissions = resourcePermissionIds.every((permissionId) => selectedPermissions.includes(permissionId))
    setSelectedPermissions((current) => {
      if (hasAllResourcePermissions) {
        return current.filter((permissionId) => !resourcePermissionIds.includes(permissionId))
      }
      return Array.from(new Set([...current, ...resourcePermissionIds]))
    })
  }

  const submit = async (event) => {
    event.preventDefault()
    if (canAssignPermissions && selectedPermissions.length === 0) {
      toast.error('Select at least one permission')
      return
    }
    setSaving(true)
    try {
      let roleId = id
      if (isEdit) {
        if (!isSystem) await roleService.update(id, { display_name: form.display_name, description: form.description })
      } else {
        const response = await roleService.create(form)
        roleId = response.data.data.id
      }
      if (canAssignPermissions) await roleService.assignPermissions(roleId, { permission_ids: selectedPermissions })
      toast.success(isEdit ? 'Role updated' : 'Role created')
      navigate(isEdit ? `/roles/${roleId}` : '/roles')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to save role'))
    } finally {
      setSaving(false)
    }
  }

  const filteredPermissions = permissions.filter((permission) => {
    const keyword = permissionSearch.trim().toLowerCase()
    if (!keyword) return true
    return [
      permission.display_name,
      permission.name,
      permission.resource,
      permission.action,
      permission.description,
    ].some((value) => String(value || '').toLowerCase().includes(keyword))
  })

  const groupedPermissions = filteredPermissions.reduce((groups, permission) => {
    const key = permission.resource || 'other'
    if (!groups[key]) groups[key] = []
    groups[key].push(permission)
    return groups
  }, {})

  if (loading) {
    return <div className="empty-cell">Loading role form...</div>
  }

  return (
    <div>
      <div className="page-toolbar">
        <div><h1>{isEdit ? 'Edit Role' : 'Create Role'}</h1><p>Role metadata and permissions.</p></div>
      </div>
      <section className="form-panel">
        <form onSubmit={submit}>
          <label className="form-label">Name</label>
          <input
            className="form-control"
            value={form.name}
            onChange={(event) => setForm({ ...form, name: event.target.value.toLowerCase().replace(/\s+/g, '') })}
            required
            disabled={isEdit || isSystem}
          />
          <label className="form-label mt-3">Display name</label>
          <input className="form-control" value={form.display_name} onChange={(event) => setForm({ ...form, display_name: event.target.value })} required disabled={isSystem} />
          <label className="form-label mt-3">Description</label>
          <textarea className="form-control" value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} disabled={isSystem} />

          {canAssignPermissions && (
            <div className="role-permission-section">
              <div className="role-permission-heading">
                <div>
                  <h2>Permissions</h2>
                  <p>{selectedPermissions.length} selected from {permissions.length} permissions.</p>
                </div>
                <div className="toolbar-actions">
                  <button className="btn btn-sm btn-outline-dark" type="button" onClick={selectAllPermissions} disabled={selectedPermissions.length === permissions.length}>Select all</button>
                  <button className="btn btn-sm btn-outline-dark" type="button" onClick={clearPermissions} disabled={!selectedPermissions.length}>Clear</button>
                </div>
              </div>

              <div className="permission-search">
                <i className="bi bi-search"></i>
                <input value={permissionSearch} onChange={(event) => setPermissionSearch(event.target.value)} placeholder="Search permission" />
              </div>

              <div className="permission-group-list">
                {Object.keys(groupedPermissions).sort().map((resource) => (
                  <div className="permission-group" key={resource}>
                    <label className="permission-group-title">
                      <span className="permission-resource-label">
                        <input
                          type="checkbox"
                          checked={groupedPermissions[resource].every((permission) => selectedPermissions.includes(permission.id))}
                          onChange={() => toggleResourcePermissions(groupedPermissions[resource])}
                        />
                        <strong>{formatResource(resource)}</strong>
                      </span>
                      <span>{groupedPermissions[resource].length}</span>
                    </label>
                    <div className="permission-check-grid">
                      {groupedPermissions[resource].map((permission) => (
                        <label className="permission-check" key={permission.id}>
                          <input
                            type="checkbox"
                            checked={selectedPermissions.includes(permission.id)}
                            onChange={() => togglePermission(permission.id)}
                          />
                          <span>
                            <strong>{permission.action || permission.name}</strong>
                            <small>{permission.display_name || permission.name}</small>
                          </span>
                        </label>
                      ))}
                    </div>
                  </div>
                ))}
                {!filteredPermissions.length && <div className="empty-cell">No permissions found</div>}
              </div>
            </div>
          )}

          <div className="d-flex gap-2 mt-4">
            <button className="btn btn-primary" disabled={saving}>{saving ? 'Saving...' : 'Save'}</button>
            <button className="btn btn-outline-secondary" type="button" onClick={() => navigate(-1)}>Cancel</button>
          </div>
        </form>
      </section>
    </div>
  )
}

export default RoleForm
