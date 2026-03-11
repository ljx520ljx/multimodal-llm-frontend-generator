'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { api } from '@/lib/api/client';
import { useAuthStore } from '@/stores/useAuthStore';
import type { Project } from '@/types';

export default function DashboardPage() {
  const router = useRouter();
  const user = useAuthStore((state) => state.user);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const logout = useAuthStore((state) => state.logout);

  const [projects, setProjects] = useState<Project[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [newProjectName, setNewProjectName] = useState('');
  const [newProjectDesc, setNewProjectDesc] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login');
      return;
    }
    loadProjects();
  }, [isAuthenticated, router]);

  const loadProjects = useCallback(async () => {
    try {
      setIsLoading(true);
      const result = await api.projects.list();
      setProjects(result.projects);
    } catch {
      // If unauthorized, redirect to login
      router.push('/login');
    } finally {
      setIsLoading(false);
    }
  }, [router]);

  const handleCreateProject = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newProjectName.trim()) return;

    setIsCreating(true);
    setErrorMessage(null);
    try {
      await api.projects.create(newProjectName, newProjectDesc);
      setShowCreateDialog(false);
      setNewProjectName('');
      setNewProjectDesc('');
      await loadProjects();
    } catch {
      setErrorMessage('创建项目失败，请重试');
    } finally {
      setIsCreating(false);
    }
  }, [newProjectName, newProjectDesc, loadProjects]);

  const handleDeleteProject = useCallback(async (id: string) => {
    setErrorMessage(null);
    try {
      await api.projects.delete(id);
      setProjects((prev) => prev.filter((p) => p.id !== id));
    } catch {
      setErrorMessage('删除项目失败，请重试');
    }
  }, []);

  const handleLogout = useCallback(() => {
    logout();
    router.push('/');
  }, [logout, router]);

  if (!isAuthenticated) return null;

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Header */}
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-3">
          <button
            onClick={() => router.push('/')}
            className="text-sm font-semibold text-slate-900 hover:text-blue-600 transition-colors"
          >
            UI to Code
          </button>
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2">
              {user?.avatar_url ? (
                <img src={user.avatar_url} alt="" className="h-6 w-6 rounded-full" />
              ) : (
                <div className="flex h-6 w-6 items-center justify-center rounded-full bg-blue-100 text-xs font-medium text-blue-600">
                  {(user?.display_name || user?.email || '?')[0].toUpperCase()}
                </div>
              )}
              <span className="text-sm text-slate-700">{user?.display_name || user?.email}</span>
            </div>
            <button
              onClick={handleLogout}
              className="rounded px-2 py-1 text-xs text-slate-500 hover:bg-slate-100 hover:text-slate-700"
            >
              退出
            </button>
          </div>
        </div>
      </header>

      {/* Main */}
      <main className="mx-auto max-w-5xl px-6 py-8">
        <div className="mb-6 flex items-center justify-between">
          <h1 className="text-lg font-semibold text-slate-900">我的项目</h1>
          <div className="flex items-center gap-2">
            <button
              onClick={() => router.push('/')}
              className="rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-700 hover:bg-slate-50 transition-colors"
            >
              新原型
            </button>
            <button
              onClick={() => setShowCreateDialog(true)}
              className="rounded-md bg-blue-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-600 transition-colors"
            >
              新建项目
            </button>
          </div>
        </div>

        {errorMessage && (
          <div className="mb-4 rounded-md bg-red-50 border border-red-200 px-4 py-2 text-sm text-red-700">
            {errorMessage}
          </div>
        )}

        {isLoading ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-20 animate-pulse rounded-lg bg-slate-200" />
            ))}
          </div>
        ) : projects.length === 0 ? (
          <div className="rounded-lg border-2 border-dashed border-slate-200 py-12 text-center">
            <p className="text-sm text-slate-500">还没有项目</p>
            <button
              onClick={() => setShowCreateDialog(true)}
              className="mt-3 rounded-md bg-blue-500 px-4 py-2 text-sm font-medium text-white hover:bg-blue-600"
            >
              创建第一个项目
            </button>
          </div>
        ) : (
          <div className="space-y-3">
            {projects.map((project) => (
              <div
                key={project.id}
                className="group flex items-center justify-between rounded-lg border border-slate-200 bg-white px-4 py-3 hover:border-blue-200 transition-colors"
              >
                <div className="min-w-0 flex-1">
                  <h3 className="text-sm font-medium text-slate-900">{project.name}</h3>
                  {project.description && (
                    <p className="mt-0.5 truncate text-xs text-slate-500">{project.description}</p>
                  )}
                  <p className="mt-1 text-xs text-slate-400">
                    更新于 {new Date(project.updated_at).toLocaleDateString('zh-CN')}
                  </p>
                </div>
                <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button
                    onClick={() => handleDeleteProject(project.id)}
                    className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50"
                  >
                    删除
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>

      {/* Create dialog */}
      {showCreateDialog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30">
          <div className="w-full max-w-sm rounded-lg bg-white p-5 shadow-lg">
            <h2 className="mb-4 text-base font-semibold text-slate-900">新建项目</h2>
            <form onSubmit={handleCreateProject} className="space-y-3">
              <div>
                <label htmlFor="projectName" className="mb-1 block text-sm font-medium text-slate-700">
                  项目名称
                </label>
                <input
                  id="projectName"
                  type="text"
                  value={newProjectName}
                  onChange={(e) => setNewProjectName(e.target.value)}
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  placeholder="例如：电商首页原型"
                  required
                  autoFocus
                />
              </div>
              <div>
                <label htmlFor="projectDesc" className="mb-1 block text-sm font-medium text-slate-700">
                  描述（可选）
                </label>
                <textarea
                  id="projectDesc"
                  value={newProjectDesc}
                  onChange={(e) => setNewProjectDesc(e.target.value)}
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  placeholder="项目简要描述"
                  rows={2}
                />
              </div>
              <div className="flex justify-end gap-2 pt-1">
                <button
                  type="button"
                  onClick={() => { setShowCreateDialog(false); setNewProjectName(''); setNewProjectDesc(''); }}
                  className="rounded-md px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-100"
                >
                  取消
                </button>
                <button
                  type="submit"
                  disabled={isCreating || !newProjectName.trim()}
                  className="rounded-md bg-blue-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-600 disabled:opacity-50"
                >
                  {isCreating ? '创建中...' : '创建'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
