import { useEffect, useMemo, useState } from "react";
import { toast } from "react-hot-toast";
import {
  getTags,
  createTag,
  updateTag,
  deleteTag,
} from "../../../services/TagService";

/* ---------- UI helpers ---------- */
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

/* ---------- inline slugify (tanpa util) ---------- */
const slugify = (input = "") =>
  String(input)
    .toLowerCase()
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .replace(/--+/g, "-");

export default function TagConfig() {
  const [rows, setRows] = useState([]);
  const [loading, setLoading] = useState(true);

  const [search, setSearch] = useState("");
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState(null);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [deleteId, setDeleteId] = useState(null);

  const [form, setForm] = useState({ name: "" });

  const openCreate = () => {
    setEditing(null);
    setForm({ name: "" });
    setModalOpen(true);
  };
  const openEdit = (t) => {
    setEditing(t);
    setForm({ name: t?.name || "" });
    setModalOpen(true);
  };
  const closeModal = () => setModalOpen(false);

  useEffect(() => {
    let alive = true;
    (async () => {
      setLoading(true);
      try {
        const list = await getTags();
        if (!alive) return;
        setRows(Array.isArray(list) ? list : []);
      } catch (e) {
        toast.error(e?.message || "Failed to load tags");
      } finally {
        if (alive) setLoading(false);
      }
    })();
    return () => {
      alive = false;
    };
  }, []);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    return rows.filter(
      (t) =>
        !q ||
        (t?.name || "").toLowerCase().includes(q) ||
        (t?.slug || "").toLowerCase().includes(q)
    );
  }, [rows, search]);

  const save = async () => {
    const name = form.name.trim();
    if (!name) return toast.error("Name is required");

    const payload = { name, slug: slugify(name) };

    try {
      if (editing) {
        const updated = await updateTag(editing.id, payload);
        setRows((prev) =>
          prev.map((r) => (r.id === editing.id ? { ...r, ...updated } : r))
        );
        toast.success("Tag updated");
      } else {
        const created = await createTag(payload);
        setRows((prev) => [
          { ...(created || payload), id: created?.id ?? Math.random() },
          ...prev,
        ]);
        toast.success("Tag created");
      }
      closeModal();
    } catch (e) {
      toast.error(e?.message || "Failed to save tag");
    }
  };

  const askDelete = (id) => {
    setDeleteId(id);
    setConfirmOpen(true);
  };

  const confirmDelete = async () => {
    const id = deleteId;
    setConfirmOpen(false);
    try {
      await deleteTag(id);
      setRows((prev) => prev.filter((r) => r.id !== id));
      toast.success("Tag deleted");
    } catch (e) {
      toast.error(e?.message || "Failed to delete tag");
    }
  };

  return (
    <div>
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-3 mb-4">
        <h2 className="text-lg font-semibold text-gray-800">Tags</h2>
        <div className="flex  justify-between gap-2">
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search name/slug…"
            className=" md:w-64 border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500"
          />
          <button
            onClick={openCreate}
            className="bg-primary hover:bg-gray-700 text-white rounded-lg px-4 py-2 text-sm font-medium shadow-sm cursor-pointer"
          >
            + Add Tag
          </button>
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full min-w-[680px] text-sm text-left text-gray-700 border-collapse">
          <thead className="bg-indigo-50 text-gray-700 uppercase text-xs font-semibold border-b border-gray-200">
            <tr>
              <th className="px-4 py-3">ID</th>
              <th className="px-4 py-3">Name</th>
              <th className="px-4 py-3">Slug</th>
              <th className="px-4 py-3 text-center">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {loading ? (
              <tr>
                <td colSpan="4" className="text-center py-8 text-gray-500">
                  Loading…
                </td>
              </tr>
            ) : filtered.length ? (
              filtered.map((t) => (
                <tr key={t.id} className="hover:bg-gray-50 transition">
                  <td className="px-4 py-3">{t.id}</td>
                  <td className="px-4 py-3 font-medium">{t.name}</td>
                  <td className="px-4 py-3">{t.slug}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center justify-center gap-2">
                      <button
                        onClick={() => openEdit(t)}
                        className="px-3 py-1.5 rounded-lg text-xs font-medium border border-gray-300 hover:bg-gray-50 cursor-pointer"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => askDelete(t.id)}
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
                <td colSpan="4" className="text-center py-8 text-gray-500">
                  No tags found
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Modal
        open={modalOpen}
        title={editing ? "Edit Tag" : "Add Tag"}
        onClose={closeModal}
      >
        <div className="space-y-3">
          <div>
            <label className="text-sm font-medium text-gray-700 mb-1 block">
              Name
            </label>
            <input
              value={form.name}
              onChange={(e) => setForm({ name: e.target.value })}
              placeholder="Magic"
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

      <Confirm
        open={confirmOpen}
        message="Delete this tag?"
        onCancel={() => setConfirmOpen(false)}
        onConfirm={confirmDelete}
      />
    </div>
  );
}
