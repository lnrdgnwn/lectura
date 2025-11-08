// pages/Admin/Novels/AdminNovelsEdit.jsx
import { Navigate, useParams } from "react-router-dom";
import NovelForm from "../../../Components/Admin/Novels/NovelForm";

export default function AdminNovelsEdit() {
  const { id } = useParams();

  if (!id) return <Navigate to="/admin/novels" replace />;
  return <NovelForm mode="edit" novelId={id} />;
}
