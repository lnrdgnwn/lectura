import { NavLink } from "react-router-dom";
import { FaHome } from "react-icons/fa";

export default function ChapterBreadcrumb({
  novelTitle,
  novelId,
  chapterTitle,
  chapterNumber,
  darkMode = false,
}) {
  if (!novelTitle) return null;

  const base = darkMode ? "text-white" : "text-gray-700";
  const title = darkMode ? "text-white" : "text-black";
  const sep = darkMode ? "text-white" : "text-black";
  const hover = darkMode ? "hover:opacity-90" : "hover:text-primary";

  return (
    <nav
      className={`flex items-center text-sm md:text-md ${base} pb-3 `}
      aria-label="Breadcrumb"
    >
      <ol className="flex items-center flex-wrap gap-1">
        <li className="flex items-center">
          <NavLink to="/" className={`flex items-center ${hover}`} title="Home">
            <FaHome className="inline-block mr-1" />
          </NavLink>
        </li>

        <span className={`mx-1 ${sep}`}>/</span>

        <li className="truncate max-w-[150px] sm:max-w-[200px] md:max-w-[250px] lg:max-w-full">
          <NavLink
            to={`/novel/${novelId}`}
            className={`flex items-center ${hover}`}
            title={novelTitle}
          >
            <span className={`${title} truncate`}>{novelTitle}</span>
          </NavLink>
        </li>

        {chapterTitle && (
          <>
            <span className={`mx-1 ${sep}`}>/</span>
            <li className="truncate ">
              <span className={`${title} font-bold `} title={chapterTitle}>
                <span>Chapter {chapterNumber ?? ""} : </span>
                {chapterTitle}
              </span>
            </li>
          </>
        )}
      </ol>
    </nav>
  );
}
