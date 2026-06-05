import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import userService from '../../services/userService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'

const UserDetail = () => {
  const { id } = useParams()
  const { hasPermission } = useAuth()
  const [user, setUser] = useState(null)

  useEffect(() => {
    userService.getById(id)
      .then((response) => setUser(response.data.data))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load user')))
  }, [id])

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>User Detail</h1>
          <p>{user?.email || id}</p>
        </div>
        <div className="toolbar-actions">
          <BackButton fallback="/users" />
          {hasPermission('users', 'update') && <Link to={`/users/${id}/edit`} className="btn btn-primary">Edit</Link>}
        </div>
      </div>

      <section className="panel">
        <div className="detail-grid detail-grid-compact">
          <span>Name</span><strong>{user?.name || '-'}</strong>
          <span>Email</span><strong>{user?.email || '-'}</strong>
          <span>Phone</span><strong>{user?.phone || '-'}</strong>
          <span>Role</span><strong>{user?.role || '-'}</strong>
          <span>Last login</span><strong>{user?.last_login_at ? new Date(user.last_login_at).toLocaleString('id-ID') : '-'}</strong>
          <span>Created</span><strong>{user?.created_at ? new Date(user.created_at).toLocaleString('id-ID') : '-'}</strong>
        </div>
      </section>
    </div>
  )
}

export default UserDetail
