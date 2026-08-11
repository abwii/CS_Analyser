import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { HomePage } from './pages/HomePage'
import AdminPage from './pages/AdminPage'
import { LoginPage } from './pages/LoginPage'
import { VetoPage } from './pages/VetoPage'
import { PlayerProfilePage } from './pages/PlayerProfilePage'
import { ComparisonPage } from './pages/ComparisonPage'
import { DemosPage } from './pages/DemosPage'
import { SteamCallbackPage } from './pages/SteamCallbackPage'
import { NavBar } from './components/NavBar'

function App() {
  return (
    <BrowserRouter>
      <NavBar />
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/admin" element={<AdminPage />} />
        <Route path="/demos" element={<DemosPage />} />
        <Route path="/auth/steam/callback" element={<SteamCallbackPage />} />
        <Route path="/veto/:sessionId" element={<VetoPage />} />
        <Route path="/players/:playerId" element={<PlayerProfilePage />} />
        <Route path="/compare/:playerAId/:playerBId" element={<ComparisonPage />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
