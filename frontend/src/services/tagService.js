import api from "./apiClient";
import { extractErrorMessage } from "./httpError";

export async function getTags() {
  try {
    const res = await api.get("/api/tags");
    return res?.data?.data ?? [];
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to fetch tags"));
  }
}

export async function getTagById(id) {
  try {
    const res = await api.get(`/api/tags/${id}`);
    return res?.data?.data ?? null;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to fetch tag"));
  }
}

export async function createTag({ name, slug }) {
  try {
    const payload = {
      name: String(name ?? "").trim(),
      slug: String(slug ?? "")
        .trim()
        .toLowerCase(),
    };
    const res = await api.post("/api/tags", payload);
    return res?.data?.data ?? null;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to create tag"));
  }
}

export async function updateTag(id, { name, slug }) {
  try {
    const payload = {
      ...(name != null ? { name: String(name).trim() } : {}),
      ...(slug != null ? { slug: String(slug).trim().toLowerCase() } : {}),
    };
    const res = await api.put(`/api/tags/${id}`, payload);
    return res?.data?.data ?? null;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to update tag"));
  }
}

/** Delete tag */
export async function deleteTag(id) {
  try {
    await api.delete(`/api/tags/${id}`);
    return true;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to delete tag"));
  }
}
