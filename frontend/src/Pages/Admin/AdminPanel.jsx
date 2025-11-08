import { useState } from "react";
import AdminSidebar from "../../Components/Admin/AdminSidebar";
import { Outlet } from "react-router-dom";

export default function AdminPanel() {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="flex min-h-screen bg-gray-50">
      {/* Sidebar */}
      <AdminSidebar isOpen={isOpen} setIsOpen={setIsOpen} />

      {/* Main workspace */}
      <main
        className={`flex-grow p-4 ml-0 transition-all duration-300 overflow-x-hidden ${
          isOpen ? "ml-64" : "md:ml-20"
        }`}
      >
        <Outlet />
      </main>
    </div>
  );
}
