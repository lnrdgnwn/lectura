import axios from "axios";

export const SKIP_AUTH_REFRESH_FLAG = "__skipAuthRefresh";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  timeout: 15000,
  withCredentials: true,
});

let refreshInFlight = null;

export const runRefresh = async () =>
  api.post("/api/auth/refresh", null, {
    [SKIP_AUTH_REFRESH_FLAG]: true,
    _retried: true,
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

    if (status === 401 && !isAuthRoute && !alreadyRetried && !skip) {
      try {
        config._retried = true;

        if (!refreshInFlight) {
          refreshInFlight = runRefresh().finally(() => {
            refreshInFlight = null;
          });
        }
        await refreshInFlight;
        return api(config);
      } catch {
        return Promise.reject(err);
      }
    }

    return Promise.reject(err);
  }
);

export default api;
