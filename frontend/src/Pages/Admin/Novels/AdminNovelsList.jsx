import { useEffect, useMemo, useState, memo } from "react";
import {
  FaPlus,
  FaEdit,
  FaTrash,
  FaSort,
  FaSortUp,
  FaSortDown,
} from "react-icons/fa";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../../../context/AuthContext";
import { getNovels, deleteNovel } from "../../../services/novelService";
import { toast } from "react-hot-toast";

const ITEMS_PER_PAGE = 8;

// Inline Delete Modal
function DeleteModal({ show, title, onClose, onConfirm }) {
  if (!show) return null;
  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 backdrop-blur-[1px] p-4">
      <div className="w-full max-w-sm rounded-lg bg-white shadow-xl border border-gray-200">
        <div className="px-5 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-800">
            Confirm Delete
          </h3>
        </div>
        <div className="px-5 py-4 text-sm text-gray-700">
          <p className="mb-2">Are you sure you want to delete:</p>
          <p className="font-medium break-words">{title || "-"}</p>
          <p className="text-gray-500 mt-2">This action cannot be undone.</p>
        </div>
        <div className="px-5 py-4 border-t border-gray-200 flex justify-end gap-3">
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-lg border border-gray-300 hover:bg-gray-50 text-sm font-medium cursor-pointer"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className="px-4 py-2 rounded-lg bg-red-600 hover:bg-red-700 text-white text-sm font-medium cursor-pointer"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  );
}

