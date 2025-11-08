import api from "./apiClient";
import { extractErrorMessage } from "./httpError";

export async function getBookmarks() {
  try {
    const res = await api.get("/api/bookmarks");
    console.log(res);
    console.log(res.data.data);
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to fetch bookmarks"));
  }
}

export async function addBookmark(novel_id) {
  try {
    const res = await api.post("/api/bookmarks", { novel_id: 
      novel_id });
    return res.data?.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to create bookmark"));
  }
}

export async function deleteBookmark(bookmarkId) {
  try {
    await api.delete(`/api/bookmarks/${bookmarkId}`);
    return true;
  } catch (err) {
    console.error("Failed to delete bookmark:", err);
    throw new Error(extractErrorMessage(err, "Failed to delete bookmark"));
  }
}
