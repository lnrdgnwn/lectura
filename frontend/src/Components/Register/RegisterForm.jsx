import { useState } from "react";
import { NavLink, useNavigate, useLocation } from "react-router-dom";
import { FaEye, FaEyeSlash } from "react-icons/fa";
import toast from "react-hot-toast";
import { useAuth } from "../../context/AuthContext";

export default function RegisterForm() {
  const { register } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const [form, setForm] = useState({
    name: "",
    email: "",
    password: "",
    confirmPassword: "",
  });
  const [touched, setTouched] = useState({});
  const [errors, setErrors] = useState({});
  const [showPass, setShowPass] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const emailRegex = /^[\w.+\-]+@([\w\-]+\.)+[A-Za-z]{2,}$/;

  const validate = ({ name, email, password, confirmPassword }) => {
    const er = {};
    if (!name || name.trim().length < 3) er.name = "Nama minimal 3 karakter.";
    if (!email) er.email = "Email wajib diisi.";
    else if (!emailRegex.test(email)) er.email = "Format email tidak valid.";
    if (!password) er.password = "Password wajib diisi.";
    else if (password.length < 8) er.password = "Minimal 8 karakter.";
    if (!confirmPassword) er.confirmPassword = "Konfirmasi password wajib.";
    else if (password !== confirmPassword)
      er.confirmPassword = "Password tidak sama.";
    return er;
  };

  const fieldError = (name) => touched[name] && errors[name];

  const handleChange = (e) => {
    const { name, value } = e.target;
    const next = { ...form, [name]: value };
    setForm(next);
    if (touched[name]) {
      const v = validate(next);
      setErrors((prev) => ({ ...prev, [name]: v[name] }));
    }
  };

  const handleBlur = (e) => {
    const { name } = e.target;
    setTouched((t) => ({ ...t, [name]: true }));
    const v = validate(form);
    setErrors((prev) => ({ ...prev, [name]: v[name] }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    const v = validate(form);
    setErrors(v);

    const nextTouched = { ...touched };
    Object.keys(v).forEach((k) => (nextTouched[k] = true));
    setTouched(nextTouched);

    if (Object.keys(v).length) return;

    setSubmitting(true);
    try {
      const user = await register({
        username: form.name.trim(),
        email: form.email.trim(),
        password: form.password,
      });

      toast.success("Registrasi berhasil! Welcome " + user?.username, {
        id: "register",
      });

      const to = location.state?.from?.pathname || "/";
      navigate(to, { replace: true });
    } catch (err) {
      const msg =
        err?.response?.data?.message ||
        err?.message ||
        "Registrasi gagal. Coba lagi.";
      toast.error(msg, { id: "register" });
    }
  };

  const disabled = submitting || Object.keys(validate(form)).length > 0;

  return (
    <div
      className="
        w-full 
        flex flex-col items-center justify-center
        px-10
        py-4 sm:py-6
      "
    >

      <div className="flex flex-col items-center mb-3 sm:mb-6 shrink-0">
        <img
          src="/img/lectura.png"
          alt="lectura"
          className="w-3/4 md:w-1/2 object-contain"
        />
        <p className="font-bold text-lg sm:text-xl md:text-2xl leading-tight mt-2">
          Register your account
        </p>
      </div>

      <div className="w-full max-w-4xl sm:max-w-2xl md:max-w-4xl ">
        <form
          onSubmit={handleSubmit}
          noValidate
          className="
            w-full flex flex-col
            space-y-2 sm:space-y-3
          "
        >
          <label
            htmlFor="register-name"
            className="font-medium text-xs sm:text-sm"
          >
            Username
          </label>
          <input
            id="register-name"
            name="name"
            type="text"
            placeholder="Your name"
            value={form.name}
            onChange={handleChange}
            onBlur={handleBlur}
            autoComplete="name"
            className={[
              "w-full rounded-xl bg-surface font-semibold text-primary placeholder-gray-400 outline-none",
              "focus:ring-2 focus:ring-primary",
              "text-sm sm:text-base",
              "py-2.5 px-4 sm:py-3 sm:px-6",
              fieldError("name")
                ? "ring-2 ring-red-400 focus:ring-red-400"
                : "",
            ].join(" ")}
            aria-invalid={!!fieldError("name")}
            aria-describedby={fieldError("name") ? "name-error" : undefined}
          />
          <p
            id="name-error"
            role="alert"
            aria-live="polite"
            className={[
              "text-xs sm:text-sm text-red-600 -mt-1 sm:-mt-2 transition-all duration-200 ease-out overflow-hidden",
              fieldError("name") ? "opacity-100 max-h-10" : "opacity-0 max-h-0",
            ].join(" ")}
            aria-hidden={!fieldError("name")}
          >
            {errors.name}
          </p>

          <label
            htmlFor="register-email"
            className="font-medium text-xs sm:text-sm"
          >
            Email Address
          </label>
          <input
            id="register-email"
            name="email"
            type="email"
            placeholder="email@gmail.com"
            value={form.email}
            onChange={handleChange}
            onBlur={handleBlur}
            autoComplete="email"
            className={[
              "w-full rounded-xl bg-surface font-semibold text-primary placeholder-gray-400 outline-none",
              "focus:ring-2 focus:ring-primary",
              "text-sm sm:text-base",
              "py-2.5 px-4 sm:py-3 sm:px-6",
              fieldError("email")
                ? "ring-2 ring-red-400 focus:ring-red-400"
                : "",
            ].join(" ")}
            aria-invalid={!!fieldError("email")}
            aria-describedby={fieldError("email") ? "email-error" : undefined}
          />
          <p
            id="email-error"
            role="alert"
            aria-live="polite"
            className={[
              "text-xs sm:text-sm text-red-600 -mt-1 sm:-mt-2 transition-all duration-200 ease-out overflow-hidden",
              fieldError("email")
                ? "opacity-100 max-h-10"
                : "opacity-0 max-h-0",
            ].join(" ")}
            aria-hidden={!fieldError("email")}
          >
            {errors.email}
          </p>

          <label
            htmlFor="register-password"
            className="font-medium text-xs sm:text-sm"
          >
            Password
          </label>
          <div className="relative">
            <input
              id="register-password"
              name="password"
              type={showPass ? "text" : "password"}
              placeholder="Enter your password"
              value={form.password}
              onChange={handleChange}
              onBlur={handleBlur}
              autoComplete="new-password"
              className={[
                "w-full rounded-xl bg-surface font-semibold text-primary placeholder-gray-400 outline-none",
                "focus:ring-2 focus:ring-primary",
                "text-sm sm:text-base",
                "py-2.5 pl-4 pr-10 sm:py-3 sm:pl-6 sm:pr-12",
                fieldError("password")
                  ? "ring-2 ring-red-400 focus:ring-red-400"
                  : "",
              ].join(" ")}
              aria-invalid={!!fieldError("password")}
              aria-describedby={
                fieldError("password") ? "password-error" : undefined
              }
            />
            <button
              type="button"
              onClick={() => setShowPass((s) => !s)}
              className="absolute inset-y-0 right-3 sm:right-4 my-auto text-gray-500 hover:text-gray-700"
              aria-label={
                showPass ? "Sembunyikan password" : "Tampilkan password"
              }
              title={showPass ? "Sembunyikan password" : "Tampilkan password"}
            >
              {showPass ? <FaEyeSlash /> : <FaEye />}
            </button>
          </div>
          <p
            id="password-error"
            role="alert"
            aria-live="polite"
            className={[
              "text-xs sm:text-sm text-red-600 -mt-1 sm:-mt-2 transition-all duration-200 ease-out overflow-hidden",
              fieldError("password")
                ? "opacity-100 max-h-10"
                : "opacity-0 max-h-0",
            ].join(" ")}
            aria-hidden={!fieldError("password")}
          >
            {errors.password}
          </p>

          <label
            htmlFor="register-confirm"
            className="font-medium text-xs sm:text-sm"
          >
            Confirm Password
          </label>
          <div className="relative">
            <input
              id="register-confirm"
              name="confirmPassword"
              type={showConfirm ? "text" : "password"}
              placeholder="Confirm your password"
              value={form.confirmPassword}
              onChange={handleChange}
              onBlur={handleBlur}
              autoComplete="new-password"
              className={[
                "w-full rounded-xl bg-surface font-semibold text-primary placeholder-gray-400 outline-none",
                "focus:ring-2 focus:ring-primary",
                "text-sm sm:text-base",
                "py-2.5 pl-4 pr-10 sm:py-3 sm:pl-6 sm:pr-12",
                fieldError("confirmPassword")
                  ? "ring-2 ring-red-400 focus:ring-red-400"
                  : "",
              ].join(" ")}
              aria-invalid={!!fieldError("confirmPassword")}
              aria-describedby={
                fieldError("confirmPassword") ? "confirm-error" : undefined
              }
            />
            <button
              type="button"
              onClick={() => setShowConfirm((s) => !s)}
              className="absolute inset-y-0 right-3 sm:right-4 my-auto text-gray-500 hover:text-gray-700"
              aria-label={
                showConfirm ? "Sembunyikan konfirmasi" : "Tampilkan konfirmasi"
              }
              title={
                showConfirm ? "Sembunyikan konfirmasi" : "Tampilkan konfirmasi"
              }
            >
              {showConfirm ? <FaEyeSlash /> : <FaEye />}
            </button>
          </div>
          <p
            id="confirm-error"
            role="alert"
            aria-live="polite"
            className={[
              "text-xs sm:text-sm text-red-600 -mt-1 sm:-mt-2 transition-all duration-200 ease-out overflow-hidden",
              fieldError("confirmPassword")
                ? "opacity-100 max-h-10"
                : "opacity-0 max-h-0",
            ].join(" ")}
            aria-hidden={!fieldError("confirmPassword")}
          >
            {errors.confirmPassword}
          </p>

          <button
            disabled={disabled}
            className="
              text-white  mb-0
              bg-primary hover:bg-primary-dark
              font-bold rounded-xl
              cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed
              transition-colors
              text-sm sm:text-base
              py-2.5 sm:py-3
            "
          >
            {submitting ? "Registering..." : "Register Now"}
          </button>

          <div className="flex items-center my-2 sm:my-4 w-full">
            <div className="flex-grow h-px bg-gray-300" />
            <span className="px-3 text-gray-400 text-xs sm:text-sm font-medium">
              OR
            </span>
            <div className="flex-grow h-px bg-gray-300" />
          </div>

          <NavLink
            to="/login"
            className="
              text-primary border hover:bg-gray-100 border-primary
              font-bold rounded-xl text-center
              text-sm sm:text-base
              py-2.5 sm:py-3
            "
          >
            Login Now
          </NavLink>
        </form>
      </div>
    </div>
  );
}
