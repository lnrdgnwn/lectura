import api from "./apiClient";
import { extractErrorMessage } from "./httpError";

export async function getGenres() {
  try {
    const res = await api.get("api/genres");
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to fetch genres"));
  }
}

export async function getHomeGenres() {
  try {
    const res = await api.get("/api/genres/home");
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to fetch home genres"));
  }
}

export async function setGenresShowOnHome(genreIds) {
  try {
    const res = await api.put("/api/admin/genres/home", {
      genre_ids: Array.isArray(genreIds) ? genreIds : [],
    });
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to update home genres"));
  }
}

export async function createGenre({ name, slug }) {
  try {
    const res = await api.post("/api/genres", { name, slug });
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to create genre"));
  }
}

export async function updateGenre(id, { name, slug }) {
  try {
    const res = await api.put(`/api/genres/${id}`, { name, slug });
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to update genre"));
  }
}

export async function deleteGenre(id) {
  try {
    const res = await api.delete(`/api/genres/${id}`);
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to delete genre"));
  }
}
