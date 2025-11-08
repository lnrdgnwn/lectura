import { useEffect, useState, useRef, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "react-hot-toast";
import CoverUploader from "./CoverUploader";
import TagSelector from "./TagSelector";
import {
  getNovelById,
  createNovel,
  updateNovel,
} from "../../../services/novelService";
import { getGenres } from "../../../services/genreService";
import { getTags } from "../../../services/TagService";

export default function NovelForm({ mode = "create", novelId }) {
  const navigate = useNavigate();

  const [title, setTitle] = useState("");
  const [synopsis, setSynopsis] = useState("");
  const [status, setStatus] = useState("ongoing");
  const [genreId, setGenreId] = useState("");
  const [selectedTags, setSelectedTags] = useState([]);
  const [coverFile, setCoverFile] = useState(null);
  const [preview, setPreview] = useState(null);

  const [genres, setGenres] = useState([]);
  const [tags, setTags] = useState([]);

  const [loading, setLoading] = useState(true);
  const [prefilling, setPrefilling] = useState(mode === "edit");
  const [submitting, setSubmitting] = useState(false);

  const lastURL = useRef(null);
  useEffect(
    () => () => {
      if (lastURL.current) URL.revokeObjectURL(lastURL.current);
    },
    []
  );

  useEffect(() => {
    (async () => {
      try {
        const [g, t] = await Promise.all([getGenres(), getTags()]);
        setGenres(g ?? []);
        setTags(
          (t ?? [])
            .map((x) => ({ id: String(x.id), name: x.name }))
            .filter((x) => x.id && x.name)
        );

        if (mode === "edit" && novelId) {
          setPrefilling(true);
          const data = await getNovelById(novelId);
          setTitle(data?.title ?? "");
          setSynopsis(data?.synopsis ?? "");
          setStatus(String(data?.status ?? "ongoing").toLowerCase());
          const gId =
            data?.genreId ??
            data?.genre_id ??
            (data?.genres && data.genres[0]?.id) ??
            "";
          setGenreId(gId ? String(gId) : "");
          setSelectedTags(
            (data?.tags ?? [])
              .map((t) => ({
                id: String(t?.id ?? t?.value ?? ""),
                name: t?.name ?? t?.label ?? "",
              }))
              .filter((x) => x.id && x.name)
          );
          setPreview(data?.cover_image ?? data?.cover_url ?? null);
        }
      } catch {
        toast.error("Failed to load form data.");
      } finally {
        setLoading(false);
        setPrefilling(false);
      }
    })();
  }, [mode, novelId]);

  // Daftar ID tag yang valid → untuk mencegah custom input
  const validTagIdSet = useMemo(
    () => new Set((tags ?? []).map((t) => String(t.id))),
    [tags]
  );

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (submitting) return;

    // 🔒 Validasi via toast
    if (!title.trim()) return toast.error("Title is required");
    if (!genreId) return toast.error("Genre is required");
    if (!synopsis.trim()) return toast.error("Synopsis is required");
    if (!selectedTags.length)
      return toast.error("At least one tag is required");
    if (mode === "create" && !(coverFile instanceof File))
      return toast.error("Cover image is required");

    setSubmitting(true);
    try {
      const slug = title
        .toLowerCase()
        .trim()
        .replace(/\s+/g, "-")
        .replace(/[^\w-]+/g, "")
        .replace(/--+/g, "-")
        .replace(/^-+|-+$/g, "");

      const payload = {
        title,
        slug,
        synopsis,
        status,
        genreId: Number(genreId),
        tagIds: (selectedTags ?? [])
          .map((t) => Number(t.id))
          .filter(Number.isFinite),
        coverFile: coverFile || null,
      };

      if (mode === "edit" && novelId) {
        await updateNovel(novelId, payload);
        toast.success("Novel updated successfully!");
      } else {
        await createNovel(payload);
        toast.success("Novel created successfully!");
      }

      navigate("/admin/novels");
    } catch (err) {
      toast.error(err?.message || "Failed to save novel.");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading || prefilling) {
    return (
      <div className="p-6 md:p-8">
        <div className="rounded-xl border border-gray-200 bg-white shadow-sm p-6 text-gray-500">
          Loading...
        </div>
      </div>
    );
  }

  return (
    <div className="p-4 md:p-6 pb-24">
      <div className="rounded-xl border border-gray-200 bg-white shadow-sm">
        <div className="px-5 py-4 border-b border-gray-200">
          <h2 className="text-lg md:text-xl font-semibold text-gray-800">
            {mode === "edit" ? "Edit Novel" : "Create New Novel"}
          </h2>
        </div>

        <form onSubmit={handleSubmit} className="p-5 md:p-6 space-y-6">
          <div className="flex flex-col md:flex-row gap-5">
            <div className="w-full md:w-1/3">
              <CoverUploader
                preview={preview}
                onFile={(file) => {
                  setCoverFile(file || null);
                  if (lastURL.current) {
                    URL.revokeObjectURL(lastURL.current);
                    lastURL.current = null;
                  }
                  if (file) {
                    const url = URL.createObjectURL(file);
                    lastURL.current = url;
                    setPreview(url);
                  }
                }}
              />
            </div>

            <div className="w-full md:flex-1 space-y-5">
              <div>
                <label className="text-sm font-medium text-gray-700 mb-1 block">
                  Title
                </label>
                <input
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="Novel Title"
                  maxLength={100}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500"
                />
                <p className="mt-1 text-xs text-gray-500">{title.length}/100</p>
              </div>

              <div>
                <label className="text-sm font-medium text-gray-700 mb-1 block">
                  Genre
                </label>
                <select
                  value={genreId}
                  onChange={(e) => setGenreId(e.target.value)}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 cursor-pointer"
                >
                  <option value="">Select Genre</option>
                  {(genres ?? []).map((g) => (
                    <option key={g.id} value={String(g.id)}>
                      {g.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          <div>
            <label className="text-sm font-medium text-gray-700 mb-1 block">
              Synopsis
            </label>
            <textarea
              rows={6}
              value={synopsis}
              onChange={(e) => setSynopsis(e.target.value)}
              placeholder="Write a short synopsis"
              maxLength={1000}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 resize min-h-[140px]"
            />
            <p className="mt-1 text-xs text-gray-500">{synopsis.length}/1000</p>
          </div>

          <TagSelector
            label="Tags"
            options={tags}
            value={selectedTags}
            onChange={(vals) => {
              const next = (vals ?? [])
                .map((v) => {
                  let id = v?.id ?? v?.value ?? null;
                  let name = v?.name ?? v?.label ?? "";
                  if (id == null) return null;
                  id = String(id);
                  if (!validTagIdSet.has(id)) return null;
                  if (!name) {
                    const found = tags.find((t) => String(t.id) === id);
                    name = found?.name ?? "";
                  }
                  return name ? { id, name } : null;
                })
                .filter(Boolean);

              setSelectedTags(next);
            }}
          />

          <div className="max-w-sm">
            <label className="text-sm font-medium text-gray-700 mb-1 block">
              Status
            </label>
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 cursor-pointer"
            >
              <option value="ongoing">Ongoing</option>
              <option value="completed">Completed</option>
              <option value="hiatus">Hiatus</option>
            </select>
          </div>

          <div className="flex flex-col sm:flex-row gap-3 pt-3 border-t border-gray-200">
            <button
              type="submit"
              disabled={submitting}
              className="inline-flex justify-center items-center bg-primary hover:bg-gray-700 text-white rounded-lg px-4 py-2.5 text-sm font-medium disabled:opacity-50 cursor-pointer"
            >
              {submitting ? "Saving..." : mode === "edit" ? "Update" : "Save"}
            </button>
            <button
              type="button"
              onClick={() => navigate("/admin/novels")}
              className="inline-flex justify-center items-center border border-gray-300 hover:bg-gray-50 rounded-lg px-4 py-2.5 text-sm font-medium text-gray-700 cursor-pointer"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
