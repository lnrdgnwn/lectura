import { useState } from "react";
import { FaChevronDown, FaChevronUp } from "react-icons/fa";
import NovelTableOfContents from "./NovelTableOfContents";

export default function NovelDescription({ novel }) {
  const [activeTab, setActiveTab] = useState("about");
  const [expanded, setExpanded] = useState(false);

  const maxLength = 300;
  const synopsis = novel?.synopsis || "";
  const isLong = synopsis.length > maxLength;
  const displayText =
    expanded || !isLong ? synopsis : synopsis.slice(0, maxLength) + "...";

  return (
    <section className="w-full flex flex-col items-center justify-center pb-3 px-5">
      {/* Tabs */}
      <div className="pt-3 w-full max-w-4xl flex flex-wrap items-center gap-6 text-base sm:text-lg font-semibold border-b border-gray-300">
        <button
          onClick={() => setActiveTab("about")}
          className={`pb-2 transition-colors ${
            activeTab === "about"
              ? "border-b-2 border-black text-black"
              : "text-gray-500 hover:text-black"
          }`}
        >
          About
        </button>

        <button
          onClick={() => setActiveTab("toc")}
          className={`pb-2 transition-colors ${
            activeTab === "toc"
              ? "border-b-2 border-black text-black"
              : "text-gray-500 hover:text-black"
          }`}
        >
          Table of Contents
        </button>
      </div>

      {/* Content */}
      <div className="w-full max-w-4xl mt-6 text-gray-800">
        {activeTab === "about" && (
          <div className="animate-fadeIn relative">
            <h3 className="font-bold text-lg sm:text-xl mb-2">Synopsis</h3>
            <p className="leading-relaxed text-sm sm:text-base whitespace-pre-line">
              {displayText}
            </p>

            {isLong && (
              <button
                type="button"
                onClick={() => setExpanded(!expanded)}
                className="absolute bottom-0 right-0 p-1 text-primary hover:text-blue-600 transition"
                aria-expanded={expanded}
                aria-label={expanded ? "Show less" : "Show more"}
              >
                {expanded ? (
                  <FaChevronUp className="text-sm" />
                ) : (
                  <FaChevronDown className="text-sm" />
                )}
              </button>
            )}
          </div>
        )}

        {activeTab === "toc" && (
          <div className="animate-fadeIn">
            <NovelTableOfContents novel={novel} />
          </div>
        )}
      </div>
    </section>
  );
}
