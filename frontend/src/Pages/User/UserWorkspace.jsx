import { useState } from "react";
import WorkspaceSidebar from "../../Components/Workspace/WorkspaceSidebar";
import { Outlet } from "react-router-dom";

export default function UserWorkspace() {
  const [isOpen, setIsOpen] = useState(true);

  return (
    <div className="flex min-h-screen bg-gray-50">
      {/* === Sidebar === */}
      <WorkspaceSidebar isOpen={isOpen} setIsOpen={setIsOpen} />

      {/* === Main Workspace === */}
      <main
        className={`flex-grow min-h-screen transition-all duration-300 overflow-x-hidden 
          ${isOpen ? "md:ml-64" : "md:ml-20"} ml-0 p-3 md:p-6`}
      >
        {/* Scroll wrapper biar konten panjang tetap bisa di-scroll */}
        <div className="max-w-full ">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
