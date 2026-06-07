import { useNavigate } from 'react-router-dom'

const BackButton = ({ fallback = '/dashboard', label = 'Back' }) => {
  const navigate = useNavigate()

  return (
    <button
      type="button"
      className="btn btn-outline-dark"
      onClick={() => navigate(fallback)}
    >
      <i className="bi bi-arrow-left me-2"></i>{label}
    </button>
  )
}

export default BackButton
