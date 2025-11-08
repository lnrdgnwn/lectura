import { useEffect, useMemo, useState, useCallback } from "react";
import { NavLink } from "react-router-dom";
import {
  getBookmarks,
  deleteBookmark,
  addBookmark,
} from "../../services/bookmarkService";
import { getNovelById, getNovels } from "../../services/novelService";
import Navbar from "../../components/Navbar";
import Footer from "../../components/Footer";

const toId = (v) => (v === 0 || v ? String(v) : null);
const normalizeBookmarks = (list) =>
  (list || [])
    .map((b) => ({ id: toId(b?.id), novel_id: toId(b?.novel_id) }))
    .filter((b) => b.id && b.novel_id);

export default function UserLibrary() {
  const [bookmarks, setBookmarks] = useState([]); // [{id, novel_id}]
  const [novels, setNovels] = useState([]); // [{id, title, cover_url, author, ...}]
  const [loading, setLoading] = useState(true);
  const [editMode, setEditMode] = useState(false);
  const [selected, setSelected] = useState(new Set()); // bookmark ids
  const [error, setError] = useState("");

  const refreshLibrary = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const rawBms = await getBookmarks();
      const bms = normalizeBookmarks(rawBms);
      setBookmarks(bms);

      const ids = bms.map((b) => b.novel_id);
      let novelsArr = await getNovels(ids);

      if ((!novelsArr || novelsArr.length === 0) && ids.length) {
        const list = await Promise.all(ids.map((id) => getNovelById(id)));
        novelsArr = list.filter(Boolean);
      }

      const safeNovels = (novelsArr || [])
        .map((n) => ({
          id: toId(n?.id),
          title: n?.title || "",
          cover_url: n?.cover_url || n?.cover_image || null,
          author:
            typeof n?.author === "string"
              ? n.author
              : n?.author?.username || n?.author?.name || null,
          progress: n?.progress ?? null,
        }))
        .filter((n) => n.id);

      setNovels(safeNovels);

      setSelected((prev) => {
        const valid = new Set(bms.map((b) => b.id));
        const next = new Set();
        for (const id of prev) if (valid.has(id)) next.add(id);
        return next;
      });
    } catch (e) {
      setError(e?.message || "Failed to load library");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refreshLibrary();
  }, [refreshLibrary]);

  const bmByNovelId = useMemo(() => {
    const map = new Map();
    for (const b of bookmarks) map.set(b.novel_id, b.id);
    return map;
  }, [bookmarks]);

  const novelsInLibrary = useMemo(() => {
    const ids = new Set(bookmarks.map((b) => b.novel_id));
    return novels.filter((n) => ids.has(n.id));
  }, [bookmarks, novels]);

  const getBookmarkIdByNovelId = (novelId) =>
    bmByNovelId.get(toId(novelId)) || null;

  function toggleSelect(bookmarkId) {
    const id = toId(bookmarkId);
    if (!id) return;
    setSelected((prev) => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  }

  function toggleSelectAll() {
    if (!editMode) return;
    const visibleBmIds = novelsInLibrary
      .map((n) => getBookmarkIdByNovelId(n.id))
      .filter(Boolean);
    if (selected.size === visibleBmIds.length) setSelected(new Set());
    else setSelected(new Set(visibleBmIds));
  }

  async function handleBulkRemove() {
    if (!selected.size) return;
    setLoading(true);
    setError("");
    const ids = Array.from(selected);
    try {
      await Promise.all(ids.map((id) => deleteBookmark(id)));
      setSelected(new Set());
      setEditMode(false);
      await refreshLibrary();
    } catch (e) {
      setError(e?.message || "Failed to remove from library");
    } finally {
      setLoading(false);
    }
  }

  async function handleAddBookmark(novelId) {
    try {
      setLoading(true);
      await addBookmark(toId(novelId));
      await refreshLibrary();
    } catch (e) {
      setError(e?.message || "Failed to add to library");
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <Navbar />
      <main className="px-4 md:px-6 py-6 min-h-screen">
        {/* Header */}
        <div className="max-w-6xl mx-auto flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-5">
          <div className="text-center sm:text-left">
            <h1 className="text-2xl font-semibold">My Library</h1>
            <p className="text-sm text-gray-500">
              {loading ? "Loading…" : `${novelsInLibrary.length} title(s)`}
            </p>
          </div>

          {novelsInLibrary.length > 0 && (
            <div className="flex items-center gap-2 justify-center sm:justify-end">
              {editMode && (
                <>
                  <button
                    className="px-3 py-2 rounded shadow-2xl hover:bg-gray-50 border border-gray-200 cursor-pointer"
                    onClick={toggleSelectAll}
                    disabled={loading}
                  >
                    {selected.size === novelsInLibrary.length
                      ? "Unselect All"
                      : "Select All"}
                  </button>
                  <button
                    className="px-3 py-2 rounded bg-red-600 text-white disabled:opacity-50"
                    disabled={!selected.size || loading}
                    onClick={handleBulkRemove}
                  >
                    Remove ({selected.size})
                  </button>
                </>
              )}
              <button
                className="px-3 py-2 rounded shadow-2xl border border-gray-200 cursor-pointer hover:bg-gray-50"
                onClick={() => {
                  setEditMode((v) => !v);
                  setSelected(new Set());
                }}
                disabled={loading || novelsInLibrary.length === 0}
              >
                {editMode ? "Done" : "Edit"}
              </button>
            </div>
          )}
        </div>

        {/* Error */}
        {error && (
          <div className="max-w-6xl mx-auto mb-4 rounded bg-red-50 text-red-700 px-3 py-2">
            {error}
          </div>
        )}

        {/* Loading skeleton */}
        {loading && (
          <div className="max-w-6xl mx-auto grid gap-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
            {Array.from({ length: 10 }).map((_, i) => (
              <div
                key={i}
                className="animate-pulse rounded-lg shadow-2xl overflow-hidden bg-white"
              >
                <div className="aspect-[3/4] bg-gray-200" />
                <div className="p-3 space-y-2">
                  <div className="h-4 bg-gray-200 rounded w-3/4" />
                  <div className="h-3 bg-gray-200 rounded w-1/2" />
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Empty */}
        {!loading && novelsInLibrary.length === 0 && (
          <div className="max-w-2xl mx-auto rounded-xl shadow-2xl bg-white p-8 text-center">
            <div className="mx-auto mb-3 h-12 w-12 rounded-full bg-blue-50 grid place-items-center text-blue-600">
              📚
            </div>
            <h2 className="text-lg font-semibold">Your library is empty</h2>
            <p className="text-gray-500 mt-1">
              Start exploring novels and add them to your library ✨
            </p>
          </div>
        )}

        {/* Grid */}
        {!loading && novelsInLibrary.length > 0 && (
          <div className="max-w-6xl mx-auto grid gap-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
            {novelsInLibrary.map((novel) => {
              const bmId = getBookmarkIdByNovelId(novel.id);
              const checked = bmId ? selected.has(bmId) : false;

              return (
                <div
                  key={novel.id}
                  className="relative rounded-lg shadow-2xl overflow-hidden bg-white hover:shadow transition"
                >
                  {/* Checkbox saat edit */}
                  {editMode && bmId && (
                    <label className="absolute top-2 left-2 bg-white/90 rounded px-2 py-1 text-sm cursor-pointer select-none shadow">
                      <input
                        type="checkbox"
                        className="mr-1 align-middle"
                        checked={checked}
                        onChange={() => toggleSelect(bmId)}
                      />
                      Select
                    </label>
                  )}

                  {/* Link ke detail novel */}
                  <NavLink
                    to={`/novel/${novel.id}`}
                    className="block focus:outline-none focus:ring-2 focus:ring-blue-400"
                  >
                    {/* Cover proporsional */}
                    <div className="aspect-[3/4] bg-gray-100">
                      {novel.cover_url ? (
                        <img
                          src={novel.cover_url}
                          alt={novel.title}
                          className="w-full h-full object-cover"
                          loading="lazy"
                        />
                      ) : (
                        <div className="w-full h-full grid place-items-center text-gray-400">
                          No Cover
                        </div>
                      )}
                    </div>

                    <div className="p-3">
                      <div className="font-medium line-clamp-2 hover:underline">
                        {novel.title}
                      </div>
                      {novel.author && (
                        <div className="text-sm text-gray-500 line-clamp-1">
                          {novel.author}
                        </div>
                      )}
                      {novel.progress != null && (
                        <div className="text-xs text-gray-400 mt-1">
                          Progress {novel.progress}
                        </div>
                      )}
                    </div>
                  </NavLink>

                  {/* Tombol add kalau belum ada di library (fallback) */}
                  {!bmId && (
                    <button
                      onClick={() => handleAddBookmark(novel.id)}
                      className="absolute bottom-2 right-2 text-sm text-blue-600 hover:underline bg-white/80 rounded px-2 py-1"
                    >
                      Add to Library
                    </button>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </main>
      <Footer />
    </>
  );
}
