import React, { useEffect, useState } from 'react';
import { api } from './api';

export default function App() {
  const [authMode, setAuthMode] = useState('login');
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const [registerForm, setRegisterForm] = useState({
    name: '',
    email: '',
    password: ''
  });
  const [loginForm, setLoginForm] = useState({
    email: '',
    password: ''
  });
  const [showResetPassword, setShowResetPassword] = useState(false);
  const [resetPasswordForm, setResetPasswordForm] = useState({
    email: '',
    new_password: ''
  });
  const [profileName, setProfileName] = useState('');

  const [tasks, setTasks] = useState([]);
  const [taskForm, setTaskForm] = useState({
    title: '',
    body: ''
  });

  useEffect(() => {
    (async () => {
      try {
        const me = await api.me();
        setUser(me);
        setProfileName(me.name);
        const list = await api.listTasks();
        setTasks(list ?? []);
      } catch {
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const handleRegister = async (e) => {
    e.preventDefault();
    setError('');
    try {
      const me = await api.register(registerForm);
      setUser(me);
      setProfileName(me.name);
      const list = await api.listTasks();
      setTasks(list ?? []);
    } catch (err) {
      setError(err.message);
    }
  };

  const handleLogin = async (e) => {
    e.preventDefault();
    setError('');
    try {
      const me = await api.login(loginForm);
      setUser(me);
      setProfileName(me.name);
      const list = await api.listTasks();
      setTasks(list ?? []);
    } catch (err) {
      setError(err.message);
    }
  };

  const handleResetPassword = async (e) => {
    e.preventDefault();
    setError('');
    try {
      await api.resetPassword({
        email: resetPasswordForm.email,
        new_password: resetPasswordForm.new_password
      });
      setError('');
      setShowResetPassword(false);
      setResetPasswordForm({ email: '', new_password: '' });
      alert('Şifre güncellendi. Yeni şifrenle giriş yapabilirsin.');
    } catch (err) {
      setError(err.message);
    }
  };

  const handleLogout = async () => {
    setError('');
    try {
      await api.logout();
      setUser(null);
      setTasks([]);
    } catch (err) {
      setError(err.message);
    }
  };

  const handleProfileSave = async (e) => {
    e.preventDefault();
    setError('');
    try {
      await api.updateMe({ name: profileName });
      setUser((u) => (u ? { ...u, name: profileName } : u));
    } catch (err) {
      setError(err.message);
    }
  };

  const handleCreateTask = async (e) => {
    e.preventDefault();
    setError('');
    if (!taskForm.title || !taskForm.body) return;
    try {
      await api.createTask({
        title: taskForm.title,
        body: taskForm.body,
        status: 'pending'
      });
      setTaskForm({ title: '', body: '' });
      const list = await api.listTasks();
      setTasks(list ?? []);
    } catch (err) {
      setError(err.message);
    }
  };

  const handleToggleDone = async (task) => {
    setError('');
    const newStatus = task.status === 'done' || task.status === 'completed' ? 'pending' : 'done';
    try {
      await api.updateTaskStatus(task.id, newStatus);
      const list = await api.listTasks();
      setTasks(list ?? []);
    } catch (err) {
      setError(err.message);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-slate-600">Yükleniyor...</div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="bg-white shadow-lg rounded-xl p-8 w-full max-w-md">
          <h1 className="text-2xl font-semibold text-center mb-6">Task App</h1>

          <div className="flex mb-6 bg-slate-100 rounded-lg p-1">
            <button
              className={`flex-1 py-2 rounded-lg text-sm font-medium ${
                authMode === 'login' ? 'bg-white shadow text-slate-900' : 'text-slate-500'
              }`}
              onClick={() => setAuthMode('login')}
            >
              Giriş yap
            </button>
            <button
              className={`flex-1 py-2 rounded-lg text-sm font-medium ${
                authMode === 'register' ? 'bg-white shadow text-slate-900' : 'text-slate-500'
              }`}
              onClick={() => setAuthMode('register')}
            >
              Kayıt ol
            </button>
          </div>

          {error && (
            <div className="mb-4 text-sm text-red-600 bg-red-50 border border-red-100 rounded-lg px-3 py-2">
              {error}
            </div>
          )}

          {authMode === 'login' ? (
            <>
              {showResetPassword ? (
                <form onSubmit={handleResetPassword} className="space-y-4">
                  <p className="text-sm text-slate-600">
                    Bu email ile kayıtlı hesabın şifresini sıfırla. Sonra yeni şifrenle giriş yapabilirsin.
                  </p>
                  <div>
                    <label className="block text-sm mb-1">Email</label>
                    <input
                      type="email"
                      className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                      value={resetPasswordForm.email}
                      onChange={(e) =>
                        setResetPasswordForm({ ...resetPasswordForm, email: e.target.value })
                      }
                      required
                    />
                  </div>
                  <div>
                    <label className="block text-sm mb-1">Yeni şifre</label>
                    <input
                      type="password"
                      className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                      value={resetPasswordForm.new_password}
                      onChange={(e) =>
                        setResetPasswordForm({ ...resetPasswordForm, new_password: e.target.value })
                      }
                      required
                    />
                  </div>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={() => {
                        setShowResetPassword(false);
                        setError('');
                      }}
                      className="flex-1 py-2 rounded-lg border text-sm font-medium text-slate-700 hover:bg-slate-50"
                    >
                      İptal
                    </button>
                    <button
                      type="submit"
                      className="flex-1 bg-amber-600 hover:bg-amber-700 text-white rounded-lg py-2 text-sm font-medium"
                    >
                      Şifreyi güncelle
                    </button>
                  </div>
                </form>
              ) : (
                <form onSubmit={handleLogin} className="space-y-4">
                  <div>
                    <label className="block text-sm mb-1">Email</label>
                    <input
                      type="email"
                      className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                      value={loginForm.email}
                      onChange={(e) => setLoginForm({ ...loginForm, email: e.target.value })}
                    />
                  </div>
                  <div>
                    <label className="block text-sm mb-1">Şifre</label>
                    <input
                      type="password"
                      className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                      value={loginForm.password}
                      onChange={(e) => setLoginForm({ ...loginForm, password: e.target.value })}
                    />
                  </div>
                  <button
                    type="submit"
                    className="w-full bg-sky-600 hover:bg-sky-700 text-white rounded-lg py-2 text-sm font-medium"
                  >
                    Giriş yap
                  </button>
                  <button
                    type="button"
                    onClick={() => setShowResetPassword(true)}
                    className="w-full text-sm text-slate-500 hover:text-sky-600"
                  >
                    Şifremi unuttum
                  </button>
                </form>
              )}
            </>
          ) : (
            <form onSubmit={handleRegister} className="space-y-4">
              <div>
                <label className="block text-sm mb-1">Ad Soyad</label>
                <input
                  type="text"
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                  value={registerForm.name}
                  onChange={(e) => setRegisterForm({ ...registerForm, name: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm mb-1">Email</label>
                <input
                  type="email"
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                  value={registerForm.email}
                  onChange={(e) => setRegisterForm({ ...registerForm, email: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm mb-1">Şifre</label>
                <input
                  type="password"
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                  value={registerForm.password}
                  onChange={(e) => setRegisterForm({ ...registerForm, password: e.target.value })}
                />
              </div>
              <button
                type="submit"
                className="w-full bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg py-2 text-sm font-medium"
              >
                Kayıt ol ve giriş yap
              </button>
            </form>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-start justify-center py-10">
      <div className="w-full max-w-4xl space-y-8">
        <header className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold">Task App</h1>
            <p className="text-sm text-slate-500">Hoş geldin, {user?.name}</p>
          </div>
          <button
            onClick={handleLogout}
            className="px-3 py-1.5 rounded-lg border text-sm text-slate-700 hover:bg-slate-100"
          >
            Çıkış yap
          </button>
        </header>

        {error && (
          <div className="mb-2 text-sm text-red-600 bg-red-50 border border-red-100 rounded-lg px-3 py-2">
            {error}
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <section className="bg-white rounded-xl shadow-sm p-5 lg:col-span-1">
            <h2 className="text-sm font-semibold mb-4">Kullanıcı Bilgileri</h2>
            <form onSubmit={handleProfileSave} className="space-y-4">
              <div>
                <label className="block text-xs mb-1 text-slate-500">Ad Soyad</label>
                <input
                  type="text"
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                  value={profileName}
                  onChange={(e) => setProfileName(e.target.value)}
                />
              </div>
              <div>
                <label className="block text-xs mb-1 text-slate-500">Email</label>
                <input
                  value={user?.email}
                  disabled
                  className="w-full border rounded-lg px-3 py-2 text-sm bg-slate-50 text-slate-500"
                />
              </div>
              <button
                type="submit"
                className="w-full bg-sky-600 hover:bg-sky-700 text-white rounded-lg py-2 text-sm font-medium"
              >
                Kaydet
              </button>
            </form>
          </section>

          <section className="bg-white rounded-xl shadow-sm p-5 lg:col-span-2 space-y-5">
            <div>
              <h2 className="text-sm font-semibold mb-3">Yeni Task</h2>
              <form onSubmit={handleCreateTask} className="space-y-3">
                <input
                  type="text"
                  placeholder="Başlık"
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                  value={taskForm.title}
                  onChange={(e) => setTaskForm({ ...taskForm, title: e.target.value })}
                />
                <textarea
                  rows={3}
                  placeholder="Detay"
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                  value={taskForm.body}
                  onChange={(e) => setTaskForm({ ...taskForm, body: e.target.value })}
                />
                <div className="flex justify-end">
                  <button
                    type="submit"
                    className="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-sm font-medium"
                  >
                    Task ekle
                  </button>
                </div>
              </form>
            </div>

            <div>
              <h2 className="text-sm font-semibold mb-3">Task Listesi</h2>
              <div className="space-y-2">
                {tasks.length === 0 && (
                  <div className="text-sm text-slate-500 bg-slate-50 rounded-lg px-3 py-2">
                    Henüz task yok. Yukarıdan ekleyebilirsin.
                  </div>
                )}
                {tasks.map((t) => (
                  <div
                    key={t.id}
                    className="flex items-start justify-between border rounded-lg px-3 py-2 bg-slate-50"
                  >
                    <div>
                      <div className="flex items-center gap-2">
                        <button
                          type="button"
                          onClick={() => handleToggleDone(t)}
                          className={`w-5 h-5 rounded border flex items-center justify-center ${
                            t.status === 'done' || t.status === 'completed'
                              ? 'bg-emerald-500 border-emerald-500 text-white'
                              : 'border-slate-300 bg-white'
                          }`}
                        >
                          {t.status === 'done' || t.status === 'completed' ? '✓' : ''}
                        </button>
                        <span
                          className={`font-medium text-sm ${
                            t.status === 'done' || t.status === 'completed'
                              ? 'line-through text-slate-400'
                              : ''
                          }`}
                        >
                          {t.title}
                        </span>
                      </div>
                      <p className="text-xs text-slate-600 mt-1 whitespace-pre-wrap">{t.body}</p>
                    </div>
                    <div className="ml-3 mt-1">
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium ${
                          t.status === 'done' || t.status === 'completed'
                            ? 'bg-emerald-50 text-emerald-700 border border-emerald-100'
                            : 'bg-amber-50 text-amber-700 border border-amber-100'
                        }`}
                      >
                        {t.status === 'done' || t.status === 'completed' ? 'Done' : 'Pending'}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}