// services/BookmarkService.js
import api from "./apiClient";
import { extractErrorMessage } from "./httpError";

/** Ambil semua bookmark milik user yang login */
export async function getBookmarks() {
  try {
    const res = await api.get("/api/bookmarks");
    // asumsi BE -> { data: [ { id, novel_id, novel? } ] }
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to fetch bookmarks"));
  }
}

/** Tambah bookmark untuk 1 novel */
export async function addBookmark(novel_id) {
  try {
    const res = await api.post("/api/bookmarks", { novel_id });
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to create bookmark"));
  }
}

/** Hapus bookmark by id */
export async function deleteBookmark(bookmarkId) {
  try {
    const res = await api.delete(`/api/bookmarks/${bookmarkId}`);
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to delete bookmark"));
  }
}
