// pages/ChapterReader.jsx
import { useMemo, useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import ChapterBreadcrumb from "../Components/Reader/ChapterBreadcrumb";
import ChapterContent from "../Components/Reader/ChapterContent";
import ReaderControl from "../Components/Reader/ReaderControl";
import ChapterMenu from "../Components/Reader/ChapterMenu";
import Navbar from "../components/Navbar";
import Footer from "../components/Footer";
import useScrollToTop from "../hooks/useScrollToTop";
import { getNovelById } from "../services/novelService";

export default function ChapterReader() {
  const { id, chapterId } = useParams();
  const navigate = useNavigate();

  const [novel, setNovel] = useState(null);
  const [chapters, setChapters] = useState([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");

  const [darkMode, setDarkMode] = useState(false);
  const [textSize, setTextSize] = useState(16);
  const [showControls, setShowControls] = useState(true);
  const [showChapters, setShowChapters] = useState(false);

  useScrollToTop();

  // Fetch novel + chapters
  useEffect(() => {
    (async () => {
      setLoading(true);
      setErr("");
      try {
        const data = await getNovelById(id);
        // Sort chapters by order_no asc (fallback created_at)
        const sorted = [...(data?.chapters ?? [])].sort((a, b) => {
          const ao = Number.isFinite(+a.order_no) ? +a.order_no : Infinity;
          const bo = Number.isFinite(+b.order_no) ? +b.order_no : Infinity;
          if (ao !== bo) return ao - bo;
          return new Date(a?.created_at || 0) - new Date(b?.created_at || 0);
        });
        setNovel(data);
        setChapters(sorted);
      } catch {
        setErr("Failed to load chapter");
      } finally {
        setLoading(false);
      }
    })();
  }, [id]);

  // Index chapter aktif
  const currentIndex = useMemo(() => {
    if (!chapters.length) return -1;
    const idx = chapters.findIndex((c) => String(c.id) === String(chapterId));
    return idx >= 0 ? idx : 0;
  }, [chapters, chapterId]);

  const currentChapter = currentIndex >= 0 ? chapters[currentIndex] : null;

  // Redirect ke chapter pertama jika chapterId invalid
  useEffect(() => {
    if (!loading && !err && chapters.length > 0) {
      const found = chapters.find((c) => String(c.id) === String(chapterId));
      if (!found) {
        navigate(`/novel/${id}/chapter/${chapters[0].id}`, { replace: true });
      }
    }
  }, [loading, err, chapters, chapterId, id, navigate]);

  // Navigasi antar chapter
  const hasPrev = useMemo(() => {
    // Tombol Back selalu aktif: kalau di first chapter, back -> '/'
    if (!currentChapter) return false;
    return currentIndex > 0 || Number(currentChapter.order_no) === 1;
  }, [currentIndex, currentChapter]);

  const hasNext = currentIndex >= 0 && currentIndex < chapters.length - 1;

  const goToNext = () => {
    if (currentIndex >= 0 && currentIndex < chapters.length - 1) {
      navigate(`/novel/${id}/chapter/${chapters[currentIndex + 1].id}`);
    }
  };

  const goToPrev = () => {
    if (!currentChapter) return;
    const orderNo = Number(currentChapter.order_no);
    if (orderNo === 1) {
      navigate("/"); // back to home kalau chapter pertama
    } else if (currentIndex > 0) {
      navigate(`/novel/${id}/chapter/${chapters[currentIndex - 1].id}`);
    }
  };

  // Toggle controls (mobile)
  const toggleControls = () => {
    if (window.innerWidth < 768) setShowControls((prev) => !prev);
  };

  const bgColor = darkMode
    ? "bg-gray-900 text-gray-100"
    : "bg-white text-gray-900";

  if (loading) {
    return (
      <section>
        <Navbar />
        <div className="min-h-screen flex items-center justify-center">
          <p className="text-gray-500">Loading…</p>
        </div>
        <Footer />
      </section>
    );
  }

  if (err || !novel) {
    return (
      <section>
        <Navbar />
        <div className="min-h-screen flex items-center justify-center">
          <p className="text-red-600">{err || "Novel not found"}</p>
        </div>
        <Footer />
      </section>
    );
  }

  return (
    <section>
      <Navbar />
      <div
        className={`${bgColor} min-h-screen flex flex-col md:flex-row transition-colors duration-300`}
      >
        <main
          className="flex-1 px-5 md:px-10 py-5 relative overflow-y-auto"
          onClick={toggleControls}
        >
          <div className="max-w-3xl mx-auto">
            <ChapterBreadcrumb
              novelTitle={novel.title}
              novelId={novel.id}
              chapterTitle={currentChapter?.title}
              chapterNumber={currentChapter?.order_no ?? currentIndex + 1}
              darkMode={darkMode}
            />

            {currentChapter ? (
              <ChapterContent
                chapter={currentChapter}
                textSize={textSize}
                onPrev={goToPrev}
                onNext={hasNext ? goToNext : undefined}
                hasPrev={hasPrev}
                hasNext={hasNext}
                darkMode={darkMode}
              />
            ) : (
              <p className="text-gray-500">No chapter content.</p>
            )}
          </div>
        </main>

        <ReaderControl
          darkMode={darkMode}
          setDarkMode={setDarkMode}
          textSize={textSize}
          setTextSize={setTextSize}
          showControls={showControls}
          setShowChapters={setShowChapters}
          goToNext={goToNext}
          goToPrev={goToPrev}
        />

        <ChapterMenu
          show={showChapters}
          setShow={setShowChapters}
          novelTitle={novel.title}
          novelCover={novel.cover_image || novel.cover_url}
          novelAuthor={novel.author?.username}
          chapters={chapters}
          currentChapterId={String(currentChapter?.id ?? "")}
          onSelectChapter={(cid) => {
            setShowChapters(false);
            navigate(`/novel/${id}/chapter/${cid}`);
          }}
        />
      </div>
      <Footer />
    </section>
  );
}
