import { useEffect, useMemo, useRef, useState } from "react";

export default function TagSelector({
  label,
  options = [],
  value = [],
  onChange,
}) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const ref = useRef();
  const inputRef = useRef(null);

  useEffect(() => {
    const handler = (e) => {
      if (ref.current && !ref.current.contains(e.target)) {
        setOpen(false);
        inputRef.current?.blur();
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  const filtered = useMemo(() => {
    const q = query.toLowerCase();
    return (options || []).filter((t) =>
      (t.name || "").toLowerCase().includes(q)
    );
  }, [options, query]);

  const toggle = (tag) => {
    const tagId = String(tag.id);
    const exists = value.some((v) => String(v.id) === tagId);
    if (exists) {
      onChange(value.filter((v) => String(v.id) !== tagId));
    } else {
      onChange([...value, { ...tag, id: tagId }]);
    }
    setQuery(""); // reset input biar gak nambah duplikat
  };

  return (
    <div ref={ref} className="relative">
      <label className="text-sm font-medium text-gray-700 mb-1 block">
        {label}
      </label>
      <div
        className="border border-gray-300 rounded-lg px-2 py-2 flex flex-wrap gap-1 items-center focus-within:ring-1 md:focus-within:ring-2 focus-within:ring-indigo-500"
        onClick={() => {
          setOpen(true);
          inputRef.current?.focus();
        }}
      >
        {value.map((t) => (
          <span
            key={t.id}
            className="flex items-center gap-1 px-2 py-1 rounded-full bg-indigo-50 border border-indigo-200 text-indigo-700 text-xs"
          >
            #{t.name}
            <button
              type="button"
              className="hover:text-indigo-900"
              onClick={(e) => {
                e.stopPropagation();
                toggle(t);
              }}
            >
              ×
            </button>
          </span>
        ))}

        <input
          ref={inputRef}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search tag..."
          className="flex-1 min-w-[110px] outline-none text-sm"
          onFocus={() => setOpen(true)}
          onKeyDown={(e) => {
            if (e.key === "Escape") {
              setOpen(false);
              inputRef.current?.blur();
            }
          }}
        />
      </div>

      {open && (
        <div className="absolute left-0 right-0 bg-white border border-gray-200 rounded-lg shadow-lg max-h-56 mt-1 overflow-y-auto z-30">
          {filtered.length === 0 ? (
            <div className="px-3 py-2 text-sm text-gray-500">No tags found</div>
          ) : (
            filtered.map((t) => {
              const active = value.some((v) => String(v.id) === String(t.id));
              return (
                <button
                  key={t.id}
                  type="button"
                  onClick={() => toggle(t)}
                  className={`w-full text-left px-3 py-2 text-sm hover:bg-gray-50 flex justify-between ${
                    active ? "bg-indigo-50 font-bold" : ""
                  }`}
                >
                  <span>{t.name}</span>
                  {active && <span>✓</span>}
                </button>
              );
            })
          )}
        </div>
      )}
    </div>
  );
}
