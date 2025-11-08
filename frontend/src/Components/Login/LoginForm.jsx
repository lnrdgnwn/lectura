import { useRef, useState } from "react";
import { NavLink, useNavigate, useLocation } from "react-router-dom";
import { FaEye, FaEyeSlash } from "react-icons/fa";
import { useAuth } from "../../context/AuthContext";
import toast from "react-hot-toast";

export default function LoginForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login } = useAuth();

  const [form, setForm] = useState({ email: "", password: "" });
  const [touched, setTouched] = useState({});
  const [errors, setErrors] = useState({});
  const [show, setShow] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const emailRef = useRef(null);
  const passwordRef = useRef(null);
  const emailRegex = /^[\w.+\-]+@([\w\-]+\.)+[A-Za-z]{2,}$/;

  const validate = ({ email, password }) => {
    const er = {};
    if (!email) er.email = "Email wajib diisi.";
    else if (!emailRegex.test(email)) er.email = "Format email tidak valid.";
    if (!password) er.password = "Password wajib diisi.";
    else if (password.length < 8) er.password = "Minimal 8 karakter.";
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

  const focusFirstInvalid = (v) => {
    if (v.email) emailRef.current?.focus();
    else if (v.password) passwordRef.current?.focus();
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    const v = validate(form);
    setErrors(v);

    const nextTouched = { ...touched };
    Object.keys(v).forEach((k) => (nextTouched[k] = true));
    setTouched(nextTouched);

    if (Object.keys(v).length) {
      focusFirstInvalid(v);
      return;
    }

    setSubmitting(true);
    try {
      const user = await login(form);
      toast.success(`Welcome back, ${user?.username || "User"}!`, {
        duration: 5000,
      });
      navigate("/", { replace: true });
    } catch (err) {
      const msg =
        err?.response?.data?.message ||
        err?.message ||
        "Login gagal. Coba lagi.";
      toast.error(msg, { duration: 3500 });
    } finally {
      setSubmitting(false);
    }
  };

  const v = validate(form);
  const disabled = submitting || Object.keys(v).length > 0;

  return (
    <div
      className="
        min-h-[100dvh] w-full
        flex flex-col items-center justify-center
        px-10
      "
    >
      <div className="flex flex-col items-center mb-3 sm:mb-6 shrink-0">
        <img
          src="img/lectura.png"
          alt="lectura"
          className="w-3/4 md:w-1/2object-contain"
        />
        <p className="font-bold text-lg sm:text-xl md:text-2xl leading-tight mt-2">
          Login to your account
        </p>
      </div>

      <div className="w-full max-w-4xl ">
        <form
          onSubmit={handleSubmit}
          noValidate
          className="w-full flex flex-col gap-2 sm:gap-3"
        >
          <label htmlFor="email" className="font-medium text-xs sm:text-sm">
            Email Address
          </label>
          <input
            ref={emailRef}
            id="email"
            type="email"
            name="email"
            placeholder="Email@gmail.com"
            value={form.email}
            onChange={handleChange}
            onBlur={handleBlur}
            autoComplete="email"
            aria-invalid={!!fieldError("email")}
            aria-describedby={fieldError("email") ? "email-error" : undefined}
            className={[
              "w-full rounded-xl bg-surface font-semibold text-primary placeholder-gray-400 outline-none",
              "focus:ring-2 focus:ring-primary",
              "text-sm sm:text-base",
              "py-2.5 px-4 sm:py-3 sm:px-6",
              fieldError("email")
                ? "ring-2 ring-red-400 focus:ring-red-400"
                : "",
            ].join(" ")}
          />
          <p
            id="email-error"
            role="alert"
            aria-live="polite"
            className={[
              "text-xs sm:text-sm text-red-600 -mt-1 sm:mt-1 transition-all duration-200 ease-out overflow-hidden",
              fieldError("email")
                ? "opacity-100 max-h-10"
                : "opacity-0 max-h-0",
            ].join(" ")}
            aria-hidden={!fieldError("email")}
          >
            {errors.email}
          </p>

          <label htmlFor="password" className="font-medium text-xs sm:text-sm">
            Password
          </label>
          <div className="relative">
            <input
              ref={passwordRef}
              id="password"
              type={show ? "text" : "password"}
              name="password"
              placeholder="Enter your password"
              value={form.password}
              onChange={handleChange}
              onBlur={handleBlur}
              autoComplete="current-password"
              aria-invalid={!!fieldError("password")}
              aria-describedby={
                fieldError("password") ? "password-error" : undefined
              }
              className={[
                "w-full rounded-xl bg-surface font-semibold text-primary placeholder-gray-400 outline-none",
                "focus:ring-2 focus:ring-primary",
                "text-sm sm:text-base",
                "py-2.5 pl-4 pr-10 sm:py-3 sm:pl-6 sm:pr-12",
                fieldError("password")
                  ? "ring-2 ring-red-400 focus:ring-red-400"
                  : "",
              ].join(" ")}
            />
            <button
              type="button"
              onClick={() => setShow((s) => !s)}
              className="absolute inset-y-0 right-3 sm:right-4 my-auto text-gray-500 hover:text-gray-700"
              aria-label={show ? "Sembunyikan password" : "Tampilkan password"}
              title={show ? "Sembunyikan password" : "Tampilkan password"}
            >
              {show ? <FaEyeSlash /> : <FaEye />}
            </button>
          </div>
          <p
            id="password-error"
            role="alert"
            aria-live="polite"
            className={[
              "text-xs sm:text-sm text-red-600 -mt-1 sm:mt-2 transition-all duration-200 ease-out overflow-hidden",
              fieldError("password")
                ? "opacity-100 max-h-10"
                : "opacity-0 max-h-0",
            ].join(" ")}
            aria-hidden={!fieldError("password")}
          >
            {errors.password}
          </p>

          <button
            type="submit"
            disabled={disabled}
            className="
              text-white
              bg-primary hover:bg-primary-dark
              font-bold rounded-xl
              cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed
              transition-colors
              text-sm sm:text-base
              py-2.5 sm:py-3
            "
          >
            {submitting ? "Logging in..." : "Login Now"}
          </button>

          <div className="flex items-center my-2 sm:my-4">
            <div className="flex-grow h-px bg-gray-300" />
            <span className="px-3 text-gray-400 text-xs sm:text-sm font-medium">
              OR
            </span>
            <div className="flex-grow h-px bg-gray-300" />
          </div>

          <NavLink
            to="/register"
            className="
              text-primary border hover:bg-gray-100 border-primary
              font-bold rounded-xl text-center
              text-sm sm:text-base
              py-2.5 sm:py-3
            "
          >
            Register Now
          </NavLink>
        </form>
      </div>
    </div>
  );
}
