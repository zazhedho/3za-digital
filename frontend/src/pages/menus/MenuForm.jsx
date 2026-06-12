import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import menuService from '../../services/menuService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'

const MenuForm = () => {
  const { id } = useParams()
  const navigate = useNavigate()
  const [form, setForm] = useState({ display_name: '', path: '', icon: '', order_index: 0, is_active: true })
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    setLoading(true)
    menuService.getById(id)
      .then((response) => {
        const item = response.data.data
        setForm({
          display_name: item.display_name || '',
          path: item.path || '',
          icon: item.icon || '',
          order_index: item.order_index || 0,
          is_active: Boolean(item.is_active),
        })
      })
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load menu')))
      .finally(() => setLoading(false))
  }, [id])

  const submit = async (event) => {
    event.preventDefault()
    setSaving(true)
    try {
      await menuService.update(id, { ...form, order_index: Number(form.order_index) })
      toast.success('Menu updated successfully')
      navigate(`/menus/${id}`, { replace: true })
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to save menu'))
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <div className="loading-fade">Loading menu settings...</div>

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>Edit Menu Entry</h1>
          <p>Modify navigation properties for <strong>{form.display_name}</strong></p>
        </div>
        <div className="toolbar-actions">
          <BackButton fallback="/menus" />
        </div>
      </div>

      <div className="content-grid single max-w-lg mx-auto">
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-layout-sidebar-inset"></i> Menu Configuration</h3>
          </div>
          <div className="luxe-card-body">
            <form onSubmit={submit} className="deposit-form-modern">
              <div className="deposit-input-group">
                <label>Display Name</label>
                <div className="auth-input m-0">
                   <i className="bi bi-fonts"></i>
                   <input 
                     value={form.display_name} 
                     onChange={(event) => setForm({ ...form, display_name: event.target.value })} 
                     placeholder="e.g. Dashboard"
                     required 
                     style={{ background: 'white' }}
                   />
                </div>
              </div>

              <div className="deposit-input-group mt-3">
                <label>Route Path</label>
                <div className="auth-input m-0">
                   <i className="bi bi-link-45deg"></i>
                   <input 
                     value={form.path} 
                     onChange={(event) => setForm({ ...form, path: event.target.value })} 
                     placeholder="e.g. /dashboard"
                     required 
                     style={{ background: 'white' }}
                   />
                </div>
              </div>

              <div className="content-grid two mt-3 gap-3">
                <div className="deposit-input-group">
                  <label>Bootstrap Icon</label>
                  <div className="auth-input m-0">
                    <i className={`bi bi-${form.icon || 'circle'}`}></i>
                    <input 
                      value={form.icon} 
                      onChange={(event) => setForm({ ...form, icon: event.target.value })} 
                      placeholder="e.g. speedometer2"
                      style={{ background: 'white' }}
                    />
                  </div>
                </div>
                <div className="deposit-input-group">
                  <label>Sort Order</label>
                  <div className="auth-input m-0">
                    <i className="bi bi-sort-numeric-down"></i>
                    <input 
                      type="number"
                      value={form.order_index} 
                      onChange={(event) => setForm({ ...form, order_index: event.target.value })} 
                      style={{ background: 'white' }}
                    />
                  </div>
                </div>
              </div>

              <div className="mt-4">
                <label className="form-check form-switch luxe-switch">
                  <input 
                    className="form-check-input" 
                    type="checkbox" 
                    role="switch"
                    checked={form.is_active} 
                    onChange={(event) => setForm({ ...form, is_active: event.target.checked })} 
                  />
                  <span className="form-check-label fw-bold ms-2">Visible in Sidebar</span>
                </label>
              </div>

              <div className="toolbar-actions justify-content-end mt-5 pt-3 border-top">
                <button className="btn btn-outline-dark px-4" type="button" onClick={() => navigate(-1)} disabled={saving}>
                  Cancel
                </button>
                <button className="btn btn-primary px-5" disabled={saving}>
                   {saving ? 'Saving...' : 'Update Menu'}
                </button>
              </div>
            </form>
          </div>
        </section>
      </div>
    </div>
  )
}

export default MenuForm
