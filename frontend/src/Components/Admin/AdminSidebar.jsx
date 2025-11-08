import { useEffect } from "react";
import { FaUser, FaBook, FaBars, FaSignOutAlt, FaEdit } from "react-icons/fa";
import { IoClose } from "react-icons/io5";
import { NavLink, useNavigate } from "react-router-dom";
import { useAuth } from "../../context/AuthContext"; // pastikan path-nya sesuai

export default function AdminSidebar({ isOpen, setIsOpen }) {
  const { logout } = useAuth();
  const navigate = useNavigate();
  const sidebarWidth = isOpen ? "w-64" : "w-20";

  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth >= 768 && isOpen) {
        setIsOpen(false);
      }
    };
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, [isOpen, setIsOpen]);

  const handleLogout = async () => {
    try {
      await logout();
      navigate("/admin-login");
    } catch (err) {
      console.error("Logout failed:", err);
    }
  };

  return (
    <>
      {/* ===== Desktop Sidebar (Kiri) ===== */}
      <aside
        className={`hidden md:flex flex-col ${sidebarWidth} fixed transition-all h-screen duration-300
        bg-primary text-white shadow-lg p-5`}
      >
        <div
          className={`flex items-center mb-6 ${
            isOpen ? "justify-between" : "justify-center"
          }`}
        >
          {isOpen && (
            <img src="/img/lectura-white.png" alt="Logo" className="w-32" />
          )}
          <button
            onClick={() => setIsOpen(!isOpen)}
            className="hidden xl:block text-white hover:bg-white/10 p-3 rounded-lg cursor-pointer "
          >
            {isOpen ? <IoClose size={22} /> : <FaBars size={20} />}
          </button>
        </div>

        <nav className="flex flex-col space-y-2">
          <NavLink
            to="/admin/users"
            className={({ isActive }) =>
              `flex items-center ${
                isOpen ? "justify-start gap-3 px-3" : "justify-center"
              } 
              p-3 rounded-lg duration-200 ${
                isActive ? "bg-white/20" : "hover:bg-white/10 text-white"
              }`
            }
          >
            <FaUser size={18} />
            {isOpen && <span className="text-sm font-medium">Users</span>}
          </NavLink>

          <NavLink
            to="/admin/novels"
            className={({ isActive }) =>
              `flex items-center ${
                isOpen ? "justify-start gap-3 px-3" : "justify-center"
              } 
              p-3 rounded-lg duration-200 ${
                isActive ? "bg-white/20" : "hover:bg-white/10 text-white"
              }`
            }
          >
            <FaBook size={18} />
            {isOpen && <span className="text-sm font-medium">Novels</span>}
          </NavLink>

          <NavLink
            to="/admin/config"
            className={({ isActive }) =>
              `flex items-center ${
                isOpen ? "justify-start gap-3 px-3" : "justify-center"
              } 
              p-3 rounded-lg duration-200 ${
                isActive ? "bg-white/20" : "hover:bg-white/10 text-white"
              }`
            }
          >
            <FaEdit size={18} />
            {isOpen && <span className="text-sm font-medium">Config</span>}
          </NavLink>
        </nav>

        <button
          onClick={handleLogout}
          className={`flex items-center ${
            isOpen ? "justify-start gap-3 px-3" : "justify-center"
          } 
            p-3 rounded-lg hover:bg-white/10 text-white mt-auto`}
        >
          <FaSignOutAlt size={18} />
          {isOpen && <span className="text-sm font-medium">Logout</span>}
        </button>
      </aside>

      {/* ===== Mobile Bottom Nav ===== */}
      <aside
        className={`md:hidden fixed bottom-0 left-0 right-0 bg-primary text-white 
        shadow-[0_-2px_10px_rgba(0,0,0,0.2)] flex justify-around items-center py-2 z-40 transition-all duration-300`}
      >
        <NavLink
          to="/admin/users"
          className={({ isActive }) =>
            `flex flex-col items-center justify-center gap-1 px-3 py-2 text-xs ${
              isActive ? "" : "text-white"
            }`
          }
          onClick={() => setIsOpen(false)}
        >
          <FaUser size={18} />
          <span>Users</span>
        </NavLink>

        <NavLink
          to="/admin/novels"
          className={({ isActive }) =>
            `flex flex-col items-center justify-center gap-1 px-3 py-2 text-xs ${
              isActive ? "" : "text-white"
            }`
          }
          onClick={() => setIsOpen(false)}
        >
          <FaBook size={18} />
          <span>Novels</span>
        </NavLink>

        <NavLink
          to="/admin/config"
          className={({ isActive }) =>
            `flex flex-col items-center justify-center gap-1 px-3 py-2 text-xs ${
              isActive ? "" : "text-white"
            }`
          }
          onClick={() => setIsOpen(false)}
        >
          <FaEdit size={18} />
          <span>Config</span>
        </NavLink>

        <button
          onClick={handleLogout}
          className="flex flex-col items-center justify-center gap-1 px-3 py-2 text-xs text-white"
        >
          <FaSignOutAlt size={18} />
          <span>Logout</span>
        </button>
      </aside>
    </>
  );
}
