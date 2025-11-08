import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

export default function ProtectedAdminRoute() {
  const { isAuthenticated, initializing } = useAuth();
  const location = useLocation();
  if (initializing) return null;
  return isAuthenticated ? (
    <Outlet />
  ) : (
    <Navigate to="/admin-login" replace state={{ from: location }} />
  );
}
