// components/NovelPreview.jsx
import { useMemo } from "react";
import { useNavigate } from "react-router-dom";
import DOMPurify from "dompurify";

export default function NovelPreview({ novel }) {
  const navigate = useNavigate();

  // Ambil chapters publish saja + urut rapi
  const publishedChapters = useMemo(() => {
    const arr = Array.isArray(novel?.chapters) ? novel.chapters : [];
    const onlyPublished = arr.filter((c) => !!c?.published_at);
    return onlyPublished.sort((a, b) => {
      const ao = Number.isFinite(+a.order_no) ? +a.order_no : Infinity;
      const bo = Number.isFinite(+b.order_no) ? +b.order_no : Infinity;
      if (ao !== bo) return ao - bo;
      return new Date(a?.created_at || 0) - new Date(b?.created_at || 0);
    });
  }, [novel?.chapters]);

  if (!publishedChapters.length) {
    return (
      <section className="w-full flex flex-col items-center justify-center pb-10 px-5">
        <div className="w-full max-w-4xl border-t border-gray-200 text-gray-600 pt-6 text-center">
          <p className="text-base sm:text-lg font-medium">
            No published chapters yet
          </p>
        </div>
      </section>
    );
  }

  const first = publishedChapters[0];
  const next = publishedChapters[1];
  const rawHtml = first.content_html ?? first.content ?? "";
  const safeHtml = DOMPurify.sanitize(rawHtml, {
    USE_PROFILES: { html: true },
  });

  const goToNext = () => {
    if (next) navigate(`/novel/${novel.id}/chapter/${next.id}`);
  };

  return (
    <section className="w-full flex flex-col items-center justify-center pb-10 px-5">
      <div className="w-full max-w-4xl border-t border-gray-200 text-gray-800 pt-4 relative">
        <h3 className="font-bold text-lg sm:text-xl mb-3 text-gray-800 break-words">
          Chapter 1: {first.title ?? "Preview"}
        </h3>

        {/* Render konten aman */}
        <div
          className="prose max-w-none prose-p:mb-4 prose-img:rounded-md"
          dangerouslySetInnerHTML={{ __html: safeHtml }}
        />

        {/* Tombol selalu tampil */}
        <div className="flex items-center justify-center pt-4 gap-3">
          <button
            onClick={() => navigate("/")}
            className="bg-white border border-primary text-primary px-8 py-2 rounded-lg hover:bg-primary hover:text-white transition cursor-pointer"
          >
            Back
          </button>

          <button
            onClick={goToNext}
            disabled={!next}
            className={`px-8 py-2 rounded-lg transition ${
              next
                ? "bg-primary text-white hover:bg-gray-800 cursor-pointer"
                : "bg-gray-300 text-gray-500 cursor-not-allowed"
            }`}
          >
            Next Chapter
          </button>
        </div>
      </div>
    </section>
  );
}
