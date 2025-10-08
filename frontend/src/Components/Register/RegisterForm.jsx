export default function RegisterForm() {
  return (
    <div className="flex flex-col justify-center items-center px-10">
      <div className="flex flex-col items-center mb-8">
        <img src="img/lectura.png" alt="lectura" className="w-3/4 lg:w-1/2" />
        <p className="font-bold text-2xl">Register your account</p>
      </div>
      <form action="" className="w-full flex flex-col space-y-3">
        <label htmlFor="email">Email Address </label>
        <input
          type="text"
          name="email"
          placeholder="Email@gmail.com"
          className="py-3 px-6 rounded-xl bg-surface font-bold text-primary"
        />
        <label htmlFor="password">Password</label>
        <input
          type="text"
          name="password"
          placeholder="Enter your password"
          className="py-3 px-6 rounded-xl bg-surface font-bold text-primary"
        />
        <label htmlFor="confirmPassword">Password</label>
        <input
          type="text"
          name="confirmPassword"
          placeholder="Confirm your password"
          className="py-3 px-6 rounded-xl bg-surface font-bold text-primary"
        />
        <button className="text-white mt-3 bg-primary hover:bg-primary-dark font-bold py-4 rounded-xl cursor-pointer">
          Register Now
        </button>
        <div class="flex items-center my-4">
          <div class="flex-grow h-px bg-gray-300"></div>
          <span class="px-3 text-gray-400 text-sm font-medium">OR</span>
          <div class="flex-grow h-px bg-gray-300"></div>
        </div>
        <button className="text-primary border hover:bg-gray-100 border-primary font-bold py-4 rounded-xl cursor-pointer">
          Login Now
        </button>
      </form>
    </div>
  );
}
