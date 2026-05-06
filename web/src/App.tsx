import { Routes, Route } from 'react-router-dom'
import { ThemeProvider } from './lib/ThemeContext'
import FrontLayout from './front/Layout'
import Home from './front/Home'
import PostDetail from './front/PostDetail'
import Treehole from './front/Treehole'
import Archive from './front/Archive'
import Search from './front/Search'
import About from './front/About'
import Links from './front/Links'
import Login from './admin/Login'
import Setup from './admin/Setup'
import AdminLayout from './admin/Layout'
import Dashboard from './admin/Dashboard'
import PostList from './admin/PostList'
import PostEditor from './admin/PostEditor'
import MediaLibrary from './admin/MediaLibrary'
import CommentList from './admin/CommentList'
import SettingsPage from './admin/Settings'
import Tools from './admin/Tools'

export default function App() {
  return (
    <ThemeProvider>
      <Routes>
        {/* Frontend */}
        <Route element={<FrontLayout />}>
          <Route path="/" element={<Home />} />
          <Route path="/post/:slug" element={<PostDetail />} />
          <Route path="/treehole" element={<Treehole />} />
          <Route path="/archive" element={<Archive />} />
          <Route path="/search" element={<Search />} />
          <Route path="/about" element={<About />} />
          <Route path="/links" element={<Links />} />
        </Route>

        {/* Admin */}
        <Route path="/setup" element={<Setup />} />
        <Route path="/admin/login" element={<Login />} />
        <Route path="/admin" element={<AdminLayout />}>
          <Route index element={<Dashboard />} />
          <Route path="posts" element={<PostList />} />
          <Route path="posts/new" element={<PostEditor />} />
          <Route path="posts/:id/edit" element={<PostEditor />} />
          <Route path="media" element={<MediaLibrary />} />
          <Route path="comments" element={<CommentList />} />
          <Route path="settings" element={<SettingsPage />} />
        <Route path="tools" element={<Tools />} />
        </Route>
      </Routes>
    </ThemeProvider>
  )
}
