import api from "./apiClient";
import { extractErrorMessage } from "./httpError";

export async function getChaptersByNovel(novelId) {
  try {
    const res = await api.get(`/api/novels/${novelId}/chapter`);
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to fetch chapters"));
  }
}

export async function getChapterById(id) {
  try {
    const res = await api.get(`/api/chapters/${id}`);
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to fetch chapter"));
  }
}


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


export async function deleteChapter(id) {
  try {
    const res = await api.delete(`/api/chapters/${id}`);
    return res.data.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to delete chapter"));
  }
}


export async function setChapterPublish(id, publish) {
  try {
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
