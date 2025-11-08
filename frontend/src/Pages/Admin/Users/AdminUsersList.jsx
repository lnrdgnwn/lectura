// pages/AdminUsersList.jsx
import { useEffect, useMemo, useState } from "react";
import { FaSort, FaSortUp, FaSortDown, FaTrash } from "react-icons/fa";
import { getUsers, deleteUser } from "../../../services/userService";

export default function AdminUsersList() {
  const [search, setSearch] = useState("");
  const [debounced, setDebounced] = useState("");
  const [roleFilter, setRoleFilter] = useState("all");

  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;

  const [sortConfig, setSortConfig] = useState({
    key: "username",
    direction: "asc",
  });

  const [allUsers, setAllUsers] = useState([]);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState("");

  // Debounce search
  useEffect(() => {
    const t = setTimeout(() => {
      setDebounced(search.trim());
      setCurrentPage(1);
    }, 300);
    return () => clearTimeout(t);
  }, [search]);

  // Fetch users
  useEffect(() => {
    let alive = true;
    (async () => {
      setLoading(true);
      setErr("");
      try {
        const res = await getUsers();
        if (!alive) return;
        const list = Array.isArray(res?.data)
          ? res.data
          : Array.isArray(res)
          ? res
          : [];
        setAllUsers(list);
      } catch (e) {
        if (!alive) return;
        setErr("Failed to fetch users");
      } finally {
        if (alive) setLoading(false);
      }
    })();
    return () => {
      alive = false;
    };
  }, []);

  // Filter
  const filteredUsers = useMemo(() => {
    const term = debounced.toLowerCase();
    return allUsers.filter((u) => {
      const roleOk =
        roleFilter === "all"
          ? true
          : (u.role ?? "").toLowerCase() === roleFilter;
      const hayUsername = (u.username ?? "").toLowerCase();
      const hayEmail = (u.email ?? "").toLowerCase();
      const searchOk =
        !term || hayUsername.includes(term) || hayEmail.includes(term);
      return roleOk && searchOk;
    });
  }, [allUsers, debounced, roleFilter]);

  // Sort
  const sortedUsers = useMemo(() => {
    const arr = [...filteredUsers];
    const { key, direction } = sortConfig;
    const dir = direction === "asc" ? 1 : -1;

    return arr.sort((a, b) => {
      const avRaw =
        key === "created_at"
          ? a.created_at ?? a.created
          : key === "updated_at"
          ? a.updated_at ?? a.updated
          : a[key];

      const bvRaw =
        key === "created_at"
          ? b.created_at ?? b.created
          : key === "updated_at"
          ? b.updated_at ?? b.updated
          : b[key];

      const av = (avRaw ?? "").toString().toLowerCase();
      const bv = (bvRaw ?? "").toString().toLowerCase();

      if (av < bv) return -1 * dir;
      if (av > bv) return 1 * dir;
      return 0;
    });
  }, [filteredUsers, sortConfig]);

  // Pagination
  const total = sortedUsers.length;
  const totalPages = Math.max(1, Math.ceil(total / itemsPerPage));
  const startIndex = (currentPage - 1) * itemsPerPage;
  const pageRows = useMemo(
    () => sortedUsers.slice(startIndex, startIndex + itemsPerPage),
    [sortedUsers, startIndex, itemsPerPage]
  );

  const handleSort = (key) => {
    setSortConfig((prev) =>
      prev.key === key
        ? { key, direction: prev.direction === "asc" ? "desc" : "asc" }
        : { key, direction: "asc" }
    );
  };

  const handlePrev = () => setCurrentPage((p) => Math.max(p - 1, 1));
  const handleNext = () => setCurrentPage((p) => Math.min(p + 1, totalPages));

  // Delete only
  const handleDelete = async (id) => {
    if (!confirm("Delete this user?")) return;
    try {
      await deleteUser(id);
      // update state lokal
      setAllUsers((prev) => prev.filter((u) => u.id !== id));
      // sesuaikan page bila kosong setelah delete
      const newTotal = total - 1;
      const newTotalPages = Math.max(1, Math.ceil(newTotal / itemsPerPage));
      setCurrentPage((p) => Math.min(p, newTotalPages));
    } catch {
      alert("Delete failed");
    }
  };

  const SortIcon = ({ column }) => {
    if (sortConfig.key !== column)
      return <FaSort className="ml-1 text-gray-400" />;
    return sortConfig.direction === "asc" ? (
      <FaSortUp className="ml-1 text-indigo-600" />
    ) : (
      <FaSortDown className="ml-1 text-indigo-600" />
    );
  };

  const defaultAvatar = "/img/default-avatar.png";
  const getAvatarSrc = (u) => u?.profile_picture || u?.avatar || defaultAvatar;

  return (
    <div className="p-0 md:p-6 pb-24">
      <div className="rounded-xl border border-gray-200 bg-white shadow-md overflow-x-auto ">
        {/* Header */}
        <div className="flex flex-wrap items-center justify-between gap-3 px-5 py-4 border-b border-gray-200">
          <h2 className="text-2xl font-semibold text-gray-800">
            User Management
          </h2>
        </div>

        {/* Filters */}
        <div className="flex flex-col md:flex-row justify-between gap-4 px-5 py-4 border-b border-gray-200 bg-gray-50">
          <div className="flex items-center gap-3">
            <label className="text-sm font-medium text-gray-700">
              Filter by Role:
            </label>
            <select
              value={roleFilter}
              onChange={(e) => {
                setRoleFilter(e.target.value);
                setCurrentPage(1);
              }}
              className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="all">All</option>
              <option value="admin">Admin</option>
              <option value="user">User</option>
            </select>
          </div>

          {/* Search */}
          <div className="flex gap-2 items-center w-full md:w-auto">
            <input
              type="text"
              placeholder="Search username or email…"
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setCurrentPage(1);
              }}
              className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 w/full md:w-72"
            />
          </div>
        </div>

        {/* Table */}
        <div className="overflow-hidden">
          <div className="overflow-x-auto max-w-full">
            <table className="w-full min-w-[760px] text-sm text-left text-gray-700 border-collapse">
              <thead className="bg-indigo-50 text-gray-700 uppercase text-xs font-semibold border-b border-gray-200">
                <tr>
                  {[
                    "avatar",
                    "username",
                    "email",
                    "role",
                    "created_at",
                    "updated_at",
                  ].map((col) => (
                    <th
                      key={col}
                      className={`px-6 py-3 font-semibold whitespace-nowrap ${
                        col !== "avatar" ? "cursor-pointer select-none" : ""
                      }`}
                      onClick={
                        col !== "avatar" ? () => handleSort(col) : undefined
                      }
                    >
                      <div className="flex items-center">
                        {col === "avatar"
                          ? "Profile"
                          : col === "created_at"
                          ? "Created At"
                          : col === "updated_at"
                          ? "Updated At"
                          : col.charAt(0).toUpperCase() + col.slice(1)}
                        {col !== "avatar" && <SortIcon column={col} />}
                      </div>
                    </th>
                  ))}
                  {/* Actions: Delete only */}
                  <th className="px-6 py-3 text-center font-semibold whitespace-nowrap">
                    Actions
                  </th>
                </tr>
              </thead>

              <tbody className="divide-y divide-gray-200">
                {loading ? (
                  <tr>
                    <td colSpan="7" className="text-center py-8 text-gray-500">
                      Loading…
                    </td>
                  </tr>
                ) : err ? (
                  <tr>
                    <td colSpan="7" className="text-center py-8 text-red-600">
                      {err}
                    </td>
                  </tr>
                ) : pageRows.length > 0 ? (
                  pageRows.map((user) => {
                    const created =
                      user.created_at?.slice?.(0, 10) || user.created || "-";
                    const updated =
                      user.updated_at?.slice?.(0, 10) || user.updated || "-";

                    return (
                      <tr key={user.id} className="hover:bg-gray-50 transition">
                        <td className="px-6 py-3">
                          <img
                            src={getAvatarSrc(user)}
                            onError={(e) => {
                              if (e.currentTarget.src !== defaultAvatar) {
                                e.currentTarget.src = defaultAvatar;
                              }
                            }}
                            alt={user.username || user.email || "User"}
                            className="w-10 h-10 rounded-full border border-gray-300 object-cover"
                            referrerPolicy="no-referrer"
                          />
                        </td>
                        <td className="px-6 py-3 font-medium whitespace-nowrap">
                          {user.username}
                        </td>
                        <td className="px-6 py-3 whitespace-nowrap">
                          {user.email}
                        </td>
                        <td className="px-6 py-3 capitalize whitespace-nowrap">
                          <span
                            className={`px-2.5 py-1 rounded-full text-xs font-semibold ${
                              user.role === "admin"
                                ? "bg-indigo-100 text-indigo-700"
                                : user.role === "editor"
                                ? "bg-yellow-100 text-yellow-700"
                                : "bg-gray-100 text-gray-700"
                            }`}
                          >
                            {user.role}
                          </span>
                        </td>
                        <td className="px-6 py-3 whitespace-nowrap">
                          {created}
                        </td>
                        <td className="px-6 py-3 whitespace-nowrap">
                          {updated}
                        </td>
                        <td className="px-6 py-3 text-center">
                          <button
                            onClick={() => handleDelete(user.id)}
                            className="text-red-600 hover:text-red-800 transition"
                            title="Delete user"
                          >
                            <FaTrash size={16} />
                          </button>
                        </td>
                      </tr>
                    );
                  })
                ) : (
                  <tr>
                    <td colSpan="7" className="text-center py-8 text-gray-500">
                      No users found
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Pagination */}
        <div className="flex flex-col sm:flex-row items-center justify-between border-t border-gray-200 px-5 py-4 bg-gray-50">
          <span className="text-sm text-gray-600">
            Showing{" "}
            <span className="font-semibold text-gray-800">
              {total === 0 ? 0 : startIndex + 1}
            </span>{" "}
            to{" "}
            <span className="font-semibold text-gray-800">
              {Math.min(startIndex + itemsPerPage, total)}
            </span>{" "}
            of <span className="font-semibold text-gray-800">{total}</span>
          </span>

          <div className="flex items-center gap-2 mt-3 sm:mt-0">
            <button
              onClick={handlePrev}
              disabled={currentPage === 1}
              className={`px-3 py-1.5 rounded-lg text-sm font-medium ${
                currentPage === 1
                  ? "text-gray-400 cursor-not-allowed"
                  : "text-indigo-600 hover:bg-indigo-50"
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
                        ? "bg-indigo-600 text-white"
                        : "text-gray-700 hover:bg-gray-100"
                    }`}
                  >
                    {i + 1}
                  </button>
                </li>
              ))}
            </ul>
            <button
              onClick={handleNext}
              disabled={currentPage >= totalPages}
              className={`px-3 py-1.5 rounded-lg text-sm font-medium ${
                currentPage >= totalPages
                  ? "text-gray-400 cursor-not-allowed"
                  : "text-indigo-600 hover:bg-indigo-50"
              }`}
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
