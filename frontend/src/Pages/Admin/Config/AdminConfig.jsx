// pages/AdminConfig.jsx
import { useState } from "react";
import GenreConfig from "../../../Components/Admin/Config/GenreConfig";
import TagConfig from "../../../Components/Admin/Config/TagConfig";

export default function AdminConfig() {
  const [activeTab, setActiveTab] = useState("genres"); // genres | tags

  const TabButton = ({ id, children }) => (
    <button
      onClick={() => setActiveTab(id)}
      className={`px-4 py-2 rounded-lg text-sm font-medium cursor-pointer
        ${
          activeTab === id
            ? "bg-primary text-white"
            : "bg-white border border-gray-300 text-gray-700 hover:bg-gray-50"
        }`}
    >
      {children}
    </button>
  );

  return (
    <section className="bg-gray-50 min-h-screen py-6 px-3 md:px-6">
      <div className="max-w-6xl mx-auto">
        <div className="mb-4">
          <h1 className="text-2xl font-semibold text-gray-800">Admin Config</h1>
        </div>

        <div className="mb-4 flex gap-2">
          <TabButton id="genres">Genres</TabButton>
          <TabButton id="tags">Tags</TabButton>
        </div>

        <div className="rounded-xl border border-gray-200 bg-white shadow-sm p-4 md:p-5">
          {activeTab === "genres" ? <GenreConfig /> : <TagConfig />}
        </div>
      </div>
    </section>
  );
}
