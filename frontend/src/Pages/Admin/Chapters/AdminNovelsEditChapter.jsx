// pages/.../UserNovelsEditChapter.jsx
import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { FaArrowLeft } from "react-icons/fa";
import { toast } from "react-hot-toast";
import ChapterEditor from "../../../Components/Admin/Novels/ChapterEditor";
import {
  getChapterById,
  updateChapter,
  setChapterPublish,
} from "../../../services/chapterService";
import { useAuth } from "../../../context/AuthContext";
import { sanitizeForBackend } from "../../../utils/sanitizeChapterHtml";

export default function AdminNovelsEditChapter() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const { id, chapterId } = useParams();

  const [chapterNo, setChapterNo] = useState(1);
  const [title, setTitle] = useState("");
  const [contentHtml, setContentHtml] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [isPublished, setIsPublished] = useState(false);

  useEffect(() => {
    (async () => {
      try {
        const ch = await getChapterById(chapterId);

        setChapterNo(Number(ch?.order_no ?? 1));
        setTitle(ch?.title ?? "");
        setContentHtml(ch?.content ?? ch?.content_html ?? "");
        setIsPublished(!!ch?.published_at);
      } catch (e) {
        toast.error(e?.message || "Failed to load chapter");
        navigate("/admin/novels", { replace: true });
      } finally {
        setLoading(false);
      }
    })();
  }, [chapterId, user?.id, navigate]);

  const saveChanges = async () => {
    if (!title.trim()) {
      toast.error("Title is required");
      return false;
    }
    if (!contentHtml.trim()) {
      toast.error("Content is empty");
      return false;
    }
    const cleanHtml = sanitizeForBackend(contentHtml);
    await updateChapter(chapterId, {
      order_no: Number(chapterNo),
      title: title.trim(),
      content: cleanHtml,
    });
    return true;
  };

  const saveAsDraft = async () => {
    if (saving) return;
    setSaving(true);
    try {
      const ok = await saveChanges();
      if (!ok) return;
      await setChapterPublish(chapterId, false);
      setIsPublished(false);
      toast.success("Saved as draft");
    } catch (e) {
      toast.error(e?.message || "Failed to save draft");
    } finally {
      setSaving(false);
    }
  };

  const publishOrUpdate = async () => {
    if (saving) return;
    setSaving(true);
    try {
      const ok = await saveChanges();
      if (!ok) return;
      if (!isPublished) {
        await setChapterPublish(chapterId, true);
        setIsPublished(true);
        toast.success("Chapter published");
      } else {
        toast.success("Chapter updated");
      }
      navigate(`/admin/novels/${id}`);
    } catch (e) {
      toast.error(
        e?.message || (isPublished ? "Failed to update" : "Failed to publish")
      );
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="p-0 md:p-6 pb-24">
        <div className="bg-white border border-gray-200 rounded-xl shadow-md p-6 text-gray-500">
          Loading…
        </div>
      </div>
    );
  }

  return (
    <div className="p-0 md:p-6 pb-24">
      <div className="bg-white border border-gray-200 rounded-xl shadow-md overflow-hidden">
        {/* Header */}
        <div className="flex flex-wrap items-center justify-between gap-3 px-5 py-4 border-b border-gray-200">
          <div className="flex items-center gap-3">
            <button
              onClick={() => navigate(-1)}
              className="text-gray-600 hover:text-indigo-600 transition cursor-pointer"
              aria-label="Back"
            >
              <FaArrowLeft size={18} />
            </button>
            <h1 className="text-lg md:text-xl font-semibold text-gray-800">
              {isPublished ? "Edit Chapter" : "Edit Draft"}
            </h1>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={saveAsDraft}
              disabled={saving}
              className="border border-gray-300 hover:bg-gray-50 text-gray-700 rounded-lg px-3 py-2 text-sm font-medium cursor-pointer disabled:opacity-60"
            >
              Save as Draft
            </button>
            <button
              onClick={publishOrUpdate}
              disabled={saving}
              className="bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg px-3 md:px-4 py-2 text-sm font-medium shadow-sm cursor-pointer disabled:opacity-60"
            >
              {isPublished ? "Update" : "Publish"}
            </button>
          </div>
        </div>

        {/* Controls: Chapter # */}
        <div className="flex flex-wrap items-center gap-3 px-5 py-3 border-b border-gray-100 text-sm">
          <label className="flex items-center gap-2">
            <span className="text-gray-600">Chapter :</span>
            <input
              type="number"
              min={1}
              value={chapterNo}
              onChange={(e) =>
                setChapterNo(Math.max(1, Number(e.target.value)))
              }
              className="w-28 border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </label>
        </div>

        {/* Editor */}
        <div className="px-5 md:px-10 py-6">
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Title Here"
            className="w-full placeholder-gray-400 text-xl md:text-2xl font-semibold text-gray-800 bg-transparent outline-none mb-4 px-0"
          />
          <ChapterEditor
            value={contentHtml}
            onChange={setContentHtml}
            minHeight={420}
          />
        </div>
      </div>
    </div>
  );
}
