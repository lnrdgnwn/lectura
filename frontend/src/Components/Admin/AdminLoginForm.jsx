// src/pages/LoginForm.jsx
import { useState } from "react";
import { NavLink, useNavigate, useLocation } from "react-router-dom";
import { FaEye, FaEyeSlash } from "react-icons/fa";
import { useAuth } from "../../context/AuthContext";

import toast from "react-hot-toast";

export default function AdminLoginForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login } = useAuth();

  const [form, setForm] = useState({ email: "", password: "" });
  const [touched, setTouched] = useState({});
  const [errors, setErrors] = useState({});
  const [show, setShow] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const emailRegex = /^[\w.+\-]+@([\w\-]+\.)+[A-Za-z]{2,}$/;

  const validate = ({ email, password }) => {
    const er = {};
    if (!email) er.email = "Email wajib diisi.";
    else if (!emailRegex.test(email)) er.email = "Format email tidak valid.";
    if (!password) er.password = "Password wajib diisi.";
    else if (password.length < 8) er.password = "Minimal 8 karakter.";
    return er;
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    const next = { ...form, [name]: value };
    setForm(next);
    if (touched[name]) setErrors(validate(next));
  };

  const handleBlur = (e) => {
    const { name } = e.target;
    setTouched((t) => ({ ...t, [name]: true }));
    setErrors(validate(form));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    const v = validate(form);
    setErrors(v);
    setTouched({ email: true, password: true });
    if (Object.keys(v).length) return;

    setSubmitting(true);
    try {
      const user = await login(form); // ini panggil loginUser() via AuthContext
      toast.success(`Welcome back, ${user?.username || "User"}!`);
      const to = location.state?.from?.pathname || "/";
      navigate(to, { replace: true });
    } catch (err) {
      const msg =
        err?.response?.data?.message ||
        err?.message ||
        "Login gagal. Coba lagi.";
      toast.error(msg);
    } finally {
      setSubmitting(false);
    }
  };

  const disabled = submitting || Object.keys(validate(form)).length > 0;

  return (
    <div className="flex flex-col justify-center items-center px-10">
      <div className="flex flex-col items-center mb-8">
        <img src="img/lectura.png" alt="lectura" className="w-3/4 lg:w-1/2" />
        <p className="font-bold text-2xl">Login to your account</p>
      </div>

      <form onSubmit={handleSubmit} className="w-full flex flex-col space-y-3">
        <label htmlFor="email" className="font-medium">
          Email Address
        </label>
        <input
          id="email"
          type="email"
          name="email"
          placeholder="Email@gmail.com"
          className="py-3 px-6 rounded-xl bg-surface font-bold text-primary outline-none focus:ring-2 focus:ring-primary"
          value={form.email}
          onChange={handleChange}
          onBlur={handleBlur}
          autoComplete="email"
          aria-invalid={!!errors.email}
          aria-describedby={errors.email ? "email-error" : undefined}
        />
        {errors.email && (
          <p id="email-error" className="text-sm text-red-600 -mt-2">
            {errors.email}
          </p>
        )}

        <label htmlFor="password" className="font-medium">
          Password
        </label>
        <div className="relative">
          <input
            id="password"
            type={show ? "text" : "password"}
            name="password"
            placeholder="Enter your password"
            className="w-full py-3 px-6 pr-12 rounded-xl bg-surface font-bold text-primary outline-none focus:ring-2 focus:ring-primary"
            value={form.password}
            onChange={handleChange}
            onBlur={handleBlur}
            autoComplete="current-password"
            aria-invalid={!!errors.password}
            aria-describedby={errors.password ? "password-error" : undefined}
          />
          <button
            type="button"
            onClick={() => setShow((s) => !s)}
            className="absolute inset-y-0 right-4 my-auto text-gray-500 hover:text-gray-700"
            aria-label={show ? "Sembunyikan password" : "Tampilkan password"}
            title={show ? "Sembunyikan password" : "Tampilkan password"}
          >
            {show ? <FaEyeSlash /> : <FaEye />}
          </button>
        </div>
        {errors.password && (
          <p id="password-error" className="text-sm text-red-600 -mt-2">
            {errors.password}
          </p>
        )}

        <button
          type="submit"
          disabled={disabled}
          className="text-white mt-3 bg-primary hover:bg-primary-dark font-bold py-4 rounded-xl cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {submitting ? "Logging in..." : "Login Now"}
        </button>

        <div className="flex items-center my-4">
          <div className="flex-grow h-px bg-gray-300" />
          <span className="px-3 text-gray-400 text-sm font-medium">OR</span>
          <div className="flex-grow h-px bg-gray-300" />
        </div>

        <NavLink
          to="/register"
          className="text-primary border hover:bg-gray-100 border-primary font-bold py-4 rounded-xl cursor-pointer text-center"
        >
          Register Now
        </NavLink>
      </form>
    </div>
  );
}
