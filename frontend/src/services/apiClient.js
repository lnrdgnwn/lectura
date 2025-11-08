// src/services/apiClient.js
import axios from "axios";

export const SKIP_AUTH_REFRESH_FLAG = "__skipAuthRefresh";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  timeout: 15000,
  withCredentials: true,
});

// Single-flight refresh promise agar tidak dobel
let refreshInFlight = null;

export const runRefresh = async () =>
  api.post("/api/auth/refresh", null, {
    [SKIP_AUTH_REFRESH_FLAG]: true, // jangan auto-refresh request ini
    _retried: true, // dan jangan di-retry ulang oleh interceptor
  });

api.interceptors.response.use(
  (res) => res,
  async (err) => {
    const { response, config } = err || {};
    if (!response || !config) return Promise.reject(err);

    const status = response.status;
    const url = config?.url || "";
    const isAuthRoute = url.includes("/api/auth/");
    const alreadyRetried = Boolean(config._retried);
    const skip = Boolean(config[SKIP_AUTH_REFRESH_FLAG]);

    // Hanya handle 401 untuk non-auth route, belum pernah di-retry, dan tidak di-skip
    if (status === 401 && !isAuthRoute && !alreadyRetried && !skip) {
      try {
        config._retried = true;

        if (!refreshInFlight) {
          refreshInFlight = runRefresh().finally(() => {
            refreshInFlight = null;
          });
        }

        await refreshInFlight;
        // ulang request asli setelah refresh berhasil
        return api(config);
      } catch {
        // gagal refresh → propagate error asli
        return Promise.reject(err);
      }
    }

    return Promise.reject(err);
  }
);

export default api;
