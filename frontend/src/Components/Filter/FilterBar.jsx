// components/Filter/FilterBar.jsx
import { FiFilter } from "react-icons/fi";

export default function FilterBar({
  genres = [],
  selectedGenre,
  setSelectedGenre,
  onOpen,
}) {
  return (
    <div className="w-full flex lg:hidden justify-between items-center bg-white border border-gray-200 rounded-xl shadow-sm p-3">
      <div className="flex gap-2 overflow-x-auto scrollbar-hide">
        {genres.map((g) => (
          <button
            key={g}
            onClick={() => setSelectedGenre(g)}
            className={`px-4 py-1.5 rounded-full whitespace-nowrap text-sm capitalize ${
              selectedGenre === g
                ? "bg-primary text-white"
                : "bg-gray-100 hover:bg-gray-200 text-gray-700"
            }`}
          >
            {g}
          </button>
        ))}
      </div>

      <button
        onClick={onOpen}
        className="flex items-center gap-2 bg-primary text-white px-4 py-2 rounded-lg  ml-3 lg:hidden cursor-pointer"
      >
        <FiFilter size={18} />
        <span>Filter</span>
      </button>
    </div>
  );
}
