import { ToastContainer } from 'react-toastify'

const ThemedToastContainer = (props) => (
  <ToastContainer position="top-right" autoClose={3000} theme="light" {...props} />
)

export default ThemedToastContainer
