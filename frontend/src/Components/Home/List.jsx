import { useEffect, useState } from "react";
import { NavLink } from "react-router-dom";
import { getNovelsByGenre } from "../../services/novelService";

export default function List({
  genreId,
  genreName = "",
  genreSlug = "",
  limit = 6,
  sort = "updated_at:desc",
  titleOverride = null,
}) {
  const [novels, setNovels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");

  useEffect(() => {
    if (!genreId) return;
    let cancelled = false;

    (async () => {
      setLoading(true);
      setErr("");
      try {
        const rows = await getNovelsByGenre(genreId);
        const [field, dirRaw] = String(sort).split(":");
        const dir = (dirRaw || "desc").toLowerCase() === "asc" ? 1 : -1;

        const sorted = rows.slice().sort((a, b) => {
          const va = a?.[field] ?? a?.updated_at ?? a?.created_at ?? "";
          const vb = b?.[field] ?? b?.updated_at ?? b?.created_at ?? "";
          const da = Date.parse(va);
          const db = Date.parse(vb);
          if (!Number.isNaN(da) && !Number.isNaN(db)) return (da - db) * dir;
          return String(va).localeCompare(String(vb)) * dir;
        });

        // --- normalisasi -> pakai cover_image dari respons kamu ---
        const normalized = sorted
          .map((n) => ({
            id: n.id,
            title: n.title,
            cover: n.cover_image || n.cover || "",
          }))
          .filter((n) => n.cover)
          .slice(0, limit);

        if (!cancelled) setNovels(normalized);
      } catch (e) {
        if (!cancelled) setErr(e.message || "Failed to load novels");
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [genreId, limit, sort]);

  const sectionTitle = titleOverride || genreName || "Section";
  const seeMorePath = `/filter/${genreSlug || (genreName || "").toLowerCase()}`;

  return (
    <section className="flex flex-col items-center justify-center w-full px-6 md:px-10 py-6">
      <div className="flex w-full max-w-6xl justify-between items-center mb-4">
        <h2 className="text-xl md:text-2xl font-bold text-gray-800 capitalize">
          {sectionTitle}
        </h2>
        <NavLink
          to={seeMorePath}
          className="text-primary hover:underline text-sm md:text-base"
        >
          See more
        </NavLink>
      </div>
      <hr className="w-full max-w-6xl border-gray-300 mb-4" />

      {loading ? (
        <p className="text-gray-500 italic">Loading…</p>
      ) : err ? (
        <p className="text-red-600">{err}</p>
      ) : novels.length === 0 ? (
        <p className="text-gray-500 italic">No novels found in this genre.</p>
      ) : (
        <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-6 gap-3 sm:gap-4 md:gap-5 w-full max-w-6xl">
          {novels.map((novel) => (
            <div
              key={novel.id}
              className="flex flex-col bg-white rounded-xl shadow-sm hover:shadow-md transition"
            >
              <NavLink to={`/novel/${novel.id}`} className="w-full">
                <img
                  src={novel.cover}
                  alt={novel.title}
                  className="w-full h-32 sm:h-40 lg:h-52 object-cover rounded-t-xl"
                />
              </NavLink>
              <div className="p-2">
                <NavLink
                  to={`/novel/${novel.id}`}
                  className="text-[11px] sm:text-xs md:text-sm font-semibold text-gray-900 hover:text-primary line-clamp-2"
                >
                  {novel.title}
                </NavLink>
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
