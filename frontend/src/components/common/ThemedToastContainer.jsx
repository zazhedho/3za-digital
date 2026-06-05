import { ToastContainer, toast } from 'react-toastify'

const originalErrorToast = toast.error.bind(toast)

if (!toast.__errorToastRouted) {
  toast.error = (content, options = {}) => originalErrorToast(content, {
    containerId: 'error-center',
    ...options,
  })
  toast.__errorToastRouted = true
}

const ThemedToastContainer = (props) => (
  <>
    <ToastContainer
      position="top-right"
      autoClose={3000}
      theme="light"
      {...props}
    />
    <ToastContainer
      containerId="error-center"
      position="top-center"
      autoClose={5000}
      theme="light"
      className="error-toast-container"
      {...props}
    />
  </>
)

export default ThemedToastContainer
