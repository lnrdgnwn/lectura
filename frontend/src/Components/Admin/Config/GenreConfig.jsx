import { useEffect, useMemo, useState } from "react";
import { toast } from "react-hot-toast";
import {
  getGenres,
  getHomeGenres,
  setGenresShowOnHome,
  createGenre,
  updateGenre,
  deleteGenre,
} from "../../../services/genreService";

function Modal({ open, title, children, onClose }) {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-end md:items-center justify-center">
      <div className="absolute inset-0 bg-black/40" onClick={onClose} />
      <div className="w-full md:max-w-md bg-white rounded-t-2xl md:rounded-2xl shadow-xl relative p-5">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-lg font-semibold text-gray-800">{title}</h3>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700 cursor-pointer"
            aria-label="Close"
          >
            ✕
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}

function Confirm({ open, message, onCancel, onConfirm }) {
  if (!open) return null;
  return (
    <Modal open title="Confirm" onClose={onCancel}>
      <p className="text-sm text-gray-700">{message}</p>
      <div className="flex justify-end gap-2 mt-4">
        <button
          onClick={onCancel}
          className="px-4 py-2 rounded-lg border border-gray-300 text-sm hover:bg-gray-50 cursor-pointer"
        >
          Cancel
        </button>
        <button
          onClick={onConfirm}
          className="px-4 py-2 rounded-lg bg-red-600 hover:bg-red-700 text-white text-sm cursor-pointer"
        >
          Delete
        </button>
      </div>
    </Modal>
  );
}

/* ---------- util ---------- */
const slugify = (text = "") =>
  String(text)
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .replace(/--+/g, "-");

const getHomeIds = (list) =>
  (list || []).filter((g) => g.show_on_home).map((g) => g.id);

