import { useEffect, useMemo, useRef, useState } from "react";

function timeAgo(dateLike) {
  if (!dateLike) return "";
  const d = new Date(dateLike);
  if (isNaN(d.getTime())) return "";
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

export default function ChapterMenu({
  show,
  setShow,
  novelTitle,
  novelCover,
  novelAuthor,
  chapters = [],
  currentChapterId,
  onSelectChapter, // (id) => void
}) {
  const [q, setQ] = useState("");
  const listRef = useRef(null);

  // sort: order_no asc (yang kosong dibelakang), lalu created_at, lalu title
  const filtered = useMemo(() => {
    const term = q.trim().toLowerCase();
    const base = chapters || [];
    const rows = term
      ? base.filter((c) => (c.title || "").toLowerCase().includes(term))
      : base;

    return [...rows].sort((a, b) => {
      const ao = Number.isFinite(+a.order_no) ? +a.order_no : Infinity;
      const bo = Number.isFinite(+b.order_no) ? +b.order_no : Infinity;
      if (ao !== bo) return ao - bo;

      const da = Date.parse(a.updated_at || a.created_at || "");
      const db = Date.parse(b.updated_at || b.created_at || "");
      if (!isNaN(da) && !isNaN(db) && da !== db) return da - db;

      return String(a.title || "").localeCompare(String(b.title || ""));
    });
  }, [q, chapters]);

  useEffect(() => {
    if (!show) return;
    const el = listRef.current?.querySelector('[data-active="true"]');
    if (el)
      listRef.current.scrollTo({ top: el.offsetTop - 120, behavior: "smooth" });
  }, [show, filtered, currentChapterId]);

  return (
    <div
      className={`fixed inset-0 bg-black/40 z-[70] ${
        show ? "opacity-100 visible" : "opacity-0 invisible"
      } transition-all duration-300`}
      onClick={() => setShow(false)}
    >
      <aside
        className="absolute left-0 top-0 h-full w-[80%] sm:w-[360px] bg-primary text-white shadow-2xl flex flex-col"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-label="Chapter list"
        aria-modal="true"
      >
        {/* Header */}
        <div className="flex gap-3 p-4 border-b border-white/20">
          {novelCover && (
            <img
              src={novelCover}
              alt={novelTitle}
              className="w-16 h-24 rounded object-cover"
            />
          )}
          <div>
            <h2 className="font-semibold text-base">{novelTitle}</h2>
            {novelAuthor && (
              <p className="text-sm mt-2 text-white/80">{novelAuthor}</p>
            )}
          </div>
        </div>

        {/* Search */}
        <div className="p-4 border-b border-white/20">
          <input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Search chapter…"
            className="w-full rounded-md border border-white/30 bg-white/10 text-white placeholder-white/70 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-white/60"
          />
        </div>

        {/* List */}
        <div ref={listRef} className="flex-1 overflow-y-auto">
          <ul className="px-2 py-3 flex flex-col gap-2">
            {filtered.map((ch, i) => {
              const isActive = String(ch.id) === String(currentChapterId);
              const order = Number.isFinite(+ch.order_no)
                ? +ch.order_no
                : i + 1;
              const when = timeAgo(ch.updated_at || ch.created_at);
              const title = ch.title || `Chapter ${order}`;

              return (
                <li key={ch.id}>
                  <button
                    type="button"
                    data-active={isActive ? "true" : "false"}
                    onClick={() => onSelectChapter(ch.id)}
                    className={`w-full rounded-md px-3 py-3 transition cursor-pointer ${
                      isActive ? "bg-white/20" : "hover:bg-white/10"
                    }`}
                  >
                    {/* ROW: left number, right title+time */}
                    <div className="flex items-start gap-3">
                      {/* left: number */}
                      <div className="w-6 flex-shrink-0 flex items-center justify-center">
                        <span className="text-sm font-semibold text-center">
                          {order}
                        </span>
                      </div>
                      {/* right: title and time (stacked) */}
                      <div className="flex-1 text-left">
                        <div className="font-medium">
                          {`Chapter ${order} - ${title}`}
                        </div>
                        {when && (
                          <div className="text-xs mt-1 text-white/80">
                            {when}
                          </div>
                        )}
                      </div>
                    </div>
                  </button>
                </li>
              );
            })}
            {!filtered.length && (
              <li className="px-3 py-3 text-sm text-white/80">
                No chapters found
              </li>
            )}
          </ul>
        </div>

        {/* Footer */}
        <div className="p-3 border-t border-white/20">
          <button
            onClick={() => setShow(false)}
            className="w-full rounded-md bg-white/20 hover:bg-white/30 text-white py-2 text-sm transition"
          >
            Close
          </button>
        </div>
      </aside>
    </div>
  );
}
