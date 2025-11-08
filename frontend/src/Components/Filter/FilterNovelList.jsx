import FilterNovelCard from "./FilterNovelCard";

export default function FilterNovelList({ novels = [] }) {
  return (
    <div className="w-full max-w-5xl flex flex-col gap-4 mt-5 lg:mt-0 lg:px-4">
      {novels.length > 0 ? (
        novels.map((novel) => <FilterNovelCard key={novel.id} novel={novel} />)
      ) : (
        <p className="text-gray-500 text-center py-6">No novels found.</p>
      )}
    </div>
  );
}
