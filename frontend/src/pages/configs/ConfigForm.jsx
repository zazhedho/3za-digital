import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import appConfigService from '../../services/appConfigService'
import { getErrorMessage } from '../../services/api'

const ConfigForm = () => {
  const { id } = useParams()
  const navigate = useNavigate()
  const [form, setForm] = useState({ value: '', is_active: true })
  const [config, setConfig] = useState(null)

  useEffect(() => {
    appConfigService.getById(id)
      .then((response) => {
        const item = response.data.data
        setConfig(item)
        setForm({ value: item.value || '', is_active: Boolean(item.is_active) })
      })
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load config')))
  }, [id])

  const submit = async (event) => {
    event.preventDefault()
    try {
      await appConfigService.update(id, form)
      toast.success('Config updated')
      navigate(`/configs/${id}`, { replace: true })
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to save config'))
    }
  }

  return (
    <div>
      <div className="page-toolbar">
        <div><h1>Edit Config</h1><p>{config?.display_name || id}</p></div>
      </div>
      <section className="form-panel">
        <form onSubmit={submit}>
          <label className="form-label">Value</label>
          <textarea className="form-control" rows="4" value={form.value} onChange={(event) => setForm({ ...form, value: event.target.value })} required />
          <label className="form-check mt-3">
            <input className="form-check-input" type="checkbox" checked={form.is_active} onChange={(event) => setForm({ ...form, is_active: event.target.checked })} />
            <span className="form-check-label">Active</span>
          </label>
          <div className="d-flex gap-2 mt-4">
            <button className="btn btn-primary">Save</button>
            <button className="btn btn-outline-secondary" type="button" onClick={() => navigate(-1)}>Cancel</button>
          </div>
        </form>
      </section>
    </div>
  )
}

export default ConfigForm
