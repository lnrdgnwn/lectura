import { NavLink } from "react-router-dom";

export default function MenuList({ onLogout }) {
  return (
    <div className="bg-white rounded-2xl shadow-sm">
      {/* Novel / Workspace */}
      <NavLink
        to="/workspace"
        className="w-full text-left px-5 py-3 flex items-center justify-between hover:bg-gray-50 cursor-pointer"
      >
        <span className="text-sm font-medium">Novel</span>
        <span className="text-sm text-gray-600 font-medium">2</span>
      </NavLink>

      {/* Change Password */}
      <NavLink
        to="/change-password"
        className="w-full text-left px-5 py-3 flex items-center justify-between hover:bg-gray-50 cursor-pointer"
      >
        <span className="text-sm font-medium">Change Password</span>
        <span className="text-gray-400">{">"}</span>
      </NavLink>

      {/* Logout */}
      <button
        onClick={onLogout}
        className="w-full text-left px-5 py-3 flex items-center justify-between hover:bg-gray-50 cursor-pointer"
      >
        <span className="text-sm font-medium text-red-600">Logout</span>
        <span className="text-gray-400">{">"}</span>
      </button>
    </div>
  );
}
