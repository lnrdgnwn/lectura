export function extractErrorMessage(err, fallback = "Request failed") {
  try {
    const arr = err?.response?.data?.errors;
    const fromArray = Array.isArray(arr) && arr.length ? (arr[0]?.message || arr[0]) : null;
    return fromArray
      || err?.response?.data?.message
      || err?.response?.data?.error
      || err?.message
      || fallback;
  } catch {
    return fallback;
  }
}