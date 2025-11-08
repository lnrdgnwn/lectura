// Components/Reader/ChapterContent.jsx
import { sanitizeForView } from "../../utils/sanitizeChapterHtml";

import DOMPurify from "dompurify";

export default function ChapterContent({
  chapter,
  textSize,
  onPrev,
  onNext,
  hasPrev,
  hasNext,
  darkMode = false,
}) {
  if (!chapter) return null;

  const rawHtml = chapter.content_html ?? chapter.content ?? "";
  const safeHtml = sanitizeForView(rawHtml);

  // class untuk konten: tambah prose-invert kalau darkMode
  const contentClass = `prose max-w-none prose-p:mb-4 prose-img:rounded-md prose-headings:scroll-mt-20 ${
    darkMode ? "prose-invert" : ""
  }`;

  return (
    <article>
      <h1
        className={`text-2xl font-bold mb-4 break-words ${
          darkMode ? "text-white" : "text-gray-900"
        }`}
      >
        Chapter {chapter.order_no ?? "-"} : {chapter.title ?? "-"}
      </h1>

      <div
        className={contentClass}
        style={{ fontSize: `${textSize}px`, lineHeight: 1.7 }}
        dangerouslySetInnerHTML={{ __html: safeHtml }}
      />

      <div className="flex items-center justify-center pt-6 gap-3">
        <button
          onClick={onPrev}
          disabled={!hasPrev}
          className={`px-8 py-2 rounded-lg transition ${
            darkMode
              ? "border border-gray-500 text-gray-100 hover:bg-gray-800"
              : "bg-white border border-primary text-primary hover:bg-primary hover:text-white"
          } ${!hasPrev ? "opacity-50 cursor-not-allowed" : "cursor-pointer"}`}
        >
          Back
        </button>

        <button
          onClick={onNext}
          disabled={!hasNext}
          className={`px-8 py-2 rounded-lg transition ${
            darkMode
              ? "bg-primary text-white hover:opacity-90"
              : "bg-primary text-white hover:bg-gray-800"
          } ${!hasNext ? "opacity-50 cursor-not-allowed" : "cursor-pointer"}`}
        >
          Next Chapter
        </button>
      </div>
    </article>
  );
}
