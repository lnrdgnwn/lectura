import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { FaArrowLeft } from "react-icons/fa";
import { toast } from "react-hot-toast";
import ChapterEditor from "../../../../Components/Workspace/ChapterEditor";
import { createChapter } from "../../../../services/chapterService";
import { getNovelById } from "../../../../services/NovelService";
import { useAuth } from "../../../../context/AuthContext";
import { sanitizeForBackend } from "../../../../utils/sanitizeChapterHtml";

export default function UserNovelsWriteChapter() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const { id } = useParams();

  const [chapterNo, setChapterNo] = useState(1);
  const [title, setTitle] = useState("");
  const [contentHtml, setContentHtml] = useState("");
  const [loading, setLoading] = useState(false);

  // ✅ cek kepemilikan & tentukan nomor urutan otomatis
  useEffect(() => {
    (async () => {
      try {
        const n = await getNovelById(id);
        const ownerId = n?.author?.id ?? n?.author_id;

        if (!user?.id || Number(ownerId) !== Number(user.id)) {
          toast.error("You can't write on this novel!");
          navigate("/workspace/novels", { replace: true });
          return;
        }

        // ✅ hitung chapter terakhir
        const chapters = Array.isArray(n?.chapters) ? n.chapters : [];
        const maxOrder = chapters.reduce((max, c) => {
          const num = Number(c?.order_no || 0);
          return num > max ? num : max;
        }, 0);

        setChapterNo(maxOrder + 1); // chapter berikutnya
      } catch (err) {
        console.error(err);
        toast.error("Failed to fetch novel");
        navigate("/workspace/novels", { replace: true });
      }
    })();
  }, [id, user?.id, navigate]);

  const create = async (publish) => {
    if (!title.trim()) return toast.error("Title is required");
    if (!contentHtml.trim()) return toast.error("Content is empty");

    const cleanHtml = sanitizeForBackend(contentHtml);
    setLoading(true);

    try {
      await createChapter({
        novel_id: Number(id),
        order_no: Number(chapterNo), // simpan nomor urut
        title: title.trim(),
        content: cleanHtml,
        publish: !!publish,
      });

      toast.success(publish ? "Chapter published!" : "Saved to draft!");
      navigate(`/workspace/novels/${id}`);
    } catch (e) {
      console.error(e);
      toast.error(e?.message || "Failed to create chapter");
    } finally {
      setLoading(false);
    }
  };

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
              Write Chapter
            </h1>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => create(false)}
              disabled={loading}
              className="border border-gray-300 hover:bg-gray-50 text-gray-700 rounded-lg px-3 py-2 text-sm font-medium cursor-pointer disabled:opacity-60"
            >
              Save to Draft
            </button>
            <button
              onClick={() => create(true)}
              disabled={loading}
              className="bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg px-3 md:px-4 py-2 text-sm font-medium shadow-sm cursor-pointer disabled:opacity-60"
            >
              Publish
            </button>
          </div>
        </div>

        {/* Top controls: Chapter # */}
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
              readOnly
              className="w-28 border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </label>
          <span className="text-gray-400">
            * Nomor otomatis diset ke urutan berikutnya.
          </span>
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
