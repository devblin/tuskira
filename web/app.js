document.addEventListener('DOMContentLoaded', () => {
  const authView = document.getElementById('auth-view');
  const dashView = document.getElementById('dashboard-view');
  const loginForm = document.getElementById('login-form');
  const registerForm = document.getElementById('register-form');
  const authMsg = document.getElementById('auth-msg');
  const tabBtns = document.querySelectorAll('.sidebar button[data-tab]');
  const panels = document.querySelectorAll('.panel');

  function showAuth() {
    authView.classList.remove('hidden');
    dashView.classList.add('hidden');
  }

  function showDashboard() {
    authView.classList.add('hidden');
    dashView.classList.remove('hidden');
    switchTab('send');
  }

  window.onAuthExpired = showAuth;

  // Auth tab toggle
  document.querySelectorAll('.auth-tabs button').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.auth-tabs button').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      const target = btn.dataset.target;
      loginForm.classList.toggle('hidden', target !== 'login');
      registerForm.classList.toggle('hidden', target !== 'register');
      authMsg.className = 'msg';
      authMsg.style.display = 'none';
    });
  });

  // Login
  loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    authMsg.style.display = 'none';
    try {
      const res = await API.login(loginForm.login_email.value, loginForm.login_password.value);
      API.setToken(res.token);
      showDashboard();
    } catch (err) {
      authMsg.className = 'msg error';
      authMsg.textContent = err.message;
      authMsg.style.display = 'block';
    }
  });

  // Register
  registerForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    authMsg.style.display = 'none';
    try {
      await API.register(registerForm.reg_email.value, registerForm.reg_password.value);
      authMsg.className = 'msg success';
      authMsg.textContent = 'Registered! You can now log in.';
      authMsg.style.display = 'block';
      // Switch to login tab
      document.querySelector('.auth-tabs button[data-target="login"]').click();
    } catch (err) {
      authMsg.className = 'msg error';
      authMsg.textContent = err.message;
      authMsg.style.display = 'block';
    }
  });

  // Logout
  document.getElementById('logout-btn').addEventListener('click', () => {
    API.clearToken();
    showAuth();
  });

  // Tab switching
  function switchTab(tab) {
    tabBtns.forEach(b => b.classList.toggle('active', b.dataset.tab === tab));
    panels.forEach(p => p.classList.toggle('active', p.id === 'tab-' + tab));

    if (tab === 'sent') Notifications.loadSent();
    if (tab === 'scheduled') Notifications.loadScheduled();
    if (tab === 'templates') Templates.loadList();
  }

  tabBtns.forEach(btn => {
    btn.addEventListener('click', () => switchTab(btn.dataset.tab));
  });

  // Init modules
  Notifications.initSendForm();
  Notifications.initScheduledTable();
  Templates.initForm();

  // Check existing session
  if (API.getToken()) {
    showDashboard();
  } else {
    showAuth();
  }
});
