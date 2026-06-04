import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'react-toastify'
import smmService from '../../services/smmService'
import { getErrorMessage, getListPayload } from '../../services/api'

const SMMOrderForm = () => {
  const [services, setServices] = useState([])
  const [form, setForm] = useState({ service_id: '', target: '', quantity: 1 })
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    smmService.getServices({ limit: 100, 'filters[is_active]': 'true' })
      .then((response) => setServices(getListPayload(response).rows))
      .catch(() => setServices([]))
  }, [])

  const submit = async (event) => {
    event.preventDefault()
    setLoading(true)
    try {
      await smmService.createOrder({ ...form, quantity: Number(form.quantity) })
      toast.success('Order created')
      navigate('/smm/orders')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to create order'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>Create SMM Order</h1>
          <p>Choose active service, target URL/user, and quantity.</p>
        </div>
      </div>

      <section className="form-panel">
        <form onSubmit={submit}>
          <label className="form-label">Service</label>
          <select className="form-select" value={form.service_id} onChange={(event) => setForm({ ...form, service_id: event.target.value })} required>
            <option value="">Select service</option>
            {services.map((service) => (
              <option value={service.id} key={service.id}>{service.name} - {service.price}</option>
            ))}
          </select>

          <label className="form-label mt-3">Target</label>
          <input className="form-control" value={form.target} onChange={(event) => setForm({ ...form, target: event.target.value })} required />

          <label className="form-label mt-3">Quantity</label>
          <input className="form-control" type="number" min="1" value={form.quantity} onChange={(event) => setForm({ ...form, quantity: event.target.value })} required />

          <div className="d-flex gap-2 mt-4">
            <button className="btn btn-primary" disabled={loading}>{loading ? 'Creating...' : 'Create order'}</button>
            <button className="btn btn-outline-secondary" type="button" onClick={() => navigate('/smm/orders')}>Cancel</button>
          </div>
        </form>
      </section>
    </div>
  )
}

export default SMMOrderForm
