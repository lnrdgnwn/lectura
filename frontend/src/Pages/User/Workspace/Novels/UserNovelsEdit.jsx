// pages/Admin/Novels/AdminNovelsEdit.jsx
import { Navigate, useParams } from "react-router-dom";
import NovelForm from "../../../../Components/Workspace/NovelForm";

export default function UserNovelsEdit() {
  const { id } = useParams();

  if (!id) return <Navigate to="/user/novels" replace />;
  return <NovelForm mode="edit" novelId={id} />;
}
