import {
  FaList,
  FaMoon,
  FaSun,
  FaFont,
  FaArrowLeft,
  FaArrowRight,
} from "react-icons/fa";

export default function ReaderControl({
  darkMode,
  setDarkMode,
  setTextSize,
  showControls,
  setShowChapters,
  goToNext,
  goToPrev,
}) {
  return (
    <aside
      className={`
        bg-primary text-white 
        transition-all duration-300 flex flex-col 
        sm:sticky sm:top-18 md:w-20 md:flex-shrink-0
        ${
          showControls
            ? "translate-y-0 opacity-100"
            : "translate-y-full opacity-0"
        }
        md:translate-y-0 md:opacity-100
        fixed bottom-0 left-0 w-full md:static
      `}
    >
      {/* Tombol kontrol */}
      <div className="flex justify-around md:flex-col md:items-center md:justify-center py-3 w-full md:py-6 md:space-y-6">
        <button
          onClick={goToPrev}
          title="Previous Chapter"
          className="cursor-pointer"
        >
          <FaArrowLeft size={18} />
        </button>

        <button
          onClick={() => setShowChapters(true)}
          title="Chapters"
          className="cursor-pointer"
        >
          <FaList size={18} />
        </button>

        <button
          onClick={() => setDarkMode(!darkMode)}
          title="Toggle Theme"
          className="cursor-pointer"
        >
          {darkMode ? <FaSun size={18} /> : <FaMoon size={18} />}
        </button>

        <button
          onClick={() => setTextSize((prev) => (prev < 24 ? prev + 1 : prev))}
          className="flex items-center cursor-pointer"
          title="Increase Font"
        >
          <FaFont />+
        </button>

        <button
          onClick={() => setTextSize((prev) => (prev > 12 ? prev - 1 : prev))}
          className="flex items-center cursor-pointer"
          title="Decrease Font"
        >
          <FaFont />-
        </button>

        <button
          onClick={goToNext}
          title="Next Chapter"
          className="cursor-pointer"
        >
          <FaArrowRight size={18} />
        </button>
      </div>
    </aside>
  );
}
