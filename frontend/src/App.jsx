import { lazy, Suspense } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import ProtectedLayout from './components/common/ProtectedLayout'
import PermissionRoute from './components/common/PermissionRoute'
import Loading from './components/common/Loading'
import ThemedToastContainer from './components/common/ThemedToastContainer'
import { AuthProvider } from './contexts/AuthContext'

const Login = lazy(() => import('./pages/auth/Login'))
const Register = lazy(() => import('./pages/auth/Register'))
const ForgotPassword = lazy(() => import('./pages/auth/ForgotPassword'))
const ResetPassword = lazy(() => import('./pages/auth/ResetPassword'))
const Dashboard = lazy(() => import('./pages/Dashboard'))
const SMMServices = lazy(() => import('./pages/smm/SMMServices'))
const SMMOrders = lazy(() => import('./pages/smm/SMMOrders'))
const SMMOrderDetail = lazy(() => import('./pages/smm/SMMOrderDetail'))
const SMMOrderForm = lazy(() => import('./pages/smm/SMMOrderForm'))
const Wallet = lazy(() => import('./pages/wallet/Wallet'))
const DepositList = lazy(() => import('./pages/wallet/DepositList'))
const DepositDetail = lazy(() => import('./pages/wallet/DepositDetail'))
const AdminWallets = lazy(() => import('./pages/admin/AdminWallets'))
const AdminDeposits = lazy(() => import('./pages/admin/AdminDeposits'))
const AdminDepositDetail = lazy(() => import('./pages/admin/AdminDepositDetail'))
const UserList = lazy(() => import('./pages/users/UserList'))
const UserForm = lazy(() => import('./pages/users/UserForm'))
const UserDetail = lazy(() => import('./pages/users/UserDetail'))
const RoleList = lazy(() => import('./pages/roles/RoleList'))
const RoleForm = lazy(() => import('./pages/roles/RoleForm'))
const RoleDetail = lazy(() => import('./pages/roles/RoleDetail'))
const ConfigList = lazy(() => import('./pages/configs/ConfigList'))
const ConfigForm = lazy(() => import('./pages/configs/ConfigForm'))
const ConfigDetail = lazy(() => import('./pages/configs/ConfigDetail'))
const MenuList = lazy(() => import('./pages/menus/MenuList'))
const MenuForm = lazy(() => import('./pages/menus/MenuForm'))
const MenuDetail = lazy(() => import('./pages/menus/MenuDetail'))
const AuditList = lazy(() => import('./pages/audits/AuditList'))
const AuditDetail = lazy(() => import('./pages/audits/AuditDetail'))
const Profile = lazy(() => import('./pages/users/Profile'))

const Guard = ({ resource, action, children }) => (
  <PermissionRoute resource={resource} action={action}>{children}</PermissionRoute>
)

const ErrorPage = ({ code, title }) => (
  <div className="empty-screen">
    <h1>{code}</h1>
    <h2>{title}</h2>
    <a href="/dashboard" className="btn btn-primary">Back to dashboard</a>
  </div>
)

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Suspense fallback={<Loading />}>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/forgot-password" element={<ForgotPassword />} />
            <Route path="/reset-password" element={<ResetPassword />} />
            <Route path="/" element={<Navigate to="/dashboard" replace />} />

            <Route element={<ProtectedLayout />}>
              <Route path="/dashboard" element={<Guard resource="dashboard" action="view"><Dashboard /></Guard>} />
              <Route path="/smm/services" element={<Guard resource="smm_services" action="list"><SMMServices /></Guard>} />
              <Route path="/smm/orders" element={<Guard resource="smm_orders" action="list"><SMMOrders /></Guard>} />
              <Route path="/smm/orders/new" element={<Guard resource="smm_orders" action="create"><SMMOrderForm /></Guard>} />
              <Route path="/smm/orders/:id" element={<Guard resource="smm_orders" action="view"><SMMOrderDetail /></Guard>} />
              <Route path="/wallet" element={<Guard resource="wallet" action="view"><Wallet /></Guard>} />
              <Route path="/deposits" element={<Guard resource="deposits" action="list"><DepositList /></Guard>} />
              <Route path="/deposits/:id" element={<Guard resource="deposits" action="view"><DepositDetail /></Guard>} />
              <Route path="/admin/wallets" element={<Guard resource="wallets" action="list"><AdminWallets /></Guard>} />
              <Route path="/admin/deposits" element={<Guard resource="admin_deposits" action="list"><AdminDeposits /></Guard>} />
              <Route path="/admin/deposits/:id" element={<Guard resource="admin_deposits" action="view"><AdminDepositDetail /></Guard>} />
              <Route path="/users" element={<Guard resource="users" action="list"><UserList /></Guard>} />
              <Route path="/users/new" element={<Guard resource="users" action="create"><UserForm /></Guard>} />
              <Route path="/users/:id" element={<Guard resource="users" action="view"><UserDetail /></Guard>} />
              <Route path="/users/:id/edit" element={<Guard resource="users" action="update"><UserForm /></Guard>} />
              <Route path="/roles" element={<Guard resource="roles" action="list"><RoleList /></Guard>} />
              <Route path="/roles/new" element={<Guard resource="roles" action="create"><RoleForm /></Guard>} />
              <Route path="/roles/:id" element={<Guard resource="roles" action="view"><RoleDetail /></Guard>} />
              <Route path="/roles/:id/edit" element={<Guard resource="roles" action="update"><RoleForm /></Guard>} />
              <Route path="/configs" element={<Guard resource="configs" action="list"><ConfigList /></Guard>} />
              <Route path="/configs/:id" element={<Guard resource="configs" action="view"><ConfigDetail /></Guard>} />
              <Route path="/configs/:id/edit" element={<Guard resource="configs" action="update"><ConfigForm /></Guard>} />
              <Route path="/menus" element={<Guard resource="menus" action="list"><MenuList /></Guard>} />
              <Route path="/menus/:id" element={<Guard resource="menus" action="view"><MenuDetail /></Guard>} />
              <Route path="/menus/:id/edit" element={<Guard resource="menus" action="update"><MenuForm /></Guard>} />
              <Route path="/audits" element={<Guard resource="audits" action="list"><AuditList /></Guard>} />
              <Route path="/audits/:id" element={<Guard resource="audits" action="view"><AuditDetail /></Guard>} />
              <Route path="/profile" element={<Profile />} />
            </Route>

            <Route path="/unauthorized" element={<ErrorPage code="403" title="Unauthorized access" />} />
            <Route path="*" element={<ErrorPage code="404" title="Page not found" />} />
          </Routes>
        </Suspense>
        <ThemedToastContainer />
      </AuthProvider>
    </BrowserRouter>
  )
}

export default App
