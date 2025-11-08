import { useEffect, useState } from "react";
import Navbar from "../components/Navbar";
import Footer from "../components/Footer";
import Slider from "../components/Home/Slider";
import List from "../components/Home/List";
import { getGenres } from "../services/genreService";

export default function Home() {
  const [homeGenres, setHomeGenres] = useState([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");

  useEffect(() => {
    (async () => {
      setLoading(true);
      setErr("");
      try {
        const genres = await getGenres();
        const filtered = (genres || []).filter((g) => g.show_on_home);
        setHomeGenres(filtered);
      } catch (e) {
        setErr(e.message || "Failed to load genres");
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  return (
    <>
      <Navbar />
      <section id="home" className="bg-gray-50 min-h-screen ">
        <Slider />

        {loading ? (
          <div className="px-6 md:px-10 py-6 text-gray-500 italic">
            Loading genres…
          </div>
        ) : err ? (
          <div className="px-6 md:px-10 py-6 text-red-600">{err}</div>
        ) : homeGenres.length === 0 ? (
          <div className="px-6 md:px-10 py-6 text-gray-500 italic">
            Belum ada genre yang diaktifkan{" "}
            <span className="font-semibold">show_on_home</span>.
          </div>
        ) : (
          homeGenres.map((g) => (
            <List
              key={g.id}
              genreId={g.id}
              genreName={g.name}
              genreSlug={g.slug}
              limit={6}
              sort="updated_at:desc"
              titleOverride={g.name}
            />
          ))
        )}

        <Footer />
      </section>
    </>
  );
}
