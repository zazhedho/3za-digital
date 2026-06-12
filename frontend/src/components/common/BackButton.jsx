import { useNavigate } from 'react-router-dom'

const BackButton = ({ fallback, label = 'Back' }) => {
  const navigate = useNavigate()

  const handleBack = () => {
    // If there is history, go back. Otherwise go to fallback or dashboard.
    if (window.history.length > 1) {
      navigate(-1)
    } else if (fallback) {
      navigate(fallback)
    } else {
      navigate('/dashboard')
    }
  }

  return (
    <button
      type="button"
      className="btn btn-outline-dark d-flex align-items-center gap-2"
      onClick={handleBack}
    >
      <i className="bi bi-arrow-left"></i>
      <span>{label}</span>
    </button>
  )
}

export default BackButton
