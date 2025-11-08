import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { FaArrowLeft, FaCog, FaBook, FaTrash } from "react-icons/fa";
import { toast } from "react-hot-toast";
import { useAuth } from "../../../context/authContext";
import { getNovelById } from "../../../services/novelService";
import { deleteChapter } from "../../../services/chapterService";

const fmtDateTime = (ts) => {
  if (!ts) return "-";
  const d = new Date(ts);
  const day = String(d.getDate()).padStart(2, "0");
  const month = d.toLocaleString("en-GB", { month: "short" }); // Nov
  const year = d.getFullYear();
  const hh = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");
  return `${day} ${month} ${year} • ${hh}:${mm}`;
};

// 1) Modal dengan state deleting
function DeleteConfirmModal({ open, chapter, deleting, onCancel, onConfirm }) {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        className="absolute inset-0 bg-black/40"
        onClick={deleting ? undefined : onCancel}
      />
      <div className="relative w-full max-w-md rounded-xl bg-white shadow-xl border border-gray-200">
        <div className="px-5 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-800">
            Delete Chapter
          </h3>
        </div>
        <div className="px-5 py-4 text-gray-700">
          <p className="mb-2">Are you sure you want to delete:</p>
          <p className="font-medium break-words">
            {chapter
              ? `Chapter ${chapter.order_no ?? ""}: ${chapter.title}`
              : "-"}
          </p>
          <p className="text-sm text-gray-500 mt-2">
            This action cannot be undone.
          </p>
        </div>
        <div className="px-5 py-4 border-t border-gray-200 flex gap-3 justify-end">
          <button
            onClick={onCancel}
            disabled={deleting}
            className="px-4 py-2 rounded-lg border border-gray-300 hover:bg-gray-50 text-sm font-medium cursor-pointer disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            disabled={deleting}
            className="px-4 py-2 rounded-lg bg-red-600 hover:bg-red-700 text-white text-sm font-medium cursor-pointer disabled:opacity-50"
          >
            {deleting ? "Deleting…" : "Delete"}
          </button>
        </div>
      </div>
    </div>
  );
}

