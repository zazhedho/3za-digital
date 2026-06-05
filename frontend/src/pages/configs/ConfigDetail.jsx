import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import appConfigService from '../../services/appConfigService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'

const ConfigDetail = () => {
  const { id } = useParams()
  const { hasPermission } = useAuth()
  const [config, setConfig] = useState(null)

  useEffect(() => {
    appConfigService.getById(id)
      .then((response) => setConfig(response.data.data))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load config')))
  }, [id])

  return (
    <div>
      <div className="page-toolbar">
        <div><h1>Config Detail</h1><p>{config?.display_name || id}</p></div>
        <div className="toolbar-actions">
          <BackButton fallback="/configs" />
          {hasPermission('configs', 'update') && <Link to={`/configs/${id}/edit`} className="btn btn-primary">Edit</Link>}
        </div>
      </div>
      <section className="panel">
        <div className="detail-grid detail-grid-compact">
          <span>Key</span><strong>{config?.config_key || '-'}</strong>
          <span>Display</span><strong>{config?.display_name || '-'}</strong>
          <span>Category</span><strong>{config?.category || '-'}</strong>
          <span>Value</span><strong>{config?.value || '-'}</strong>
          <span>Active</span><strong>{config?.is_active ? 'Yes' : 'No'}</strong>
          <span>Description</span><strong>{config?.description || '-'}</strong>
        </div>
      </section>
    </div>
  )
}

export default ConfigDetail
