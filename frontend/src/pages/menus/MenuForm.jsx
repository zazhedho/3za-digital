import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import menuService from '../../services/menuService'
import { getErrorMessage } from '../../services/api'

const MenuForm = () => {
  const { id } = useParams()
  const navigate = useNavigate()
  const [form, setForm] = useState({ display_name: '', path: '', icon: '', order_index: 0, is_active: true })

  useEffect(() => {
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
  }, [id])

  const submit = async (event) => {
    event.preventDefault()
    try {
      await menuService.update(id, { ...form, order_index: Number(form.order_index) })
      toast.success('Menu updated')
      navigate(`/menus/${id}`)
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to save menu'))
    }
  }

  return (
    <div>
      <div className="page-toolbar">
        <div><h1>Edit Menu</h1><p>Navigation metadata.</p></div>
      </div>
      <section className="form-panel">
        <form onSubmit={submit}>
          <label className="form-label">Display name</label>
          <input className="form-control" value={form.display_name} onChange={(event) => setForm({ ...form, display_name: event.target.value })} required />
          <label className="form-label mt-3">Path</label>
          <input className="form-control" value={form.path} onChange={(event) => setForm({ ...form, path: event.target.value })} required />
          <label className="form-label mt-3">Icon</label>
          <input className="form-control" value={form.icon} onChange={(event) => setForm({ ...form, icon: event.target.value })} />
          <label className="form-label mt-3">Order index</label>
          <input className="form-control" type="number" value={form.order_index} onChange={(event) => setForm({ ...form, order_index: event.target.value })} />
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

export default MenuForm
