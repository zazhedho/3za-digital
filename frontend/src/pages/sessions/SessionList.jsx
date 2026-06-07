import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'react-toastify'
import ConfirmationModal from '../../components/common/ConfirmationModal'
import { useAuth } from '../../contexts/AuthContext'
import { getErrorMessage } from '../../services/api'
import sessionService from '../../services/sessionService'

const formatDate = (value) => {
  if (!value) return '-'
  return new Intl.DateTimeFormat('en-US', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

const relativeTime = (value) => {
  if (!value) return '-'
  const diffMs = Date.now() - new Date(value).getTime()
  if (Number.isNaN(diffMs)) return '-'
  const minutes = Math.floor(diffMs / 60000)
  const hours = Math.floor(diffMs / 3600000)
  const days = Math.floor(diffMs / 86400000)
  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes} minute${minutes === 1 ? '' : 's'} ago`
  if (hours < 24) return `${hours} hour${hours === 1 ? '' : 's'} ago`
  return `${days} day${days === 1 ? '' : 's'} ago`
}

const deviceIcon = (deviceInfo = '') => {
  const value = deviceInfo.toLowerCase()
  if (value.includes('mobile') || value.includes('android') || value.includes('iphone')) return 'bi-phone'
  if (value.includes('tablet') || value.includes('ipad')) return 'bi-tablet'
  if (value.includes('mac') || value.includes('windows') || value.includes('linux')) return 'bi-laptop'
  return 'bi-display'
}

const SessionList = () => {
  const navigate = useNavigate()
  const { logout } = useAuth()
  const [sessions, setSessions] = useState([])
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState(false)
  const [confirmSession, setConfirmSession] = useState(null)
  const [confirmAll, setConfirmAll] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const response = await sessionService.getActiveSessions()
      setSessions(response.data.data?.sessions || [])
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to load sessions'))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const sortedSessions = useMemo(() => (
    [...sessions].sort((a, b) => {
      if (a.is_current_session) return -1
      if (b.is_current_session) return 1
      return new Date(b.last_activity).getTime() - new Date(a.last_activity).getTime()
    })
  ), [sessions])

  const otherSessionsCount = sessions.filter((session) => !session.is_current_session).length

  const revokeSession = async () => {
    if (!confirmSession?.session_id) return
    setActionLoading(true)
    try {
      await sessionService.revokeSession(confirmSession.session_id)
      toast.success(confirmSession.is_current_session ? 'Session ended' : 'Session revoked')
      setConfirmSession(null)
      if (confirmSession.is_current_session) {
        await logout()
        navigate('/login', { replace: true })
        return
      }
      await load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to revoke session'))
    } finally {
      setActionLoading(false)
    }
  }

  const revokeAllOtherSessions = async () => {
    setActionLoading(true)
    try {
      await sessionService.revokeAllOtherSessions()
      toast.success('Other sessions revoked')
      setConfirmAll(false)
      await load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to revoke sessions'))
    } finally {
      setActionLoading(false)
    }
  }

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>Active Sessions</h1>
          <p>Manage login sessions across your devices.</p>
        </div>
        {otherSessionsCount > 0 && (
          <button className="btn btn-danger" type="button" onClick={() => setConfirmAll(true)}>
            <i className="bi bi-box-arrow-right me-2"></i>Logout other devices
          </button>
        )}
      </div>

      {loading ? (
        <section className="panel session-empty-panel">
          <div className="spinner-border text-primary" role="status" aria-label="Loading"></div>
        </section>
      ) : !sortedSessions.length ? (
        <section className="panel session-empty-panel">
          <i className="bi bi-globe2"></i>
          <h2>No Active Sessions</h2>
          <p>You do not have any active sessions.</p>
        </section>
      ) : (
        <section className="session-list">
          {sortedSessions.map((session) => (
            <article className={`session-card ${session.is_current_session ? 'current' : ''}`} key={session.session_id}>
              <div className="session-card-icon">
                <i className={`bi ${deviceIcon(session.device_info)}`}></i>
              </div>
              <div className="session-card-body">
                <div className="session-card-title">
                  <h2>{session.device_info || 'Unknown device'}</h2>
                  {session.is_current_session && (
                    <span className="status-badge status-paid">
                      Current Session
                    </span>
                  )}
                </div>
                <div className="session-meta-grid">
                  <span><i className="bi bi-globe2"></i>IP {session.ip || '-'}</span>
                  <span><i className="bi bi-clock-history"></i>Login {formatDate(session.login_at)}</span>
                  <span><i className="bi bi-activity"></i>Last activity {relativeTime(session.last_activity)}</span>
                </div>
              </div>
              <div className="session-card-action">
                <button
                  className={`btn btn-sm ${session.is_current_session ? 'btn-outline-dark' : 'btn-outline-danger'}`}
                  type="button"
                  onClick={() => setConfirmSession(session)}
                >
                  <i className="bi bi-trash me-2"></i>{session.is_current_session ? 'Logout' : 'Revoke'}
                </button>
              </div>
            </article>
          ))}
        </section>
      )}

      <ConfirmationModal
        show={Boolean(confirmSession)}
        title={confirmSession?.is_current_session ? 'Logout From This Device' : 'Revoke Session'}
        message={confirmSession?.is_current_session
          ? 'You will be logged out from this device and redirected to the login page.'
          : 'This device will be logged out and must sign in again.'}
        confirmLabel={confirmSession?.is_current_session ? 'Logout' : 'Revoke'}
        confirmClassName="btn-danger"
        loading={actionLoading}
        onCancel={() => setConfirmSession(null)}
        onConfirm={revokeSession}
      />

      <ConfirmationModal
        show={confirmAll}
        title="Logout Other Devices"
        message={`This will log you out from ${otherSessionsCount} other device${otherSessionsCount === 1 ? '' : 's'}. Current session stays active.`}
        confirmLabel="Logout others"
        confirmClassName="btn-danger"
        loading={actionLoading}
        onCancel={() => setConfirmAll(false)}
        onConfirm={revokeAllOtherSessions}
      />
    </div>
  )
}

export default SessionList
