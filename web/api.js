const API = (() => {
  const TOKEN_KEY = 'tuskira_token';

  function getToken() {
    return localStorage.getItem(TOKEN_KEY);
  }

  function setToken(token) {
    localStorage.setItem(TOKEN_KEY, token);
  }

  function clearToken() {
    localStorage.removeItem(TOKEN_KEY);
  }

  async function apiFetch(path, options = {}) {
    const headers = options.headers || {};
    const token = getToken();
    if (token) {
      headers['Authorization'] = 'Bearer ' + token;
    }
    if (options.body && typeof options.body === 'string') {
      headers['Content-Type'] = 'application/json';
    }
    const res = await fetch(path, { ...options, headers });
    if (res.status === 401) {
      clearToken();
      if (window.onAuthExpired) window.onAuthExpired();
      throw new Error('Session expired');
    }
    const data = await res.json().catch(() => null);
    if (!res.ok) {
      throw new Error(data?.error || `Request failed (${res.status})`);
    }
    return data;
  }

  return {
    getToken,
    setToken,
    clearToken,

    register(email, password) {
      return apiFetch('/api/v1/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });
    },

    login(email, password) {
      return apiFetch('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });
    },

    sendNotification(data) {
      return apiFetch('/api/v1/notifications', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    },

    listScheduled() {
      return apiFetch('/api/v1/notifications/scheduled');
    },

    getNotification(id) {
      return apiFetch('/api/v1/notifications/' + id);
    },

    triggerSend(id) {
      return apiFetch('/api/v1/notifications/' + id + '/send', { method: 'POST' });
    },

    reschedule(id, scheduleAt) {
      return apiFetch('/api/v1/notifications/' + id + '/schedule', {
        method: 'PATCH',
        body: JSON.stringify({ schedule_at: scheduleAt }),
      });
    },

    cancelNotification(id) {
      return apiFetch('/api/v1/notifications/' + id + '/cancel', { method: 'POST' });
    },

    listSent() {
      return apiFetch('/api/v1/notifications/sent');
    },

    listByRecipient(recipient) {
      return apiFetch('/api/v1/notifications?recipient=' + encodeURIComponent(recipient));
    },

    createTemplate(data) {
      return apiFetch('/api/v1/templates', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    },

    listTemplates() {
      return apiFetch('/api/v1/templates');
    },
  };
})();
