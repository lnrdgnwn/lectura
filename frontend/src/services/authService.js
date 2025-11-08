import api, { SKIP_AUTH_REFRESH_FLAG, runRefresh } from "./apiClient";
import { extractErrorMessage } from "./httpError";

const data = (res) => res?.data?.data ?? res?.data ?? null;

export async function getMe() {
  try {
    const res = await api.get("/api/users/me", {
      [SKIP_AUTH_REFRESH_FLAG]: true, 
      _retried: true, 
    });
    return data(res);
  } catch (err) {
    if (err?.response?.status === 401) return null; // belum login
    throw new Error(extractErrorMessage(err, "Failed to get profile"));
  }
}


export async function loginUser({ email, password }) {
  try {
    await api.post("/api/auth/login", { email, password });
    return await getMe();
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to login"));
  }
}


export async function registerUser({ username, email, password }) {
  try {
    await api.post("/api/auth/register", { username, email, password });
    return await loginUser({ email, password });
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to register"));
  }
}


export async function logoutUser() {
  try {
    const res = await api.post("/api/auth/logout");
    return data(res);
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to logout"));
  }
}

export async function refreshSession() {
  try {
    await runRefresh();
    const user = await getMe();
    return user;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to refresh session"));
  }
}

export async function bootstrapSession() {
  try {
    const first = await getMe();
    if (first) return first;

    await runRefresh();
    const after = await getMe();
    return after;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to bootstrap session"));
  }
}


export async function adminDeleteUser(id) {
  try {
    const res = await api.delete(`/api/admin/users/${id}`);
    return data(res);
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to delete user"));
  }
}
