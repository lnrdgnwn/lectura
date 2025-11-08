// src/pages/UserSettings.jsx
import { useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { FiArrowLeft, FiCheck } from "react-icons/fi";
import Navbar from "../../components/Navbar";
import Footer from "../../components/Footer";
import { useAuth } from "../../context/AuthContext";
import { updateMyProfile } from "../../services/userService";
import useScrollToTop from "../../hooks/useScrollToTop";
import toast from "react-hot-toast";

export default function UserSettings() {
  useScrollToTop();
  const { user, reloadProfile } = useAuth();
  const navigate = useNavigate();
  const formRef = useRef(null);

  const [username, setUsername] = useState(user?.username || "");
  const [email, setEmail] = useState(user?.email || "");
  const [avatarFile, setAvatarFile] = useState(null);
  const [preview, setPreview] = useState(user?.profile_picture);
  const [saving, setSaving] = useState(false);

  if (!user) return null;

  const onPick = (e) => {
    const f = e.target.files?.[0];
    if (f) {
      setAvatarFile(f);
      setPreview(URL.createObjectURL(f));
    }
  };

  const submit = async (e) => {
    e?.preventDefault?.();
    if (saving) return;
    setSaving(true);
    try {
      await toast.promise(updateMyProfile({ username, email, avatarFile }), {
        loading: "Updating profile…",
        success: "Profile updated!",
        error: (err) => err?.message || "Failed to update profile",
      });
      await reloadProfile();
    } finally {
      setSaving(false);
    }
  };

  return (
    <>
      <Navbar />
      <main className="min-h-screen bg-gray-50 pb-10">
        <div className="mx-auto md:p-16 w-full max-w-5xl px-4">
          {/* Mobile sub-navbar */}
          <div className="md:hidden sticky top-[65px] z-40 -mx-4 mb-4 px-4 py-3 bg-primary text-white flex items-center justify-between">
            <button
              onClick={() => navigate("/profile")}
              className="text-white/90 text-xl p-1"
              aria-label="Back"
            >
              <FiArrowLeft />
            </button>
            <h1 className="text-base font-semibold">Edit profile</h1>
            <button
              onClick={() => formRef.current?.requestSubmit()}
              disabled={saving}
              className="text-xl p-1 disabled:opacity-60"
              aria-label="Save"
            >
              <FiCheck />
            </button>
          </div>

          <div className="grid md:grid-cols-12 gap-6">
            {/* Left */}
            <aside className="md:col-span-4">
              <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-6 flex flex-col items-center">
                <img
                  src={preview}
                  alt={username}
                  className="h-28 w-28 rounded-full object-cover ring-1 ring-gray-200"
                />
                <label className="mt-4 inline-flex items-center justify-center h-10 px-4 rounded-full border border-gray-300 text-sm font-medium text-gray-700 hover:bg-gray-50 cursor-pointer">
                  <input
                    type="file"
                    className="hidden"
                    accept="image/*"
                    onChange={onPick}
                  />
                  Upload picture
                </label>
                {user.id && (
                  <p className="mt-5 text-sm text-gray-500">ID {user.id}</p>
                )}
              </div>
            </aside>

            {/* Right */}
            <section className="md:col-span-8">
              <div className="bg-white rounded-2xl shadow-sm border border-gray-100">
                <form
                  ref={formRef}
                  onSubmit={submit}
                  className="px-5 md:px-6 py-6 space-y-5"
                >
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Username
                    </label>
                    <input
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      className="w-full h-11 rounded-xl border border-gray-300 px-4 focus:outline-none focus:ring-2 focus:ring-primary"
                      required
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Email address
                    </label>
                    <input
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      className="w-full h-11 rounded-xl border border-gray-300 px-4 focus:outline-none focus:ring-2 focus:ring-primary"
                      required
                    />
                  </div>

                  {/* Desktop actions */}
                  <div className="hidden md:flex justify-end gap-2 mt-6">
                    <button
                      type="button"
                      onClick={() => navigate("/profile")}
                      className="h-10 px-5 rounded-full border border-gray-300 text-gray-700 hover:bg-gray-50"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={saving}
                      className="h-10 px-6 rounded-full bg-primary text-white font-medium hover:opacity-90 disabled:opacity-60"
                    >
                      {saving ? "Saving..." : "Save"}
                    </button>
                  </div>
                </form>
              </div>
            </section>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
