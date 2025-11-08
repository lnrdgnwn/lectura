import { NavLink } from "react-router-dom";

export default function ProfileHeader({ username, id, avatar, onEdit }) {
  return (
    <section className="bg-gradient-to-tr from-primary to-blue-400 rounded-2xl p-5 text-white shadow-md">
      <div className="flex items-center gap-4">
        <img
          src={avatar}
          alt={username}
          className="w-16 h-16 rounded-full border-2 border-white object-cover"
        />
        <div className="flex-1">
          <h2 className="text-lg font-semibold">{username}</h2>
          {id && <p className="text-xs opacity-80">ID : {id}</p>}
        </div>
        <NavLink
          to={"/settings"}
          onClick={onEdit}
          className="bg-white text-primary text-sm font-medium px-3 py-1 rounded-full shadow-sm cursor-pointer"
        >
          Edit Profile
        </NavLink>
      </div>
    </section>
  );
}
