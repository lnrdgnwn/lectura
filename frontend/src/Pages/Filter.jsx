import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import Navbar from "../components/Navbar";
import Footer from "../components/Footer";
import FilterSidebar from "../components/Filter/FilterSidebar";
import FilterNovelList from "../components/Filter/FilterNovelList";
import FilterPagination from "../components/Filter/FilterPagination";
import FilterBar from "../components/Filter/FilterBar";
import FilterModal from "../components/Filter/FilterModal";
import useScrollToTop from "../hooks/useScrollToTop";
import { getNovels } from "../services/novelService";
import { getGenres } from "../services/genreService";
import { getTags } from "../services/TagService";

const normSlug = (s) => {
  const v = (s ?? "").toString().toLowerCase();
  return !v || v === "all" ? "All" : v;
};

export default function Filter() {
  useScrollToTop();

  const navigate = useNavigate();
  const { genreSlug } = useParams();
  const [sp] = useSearchParams();

  const [novels, setNovels] = useState([]);
  const [genres, setGenres] = useState([]);
  const [tags, setTags] = useState([]);

  const [selectedGenreSlug, setSelectedGenreSlug] = useState("All");
  const [selectedTags, setSelectedTags] = useState([]);
  const [page, setPage] = useState(1);
  const [modalOpen, setModalOpen] = useState(false);

  const novelsPerPage = 8;

  useEffect(() => {
    (async () => {
      try {
        const [g, t, n] = await Promise.all([
          getGenres(),
          getTags(),
          getNovels(),
        ]);
        setGenres(g || []);
        setTags(t || []);
        setNovels(n || []);
      } catch {}
    })();
  }, []);

  useEffect(() => {
    if ((genreSlug ?? "").toLowerCase() === "all") {
      navigate("/filter", { replace: true });
    }
  }, [genreSlug, navigate]);

  const { genreNames, slugToName, nameToSlug } = useMemo(() => {
    const names = (genres || []).map((g) => g?.name || g?.slug || "");
    const s2n = Object.create(null);
    const n2s = Object.create(null);
    (genres || []).forEach((g) => {
      const slug = String(g?.slug || "").toLowerCase();
      const name = g?.name || slug;
      if (slug) s2n[slug] = name;
      if (name) n2s[name] = slug;
    });
    return { genreNames: ["All", ...names], slugToName: s2n, nameToSlug: n2s };
  }, [genres]);

  const tagNames = useMemo(
    () => (tags || []).map((t) => t?.name ?? t?.slug ?? String(t)),
    [tags]
  );

  const query = (sp.get("q") || "").trim().toLowerCase();

  useEffect(() => {
    setSelectedGenreSlug((prev) => {
      const incoming = normSlug(genreSlug);
      return prev === incoming ? prev : incoming;
    });
  }, [genreSlug]);

  useEffect(() => {
    const q = sp.get("q");
    const base = "/filter";
    const path =
      selectedGenreSlug === "All" ? base : `${base}/${selectedGenreSlug}`;
    const current = window.location.pathname + (window.location.search || "");
    const next = q ? `${path}?q=${encodeURIComponent(q)}` : path;
    if (current !== next) navigate(next, { replace: false });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedGenreSlug]);

  const selectedGenreName =
    selectedGenreSlug === "All"
      ? "All"
      : slugToName[selectedGenreSlug] || selectedGenreSlug;

  const filtered = useMemo(() => {
    const list = novels || [];

    const matchQuery = (n) => {
      if (!query) return true;
      const title = (n?.title || "").toLowerCase();
      const author = (n?.author?.username || n?.author || "").toLowerCase();
      return title.includes(query) || author.includes(query);
    };

    const matchGenre = (n) => {
      if (selectedGenreSlug === "All") return true;
      const gs = (n?.genres || []).map((g) =>
        String(g?.slug || g?.name || "").toLowerCase()
      );
      return gs.includes(String(selectedGenreSlug).toLowerCase());
    };

    const matchTags = (n) => {
      if (!selectedTags.length) return true;
      const ts = (n?.tags || []).map((t) =>
        String(t?.name ?? t?.slug ?? t).toLowerCase()
      );
      return selectedTags.every((t) => ts.includes(String(t).toLowerCase()));
    };

    return list.filter((n) => matchQuery(n) && matchGenre(n) && matchTags(n));
  }, [novels, query, selectedGenreSlug, selectedTags]);

  const totalPages = Math.max(1, Math.ceil(filtered.length / novelsPerPage));
  const paginated = useMemo(() => {
    const start = (page - 1) * novelsPerPage;
    return filtered.slice(start, start + novelsPerPage);
  }, [filtered, page]);

  useEffect(() => {
    setPage(1);
  }, [query, selectedGenreSlug, selectedTags]);

  const handleGenreSelectByName = (name) => {
    if (name === "All") return setSelectedGenreSlug("All");
    const slug = nameToSlug[name];
    setSelectedGenreSlug(slug || "All");
  };

  const handleTagToggle = (tag) =>
    setSelectedTags((prev) =>
      prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag]
    );

  return (
    <>
      <Navbar />
      <div className="flex justify-center items-start min-h-screen bg-gray-50 py-6 px-2">
        <FilterSidebar
          genres={genreNames}
          tags={tagNames}
          selectedGenre={selectedGenreName}
          selectedTags={selectedTags}
          onGenreSelect={handleGenreSelectByName}
          onTagToggle={handleTagToggle}
        />

        <div className="flex flex-col items-center w-full max-w-4xl">
          <FilterBar
            genres={genreNames}
            selectedGenre={selectedGenreName}
            setSelectedGenre={handleGenreSelectByName}
            onOpen={() => setModalOpen(true)}
          />

          <FilterNovelList novels={paginated} />

          <FilterPagination
            current={page}
            total={totalPages}
            hasData={filtered.length > 0}
            onChange={(p) => p >= 1 && p <= totalPages && setPage(p)}
          />
        </div>
      </div>

      <FilterModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        genres={genreNames}
        tags={tagNames}
        selectedGenre={selectedGenreName}
        setSelectedGenre={handleGenreSelectByName}
        selectedTags={selectedTags}
        setSelectedTags={setSelectedTags}
      />
      <Footer />
    </>
  );
}
