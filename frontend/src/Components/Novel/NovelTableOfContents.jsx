import { Link } from "react-router-dom";

// helper: format "x months ago"
function timeAgo(dateLike) {
  if (!dateLike) return "";
  const d = new Date(dateLike);
  if (Number.isNaN(d.getTime())) return "";
  const diff = Date.now() - d.getTime();

  const sec = Math.floor(diff / 1000);
  if (sec < 60) return "just now";
  const min = Math.floor(sec / 60);
  if (min < 60) return `${min} minute${min > 1 ? "s" : ""} ago`;
  const hour = Math.floor(min / 60);
  if (hour < 24) return `${hour} hour${hour > 1 ? "s" : ""} ago`;
  const day = Math.floor(hour / 24);
  if (day < 30) return `${day} day${day > 1 ? "s" : ""} ago`;
  const month = Math.floor(day / 30);
  if (month < 12) return `${month} month${month > 1 ? "s" : ""} ago`;
  const year = Math.floor(month / 12);
  return `${year} year${year > 1 ? "s" : ""} ago`;
}

function isValidDateLike(v) {
  if (!v) return false;
  const d = new Date(v);
  return !Number.isNaN(d.getTime());
}

export default function NovelTableOfContents({ novel }) {
  const chapters = Array.isArray(novel?.chapters) ? novel.chapters : [];

  // 1) Ambil yang publish saja:
  // - published_at berisi tanggal valid, ATAU
  // - published === true (boolean). (Kalau hanya boolean tanpa tanggal, tetap ditampilkan, tapi tanpa "time ago")
  const publishedChapters = chapters.filter((c) => {
    if (typeof c?.published === "boolean") return c.published === true;
    return isValidDateLike(c?.published_at);
  });

  // 2) Urutkan: order_no naik (kosong di belakang), lalu title
  const sorted = [...publishedChapters].sort((a, b) => {
    const ao = Number.isFinite(+a.order_no) ? +a.order_no : Infinity;
    const bo = Number.isFinite(+b.order_no) ? +b.order_no : Infinity;
    if (ao !== bo) return ao - bo;
    return String(a.title || "").localeCompare(String(b.title || ""));
  });

  if (!sorted.length) {
    return <p className="text-gray-500">No chapters available yet.</p>;
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
      {sorted.map((ch, i) => {
        const order = Number.isFinite(+ch.order_no) ? +ch.order_no : i + 1;
        const href = `/novel/${novel.id}/chapter/${ch.id}`;
        const when = isValidDateLike(ch?.published_at)
          ? timeAgo(ch.published_at)
          : "";

        return (
          <Link
            key={ch.id}
            to={href}
            className="flex items-start gap-3 rounded-lg border border-gray-200 bg-white hover:bg-gray-50 transition p-3"
          >
            <span className="w-6 text-right font-semibold text-gray-700">
              {order}
            </span>
            <div className="flex-1">
              <div className="text-sm sm:text-base font-medium text-gray-900">
                {`Chapter ${order} - ${ch.title || `Chapter ${order}`}`}
              </div>
              {when && (
                <div className="text-xs text-gray-500 mt-0.5">{when}</div>
              )}
            </div>
          </Link>
        );
      })}
    </div>
  );
}
