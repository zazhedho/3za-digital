import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import appConfigService from '../../services/appConfigService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'

const ConfigForm = () => {
  const { id } = useParams()
  const navigate = useNavigate()
  const [form, setForm] = useState({ value: '', is_active: true })
  const [config, setConfig] = useState(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    setLoading(true)
    appConfigService.getById(id)
      .then((response) => {
        const item = response.data.data
        setConfig(item)
        setForm({ value: item.value || '', is_active: Boolean(item.is_active) })
      })
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load config')))
      .finally(() => setLoading(false))
  }, [id])

  const submit = async (event) => {
    event.preventDefault()
    setSaving(true)
    try {
      await appConfigService.update(id, form)
      toast.success('Configuration updated successfully')
      navigate(`/configs/${id}`, { replace: true })
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to save configuration'))
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <div className="loading-fade">Loading configuration...</div>

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>Edit Configuration</h1>
          <p>Update runtime parameters for <strong>{config?.display_name || config?.config_key}</strong></p>
        </div>
        <div className="toolbar-actions">
          <BackButton fallback="/configs" />
        </div>
      </div>

      <div className="content-grid single max-w-lg mx-auto">
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-pencil-square"></i> Update Value</h3>
          </div>
          <div className="luxe-card-body">
            <form onSubmit={submit} className="deposit-form-modern">
              <div className="deposit-input-group">
                <label>Config Value</label>
                <textarea 
                  className="form-control" 
                  rows="5" 
                  value={form.value} 
                  onChange={(event) => setForm({ ...form, value: event.target.value })}
                  placeholder="Enter configuration value..."
                  style={{ borderRadius: '14px', padding: '16px', background: '#fcfcfc' }}
                />
                <div className="form-hint">Ensure the value format matches the expected type (JSON, string, or number).</div>
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
                  <span className="form-check-label fw-bold ms-2">Enable this configuration</span>
                </label>
              </div>

              <div className="toolbar-actions justify-content-end mt-5 pt-3 border-top">
                <button className="btn btn-outline-dark px-4" type="button" onClick={() => navigate(-1)} disabled={saving}>
                  Cancel
                </button>
                <button className="btn btn-primary px-5" disabled={saving}>
                  {saving ? (
                    <>
                      <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                      Saving...
                    </>
                  ) : 'Save Changes'}
                </button>
              </div>
            </form>
          </div>
        </section>

        <section className="luxe-detail-card bg-light border-0 shadow-none">
           <div className="luxe-card-body">
              <div className="luxe-grid">
                 <div className="luxe-item">
                    <span className="luxe-label">Config Key</span>
                    <code className="luxe-value-code">{config?.config_key}</code>
                 </div>
                 <div className="luxe-item">
                    <span className="luxe-label">Category</span>
                    <span className="luxe-value text-capitalize">{config?.category}</span>
                 </div>
              </div>
           </div>
        </section>
      </div>
    </div>
  )
}

export default ConfigForm
