import { useState, useRef, useEffect } from "react";
import { NavLink, useNavigate, useLocation } from "react-router-dom";
import {
  FaChevronDown,
  FaSearch,
  FaBook,
  FaCompass,
  FaPenNib,
  FaBars,
  FaTimes,
} from "react-icons/fa";
import { useAuth } from "../context/AuthContext";
import { getGenres } from "../services/genreService";
import { getNovels } from "../services/novelService";

export default function Navbar() {
  const { user, isAuthenticated } = useAuth();

  const [menuOpen, setMenuOpen] = useState(false);
  const [browseOpen, setBrowseOpen] = useState(false);
  const [mobileBrowseOpen, setMobileBrowseOpen] = useState(true);
  const [searchOpen, setSearchOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [genres, setGenres] = useState([]);
  const [novels, setNovels] = useState([]);

  const navigate = useNavigate();
  const location = useLocation();

  const searchRef = useRef(null);
  const profileRef = useRef(null);
  const browseRef = useRef(null);
  const hoverTimer = useRef(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [genreData, novelData] = await Promise.all([
          getGenres(),
          getNovels(),
        ]);
        setGenres(genreData);
        setNovels(novelData);
      } catch (err) {
        console.error("Failed to load navbar data:", err);
      }
    };
    fetchData();
  }, []);

  // Filter hasil pencarian
  const filteredResults = query.trim()
    ? novels.filter((n) => {
        const q = query.toLowerCase();

        const title = (n?.title || "").toLowerCase();
        const author = n?.author?.username || n?.author?.name || n?.author;

        const genres = (n?.genres || [])
          .map((g) => (g?.slug || g?.name || "").toLowerCase())
          .join(" ");

        const tags = (n?.tags || [])
          .map((t) => (t?.name || t?.slug || "").toLowerCase())
          .join(" ");

        const haystack = [title, author, genres, tags].join(" ");
        return haystack.includes(q);
      })
    : [];

  // Tutup dropdown saat klik di luar
  useEffect(() => {
    const handleClickOutside = (e) => {
      if (searchRef.current && !searchRef.current.contains(e.target)) {
        setSearchOpen(false);
      }
      if (browseRef.current && !browseRef.current.contains(e.target)) {
        setBrowseOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Hover-intent (desktop): delay kecil supaya nggak kedip
  const handleBrowseEnter = () => {
    clearTimeout(hoverTimer.current);
    hoverTimer.current = setTimeout(() => setBrowseOpen(true), 80);
  };
  const handleBrowseLeave = () => {
    clearTimeout(hoverTimer.current);
    hoverTimer.current = setTimeout(() => setBrowseOpen(false), 120);
  };

  // Navigasi ke library, login jika belum login
  const handleGoLibrary = () => {
    if (isAuthenticated) navigate("/library");
    else navigate("/login", { replace: true, state: { from: location } });
  };

  return (
    <nav className="sticky top-0 left-0 z-50 w-full bg-white border-b border-gray-200 font-urbanist shadow-sm">
      <div className="max-w-7xl mx-auto px-6 py-3 flex items-center justify-between relative">
        {/* MOBILE: Hamburger */}
        <button
          className="block lg:hidden text-gray-700 text-xl mr-3"
          onClick={() => setMenuOpen(true)}
          aria-label="Open menu"
        >
          <FaBars />
        </button>

        {/* LEFT (Desktop) */}
        <div className="hidden lg:flex items-center gap-4">
          <NavLink to="/" className="flex items-center cursor-pointer">
            <img src="/img/lectura.png" alt="Logo" className="h-12" />
          </NavLink>

          <ul className="flex items-center gap-4 text-base font-medium">
            {/* BROWSE: Mega menu */}
            <li
              ref={browseRef}
              className="relative"
              onMouseEnter={handleBrowseEnter}
              onMouseLeave={handleBrowseLeave}
            >
              <button
                onClick={() => setBrowseOpen((v) => !v)}
                className={[
                  "group flex items-center gap-2 px-3 py-1.5 rounded-xl transition-colors duration-200 cursor-pointer",
                  browseOpen
                    ? "bg-primary text-white"
                    : "text-gray-700 hover:bg-primary hover:text-white",
                ].join(" ")}
                aria-haspopup="menu"
                aria-expanded={browseOpen}
              >
                <FaCompass size={14} />
                <span>Browse</span>
                <FaChevronDown
                  size={12}
                  className={`mt-[2px] transition-transform ${
                    browseOpen ? "rotate-180" : ""
                  }`}
                />
              </button>

              <div
                className={[
                  "absolute left-0 top-full w-[640px] bg-white border border-gray-200 rounded-xl shadow-xl overflow-hidden",
                  "transition duration-150 ease-out origin-top",
                  browseOpen
                    ? "opacity-100 scale-100 pointer-events-auto"
                    : "opacity-0 scale-95 pointer-events-none",
                ].join(" ")}
              >
                <section className="p-4 max-h-80 overflow-y-auto">
                  <div className="mb-2 px-1 font-semibold text-gray-600 text-xs uppercase tracking-wide">
                    Genres
                  </div>
                  <div className="grid grid-cols-3 gap-x-6 gap-y-1.5">
                    {genres.length > 0 ? (
                      genres.map((g) => (
                        <NavLink
                          key={g.id}
                          to={`/filter/${g.slug}`}
                          className="rounded-md px-2 py-1 text-gray-800 hover:bg-primary/10 hover:text-primary transition-colors capitalize"
                          onClick={() => setBrowseOpen(false)}
                        >
                          {g.name}
                        </NavLink>
                      ))
                    ) : (
                      <span className="px-1 text-gray-400 text-sm col-span-3">
                        Genre belum tersedia
                      </span>
                    )}
                  </div>
                </section>
              </div>
            </li>

            <li>
              <NavLink
                to="/workspace"
                className={({ isActive }) =>
                  [
                    "flex items-center gap-2 px-3 py-1.5 rounded-xl transition-colors duration-200",
                    isActive
                      ? "bg-primary text-white"
                      : "text-gray-700 hover:bg-primary hover:text-white",
                  ].join(" ")
                }
              >
                <FaPenNib size={14} />
                <span>Create</span>
              </NavLink>
            </li>
          </ul>
        </div>

        {/* CENTER: Search */}
        <div
          className="flex-1 flex justify-center px-4 relative"
          ref={searchRef}
        >
          <div className="flex items-center bg-gray-100 rounded-full px-3 py-1.5 w-full">
            <FaSearch size={14} className="text-gray-500" />
            <input
              type="text"
              placeholder="Search novel..."
              value={query}
              onChange={(e) => {
                setQuery(e.target.value);
                setSearchOpen(true);
              }}
              onFocus={() => setSearchOpen(true)}
              className="bg-transparent outline-none text-sm ml-2 w-full placeholder-gray-500"
              aria-label="Search novels"
            />
          </div>
          {searchOpen && (
            <div className="absolute top-12 left-0 w-full bg-white border border-gray-200 rounded-lg shadow-lg z-40 max-h-72 overflow-y-auto">
              {filteredResults.length > 0 ? (
                filteredResults.map((novel) => {
                  const authorText = novel?.author?.username;

                  return (
                    <NavLink
                      key={novel.id}
                      to={`/novel/${novel.id}`}
                      className="flex items-center gap-3 px-4 py-2 hover:bg-gray-100 transition-colors"
                      onClick={() => setSearchOpen(false)}
                    >
                      {/* cover kecil (opsional) */}
                      {novel?.cover_image ? (
                        <img
                          src={novel.cover_image}
                          alt={novel.title}
                          className="w-8 h-10 rounded object-cover flex-shrink-0"
                        />
                      ) : (
                        <div className="w-8 h-10 rounded bg-gray-100 text-gray-500 text-xs flex items-center justify-center flex-shrink-0">
                          {(novel.title || "N").slice(0, 1)}
                        </div>
                      )}

                      <div className="min-w-0">
                        {" "}
                        {/* penting untuk ellipsis */}
                        <div className="font-medium text-gray-800 truncate">
                          {novel.title}
                        </div>
                        {authorText && (
                          <div className="text-sm text-gray-500 truncate capitalize">
                            by {authorText}
                          </div>
                        )}
                      </div>
                    </NavLink>
                  );
                })
              ) : query ? (
                <div className="px-4 py-3 text-gray-500 text-sm">
                  No results found for “{query}”
                </div>
              ) : (
                <div className="px-4 py-3 text-gray-500 text-sm">
                  Start typing to search novels...
                </div>
              )}
            </div>
          )}
        </div>

        {/* RIGHT */}
        <div className="flex items-center gap-4">
          <div className="hidden lg:flex items-center gap-4">
            <button
              onClick={handleGoLibrary}
              className="flex cursor-pointer font-medium items-center gap-2 px-3 py-1.5 rounded-xl transition-colors duration-200 text-gray-700 hover:bg-primary hover:text-white"
            >
              <FaBook size={15} />
              <span>Library</span>
            </button>

            <div className="flex" ref={profileRef}>
              {isAuthenticated ? (
                <button
                  onClick={() => navigate(`/profile`)}
                  className="focus:outline-none rounded-full"
                  aria-label="Open profile"
                >
                  <img
                    src={user?.profile_picture || "/img/default-avatar.png"}
                    alt="User Avatar"
                    className="w-10 h-10 rounded-full object-cover border border-gray-300 cursor-pointer hover:shadow-lg"
                  />
                </button>
              ) : (
                <NavLink
                  to="/login"
                  className="bg-primary hover:bg-primary/90 text-white text-base px-5 py-1.5 rounded-md font-semibold transition-colors"
                >
                  LOGIN
                </NavLink>
              )}
            </div>
          </div>

          {/* Mobile Right icons */}
          <div className="block lg:hidden">
            <div className="flex items-center gap-3">
              <button
                onClick={handleGoLibrary}
                className="p-2 rounded-md text-gray-700 hover:bg-gray-100"
                aria-label="Library"
                title="Library"
              >
                <FaBook className="text-xl" />
              </button>

              {isAuthenticated ? (
                <button
                  onClick={() => navigate("/profile")}
                  className="rounded-full"
                  aria-label="Open settings"
                >
                  <img
                    src={user?.profile_picture || "/img/default-avatar.png"}
                    alt="User Avatar"
                    className="w-10 h-10 rounded-full object-cover border border-gray-300"
                  />
                </button>
              ) : (
                <NavLink
                  to="/login"
                  className="bg-primary hover:bg-primary/90 text-white text-base px-5 py-1.5 rounded-md font-semibold transition-colors"
                >
                  LOGIN
                </NavLink>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Mobile Sidebar */}
      <div
        className={`fixed top-0 left-0 h-screen w-80 bg-white text-gray-800 shadow-2xl z-[60] transform transition-transform duration-300 ${
          menuOpen ? "translate-x-0" : "-translate-x-full"
        }`}
        aria-hidden={!menuOpen}
      >
        <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200">
          <img src="/img/lectura.png" className="h-10" alt="Lectura" />
          <FaTimes
            className="text-gray-600 cursor-pointer text-xl"
            onClick={() => setMenuOpen(false)}
            aria-label="Close menu"
          />
        </div>

        <div className="py-2">
          {/* Browse */}
          <div className="px-3">
            <button
              className="w-full flex items-center justify-between rounded-md px-3 py-3 hover:bg-gray-100 transition-colors cursor-pointer"
              onClick={() => setMobileBrowseOpen((v) => !v)}
            >
              <span className="font-semibold text-primary">Browse</span>
              <FaChevronDown
                className={`text-primary transition-transform ${
                  mobileBrowseOpen ? "rotate-180" : ""
                }`}
              />
            </button>

            {mobileBrowseOpen && (
              <div className="px-1 pt-2 pb-4">
                <div className="flex flex-wrap gap-2">
                  {genres.length > 0 ? (
                    genres.map((g) => (
                      <NavLink
                        key={g.id}
                        to={`/filter/${g.slug}`}
                        className="px-3 py-1.5 rounded-full bg-primary text-white text-sm hover:bg-primary/90 transition-colors"
                        onClick={() => setMenuOpen(false)}
                      >
                        {g.name}
                      </NavLink>
                    ))
                  ) : (
                    <span className="px-2 text-gray-400 text-sm">
                      Genre belum tersedia
                    </span>
                  )}
                </div>
              </div>
            )}
          </div>

          <div className="mx-3 border-t border-gray-200" />

          {/* Create */}
          <NavLink
            to="/workspace"
            onClick={() => setMenuOpen(false)}
            className="block px-6 py-3 hover:bg-gray-100 transition-colors"
          >
            <span className="text-primary font-semibold">Create</span>
          </NavLink>

          <div className="mx-3 border-t border-gray-200" />

          {/* Library */}
          <button
            onClick={() => {
              setMenuOpen(false);
              handleGoLibrary();
            }}
            className="w-full text-left px-6 py-3 hover:bg-gray-100 transition-colors"
          >
            <span className="text-primary font-semibold">Library</span>
          </button>
        </div>
      </div>

      {menuOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-[55]"
          onClick={() => setMenuOpen(false)}
        />
      )}
    </nav>
  );
}
