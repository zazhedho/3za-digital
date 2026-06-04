import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'react-toastify'
import smmService from '../../services/smmService'
import { getErrorMessage, getListPayload } from '../../services/api'
import SearchableSelect from '../../components/common/SearchableSelect'

const formatMoney = (value) => new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
}).format(Number(value || 0))

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
    if (!form.service_id) {
      toast.error('Service is required')
      return
    }
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
          <SearchableSelect
            value={form.service_id}
            onChange={(serviceId) => setForm({ ...form, service_id: serviceId })}
            placeholder="Select service"
            searchPlaceholder="Search service name, platform, category"
            options={services.map((service) => ({
              value: service.id,
              label: service.name,
              description: service.category || service.brand || 'SMM service',
              meta: [
                { label: 'ID', value: service.provider_service_id || '-' },
                { label: 'Platform', value: service.platform || '-' },
                { label: 'Price/1k', value: formatMoney(service.price) },
              ],
            }))}
          />

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
