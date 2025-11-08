import api from "./apiClient";
import { extractErrorMessage } from "./httpError";

export async function getNovels() {
  try {
    const res = await api.get("/api/novels");
    return res?.data?.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to fetch novels"));
  }
}

export async function getMyNovels() {
  try {
    const res = await api.get("/api/novels/me");
    return res?.data?.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to fetch novels"));
  }
}

export async function getNovelsByGenre(genreId) {
  try {
    const res = await api.get(`/api/novels/by-genre/${genreId}`);
    return res?.data?.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to fetch novels"));
  }
}

export async function getNovelById(id) {
  try {
    const res = await api.get(`/api/novels/${id}`);
    return res?.data?.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to fetch novel"));
  }
}

export async function createNovel(payload) {
  try {
    const {
      title,
      slug,
      synopsis,
      status,
      genreId,
      tagIds = [],
      coverFile,
    } = payload;
    const fd = new FormData();
    fd.append("title", title);
    fd.append("slug", slug);
    fd.append("synopsis", synopsis);
    fd.append("status", status);
    fd.append("genre_ids", JSON.stringify([Number(genreId)].filter(Boolean)));
    fd.append(
      "tag_ids",
      JSON.stringify(tagIds.map(Number).filter(Number.isFinite))
    );
    if (coverFile instanceof File) fd.append("cover_image", coverFile);

    const res = await api.post("/api/novels", fd, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return res?.data?.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to create novel"));
  }
}

export async function updateNovel(id, payload) {
  try {
    const {
      title,
      slug,
      synopsis,
      status,
      genreId,
      tagIds = [],
      coverFile,
    } = payload || {};

    const fd = new FormData();
    fd.append("title", title);
    fd.append("slug", slug);
    fd.append("synopsis", synopsis);
    fd.append("status", status);
    fd.append("genre_ids", JSON.stringify([Number(genreId)].filter(Boolean)));
    fd.append(
      "tag_ids",
      JSON.stringify(tagIds.map(Number).filter(Number.isFinite))
    );
    if (coverFile instanceof File) fd.append("cover_image", coverFile);

    const res = await api.put(`/api/novels/${id}`, fd, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return res?.data?.data;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to update novel"));
  }
}

export async function deleteNovel(id) {
  try {
    const res = await api.delete(`/api/novels/${id}`);
    return res.data?.data ?? true;
  } catch (err) {
    throw new Error(extractErrorMessage(err, "Failed to delete novel"));
  }
}
