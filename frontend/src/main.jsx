import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";

import { AuthProvider } from "./context/AuthContext";

import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import ProtectedRoute from "./utils/ProtectedRoute";
import ProtectedAdminRoute from "./utils/ProtectedAdminRoute";
import { Toaster } from "react-hot-toast";

import Home from "./Pages/Home";
import Register from "./Pages/Register";
import Login from "./Pages/Login";

import UserProfile from "./Pages/User/UserProfile";
import UserSettings from "./Pages/User/UserSettings";
import UserChangePassword from "./Pages/User/UserChangePassword";
import UserLibrary from "./Pages/User/UserLibrary";

import AdminLogin from "./Pages/Admin/AdminLogin";
import AdminPanel from "./Pages/Admin/AdminPanel";
import AdminConfig from "./Pages/Admin/Config/AdminConfig";
import AdminUsers from "./Pages/Admin/Users/AdminUsersList";
import AdminNovelsList from "./Pages/Admin/Novels/AdminNovelsList";
import AdminNovelsCreate from "./Pages/Admin/Novels/AdminNovelsCreate";
import AdminNovelsDetail from "./Pages/Admin/Novels/AdminNovelsDetail";
import AdminNovelsEdit from "./Pages/Admin/Novels/AdminNovelsEdit";
import AdminNovelsWrite from "./Pages/Admin/Chapters/AdminNovelsWriteChapter";
import AdminNovelsEditChapter from "./Pages/Admin/Chapters/AdminNovelsEditChapter";

import UserWorkspace from "./Pages/User/UserWorkspace";
import UserNovelsList from "./Pages/User/Workspace/Novels/UserNovelsList";
import UserNovelsCreate from "./Pages/User/Workspace/Novels/UserNovelsCreate";
import UserNovelsDetail from "./Pages/User/Workspace/Novels/UserNovelsDetail";
import UserNovelsEdit from "./Pages/User/Workspace/Novels/UserNovelsEdit";
import UserNovelsEditChapter from "./Pages/User/Workspace/Chapters/UserNovelsEditChapter";
import UserNovelsWrite from "./Pages/User/Workspace/Chapters/UserNovelsWriteChapter";

import NovelDetail from "./Pages/NovelDetail";
import ChapterReader from "./Pages/ChapterReader";

import Filter from "./Pages/Filter";

import NotFound from "./Pages/NotFound";

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <AuthProvider>
      <BrowserRouter>
        <Toaster position="top-center" reverseOrder={false} />
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="filter" element={<Filter />}></Route>
          <Route path="filter/:genreSlug" element={<Filter />}></Route>
          <Route path="register" element={<Register />}></Route>
          <Route path="login" element={<Login />}></Route>
          <Route path="admin-login" element={<AdminLogin />}></Route>
          <Route path="novel/:id" element={<NovelDetail />} />
          <Route
            path="novel/:id/chapter/:chapterId"
            element={<ChapterReader />}
          />
          <Route element={<ProtectedRoute />}>
            <Route path="/profile/" element={<UserProfile />} />
            <Route path="/settings" element={<UserSettings />} />
            <Route path="/change-password" element={<UserChangePassword />} />
            <Route path="/library" element={<UserLibrary />} />
            <Route path="/workspace" element={<UserWorkspace />}>
              <Route index element={<Navigate to="novels" replace />} />
              <Route path="novels">
                <Route index element={<UserNovelsList />} />
                <Route path="create" element={<UserNovelsCreate />} />
                <Route path=":id" element={<UserNovelsDetail />} />
                <Route path="edit/:id" element={<UserNovelsEdit />} />
                <Route path=":id/write" element={<UserNovelsWrite />} />
                <Route
                  path=":id/edit-chapter/:chapterId"
                  element={<UserNovelsEditChapter />}
                />
              </Route>
            </Route>
          </Route>

          <Route element={<ProtectedAdminRoute />}>
            <Route path="admin" element={<AdminPanel />}>
              <Route index element={<Navigate to="users" replace />} />
              <Route path="users" element={<AdminUsers />} />
              <Route path="novels">
                <Route index element={<AdminNovelsList />} />
                <Route path="create" element={<AdminNovelsCreate />} />
                <Route path=":id" element={<AdminNovelsDetail />} />
                <Route path="edit/:id" element={<AdminNovelsEdit />} />
                <Route path=":id/write" element={<AdminNovelsWrite />} />
                <Route
                  path=":id/edit-chapter/:chapterId"
                  element={<AdminNovelsEditChapter />}
                />
              </Route>
              <Route path="config" element={<AdminConfig />} />
            </Route>
          </Route>

          <Route path="*" element={<NotFound />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  </StrictMode>
);
