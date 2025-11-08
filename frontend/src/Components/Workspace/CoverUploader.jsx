import { useEffect, useRef } from "react";

export default function CoverUploader({ preview, onFile, error }) {
  const inputRef = useRef(null);

  const handleChange = (e) => {
    const file = e.target.files?.[0];
    if (file) {
      onFile(file);
    } else {
      onFile(null);
    }
  };

  return (
    <div>
      <label className="text-sm font-medium text-gray-700 mb-1 block">
        Cover Image
      </label>
      <div className="flex items-center gap-4 flex-wrap">
        <div
          className={`w-28 h-36 md:w-32 md:h-40 border rounded-lg overflow-hidden bg-gray-50 flex items-center justify-center ${
            error ? "border-red-300" : "border-gray-300"
          }`}
        >
          {preview ? (
            <img
              src={preview}
              alt="Preview"
              className="block w-full h-full object-cover"
            />
          ) : (
            <span className="text-gray-400 text-xs">No image</span>
          )}
        </div>
        <button
          type="button"
          onClick={() => inputRef.current?.click()}
          className="border border-gray-300 rounded-lg px-3 py-2 text-sm font-medium hover:bg-gray-50"
        >
          Upload Image
        </button>
        <input
          ref={inputRef}
          type="file"
          accept="image/*"
          onChange={handleChange}
          className="hidden"
        />
      </div>
      {error && <p className="mt-1 text-xs text-red-600">{error}</p>}
    </div>
  );
}
