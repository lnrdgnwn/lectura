export default function FilterPagination({
  current,
  total,
  onChange,
  hasData = true,
}) {
  if (!hasData) return null;
  const getPages = () => {
    const pages = [];
    for (let i = 1; i <= total; i++) {
      if (i === 1 || i === total || Math.abs(current - i) <= 1) {
        pages.push(i);
      } else if (
        (i === 2 && current > 3) ||
        (i === total - 1 && current < total - 2)
      ) {
        pages.push("...");
      }
    }
    return pages;
  };

  return (
    <div className="flex justify-center items-center gap-2 mt-6 select-none">
      <button
        onClick={() => onChange(current - 1)}
        disabled={current === 1}
        className={`px-3 py-1 border rounded-md text-sm transition ${
          current === 1
            ? "opacity-50 cursor-not-allowed bg-gray-100"
            : "hover:bg-gray-100"
        }`}
      >
        Prev
      </button>

      {getPages().map((p, i) =>
        p === "..." ? (
          <span key={i} className="px-2 text-gray-400">
            ...
          </span>
        ) : (
          <button
            key={p}
            onClick={() => onChange(p)}
            className={`px-3 py-1 border rounded-md  cursor-pointer text-sm transition ${
              p === current
                ? "bg-primary text-white border-primary"
                : "hover:bg-gray-100"
            }`}
          >
            {p}
          </button>
        )
      )}

      <button
        onClick={() => onChange(current + 1)}
        disabled={current === total}
        className={`px-3 py-1 border rounded-md cursor-pointer text-sm transition ${
          current === total
            ? "opacity-50 cursor-not-allowed bg-gray-100"
            : "hover:bg-gray-100"
        }`}
      >
        Next
      </button>
    </div>
  );
}