export default function AdminNovelsList() {
  const navigate = useNavigate();
  const { user, isAuthenticated, isAdmin, initializing } = useAuth();

  const [myNovels, setMyNovels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(null);

  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [sortConfig, setSortConfig] = useState({
    key: "title",
    direction: "asc",
  });

  // Delete modal state
  const [deleteTarget, setDeleteTarget] = useState(null);
  const openDelete = (novel) => setDeleteTarget(novel);
  const closeDelete = () => setDeleteTarget(null);

  useEffect(() => {
    async function load() {
      if (initializing) return; // ⬅️ tunggu
      if (!isAuthenticated || !user || !isAdmin) {
        setMyNovels([]);
        setLoading(false);
        return;
      }
      setLoading(true);
      setLoadError(null);
      try {
        const list = await getNovels();
        setMyNovels(list || []);
      } catch (e) {
        setLoadError(e?.message || "Failed to load novels");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [initializing, isAuthenticated, isAdmin, user?.id]);

  const handleSort = (key) => {
    setSortConfig((prev) =>
      prev.key === key
        ? { key, direction: prev.direction === "asc" ? "desc" : "asc" }
        : { key, direction: "asc" }
    );
    setCurrentPage(1);
  };

  const confirmDelete = async () => {
    if (!deleteTarget) return;
    try {
      await deleteNovel(deleteTarget.id); // panggil API
      setMyNovels((prev) => prev.filter((n) => n.id !== deleteTarget.id));
      toast.success("Novel deleted");
    } catch (e) {
      toast.error(e?.message || "Failed to delete novel");
    } finally {
      closeDelete();
    }
  };

  const getChapterCount = (n) =>
    n?.chapters?.length ?? n?.chaptersCount ?? n?.chapters_count ?? 0;
  const fmtDate = (v) => (v ? String(v).slice(0, 10) : "-");

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    return (myNovels || []).filter((n) => {
      const status = String(n?.status || "").toLowerCase();
      const matchStatus = statusFilter === "all" || status === statusFilter;
      const matchSearch = String(n?.title || "")
        .toLowerCase()
        .includes(q);
      return matchStatus && matchSearch;
    });
  }, [myNovels, search, statusFilter]);

  const sorted = useMemo(() => {
    const arr = [...filtered];
    const { key, direction } = sortConfig;

    return arr.sort((a, b) => {
      if (key === "id") {
        const av = Number(a?.id ?? 0);
        const bv = Number(b?.id ?? 0);
        return direction === "asc" ? av - bv : bv - av;
      }
      if (key === "chapters") {
        const av = getChapterCount(a);
        const bv = getChapterCount(b);
        return direction === "asc" ? av - bv : bv - av;
      }
      const as = String(a?.[key] ?? "").toLowerCase();
      const bs = String(b?.[key] ?? "").toLowerCase();
      if (as < bs) return direction === "asc" ? -1 : 1;
      if (as > bs) return direction === "asc" ? 1 : -1;
      return 0;
    });
  }, [filtered, sortConfig]);

  const totalPages = Math.max(1, Math.ceil(sorted.length / ITEMS_PER_PAGE));
  const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
  const paginated = sorted.slice(startIndex, startIndex + ITEMS_PER_PAGE);

  const SortIcon = memo(function SortIcon({ column }) {
    if (sortConfig.key !== column)
      return <FaSort className="ml-1 text-gray-400" />;
    return sortConfig.direction === "asc" ? (
      <FaSortUp className="ml-1 text-primary" />
    ) : (
      <FaSortDown className="ml-1 text-primary" />
    );
  });

  const coverSrc = (n) => n?.cover_image || n?.cover_url || "";

  return (
    <div className="ml-0 p-3 md:p-6 pb-28">
      <div className="rounded-xl border border-gray-200 bg-white shadow-md overflow-hidden">
        {/* Header */}
        <div className="px-4 md:px-5 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h2 className="text-xl md:text-2xl font-semibold text-gray-800">
              My Novels
            </h2>
            <button
              onClick={() => navigate("/admin/novels/create")}
              className="hidden sm:inline-flex bg-primary hover:bg-gray-700 text-white rounded-lg px-4 py-2.5 items-center gap-2 text-sm font-medium shadow-sm transition cursor-pointer"
            >
              <FaPlus size={14} />
              Create Novel
            </button>
          </div>

          {/* Filters – stack on mobile */}
          <div className="mt-3 grid grid-cols-1 sm:grid-cols-3 gap-2 sm:gap-3">
            <input
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setCurrentPage(1);
              }}
              placeholder="Search by title"
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500"
            />
            <select
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value);
                setCurrentPage(1);
              }}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 cursor-pointer"
            >
              <option value="all">All</option>
              <option value="ongoing">Ongoing</option>
              <option value="completed">Completed</option>
              <option value="hiatus">Hiatus</option>
            </select>

            {/* Create on mobile */}
            <button
              onClick={() => navigate("/admin/novels/create")}
              className="sm:hidden bg-primary hover:bg-gray-700 text-white rounded-lg px-4 py-2.5 flex items-center justify-center gap-2 text-sm font-medium shadow-sm transition cursor-pointer"
            >
              <FaPlus size={14} />
              Create Novel
            </button>
          </div>
        </div>

        {/* Loading / Error */}
        {loading && <div className="px-5 py-6 text-gray-500">Loading…</div>}
        {loadError && !loading && (
          <div className="px-5 py-6 text-red-600">Error: {loadError}</div>
        )}

        {!loading && !loadError && (
          <div className="relative">
            {/* Mobile cards */}
            <div className="block md:hidden px-4 py-4 space-y-3">
              {paginated.length ? (
                paginated.map((novel) => (
                  <div
                    key={novel.id}
                    className="flex gap-3 rounded-lg border border-gray-200 p-3 shadow-sm"
                  >
                    <div className="flex-shrink-0">
                      {coverSrc(novel) ? (
                        <img
                          src={coverSrc(novel)}
                          alt={novel.title}
                          className="w-14 h-20 rounded-md object-cover border border-gray-200"
                        />
                      ) : (
                        <div className="w-14 h-20 rounded-md border border-gray-200 bg-gray-50" />
                      )}
                    </div>

                    <div className="min-w-0 flex-1">
                      <div
                        onClick={() => navigate(`/admin/novels/${novel.id}`)}
                        className="font-semibold text-primary hover:text-indigo-800 cursor-pointer line-clamp-2"
                      >
                        {novel.title}
                      </div>

                      <div className="mt-1 text-xs text-gray-500 flex flex-wrap gap-x-2 gap-y-1">
                        <span>ID: {novel.id}</span>
                        <span>•</span>
                        <span>Chapters: {getChapterCount(novel)}</span>
                      </div>

                      <div className="mt-1">
                        <span
                          className={`px-2 py-0.5 rounded-full text-xs font-semibold capitalize ${
                            String(novel.status || "").toLowerCase() ===
                            "ongoing"
                              ? "bg-blue-100 text-blue-700"
                              : String(novel.status || "").toLowerCase() ===
                                "completed"
                              ? "bg-green-100 text-green-700"
                              : "bg-yellow-100 text-yellow-700"
                          }`}
                        >
                          {novel.status}
                        </span>
                      </div>

                      <div className="mt-2 text-[11px] text-gray-500">
                        Created: {fmtDate(novel.created_at)} • Updated:{" "}
                        {fmtDate(novel.updated_at)}
                      </div>

                      <div className="mt-2 flex items-center gap-3">
                        <button
                          onClick={() =>
                            navigate(`/admin/novels/edit/${novel.id}`)
                          }
                          className="text-primary hover:text-indigo-800 cursor-pointer"
                          title="Edit metadata"
                        >
                          <FaEdit size={16} />
                        </button>
                        <button
                          onClick={() => openDelete(novel)}
                          className="text-red-600 hover:text-red-800 cursor-pointer"
                          title="Delete"
                        >
                          <FaTrash size={16} />
                        </button>
                      </div>
                    </div>
                  </div>
                ))
              ) : (
                <div className="text-center py-8 text-gray-500">
                  No novels found
                </div>
              )}
            </div>

            {/* Desktop table */}
            <div className="hidden md:block overflow-x-auto">
              <div className="max-h-[60vh] overflow-y-auto">
                <table className="w-full min-w-[900px] text-sm text-left text-gray-700 border-collapse">
                  <thead className="bg-indigo-50 text-gray-700 uppercase text-xs font-semibold border-b border-gray-200 sticky top-0 z-10">
                    <tr>
                      {[
                        "cover_image",
                        "id",
                        "title",
                        "chapters",
                        "status",
                        "created_at",
                        "updated_at",
                      ].map((col) => (
                        <th
                          key={col}
                          onClick={
                            col !== "cover_image"
                              ? () => handleSort(col)
                              : undefined
                          }
                          className={`px-6 py-3 font-semibold ${
                            col !== "cover_image"
                              ? "cursor-pointer select-none"
                              : "cursor-default"
                          }`}
                        >
                          <div className="flex items-center">
                            {col === "chapters"
                              ? "Chapters"
                              : String(col).replace("_", " ")}
                            {col !== "cover_image" && <SortIcon column={col} />}
                          </div>
                        </th>
                      ))}
                      <th className="px-6 py-3 text-center">Actions</th>
                    </tr>
                  </thead>

                  <tbody className="divide-y divide-gray-200">
                    {paginated.length > 0 ? (
                      paginated.map((novel) => (
                        <tr
                          key={novel.id}
                          className="hover:bg-gray-50 transition"
                        >
                          <td className="px-6 py-3">
                            {coverSrc(novel) ? (
                              <img
                                src={coverSrc(novel)}
                                alt={novel.title}
                                className="w-12 h-16 rounded-md object-cover border border-gray-200"
                              />
                            ) : (
                              <div className="w-12 h-16 rounded-md border border-gray-200 bg-gray-50" />
                            )}
                          </td>
                          <td className="px-6 py-3 font-medium">{novel.id}</td>
                          <td
                            className="px-6 py-3 font-medium text-primary hover:text-indigo-800 cursor-pointer whitespace-nowrap"
                            onClick={() =>
                              navigate(`/admin/novels/${novel.id}`)
                            }
                          >
                            {novel.title}
                          </td>
                          <td className="px-6 py-3 text-center">
                            {getChapterCount(novel)}
                          </td>
                          <td className="px-6 py-3 capitalize">
                            <span
                              className={`px-2.5 py-1 rounded-full text-xs font-semibold ${
                                String(novel.status || "").toLowerCase() ===
                                "ongoing"
                                  ? "bg-blue-100 text-blue-700"
                                  : String(novel.status || "").toLowerCase() ===
                                    "completed"
                                  ? "bg-green-100 text-green-700"
                                  : "bg-yellow-100 text-yellow-700"
                              }`}
                            >
                              {novel.status}
                            </span>
                          </td>
                          <td className="px-6 py-3 whitespace-nowrap">
                            {fmtDate(novel.created_at)}
                          </td>
                          <td className="px-6 py-3 whitespace-nowrap">
                            {fmtDate(novel.updated_at)}
                          </td>
                          <td className="px-6 py-3 text-center">
                            <div className="flex justify-center gap-3">
                              <button
                                onClick={() =>
                                  navigate(`/admin/novels/edit/${novel.id}`)
                                }
                                className="text-primary hover:text-indigo-800 transition cursor-pointer"
                                title="Edit metadata"
                              >
                                <FaEdit size={16} />
                              </button>
                              <button
                                onClick={() => openDelete(novel)}
                                className="text-red-600 hover:text-red-800 transition cursor-pointer"
                                title="Delete"
                              >
                                <FaTrash size={16} />
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))
                    ) : (
                      <tr>
                        <td
                          colSpan={10}
                          className="text-center py-8 text-gray-500"
                        >
                          No novels found
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Pagination (desktop & mobile sama agar konsisten) */}
            <div className="border-t border-gray-200 bg-gray-50 px-4 md:px-5 py-4 sticky bottom-0">
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600">
                  Showing{" "}
                  <span className="font-semibold text-gray-800">
                    {sorted.length ? startIndex + 1 : 0}
                  </span>{" "}
                  to{" "}
                  <span className="font-semibold text-gray-800">
                    {Math.min(startIndex + ITEMS_PER_PAGE, sorted.length)}
                  </span>{" "}
                  of{" "}
                  <span className="font-semibold text-gray-800">
                    {sorted.length}
                  </span>
                </span>

                <div className="flex items-center gap-2">
                  <button
                    onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                    disabled={currentPage === 1}
                    className={`px-3 py-1.5 rounded-lg text-sm font-medium ${
                      currentPage === 1
                        ? "text-gray-400 cursor-not-allowed"
                        : "text-primary hover:bg-indigo-50 cursor-pointer"
                    }`}
                  >
                    Prev
                  </button>
                  <ul className="flex items-center gap-1">
                    {Array.from({ length: totalPages }, (_, i) => (
                      <li key={i}>
                        <button
                          onClick={() => setCurrentPage(i + 1)}
                          className={`px-3 py-1.5 rounded-lg text-sm font-medium ${
                            currentPage === i + 1
                              ? "bg-primary text-white"
                              : "text-gray-700 hover:bg-gray-100"
                          } cursor-pointer`}
                        >
                          {i + 1}
                        </button>
                      </li>
                    ))}
                  </ul>
                  <button
                    onClick={() =>
                      setCurrentPage((p) => Math.min(totalPages, p + 1))
                    }
                    disabled={currentPage === totalPages}
                    className={`px-3 py-1.5 rounded-lg text-sm font-medium ${
                      currentPage === totalPages
                        ? "text-gray-400 cursor-not-allowed"
                        : "text-primary hover:bg-indigo-50 cursor-pointer"
                    }`}
                  >
                    Next
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Inline Delete Modal */}
      <DeleteModal
        show={!!deleteTarget}
        title={deleteTarget?.title}
        onClose={closeDelete}
        onConfirm={confirmDelete}
      />
    </div>
  );
}
