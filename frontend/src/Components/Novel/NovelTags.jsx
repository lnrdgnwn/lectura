import { useState } from "react";
import { FaChevronDown, FaChevronUp } from "react-icons/fa";
import { NavLink } from "react-router-dom";

export default function NovelTags({ tags = [] }) {
  const [expanded, setExpanded] = useState(false);

  const visibleCount = 6;
  const isExpandable = tags.length > visibleCount;

  const labelOf = (t) => (typeof t === "string" ? t : t?.name ?? "");
  const slugOf = (t) => (typeof t === "string" ? t : t?.slug ?? "");
  const keyOf = (t, i) => (typeof t === "string" ? `${t}-${i}` : t?.id ?? i);

  const renderTag = (t, i) => {
    const label = labelOf(t);
    return (
      <NavLink
        key={keyOf(t, i)}
        to={`/tag/${slugOf(t)}`}
        className="px-3 py-1 bg-gray-100 text-gray-700 rounded-full text-sm hover:bg-gray-200 transition"
        title={label}
      >
        #{label}
      </NavLink>
    );
  };

  return (
    <section className="w-full flex flex-col items-center justify-center pb-3 px-4">
      <div className="w-full max-w-4xl">
        <h3 className="font-bold text-lg sm:text-xl mb-3 text-gray-800">
          Tags
        </h3>

        <div className="flex w-full flex-wrap items-center relative">
          {/* Mobile: expandable */}
          <div className="flex w-full sm:hidden flex-wrap items-center gap-2">
            {tags
              .slice(0, expanded ? tags.length : visibleCount)
              .map(renderTag)}
          </div>

          {/* Desktop: show all */}
          <div className="hidden sm:flex flex-wrap gap-2">
            {tags.map(renderTag)}
          </div>

          {isExpandable && (
            <button
              type="button"
              onClick={() => setExpanded(!expanded)}
              className="ml-auto p-1 text-primary hover:text-blue-600 transition sm:hidden"
              aria-expanded={expanded}
              aria-label={expanded ? "Tutup" : "Lihat semua tag"}
            >
              {expanded ? (
                <FaChevronUp className="text-sm" />
              ) : (
                <FaChevronDown className="text-sm" />
              )}
            </button>
          )}
        </div>
      </div>
    </section>
  );
}
