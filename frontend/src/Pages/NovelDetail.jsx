import { useParams, Navigate } from "react-router-dom";
import { useEffect, useState } from "react";
import Navbar from "../components/Navbar";
import Footer from "../components/Footer";
import NovelCard from "../Components/Novel/NovelCard";
import NovelDescription from "../Components/Novel/NovelDescription";
import NovelTags from "../Components/Novel/NovelTags";
import NovelPreview from "../Components/Novel/NovelPreview";
import useScrollToTop from "../hooks/useScrollToTop";
import { getNovelById } from "../services/novelService";

export default function NovelDetail() {
  useScrollToTop();
  const { id } = useParams();
  const [novel, setNovel] = useState(null);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");

  useEffect(() => {
    let cancelled = false;

    (async () => {
      setLoading(true);
      setErr("");
      try {
        const data = await getNovelById(id);
        console.log(novel);
        if (!cancelled) setNovel(data);
      } catch (e) {
        if (!cancelled) setErr(e?.message || "Failed to load novel");
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [id]);

  if (loading) {
    return (
      <>
        <Navbar />
        <section className="px-6 py-10 text-gray-600">Loading…</section>
        <Footer />
      </>
    );
  }

  if (err) {
    return (
      <>
        <Navbar />
        <section className="px-6 py-10 text-red-600">{err}</section>
        <Footer />
      </>
    );
  }

  if (!novel) return <Navigate to="*" replace />;

  return (
    <>
      <Navbar />
      <section id="novel-detail">
        <NovelCard novel={novel} />
        <NovelDescription novel={novel} />
        <NovelTags tags={novel.tags} />
        <NovelPreview novel={novel} />
      </section>
      <Footer />
    </>
  );
}
