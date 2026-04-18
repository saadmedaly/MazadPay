import { Routes, Route, Navigate } from 'react-router-dom'
import { AdminLayout } from './components/layout/AdminLayout'
import { LoginPage } from './pages/LoginPage'
import { DashboardPage } from './pages/DashboardPage'
import { AuctionsPage } from './pages/AuctionsPage'
import { AuctionDetailPage } from './pages/AuctionDetailPage'
import { TransactionsPage } from './pages/TransactionsPage'
import { TransactionDetailPage } from './pages/TransactionDetailPage'
import { UsersPage } from './pages/UsersPage'
import { UserDetailPage } from './pages/UserDetailPage'
import { ReportsPage } from './pages/ReportsPage'
import { BannersPage } from './pages/BannersPage'
import { KYCPage } from './pages/KYCPage'
import { FAQPage } from './pages/FAQPage'
import { TutorialsPage } from './pages/TutorialsPage'
import { NotificationsPage } from './pages/NotificationsPage'
import { ProfilePage } from './pages/ProfilePage'
import { AdminInvitePage } from './pages/AdminInvitePage'
import { CategoriesPage } from './pages/CategoriesPage'
import { LocationsPage } from './pages/LocationsPage'
import { useAuthStore } from './stores/authStore'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, user } = useAuthStore()
  
  if (!isAuthenticated) return <Navigate to="/login" replace />
  if (user?.role !== 'admin') return <Navigate to="/login" replace />
  
  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      
      <Route path="/" element={
        <ProtectedRoute>
          <AdminLayout />
        </ProtectedRoute>
      }>
        <Route index element={<DashboardPage />} />
        
        <Route path="auctions" element={<AuctionsPage />} />
        <Route path="auctions/:id" element={<AuctionDetailPage />} />
        
        <Route path="transactions" element={<TransactionsPage />} />
        <Route path="transactions/:id" element={<TransactionDetailPage />} />
        
        <Route path="users" element={<UsersPage />} />
        <Route path="users/:id" element={<UserDetailPage />} />
        
        <Route path="reports" element={<ReportsPage />} />
        <Route path="banners" element={<BannersPage />} />
        <Route path="kyc" element={<KYCPage />} />
        <Route path="faq" element={<FAQPage />} />
        <Route path="tutorials" element={<TutorialsPage />} />
        <Route path="categories" element={<CategoriesPage />} />
        <Route path="locations" element={<LocationsPage />} />
        <Route path="notifications" element={<NotificationsPage />} />
        <Route path="profile" element={<ProfilePage />} />
      </Route>

      <Route path="/admin/register-admin" element={<AdminInvitePage />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
