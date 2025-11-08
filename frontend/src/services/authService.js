// src/services/AuthService.js
import api, { SKIP_AUTH_REFRESH_FLAG, runRefresh } from "./apiClient";
import { extractErrorMessage } from "./httpError";

// Helper akses payload utama
const data = (res) => res?.data?.data ?? res?.data ?? null;

// Profil user yang sedang login
export async function getMe() {
  try {
    const res = await api.get("/api/users/me", {
      [SKIP_AUTH_REFRESH_FLAG]: true, // jangan auto-refresh
      _retried: true, // jangan di-retry oleh interceptor
    });
    return data(res);
  } catch (err) {
    if (err?.response?.status === 401) return null; // belum login
    throw new Error(extractErrorMessage(err, "Failed to get profile"));
  }
}

// Login
export async function loginUser({ email, password }) {
  try {
    await api.post("/api/auth/login", { email, password });
    return await getMe();
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to login"));
  }
}

// Register → auto login
export async function registerUser({ username, email, password }) {
  try {
    await api.post("/api/auth/register", { username, email, password });
    return await loginUser({ email, password });
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to register"));
  }
}

// Logout
export async function logoutUser() {
  try {
    const res = await api.post("/api/auth/logout");
    return data(res);
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to logout"));
  }
}

// Refresh session (cookie) → ambil user
export async function refreshSession() {
  try {
    await runRefresh();
    const user = await getMe();
    return user; // bisa null kalau memang tidak ada sesi
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to refresh session"));
  }
}

// Bootstrap di awal load: coba /me dulu, kalau null baru 1x refresh
export async function bootstrapSession() {
  try {
    const first = await getMe();
    if (first) return first;

    await runRefresh(); // coba sekali
    const after = await getMe(); // bisa null kalau tetap tidak ada user
    return after;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to bootstrap session"));
  }
}

// Admin delete
export async function adminDeleteUser(id) {
  try {
    const res = await api.delete(`/api/admin/users/${id}`);
    return data(res);
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to delete user"));
  }
}