export default function AdminNovelsDetail() {
  const navigate = useNavigate();
  const { id } = useParams();
  const { user, isAuthenticated } = useAuth();

  const [novel, setNovel] = useState(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState("draft"); // 'draft' | 'published'
  const [synopsisExpanded, setSynopsisExpanded] = useState(false);

  const [delOpen, setDelOpen] = useState(false);
  const [delChapter, setDelChapter] = useState(null);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    (async () => {
      if (!isAuthenticated) return;
      setLoading(true);
      try {
        const data = await getNovelById(id);

        const ownerId = data?.author?.id ?? data?.author_id;
        if (ownerId && user?.id && Number(ownerId) !== Number(user.id)) {
          toast.error("You don't have access to that novel.");
          navigate("/admin/novels", { replace: true });
          return;
        }
        setNovel(data);
      } catch (e) {
        toast.error(e?.message || "Failed to load novel.");
        navigate("/admin/novels", { replace: true });
      } finally {
        setLoading(false);
      }
    })();
  }, [id, isAuthenticated, user?.id, navigate]);

  const chapters = novel?.chapters ?? [];

  const sorted = useMemo(() => {
    return [...chapters].sort((a, b) => {
      const ao = Number(a?.order_no ?? 999999);
      const bo = Number(b?.order_no ?? 999999);
      if (ao !== bo) return ao - bo;
      return new Date(a?.created_at || 0) - new Date(b?.created_at || 0);
    });
  }, [chapters]);

  const draftChapters = useMemo(
    () => sorted.filter((c) => !c?.published_at),
    [sorted]
  );
  const publishedChapters = useMemo(
    () => sorted.filter((c) => !!c?.published_at),
    [sorted]
  );

  const showSynopsisArrow =
    (novel?.synopsis?.length || 0) > 300 && !synopsisExpanded;
  const synopsisText = !novel?.synopsis
    ? "-"
    : synopsisExpanded || !showSynopsisArrow
    ? novel.synopsis
    : novel.synopsis.slice(0, 300) + "...";

  const goEditChapter = (ch) =>
    navigate(`/admin/novels/${novel.id}/edit-chapter/${ch.id}`);

  const askDelete = (ch) => {
    setDelChapter(ch);
    setDelOpen(true);
  };

  const confirmDelete = async () => {
    if (!delChapter?.id) return;
    try {
      setDeleting(true);
      await deleteChapter(delChapter.id);
      setNovel((prev) =>
        !prev
          ? prev
          : {
              ...prev,
              chapters: (prev.chapters || []).filter(
                (c) => c.id !== delChapter.id
              ),
            }
      );
      toast.success("Chapter deleted.");
    } catch (e) {
      toast.error(e?.message || "Failed to delete chapter.");
    } finally {
      setDeleting(false);
      setDelOpen(false);
      setDelChapter(null);
    }
  };

  if (loading) {
    return (
      <div className="p-4 md:p-6">
        <div className="bg-white border rounded-xl shadow-sm p-6 text-gray-500">
          Loading…
        </div>
      </div>
    );
  }
  if (!novel) return null;

  return (
    <div className="p-3 md:p-6 pb-24">
      <div className="bg-white border border-gray-200 rounded-xl shadow-sm">
        {/* Header */}
        <div className="flex flex-wrap gap-3 justify-between items-center px-4 md:px-5 py-4 border-b border-gray-200">
          <div className="flex items-center gap-3 min-w-0 flex-1">
            <button
              onClick={() => navigate("/admin/novels")}
              className="text-gray-600 hover:text-primary transition cursor-pointer"
              aria-label="Back"
            >
              <FaArrowLeft size={18} />
            </button>
            <h1 className="text-lg md:text-2xl font-semibold line-clamp-2 text-gray-800 whitespace-normal break-words leading-snug">
              {novel.title}
            </h1>
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              onClick={() => navigate(`/admin/novels/edit/${novel.id}`)}
              className="p-2.5 border border-gray-300 rounded-lg hover:bg-gray-50 cursor-pointer w-full sm:w-auto"
              aria-label="Settings"
              title="Edit Novel"
            >
              <div className="flex items-center justify-center gap-2">
                <FaCog />
                <span className="text-sm hidden sm:inline">Settings</span>
              </div>
            </button>
            <button
              onClick={() => navigate(`/admin/novels/${novel.id}/write`)}
              className="bg-primary hover:bg-gray-700 text-white font-medium rounded-lg px-4 py-2.5 flex items-center justify-center gap-2 text-sm shadow-sm cursor-pointer w-full sm:w-auto"
            >
              <FaBook size={14} />
              Create Chapter
            </button>
          </div>
        </div>

        {/* Novel Info */}
        <div className="flex flex-col md:flex-row gap-5 md:gap-6 px-4 md:px-5 py-6">
          <div className="flex-shrink-0 self-center md:self-start">
            <img
              src={novel.cover_image || novel.cover_url}
              alt={novel.title}
              className="w-32 h-44 md:w-40 md:h-56 object-cover rounded-md border border-gray-200"
            />
          </div>

          <div className="flex-1 min-w-0">
            <div className="flex flex-col items-center justify-center md:items-start">
              <h2 className="text-base md:text-lg font-semibold text-gray-900 mb-1 text-center md:text-left">
                {novel.title}
              </h2>

              <p className="text-gray-500 text-sm mb-2 flex flex-wrap gap-x-1">
                <span>BY</span>
                <span className="text-gray-800 font-medium">
                  {novel?.author?.username || "-"}
                </span>
                <span>/ IN</span>
                <span className="text-gray-800 font-medium">
                  {novel?.genres?.[0]?.name || "-"}
                </span>
              </p>
            </div>

            {/* Synopsis + Arrow bar */}
            <div className="relative">
              <div className="flex items-start gap-2 mb-2">
                <span className="text-gray-700 font-medium">Synopsis:</span>
              </div>

              <p className="text-gray-600 mb-5 whitespace-pre-wrap break-words">
                {synopsisText}
              </p>

              {(novel?.synopsis?.length || 0) > 300 && (
                <button
                  type="button"
                  onClick={() => setSynopsisExpanded((s) => !s)}
                  className="absolute -bottom-2 right-0 px-2 py-1 text-primary hover:text-gray-700 transition cursor-pointer"
                  aria-expanded={synopsisExpanded}
                  aria-label={synopsisExpanded ? "Show less" : "Show more"}
                >
                  <span className="text-sm">
                    {synopsisExpanded ? "▲" : "▼"}
                  </span>
                </button>
              )}
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-sm text-gray-700 border-t pt-3">
              <div>
                <p className="text-xs text-gray-500">STATUS</p>
                <p className="font-medium capitalize">
                  {String(novel.status || "").toLowerCase()}
                </p>
              </div>
              <div>
                <p className="text-xs text-gray-500">CHAPTERS</p>
                <p className="font-medium">{chapters.length}</p>
              </div>
              <div className="min-w-0">
                <p className="text-xs text-gray-500">TAGS</p>
                <p className="font-medium whitespace-normal break-words">
                  {(novel.tags || []).map((t) => t.name).join(", ") || "-"}
                </p>
              </div>
              <div>
                <p className="text-xs text-gray-500">CREATED</p>
                <p className="font-medium">
                  {fmtDateTime(novel.created_at).split(" • ")[0]}
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="px-4 md:px-5 border-t border-gray-200">
          <div className="flex gap-6 border-b border-gray-200 text-sm font-medium mt-2 md:mt-4 overflow-x-auto">
            {[
              { key: "draft", label: "DRAFT", count: draftChapters.length },
              {
                key: "published",
                label: "PUBLISHED",
                count: publishedChapters.length,
              },
            ].map(({ key, label, count }) => (
              <button
                key={key}
                onClick={() => setActiveTab(key)}
                className={`py-3 px-1 border-b-2 transition cursor-pointer ${
                  activeTab === key
                    ? "border-primary text-primary"
                    : "border-transparent text-gray-600 hover:text-gray-800"
                }`}
              >
                {label} <span className="text-gray-400">({count})</span>
              </button>
            ))}
          </div>

          <div className="py-8 md:py-10 flex flex-col items-center text-center">
            {/* DRAFT */}
            {activeTab === "draft" && (
              <>
                {draftChapters.length === 0 ? (
                  <>
                    <p className="text-gray-400 mb-2 font-medium">NO DRAFTS!</p>
                    <p className="text-gray-500 text-sm mb-4"></p>
                    <button
                      onClick={() =>
                        navigate(`/admin/novels/${novel.id}/write`)
                      }
                      className="bg-primary hover:bg-gray-700 text-white px-4 py-2 rounded-lg text-sm font-medium cursor-pointer"
                    >
                      CREATE NOW
                    </button>
                  </>
                ) : (
                  <div className="w-full max-w-3xl text-left">
                    {draftChapters.map((ch, idx) => (
                      <div
                        key={ch.id}
                        className="border border-gray-200 rounded-lg p-4 mb-2 hover:bg-gray-50"
                      >
                        <div className="flex items-center gap-3">
                          {/* LEFT: delete */}

                          {/* MIDDLE: title -> edit */}
                          <button
                            onClick={() => goEditChapter(ch)}
                            className="text-left flex-1 font-medium text-gray-800 whitespace-normal break-words cursor-pointer hover:underline"
                          >
                            {`Chapter ${ch?.order_no ?? idx + 1}: ${ch.title}`}
                          </button>

                          <button
                            onClick={() => askDelete(ch)}
                            className="text-red-600 hover:text-red-700 cursor-pointer"
                            title="Delete"
                            aria-label="Delete"
                          >
                            <FaTrash size={14} />
                          </button>

                          {/* RIGHT: time always visible */}
                          <span className="text-xs text-gray-500 shrink-0">
                            Saved {fmtDateTime(ch.updated_at)}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </>
            )}

            {/* PUBLISHED */}
            {activeTab === "published" && (
              <>
                {publishedChapters.length === 0 ? (
                  <p className="text-gray-400 italic">
                    No published chapters yet.
                  </p>
                ) : (
                  <div className="w-full max-w-3xl text-left">
                    {publishedChapters.map((ch, idx) => (
                      <div
                        key={ch.id}
                        className="border border-gray-200 rounded-lg p-4 mb-2 hover:bg-gray-50"
                      >
                        <div className="flex items-center gap-3">
                          {/* LEFT: delete */}

                          {/* MIDDLE: title -> edit */}
                          <button
                            onClick={() => goEditChapter(ch)}
                            className="text-left flex-1 font-medium text-gray-800 whitespace-normal break-words cursor-pointer hover:underline"
                          >
                            {`Chapter ${ch?.order_no ?? idx + 1}: ${ch.title}`}
                          </button>

                          <button
                            onClick={() => askDelete(ch)}
                            className="text-red-600 hover:text-red-700 cursor-pointer"
                            title="Delete"
                            aria-label="Delete"
                          >
                            <FaTrash size={14} />
                          </button>

                          {/* RIGHT: status + time always visible */}
                          <div className="flex items-center gap-2 shrink-0">
                            <span className="text-xs text-gray-500">
                              {fmtDateTime(ch.published_at)}
                            </span>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      </div>

      <DeleteConfirmModal
        open={delOpen}
        chapter={delChapter}
        deleting={deleting}
        onCancel={() => {
          if (deleting) return;
          setDelOpen(false);
          setDelChapter(null);
        }}
        onConfirm={confirmDelete}
      />
    </div>
  );
}
