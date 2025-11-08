import { NavLink } from "react-router-dom";

export default function FilterNovelCard({ novel }) {
  return (
    <NavLink to={`/novel/${novel.id}`}>
      <div className="flex bg-white rounded-xl shadow-sm hover:shadow-md transition p-3 w-full cap ">
        {/* Cover */}
        <div className="w-28 h-36 flex-shrink-0 rounded-lg overflow-hidden bg-blue-50 border border-blue-100">
          {novel.cover_image ? (
            <img
              src={novel.cover_image}
              alt={novel.title}
              className="w-full h-full object-cover"
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center text-blue-700 font-bold text-lg">
              {novel.title ? novel.title.charAt(0) : "N"}
            </div>
          )}
        </div>

        <div className="flex flex-col justify-between px-4 w-full">
          <div>
            <h3 className="font-semibold text-gray-800 text-lg line-clamp-1">
              {novel.title}
            </h3>

            {novel.genres && novel.genres.length > 0 && (
              <div className="flex flex-wrap gap-1 mb-1 py-1">
                {novel.genres.map((g) => (
                  <span
                    key={g.id ?? g.slug}
                    className="text-xs bg-gray-100 text-gray-700 px-2 py-0.5 rounded-full capitalize"
                  >
                    {g.slug || g.name}
                  </span>
                ))}
              </div>
            )}

            {novel.tags && novel.tags.length > 0 && (
              <div className="flex flex-wrap gap-1 mb-2">
                {novel.tags.map((tag) => (
                  <span
                    key={tag.id ?? tag.slug}
                    className="text-xs bg-blue-100 text-blue-700 px-2 py-0.5 rounded-full capitalize"
                  >
                    {tag.name || tag.slug}
                  </span>
                ))}
              </div>
            )}

            {novel.synopsis && (
              <p className="text-sm text-gray-600 line-clamp-2">
                {novel.synopsis}
              </p>
            )}
          </div>
        </div>
      </div>
    </NavLink>
  );
}
