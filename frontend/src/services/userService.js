import api from "./apiClient";
import { extractErrorMessage } from "./httpError";

export async function getUsers() {
  try {
    const { data } = await api.get("/api/admin/users");
    return data;
  } catch (err) {
    const msg = extractErrorMessage(err, "Failed to fetch users");
    console.error("[getUsers]", msg, err);
    throw new Error(msg);
  }
}

export async function deleteUser(id) {
  try {
    await api.delete(`/api/admin/users/${id}`);
    return true;
  } catch (err) {
    const msg = extractErrorMessage(err, "Failed to delete user");
    console.error("[deleteUser]", msg, err);
    throw new Error(msg);
  }
}

export async function getMyProfile() {
  try {
    const res = await api.get("/api/users/me");
    return res.data.data;
  } catch (err) {
    const msg = extractErrorMessage(err, "Failed to fetch profile");
    console.error("[getMyProfile]", msg, err);
    throw new Error(msg);
  }
}

export async function updateMyProfile({ username, email, avatarFile } = {}) {
  try {
    if (avatarFile) {
      const form = new FormData();
      if (username !== undefined) form.append("username", username);
      if (email !== undefined) form.append("email", email);
      form.append("profile_picture", avatarFile);

      const res = await api.put("/api/users/me", form, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      return res?.data?.data ?? res?.data ?? null;
    }

    const res = await api.put("/api/users/me", { username, email });
    return res?.data?.data ?? res?.data ?? null;
  } catch (err) {
    const msg = extractErrorMessage(err, "Failed to update profile");
    console.error("[updateMyProfile]", msg, err);
    throw new Error(msg);
  }
}

export async function changeMyPassword({ old_password, new_password }) {
  try {
    const res = await api.put("/api/users/me/changepassword", {
      old_password,
      new_password,
    });
    return res.data.success;
  } catch (err) {
    const msg = extractErrorMessage(err, "Failed to change password");
    console.error("[changeMyPassword]", msg, err);
    throw new Error(msg);
  }
}
