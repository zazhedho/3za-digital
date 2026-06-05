import { useCallback, useEffect, useState } from 'react'
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

const getMinQuantity = (service) => Math.max(1, Number(service?.min_quantity || 1))

const getMaxQuantity = (service) => {
  const max = Number(service?.max_quantity || 0)
  return max > 0 ? max : undefined
}

const clampQuantity = (value, service) => {
  const min = getMinQuantity(service)
  const max = getMaxQuantity(service)
  const quantity = Math.trunc(Number(value))

  if (!Number.isFinite(quantity) || quantity < min) return min
  if (max && quantity > max) return max
  return quantity
}

const platformOptions = [
  { value: '', label: 'All platforms' },
  { value: 'instagram', label: 'Instagram' },
  { value: 'tiktok', label: 'TikTok' },
  { value: 'youtube', label: 'YouTube' },
  { value: 'facebook', label: 'Facebook' },
]

const SMMOrderForm = () => {
  const [services, setServices] = useState([])
  const [serviceSearch, setServiceSearch] = useState('')
  const [serviceLoading, setServiceLoading] = useState(false)
  const [platform, setPlatform] = useState('')
  const [form, setForm] = useState({ service_id: '', target: '', quantity: 1 })
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const selectedService = services.find((service) => service.id === form.service_id)
  const minQuantity = getMinQuantity(selectedService)
  const maxQuantity = getMaxQuantity(selectedService)

  const loadServices = useCallback(async (keyword = '') => {
    setServiceLoading(true)
    try {
      const response = await smmService.getServices({
        search: keyword,
        limit: 50,
        'filters[platform]': platform || undefined,
        'filters[is_active]': 'true',
      })
      setServices(getListPayload(response).rows)
    } catch {
      setServices([])
    } finally {
      setServiceLoading(false)
    }
  }, [platform])

  useEffect(() => {
    const timer = setTimeout(() => {
      loadServices(serviceSearch)
    }, 300)
    return () => clearTimeout(timer)
  }, [loadServices, serviceSearch])

  const handleServiceChange = (serviceId) => {
    const service = services.find((item) => item.id === serviceId)
    setForm((current) => ({
      ...current,
      service_id: serviceId,
      quantity: getMinQuantity(service),
    }))
  }

  const handlePlatformChange = (event) => {
    setPlatform(event.target.value)
    setServiceSearch('')
    setForm((current) => ({ ...current, service_id: '', quantity: 1 }))
  }

  const handleQuantityChange = (event) => {
    setForm((current) => ({
      ...current,
      quantity: clampQuantity(event.target.value, selectedService),
    }))
  }

  const submit = async (event) => {
    event.preventDefault()
    if (!form.service_id) {
      toast.error('Service is required')
      return
    }
    const quantity = clampQuantity(form.quantity, selectedService)
    setLoading(true)
    try {
      await smmService.createOrder({ ...form, quantity })
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
          <label className="form-label">Platform</label>
          <select className="form-select" value={platform} onChange={handlePlatformChange}>
            {platformOptions.map((option) => (
              <option key={option.value || 'all'} value={option.value}>{option.label}</option>
            ))}
          </select>

          <label className="form-label mt-3">Service</label>
          <SearchableSelect
            key={platform || 'all-platforms'}
            value={form.service_id}
            onChange={handleServiceChange}
            onSearch={setServiceSearch}
            loading={serviceLoading}
            placeholder="Select service"
            searchPlaceholder="Search service name or provider ID"
            emptyLabel={serviceSearch ? 'No matching service found' : 'No active services found'}
            options={services.map((service) => ({
              value: service.id,
              label: service.name,
              description: service.category || service.brand || 'SMM service',
              meta: [
                { label: 'ID', value: service.provider_service_id || '-' },
                { label: 'Platform', value: service.platform || '-' },
                { label: 'Price/1k', value: formatMoney(service.price) },
                { label: 'Qty', value: `${getMinQuantity(service)}-${getMaxQuantity(service) || '-'}` },
              ],
            }))}
          />

          <label className="form-label mt-3">Target URL</label>
          <input
            className="form-control"
            type="url"
            pattern="https?://.+"
            placeholder="https://instagram.com/username"
            value={form.target}
            onChange={(event) => setForm({ ...form, target: event.target.value })}
            required
          />

          <label className="form-label mt-3">Quantity</label>
          <input
            className="form-control"
            type="number"
            min={minQuantity}
            max={maxQuantity}
            value={form.quantity}
            onChange={handleQuantityChange}
            required
          />
          {selectedService && (
            <div className="form-text">
              Min {minQuantity}{maxQuantity ? ` / Max ${maxQuantity}` : ''}
            </div>
          )}

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