export default function GenreConfig() {
  const [rows, setRows] = useState([]);
  const [loading, setLoading] = useState(true);

  const [search, setSearch] = useState("");
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState(null);

  const [confirmOpen, setConfirmOpen] = useState(false);
  const [deleteId, setDeleteId] = useState(null);

  const [form, setForm] = useState({ name: "", show_on_home: false });

  // untuk disable checkbox saat request
  const [busyId, setBusyId] = useState(null);

  /* ---------- modal ---------- */
  const openCreate = () => {
    setEditing(null);
    setForm({ name: "", show_on_home: false });
    setModalOpen(true);
  };
  const openEdit = (g) => {
    setEditing(g);
    setForm({ name: g?.name || "", show_on_home: !!g?.show_on_home });
    setModalOpen(true);
  };
  const closeModal = () => setModalOpen(false);

  /* ---------- load ---------- */
  useEffect(() => {
    let alive = true;
    (async () => {
      setLoading(true);
      try {
        const [all, home] = await Promise.all([getGenres(), getHomeGenres()]);
        if (!alive) return;

        const homeSet = new Set((home || []).map((h) => h.id));
        const merged = (all || []).map((g) => ({
          ...g,
          show_on_home: homeSet.has(g.id),
        }));

        setRows(merged);
      } catch (e) {
        toast.error(e?.message || "Failed to load genres");
      } finally {
        if (alive) setLoading(false);
      }
    })();
    return () => {
      alive = false;
    };
  }, []);

  /* ---------- filter ---------- */
  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    return rows.filter(
      (g) =>
        !q ||
        (g?.name || "").toLowerCase().includes(q) ||
        (g?.slug || "").toLowerCase().includes(q)
    );
  }, [rows, search]);

  /* ---------- save (create/update) ---------- */
  const save = async () => {
    const name = form.name.trim();
    if (!name) return toast.error("Name is required");

    const payload = { name, slug: slugify(name) };

    try {
      if (editing) {
        const updated = await updateGenre(editing.id, payload);

        // gabungkan hasil update + flag dari form
        const merged = rows.map((r) =>
          r.id === editing.id
            ? { ...r, ...updated, show_on_home: !!form.show_on_home }
            : r
        );
        setRows(merged);

        // sinkronkan flag global ke server
        await setGenresShowOnHome(getHomeIds(merged));
        toast.success("Genre updated");
      } else {
        const created = await createGenre(payload);

        const merged = [
          {
            ...(created || payload),
            id: created?.id ?? Math.random(),
            show_on_home: !!form.show_on_home,
          },
          ...rows,
        ];
        setRows(merged);

        // sinkronkan flag global ke server
        await setGenresShowOnHome(getHomeIds(merged));
        toast.success("Genre created");
      }

      closeModal();
    } catch (e) {
      toast.error(e?.message || "Failed to save genre");
    }
  };

  /* ---------- delete ---------- */
  const askDelete = (id) => {
    setDeleteId(id);
    setConfirmOpen(true);
  };

  const confirmDelete = async () => {
    const id = deleteId;
    setConfirmOpen(false);
    try {
      const merged = rows.filter((r) => r.id !== id);
      setRows(merged);

      await deleteGenre(id);
      // sinkronkan flag home (kalau yang dihapus tadinya ON)
      await setGenresShowOnHome(getHomeIds(merged));

      toast.success("Genre deleted");
    } catch (e) {
      toast.error(e?.message || "Failed to delete genre");
    }
  };

  /* ---------- toggle show_on_home (bulk endpoint) ---------- */
  async function toggleShow(g) {
    const prev = rows;
    const next = rows.map((r) =>
      r.id === g.id ? { ...r, show_on_home: !g.show_on_home } : r
    );

    setBusyId(g.id);
    setRows(next); // optimistic

    try {
      await setGenresShowOnHome(getHomeIds(next));
      toast.success("Home genres updated");
    } catch (e) {
      setRows(prev); // rollback
      toast.error(e?.message || "Failed to update home genres");
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div>
      {/* header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-3 mb-4">
        <h2 className="text-lg font-semibold text-gray-800">Genres</h2>
        <div className="flex justify-between gap-2">
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search name/slug…"
            className="md:w-64 border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500"
          />
          <button
            onClick={openCreate}
            className="bg-primary hover:bg-gray-700 text-white rounded-lg px-4 py-2 text-sm font-medium shadow-sm cursor-pointer"
          >
            + Add Genre
          </button>
        </div>
      </div>

      {/* table */}
      <div className="overflow-x-auto">
        <table className="w-full min-w-[720px] text-sm text-left text-gray-700 border-collapse">
          <thead className="bg-indigo-50 text-gray-700 uppercase text-xs font-semibold border-b border-gray-2 00">
            <tr>
              <th className="px-4 py-3">ID</th>
              <th className="px-4 py-3">Name</th>
              <th className="px-4 py-3">Slug</th>
              <th className="px-4 py-3 text-center">Show on Home</th>
              <th className="px-4 py-3 text-center">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {loading ? (
              <tr>
                <td colSpan="5" className="text-center py-8 text-gray-500">
                  Loading…
                </td>
              </tr>
            ) : filtered.length ? (
              filtered.map((g) => (
                <tr key={g.id} className="hover:bg-gray-50 transition">
                  <td className="px-4 py-3">{g.id}</td>
                  <td className="px-4 py-3 font-medium">{g.name}</td>
                  <td className="px-4 py-3">{g.slug}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center justify-center">
                      <label className="inline-flex items-center gap-2 cursor-pointer select-none">
                        <input
                          type="checkbox"
                          className="w-4 h-4"
                          checked={!!g.show_on_home}
                          disabled={busyId === g.id}
                          onChange={() => toggleShow(g)}
                        />
                        <span className="text-xs">
                          {g.show_on_home ? "Yes" : "No"}
                        </span>
                      </label>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center justify-center gap-2">
                      <button
                        onClick={() => openEdit(g)}
                        className="px-3 py-1.5 rounded-lg text-xs font-medium border border-gray-300 hover:bg-gray-50 cursor-pointer"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => askDelete(g.id)}
                        className="px-3 py-1.5 rounded-lg text-xs font-medium border border-red-300 text-red-600 hover:bg-red-50 cursor-pointer"
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan="5" className="text-center py-8 text-gray-500">
                  No genres found
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* modal form */}
      <Modal
        open={modalOpen}
        title={editing ? "Edit Genre" : "Add Genre"}
        onClose={closeModal}
      >
        <div className="space-y-3">
          <div>
            <label className="text-sm font-medium text-gray-700 mb-1 block">
              Name
            </label>
            <input
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              placeholder="Male Lead"
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500"
            />
          </div>

          <div>
            <label className="text-sm font-medium text-gray-700 mb-1 block">
              Slug (auto)
            </label>
            <input
              value={slugify(form.name)}
              readOnly
              className="w-full border border-gray-200 bg-gray-50 rounded-lg px-3 py-2 text-sm"
            />
          </div>

          <label className="inline-flex items-center gap-2 text-sm text-gray-700 select-none cursor-pointer">
            <input
              type="checkbox"
              checked={!!form.show_on_home}
              onChange={(e) =>
                setForm((f) => ({ ...f, show_on_home: e.target.checked }))
              }
              className="w-4 h-4 rounded border-gray-300"
            />
            Show on Home
          </label>

          <div className="flex justify-end gap-2 pt-2">
            <button
              onClick={closeModal}
              className="px-4 py-2 rounded-lg border border-gray-300 text-sm hover:bg-gray-50 cursor-pointer"
            >
              Cancel
            </button>
            <button
              onClick={save}
              className="px-4 py-2 rounded-lg bg-primary hover:bg-gray-700 text-white text-sm cursor-pointer"
            >
              {editing ? "Save Changes" : "Create"}
            </button>
          </div>
        </div>
      </Modal>

      {/* confirm delete */}
      <Confirm
        open={confirmOpen}
        message="Delete this genre?"
        onCancel={() => setConfirmOpen(false)}
        onConfirm={confirmDelete}
      />
    </div>
  );
}
