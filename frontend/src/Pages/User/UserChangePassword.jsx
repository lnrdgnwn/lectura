import { useState, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { FiArrowLeft, FiCheck, FiEye, FiEyeOff } from "react-icons/fi";
import Navbar from "../../components/Navbar";
import Footer from "../../components/Footer";
import { useAuth } from "../../context/AuthContext";
import { changeMyPassword } from "../../services/userService";
import toast from "react-hot-toast";
import useScrollToTop from "../../hooks/useScrollToTop";

export default function UserChangePassword() {
  useScrollToTop();
  const { user } = useAuth();
  const navigate = useNavigate();
  const formRef = useRef(null);

  const [oldPassword, setOld] = useState("");
  const [newPassword, setNew] = useState("");
  const [confirmPassword, setConfirm] = useState("");
  const [loading, setLoading] = useState(false);

  // state buat toggle hide/show
  const [showOld, setShowOld] = useState(false);
  const [showNew, setShowNew] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);

  if (!user) return null;

  const submit = async (e) => {
    e?.preventDefault?.();
    if (loading) return;

    if (newPassword.length < 8) return toast.error("Min 8 characters.");
    if (newPassword !== confirmPassword)
      return toast.error("Passwords do not match.");

    setLoading(true);
    try {
      await toast.promise(
        changeMyPassword({
          old_password: oldPassword,
          new_password: newPassword,
        }),
        {
          loading: "Changing password…",
          success: "Password changed!",
          error: (err) => err?.message || "Failed to change password",
        }
      );
      setOld("");
      setNew("");
      setConfirm("");
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Navbar />
      <main className="min-h-screen bg-gray-50 pb-10">
        <div className="mx-auto md:p-16 w-full max-w-5xl px-4">
          {/* Mobile sub-navbar */}
          <div className="md:hidden sticky top-[65px] z-40 -mx-4 mb-4 px-4 py-3 bg-primary text-white flex items-center justify-between cursor-pointer">
            <button
              onClick={() => navigate("/profile")}
              className="text-white/90 text-xl p-1"
              aria-label="Back"
            >
              <FiArrowLeft />
            </button>
            <h1 className="text-base font-semibold">Change password</h1>
            <button
              onClick={() => formRef.current?.requestSubmit()}
              disabled={loading}
              className="text-xl p-1 disabled:opacity-60 cursor-pointer"
              aria-label="Save"
            >
              <FiCheck />
            </button>
          </div>

          {/* Layout */}
          <div className="grid md:grid-cols-12 gap-6">
            <aside className="block md:col-span-4">
              <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
                <div className="flex items-center gap-4">
                  <img
                    src={user.profile_picture}
                    alt={user.username}
                    className="h-16 w-16 rounded-full object-cover ring-1 ring-gray-200"
                  />
                  <div>
                    <p className="font-semibold text-gray-900">
                      {user.username}
                    </p>
                    {user.id && (
                      <p className="text-xs text-gray-500">ID {user.id}</p>
                    )}
                  </div>
                </div>
                <ul className="mt-5 text-sm text-gray-600 space-y-2 list-disc list-inside">
                  <li>Use at least 8 characters.</li>
                  <li>Avoid reusing old passwords.</li>
                </ul>
              </div>
            </aside>

            <div className="md:col-span-8">
              <div className="bg-white rounded-2xl shadow-sm border border-gray-100">
                <form
                  ref={formRef}
                  onSubmit={submit}
                  className="px-5 md:px-6 py-6 space-y-5"
                >
                  {/* Current password */}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Current password
                    </label>
                    <div className="relative">
                      <input
                        type={showOld ? "text" : "password"}
                        value={oldPassword}
                        onChange={(e) => setOld(e.target.value)}
                        className="w-full h-11 rounded-xl border border-gray-300 px-4 pr-10 focus:outline-none focus:ring-2 focus:ring-primary"
                        required
                      />
                      <button
                        type="button"
                        onClick={() => setShowOld((v) => !v)}
                        className="absolute inset-y-0 right-3 flex items-center text-gray-500 hover:text-gray-700 cursor-pointer"
                      >
                        {showOld ? <FiEyeOff /> : <FiEye />}
                      </button>
                    </div>
                  </div>

                  {/* New + confirm password */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        New password
                      </label>
                      <div className="relative">
                        <input
                          type={showNew ? "text" : "password"}
                          value={newPassword}
                          onChange={(e) => setNew(e.target.value)}
                          className="w-full h-11 rounded-xl border border-gray-300 px-4 pr-10 focus:outline-none focus:ring-2 focus:ring-primary"
                          required
                        />
                        <button
                          type="button"
                          onClick={() => setShowNew((v) => !v)}
                          className="absolute inset-y-0 right-3 flex items-center text-gray-500 hover:text-gray-700 cursor-pointer"
                        >
                          {showNew ? <FiEyeOff /> : <FiEye />}
                        </button>
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Confirm new password
                      </label>
                      <div className="relative">
                        <input
                          type={showConfirm ? "text" : "password"}
                          value={confirmPassword}
                          onChange={(e) => setConfirm(e.target.value)}
                          className="w-full h-11 rounded-xl border border-gray-300 px-4 pr-10 focus:outline-none focus:ring-2 focus:ring-primary"
                          required
                        />
                        <button
                          type="button"
                          onClick={() => setShowConfirm((v) => !v)}
                          className="absolute inset-y-0 right-3 flex items-center text-gray-500 hover:text-gray-700 cursor-pointer"
                        >
                          {showConfirm ? <FiEyeOff /> : <FiEye />}
                        </button>
                      </div>
                    </div>
                  </div>

                  {/* Desktop actions */}
                  <div className="hidden md:flex justify-end gap-2 mt-6">
                    <button
                      type="button"
                      onClick={() => navigate("/profile")}
                      className="h-10 px-5 rounded-full border border-gray-300 text-gray-700 hover:bg-gray-50 cursor-pointer"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={loading}
                      className="h-10 px-6 rounded-full bg-primary text-white font-medium hover:opacity-90 disabled:opacity-60 cursor-pointer"
                    >
                      {loading ? "Saving…" : "Save"}
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
