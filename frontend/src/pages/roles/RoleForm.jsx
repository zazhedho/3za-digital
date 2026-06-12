import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import roleService from '../../services/roleService'
import permissionService from '../../services/permissionService'
import { getErrorMessage, getListPayload } from '../../services/api'
import { useAuth } from '../../contexts/AuthContext'
import BackButton from '../../components/common/BackButton'

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
    if (isEdit) return
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
      toast.error('Please select at least one permission')
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
      toast.success(isEdit ? 'Role updated' : 'New role created')
      navigate(isEdit ? `/roles/${roleId}` : '/roles', { replace: isEdit })
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

  if (loading) return <div className="loading-fade">Loading role permissions...</div>

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>{isEdit ? 'Edit Role' : 'Create Role'}</h1>
          <p>Define access levels and system permissions.</p>
        </div>
        <div className="toolbar-actions">
           <BackButton fallback="/roles" />
        </div>
      </div>

      <form onSubmit={submit}>
        <div className="content-grid two">
          <section className="luxe-detail-card h-fit">
            <div className="luxe-card-header">
              <h3><i className="bi bi-info-circle"></i> Role Metadata</h3>
            </div>
            <div className="luxe-card-body">
              <div className="deposit-form-modern">
                <div className="deposit-input-group">
                  <label>Unique Name</label>
                  <div className="auth-input m-0">
                    <i className="bi bi-tag"></i>
                    <input
                      value={form.name}
                      onChange={(event) => setForm({ ...form, name: event.target.value.toLowerCase().replace(/\s+/g, '') })}
                      required
                      disabled={isEdit || isSystem}
                      placeholder="e.g. support_staff"
                      style={{ background: isEdit || isSystem ? '#f8fafc' : 'white' }}
                    />
                  </div>
                </div>

                <div className="deposit-input-group mt-3">
                  <label>Display Name</label>
                  <div className="auth-input m-0">
                    <i className="bi bi-type-strikethrough"></i>
                    <input 
                      value={form.display_name} 
                      onChange={(event) => setForm({ ...form, display_name: event.target.value })} 
                      required 
                      disabled={isSystem}
                      placeholder="e.g. Support Staff"
                      style={{ background: isSystem ? '#f8fafc' : 'white' }}
                    />
                  </div>
                </div>

                <div className="deposit-input-group mt-3">
                  <label>Description</label>
                  <textarea 
                    className="form-control" 
                    rows="3"
                    value={form.description} 
                    onChange={(event) => setForm({ ...form, description: event.target.value })} 
                    disabled={isSystem}
                    placeholder="Describe what this role can do..."
                    style={{ borderRadius: '14px', padding: '12px', background: isSystem ? '#f8fafc' : 'white' }}
                  />
                </div>
                
                {isSystem && (
                   <div className="auth-alert mt-4 mb-0">
                      <i className="bi bi-shield-lock me-2"></i>
                      System roles are protected and cannot be fully modified.
                   </div>
                )}
              </div>
            </div>
          </section>

          {canAssignPermissions && (
            <section className="luxe-detail-card">
              <div className="luxe-card-header">
                <div>
                  <h3><i className="bi bi-shield-check"></i> Permissions</h3>
                  <small className="text-muted">{selectedPermissions.length} / {permissions.length} selected</small>
                </div>
                <div className="toolbar-actions">
                  <button className="btn btn-sm btn-outline-dark" type="button" onClick={selectAllPermissions}>All</button>
                  <button className="btn btn-sm btn-outline-dark" type="button" onClick={clearPermissions}>Clear</button>
                </div>
              </div>

              <div className="luxe-card-body p-0">
                <div className="px-4 pt-3 pb-2 border-bottom bg-light">
                   <div className="auth-input m-0" style={{ height: '40px' }}>
                      <i className="bi bi-search" style={{ fontSize: '14px' }}></i>
                      <input 
                         value={permissionSearch} 
                         onChange={(event) => setPermissionSearch(event.target.value)} 
                         placeholder="Filter permissions..." 
                         style={{ fontSize: '14px' }}
                      />
                   </div>
                </div>

                <div className="permission-group-list-luxe" style={{ maxHeight: '600px', overflow: 'auto', padding: '12px 24px' }}>
                  {Object.keys(groupedPermissions).sort().map((resource) => (
                    <div className="permission-group-box mb-4" key={resource}>
                      <label className="d-flex justify-content-between align-items-center mb-2 pb-2 border-bottom">
                        <div className="d-flex align-items-center gap-2 cursor-pointer">
                          <input
                            className="form-check-input m-0"
                            type="checkbox"
                            checked={groupedPermissions[resource].every((permission) => selectedPermissions.includes(permission.id))}
                            onChange={() => toggleResourcePermissions(groupedPermissions[resource])}
                          />
                          <strong className="text-capitalize" style={{ fontSize: '15px' }}>{formatResource(resource)}</strong>
                        </div>
                        <span className="status-badge status-badge-sm info">{groupedPermissions[resource].length}</span>
                      </label>
                      <div className="d-grid gap-2" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))' }}>
                        {groupedPermissions[resource].map((permission) => (
                          <label className={`permission-item-luxe ${selectedPermissions.includes(permission.id) ? 'selected' : ''}`} key={permission.id}>
                            <input
                              type="checkbox"
                              hidden
                              checked={selectedPermissions.includes(permission.id)}
                              onChange={() => togglePermission(permission.id)}
                            />
                            <div className="d-flex align-items-center gap-2">
                               <i className={`bi ${selectedPermissions.includes(permission.id) ? 'bi-check-circle-fill text-success' : 'bi-circle text-muted'}`}></i>
                               <div className="d-flex flex-column">
                                  <strong style={{ fontSize: '13px' }}>{permission.action || permission.name}</strong>
                                  <small className="text-muted" style={{ fontSize: '11px' }}>{permission.display_name}</small>
                               </div>
                            </div>
                          </label>
                        ))}
                      </div>
                    </div>
                  ))}
                  {!filteredPermissions.length && (
                    <div className="text-center py-5">
                       <i className="bi bi-search text-muted display-4"></i>
                       <p className="mt-2 text-muted">No permissions matching your search.</p>
                    </div>
                  )}
                </div>
              </div>
            </section>
          )}
        </div>

        <div className="toolbar-actions justify-content-center mt-4">
           <button className="btn btn-outline-dark px-5" type="button" onClick={() => navigate(-1)} disabled={saving}>Cancel</button>
           <button className="btn btn-primary px-5" disabled={saving}>
              {saving ? 'Saving...' : 'Save Role Configuration'}
           </button>
        </div>
      </form>
    </div>
  )
}

export default RoleForm
