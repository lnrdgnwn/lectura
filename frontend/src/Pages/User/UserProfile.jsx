import Navbar from "../../components/Navbar";
import Footer from "../../components/Footer";
import ProfileHeader from "../../Components/Profile/ProfileHeader";
import MenuList from "../../Components/Profile/MenuList";
import useScrollToTop from "../../hooks/useScrollToTop";
import { useAuth } from "../../context/AuthContext";
import { useNavigate } from "react-router-dom";

export default function UserProfile() {
  useScrollToTop();
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  if (!user) return null;

  const avatar = user.profile_picture;

  const handleLogout = async () => {
    try {
      await logout();
    } finally {
      navigate("/", { replace: true });
    }
  };

  return (
    <>
      <Navbar />
      <main className="min-h-screen bg-gray-50 pb-8">
        <div className="mx-auto w-full max-w-md sm:max-w-2xl lg:max-w-5xl px-4 py-6 sm:py-10">
          <ProfileHeader
            username={user.username}
            id={user.id}
            avatar={avatar}
            onEdit={() => navigate("/settings")}
          />
          <div className="mt-5">
            <MenuList onLogout={handleLogout} />
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
