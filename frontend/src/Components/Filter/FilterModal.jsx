import { FiX } from "react-icons/fi";

export default function FilterModal({
  open,
  onClose,
  genres,
  tags,
  selectedGenre,
  setSelectedGenre,
  selectedTags,
  setSelectedTags,
}) {
  if (!open) return null;

  const toggleTag = (tag) => {
    setSelectedTags((prev) =>
      prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag]
    );
  };

  return (
    <div
      className="fixed inset-0 bg-black/50 z-50 flex justify-center items-end lg:hidden"
      onClick={onClose}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        className="bg-white w-full  rounded-t-3xl md:rounded-b-3xl p-6 shadow-lg animate-slide-up"
      >
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-semibold text-gray-800">Filter Novel</h2>
          <button onClick={onClose}>
            <FiX size={22} className="text-gray-600 cursor-pointer" />
          </button>
        </div>

        <div>
          <h3 className="font-medium text-gray-700 mb-2">Genre</h3>
          <div className="flex flex-wrap gap-2 mb-4">
            {genres.map((g) => (
              <button
                key={g}
                onClick={() => setSelectedGenre(g)}
                className={`px-3 py-1 rounded-full text-sm capitalize ${
                  selectedGenre === g
                    ? "bg-primary text-white"
                    : "bg-gray-100 text-gray-700 hover:bg-gray-200"
                }`}
              >
                {g}
              </button>
            ))}
          </div>

          <h3 className="font-medium text-gray-700 mb-2">Tags</h3>
          <div className="flex flex-wrap gap-2">
            {tags.map((t) => (
              <button
                key={t}
                onClick={() => toggleTag(t)}
                className={`px-3 py-1 rounded-full text-sm ${
                  selectedTags.includes(t)
                    ? "bg-primary text-white"
                    : "bg-gray-100 text-gray-700 hover:bg-gray-200"
                }`}
              >
                {t}
              </button>
            ))}
          </div>

          <button
            onClick={onClose}
            className="mt-6 w-full bg-primary text-white py-2 rounded-xl font-medium"
          >
            Apply
          </button>
        </div>
      </div>
    </div>
  );
}
