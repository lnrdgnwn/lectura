// components/Filter/FilterSidebar.jsx
export default function FilterSidebar({
  genres = [],
  tags = [],
  selectedGenre,
  selectedTags,
  onGenreSelect,
  onTagToggle,
}) {
  return (
    <aside className="hidden lg:flex flex-col justify-center rounded-lg shadow-sm items-start max-w-64 bg-white p-6">
      <div className="mb-6">
        <h2 className="font-bold text-xl text-black mb-3 ">Genre</h2>
        <div className="flex flex-wrap gap-2">
          {genres.map((g) => (
            <button
              key={g}
              onClick={() => onGenreSelect(g)}
              className={`px-3 py-1 rounded-full text-sm cursor-pointer capitalize ${
                selectedGenre === g
                  ? "bg-primary text-white"
                  : "bg-gray-100 hover:bg-gray-200 text-gray-700"
              }`}
            >
              {g}
            </button>
          ))}
        </div>
      </div>

      <div>
        <h2 className="font-bold text-xl text-black mb-3">Tags</h2>
        <div className="flex flex-wrap gap-2">
          {tags.map((t) => (
            <button
              key={t}
              onClick={() => onTagToggle(t)}
              className={`px-3 py-1 rounded-full text-sm cursor-pointer capitalize ${
                selectedTags.includes(t)
                  ? "bg-primary text-white"
                  : "bg-gray-100 hover:bg-gray-200 text-gray-700"
              }`}
            >
              {t}
            </button>
          ))}
        </div>
      </div>
    </aside>
  );
}
