import { useEffect, useState, useCallback } from "react";
import { NavLink, useNavigate } from "react-router-dom";
import { FaBookOpen, FaHome, FaPlusCircle } from "react-icons/fa";
import { toast } from "react-hot-toast";
import {
  getBookmarks,
  addBookmark,
  deleteBookmark,
} from "../../services/bookmarkService";

export default function NovelCard({ novel }) {
  const navigate = useNavigate();
  if (!novel) return null;

  const chapterCount = novel.chapters?.length ?? 0;
  const firstGenreName = novel.genres?.[0]?.name ?? "Unknown";
  const authorName = novel.author?.username ?? `ID: ${novel.author_id}`;
  const firstChapterNumber = 1;

  const toInt = (v) => (v === 0 || v ? Number(v) : null);

  const [bookmarkId, setBookmarkId] = useState(null); // number|null
  const [bmLoading, setBmLoading] = useState(false);
  const [bmError, setBmError] = useState("");
  const isBookmarked = Number.isInteger(bookmarkId);

  const normalizeBookmarks = (list) =>
    (list || [])
      .map((b) => ({ id: toInt(b?.id), novel_id: toInt(b?.novel_id) }))
      .filter((b) => Number.isInteger(b.id) && Number.isInteger(b.novel_id));

  const checkBookmark = useCallback(async () => {
    try {
      setBmError("");
      const raw = await getBookmarks();
      const list = normalizeBookmarks(raw);
      const nid = toInt(novel.id);
      const found = list.find((b) => b.novel_id === nid);
      setBookmarkId(found ? found.id : null);
    } catch (e) {
      setBmError(e?.message || "Failed to check bookmark");
    }
  }, [novel?.id]);

  useEffect(() => {
    if (!novel?.id) return;
    checkBookmark();
  }, [novel?.id, checkBookmark]);

  const handleAdd = async () => {
    try {
      setBmLoading(true);
      setBmError("");
      const res = await addBookmark(novel.id);
      setBookmarkId(toInt(res?.id));
      toast.success("Added to Library");
    } catch (e) {
      setBmError(e?.message || "Failed to add to library");
      if (e.message == "unauthorized") {
        navigate("/login");
      }
    } finally {
      setBmLoading(false);
    }
  };

  const handleRemove = async () => {
    if (!Number.isInteger(bookmarkId)) return;
    try {
      setBmLoading(true);
      setBmError("");
      await deleteBookmark(bookmarkId);
      setBookmarkId(null);
      toast.success("Removed from Library");
    } catch (e) {
      setBmError(e?.message || "Failed to remove from library");
      toast.error(e?.message || "Failed to remove from library");
    } finally {
      setBmLoading(false);
    }
  };

  return (
    <section className="flex w-full items-center justify-center bg-[#F5F6FC]">
      <div className="w-full max-w-4xl px-5 lg:px-0 pb-4">
        {/* Breadcrumb */}
        <p className="items-center text-md gap-1 text-primary font-bold py-6 hidden md:flex">
          <NavLink
            to="/"
            className="hover:underline hover:text-gray-500 flex items-center"
            aria-label="Home"
            title="Home"
          >
            <FaHome />
          </NavLink>
          &nbsp;/&nbsp;{novel.title || ""}
        </p>

        <div className="flex md:flex-row gap-6 pt-5 md:pt-0">
          {/* Cover */}
          <div className="flex">
            <div className="w-30 sm:w-44 md:w-52 lg:w-56 xl:w-64">
              <div className="aspect-[3/4] rounded-xl overflow-hidden shadow-md bg-gray-100">
                <img
                  src={novel.cover_image || novel.cover_url}
                  alt={novel.title}
                  className="w-full h-full object-cover"
                />
              </div>
            </div>
          </div>

          {/* Info */}
          <div className="flex-1 flex flex-col justify-between">
            <div>
              <h1 className="text-base sm:text-xl font-bold mb-2">
                {novel.title}
              </h1>
              {/* Meta */}
              <div className="flex flex-wrap items-center gap-4 mb-2 text-gray-600">
                <span className="bg-blue-100 text-primary px-3 py-1 rounded-full text-sm font-medium capitalize">
                  {firstGenreName}
                </span>

                <span className="items-center gap-2 text-sm hidden sm:flex">
                  <FaBookOpen /> {chapterCount} Chapters
                </span>

                <span className="text-sm hidden sm:flex items-center gap-1">
                  <span className="font-medium text-gray-700">Status:</span>{" "}
                  <span className="capitalize">
                    {novel.status || "ongoing"}
                  </span>
                </span>
              </div>
              <div className="text-gray-700">
                Author: <span className="font-medium">{authorName}</span>
              </div>
            </div>

            <div className="gap-3 hidden sm:flex">
              <NavLink
                to={`/novel/${novel.id}/chapter/${firstChapterNumber}`}
                className="bg-primary text-white px-4 py-2 rounded-lg hover:bg-gray-700 transition"
              >
                READ
              </NavLink>

              {isBookmarked ? (
                <button
                  className="border border-primary text-primary px-4 py-2 rounded-lg hover:bg-gray-200 transition disabled:opacity-50 cursor-pointer"
                  onClick={handleRemove}
                  disabled={bmLoading}
                  aria-label="Remove from Library"
                  title="Remove from Library"
                >
                  {bmLoading ? "Removing..." : "Remove From Library"}
                </button>
              ) : (
                <button
                  className="border border-primary text-primary px-4 py-2 rounded-lg hover:bg-gray-200 transition cursor-pointer disabled:opacity-50"
                  onClick={handleAdd}
                  disabled={bmLoading}
                  aria-label="Add to Library"
                  title="Add to Library"
                >
                  <FaPlusCircle className="inline-block mr-2" />
                  {bmLoading ? "ADDING..." : "ADD TO LIBRARY"}
                </button>
              )}
            </div>
          </div>
        </div>

        {/* Mobile quick stats */}
        <div className="flex sm:hidden items-stretch justify-between mt-3 rounded-2xl bg-white shadow-sm border border-gray-200 overflow-hidden">
          <div className=" flex flex-col items-center justify-center py-3 px-2">
            <span className="text-base font-semibold text-gray-900">
              {chapterCount}
            </span>
            <span className="text-xs text-gray-500 uppercase tracking-wide">
              Chapters
            </span>
          </div>

          <div className="w-px bg-gray-200" />

          <div className="flex-1 flex flex-col items-center justify-center py-3 px-2">
            <span className="text-base font-semibold text-gray-900 capitalize">
              {novel.status || "ongoing"}
            </span>
            <span className="text-xs text-gray-500 uppercase tracking-wide">
              Status
            </span>
          </div>

          <div className="w-px bg-gray-200" />

          {isBookmarked ? (
            <button
              className="flex-1 flex flex-col items-center justify-center py-3 px-2 text-white disabled:opacity-50 cursor-pointer bg-gray-500"
              onClick={handleRemove}
              disabled={bmLoading}
              aria-label="Remove from Library"
              title="Remove from Library"
            >
              <span className="text-xs font-medium">In Library</span>
            </button>
          ) : (
            <button
              className="flex-1 flex flex-col items-center justify-center py-3 px-2 text-primary disabled:opacity-50"
              onClick={handleAdd}
              disabled={bmLoading}
              aria-label="Add to Library"
              title="Add to Library"
            >
              <FaPlusCircle className="text-xl mb-1" />
              <span className="text-xs font-medium text-gray-500">Library</span>
            </button>
          )}
        </div>
      </div>
    </section>
  );
}
