// services/ChapterService.js
import api from "./apiClient";
import { extractErrorMessage } from "./httpError";

/** List chapters untuk 1 novel (urut dari BE kalau sudah diset) */
export async function getChaptersByNovel(novelId) {
  try {
    // di sidebar kamu ada: GET /novels/{id}/chapter
    const res = await api.get(`/api/novels/${novelId}/chapter`);
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to fetch chapters"));
  }
}

/** Ambil 1 chapter by id */
export async function getChapterById(id) {
  try {
    const res = await api.get(`/api/chapters/${id}`);
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to fetch chapter"));
  }
}

/** Buat chapter baru
 *  payload contoh dari kamu:
 *  { novel_id: 1, title: "Chapter 1", content: "....", publish: false|true }
 */
export async function createChapter({
  novel_id,
  title,
  content,
  publish = false,
}) {
  try {
    const body = { novel_id, title, content, publish: !!publish };
    const res = await api.post("/api/chapters", body);
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to create chapter"));
  }
}

/** Update chapter (order_no/title/content) — sesuai contohmu */
export async function updateChapter(id, { order_no, title, content }) {
  try {
    const body = {};
    if (order_no != null) body.order_no = order_no;
    if (title != null) body.title = title;
    if (content != null) body.content = content;

    const res = await api.put(`/api/chapters/${id}`, body);
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to update chapter"));
  }
}

/** Hapus chapter */
export async function deleteChapter(id) {
  try {
    const res = await api.delete(`/api/chapters/${id}`);
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to delete chapter"));
  }
}

/** Convenience: publish / unpublish (kalau BE menerima flag `publish`) */
export async function setChapterPublish(id, publish) {
  try {
    // Jika BE-mu tidak menerima ini via PUT, hapus fungsi ini.
    const res = await api.put(`/api/chapters/${id}`, { publish: !!publish });
    return res.data.data;
  } catch (err) {
    throw new Error(
      extractErrorMessage(
        err,
        `Failed to ${publish ? "publish" : "unpublish"} chapter`
      )
    );
  }
}
