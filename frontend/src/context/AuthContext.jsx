// context/AuthContext.jsx
import { createContext, useContext, useEffect, useMemo, useState } from "react";
import {
  loginUser,
  registerUser,
  logoutUser,
  bootstrapSession,
  refreshSession,
} from "../services/AuthService"; // <-- pastikan kapitalisasi sama dengan file-mu
// (opsional) hapus getMyProfile kalau tidak dipakai

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
    const me = await loginUser(credentials); // returns user
    setUser(me);
    setIsAuthenticated(true);
    return me;
  };

  const register = async (payload) => {
    const me = await registerUser(payload); // returns user
    setUser(me);
    setIsAuthenticated(true);
    return me;
  };

  const refresh = async () => {
    const me = await refreshSession(); // returns user | null
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

  const value = useMemo(
    () => ({
      user,
      isAuthenticated,
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
