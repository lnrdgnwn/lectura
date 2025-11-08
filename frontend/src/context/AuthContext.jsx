// context/AuthContext.jsx
import { createContext, useContext, useEffect, useMemo, useState } from "react";
import {
  loginUser,
  registerUser,
  logoutUser,
  bootstrapSession,
  refreshSession,
} from "../services/AuthService";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [initializing, setInitializing] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        const me = await bootstrapSession(); // returns user | null
        if (me) {
          setUser(me);
          setIsAuthenticated(true);
        } else {
          setUser(null);
          setIsAuthenticated(false);
        }
      } finally {
        setInitializing(false);
      }
    })();
  }, []);

  const login = async (credentials) => {
    const me = await loginUser(credentials);
    setUser(me);
    setIsAuthenticated(true);
    return me;
  };

  const register = async (payload) => {
    const me = await registerUser(payload);
    setUser(me);
    setIsAuthenticated(true);
    return me;
  };

  const refresh = async () => {
    const me = await refreshSession();
    if (me) {
      setUser(me);
      setIsAuthenticated(true);
      return me;
    }
    setUser(null);
    setIsAuthenticated(false);
    return null;
  };

  const logout = async () => {
    await logoutUser();
    setUser(null);
    setIsAuthenticated(false);
  };

  const isAdmin = !!(user?.role === "admin" || user?.is_admin === true);

  const value = useMemo(
    () => ({
      user,
      isAuthenticated,
      isAdmin,
      initializing,
      login,
      register,
      refresh,
      logout,
      setUser,
    }),
    [user, isAuthenticated, initializing]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
