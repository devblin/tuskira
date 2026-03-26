const Channels = (() => {
  let _eventSource = null;

  function init() {
    document.getElementById('channel-cards').addEventListener('click', (e) => {
      const btn = e.target.closest('button');
      if (!btn) return;
      const card = btn.closest('.channel-card');
      if (!card) return;
      const channel = card.dataset.channel;

      if (btn.classList.contains('ch-save')) {
        saveChannel(channel, card);
      } else if (btn.classList.contains('ch-delete')) {
        deleteChannel(channel);
      } else if (btn.classList.contains('slack-add-ch')) {
        addSlackChannelRow(card.querySelector('.slack-channels'));
      } else if (btn.classList.contains('btn-remove')) {
        btn.parentElement.remove();
      }
    });

    // Email add provider button
    document.querySelector('.email-add-provider').addEventListener('click', () => {
      addEmailProviderBlock(document.getElementById('email-providers'));
    });

    // In-App SSE controls
    document.getElementById('inapp-connect-btn').addEventListener('click', () => {
      const connId = document.getElementById('inapp-connection-id').value;
      if (connId) connectSSE(connId);
    });

    document.getElementById('inapp-disconnect-btn').addEventListener('click', () => {
      disconnectSSE();
    });

    document.getElementById('inapp-copy-btn').addEventListener('click', () => {
      const input = document.getElementById('inapp-connection-id');
      navigator.clipboard.writeText(input.value).then(() => {
        const btn = document.getElementById('inapp-copy-btn');
        btn.textContent = 'Copied!';
        setTimeout(() => { btn.textContent = 'Copy'; }, 1500);
      });
    });

    document.getElementById('inapp-clear-feed').addEventListener('click', () => {
      document.getElementById('inapp-feed-items').innerHTML = '';
    });
  }

  async function load() {
    const msg = document.getElementById('channels-msg');
    msg.style.display = 'none';

    try {
      const list = await API.listChannelConfigs();
      if (!list) return;
      list.forEach(cfg => populateCard(cfg));
    } catch (err) {
      showMsg(msg, 'error', err.message);
    }
  }

  function populateCard(cfg) {
    const card = document.querySelector(`.channel-card[data-channel="${cfg.channel}"]`);
    if (!card) return;

    card.querySelector('.ch-enabled').checked = cfg.enabled;

    const config = cfg.config || {};

    if (cfg.channel === 'email') {
      const container = document.getElementById('email-providers');
      container.innerHTML = '';
      const providers = config.providers || [];
      providers.forEach(p => addEmailProviderBlock(container, p));
    } else if (cfg.channel === 'slack') {
      const tokenInput = card.querySelector('.ch-field[data-key="bot_token"]');
      if (tokenInput) tokenInput.value = config.bot_token || '';

      const container = card.querySelector('.slack-channels');
      container.innerHTML = '';
      if (config.channels && config.channels.length) {
        config.channels.forEach(ch => addSlackChannelRow(container, ch.id, ch.name));
      }
    } else if (cfg.channel === 'inapp') {
      const connSection = document.getElementById('inapp-connection');
      const connInput = document.getElementById('inapp-connection-id');

      if (config.connection_id) {
        connInput.value = config.connection_id;
        connSection.classList.remove('hidden');
      } else {
        connInput.value = '';
        connSection.classList.add('hidden');
      }
    }
  }

  function collectConfig(channel, card) {
    if (channel === 'email') {
      const providers = [];
      card.querySelectorAll('.email-provider-block').forEach(block => {
        const p = {};
        block.querySelectorAll('.ep-field').forEach(input => {
          p[input.dataset.key] = input.value.trim();
        });
        const tlsCheck = block.querySelector('.ep-field-bool[data-key="tls"]');
        if (tlsCheck) p.tls = tlsCheck.checked;
        providers.push(p);
      });
      return { providers };
    }

    if (channel === 'slack') {
      const config = {};
      const tokenInput = card.querySelector('.ch-field[data-key="bot_token"]');
      config.bot_token = tokenInput ? tokenInput.value.trim() : '';
      config.channels = [];
      card.querySelectorAll('.slack-ch-row').forEach(row => {
        const id = row.querySelector('.slack-ch-id').value.trim();
        const name = row.querySelector('.slack-ch-name').value.trim();
        if (id) config.channels.push({ id, name });
      });
      return config;
    }

    return {};
  }

  async function saveChannel(channel, card) {
    const msg = document.getElementById('channels-msg');
    msg.style.display = 'none';

    const data = {
      channel: channel,
      enabled: card.querySelector('.ch-enabled').checked,
      config: collectConfig(channel, card),
    };

    try {
      await API.upsertChannelConfig(data);
      showMsg(msg, 'success', channel + ' channel saved.');

      // For in-app, re-fetch to get the auto-generated connection_id
      if (channel === 'inapp') {
        const cfg = await API.getChannelConfig('inapp');
        if (cfg) populateCard(cfg);
      }
    } catch (err) {
      showMsg(msg, 'error', err.message);
    }
  }

  async function deleteChannel(channel) {
    const msg = document.getElementById('channels-msg');
    msg.style.display = 'none';

    try {
      await API.deleteChannelConfig(channel);
      showMsg(msg, 'success', channel + ' channel config deleted.');
      resetCard(channel);
    } catch (err) {
      showMsg(msg, 'error', err.message);
    }
  }

  function resetCard(channel) {
    const card = document.querySelector(`.channel-card[data-channel="${channel}"]`);
    if (!card) return;
    card.querySelector('.ch-enabled').checked = false;
    card.querySelectorAll('.ch-field').forEach(input => {
      if (input.tagName === 'SELECT') {
        input.selectedIndex = 0;
      } else {
        input.value = '';
      }
    });
    const tlsCheck = card.querySelector('.ch-field-bool[data-key="tls"]');
    if (tlsCheck) tlsCheck.checked = false;
    if (channel === 'email') document.getElementById('email-providers').innerHTML = '';
    const slackCh = card.querySelector('.slack-channels');
    if (slackCh) slackCh.innerHTML = '';

    if (channel === 'inapp') {
      disconnectSSE();
      document.getElementById('inapp-connection-id').value = '';
      document.getElementById('inapp-connection').classList.add('hidden');
    }
  }

  function connectSSE(connectionId) {
    disconnectSSE();

    const token = API.getToken();
    const url = '/api/v1/notifications/stream?connection_id=' + encodeURIComponent(connectionId);

    // EventSource doesn't support custom headers, so use fetch with streaming
    const feedItems = document.getElementById('inapp-feed-items');
    const feed = document.getElementById('inapp-feed');
    const connectBtn = document.getElementById('inapp-connect-btn');
    const disconnectBtn = document.getElementById('inapp-disconnect-btn');
    const status = document.getElementById('inapp-status');

    feed.classList.remove('hidden');
    connectBtn.classList.add('hidden');
    disconnectBtn.classList.remove('hidden');
    status.textContent = 'Connecting...';
    status.className = 'sse-status';

    const abortController = new AbortController();
    _eventSource = { abort: () => abortController.abort() };

    fetch(url, {
      headers: { 'Authorization': 'Bearer ' + token },
      signal: abortController.signal,
    }).then(res => {
      if (!res.ok) {
        throw new Error('Stream failed (' + res.status + ')');
      }

      status.textContent = 'Connected';
      status.className = 'sse-status connected';

      const reader = res.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      function read() {
        reader.read().then(({ done, value }) => {
          if (done) {
            disconnectSSE();
            return;
          }

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split('\n');
          buffer = lines.pop(); // keep incomplete line in buffer

          for (const line of lines) {
            if (line.startsWith('data: ')) {
              try {
                const msg = JSON.parse(line.slice(6));
                appendNotification(msg);
              } catch (_) {
                // ignore parse errors
              }
            }
          }

          read();
        }).catch(err => {
          if (err.name !== 'AbortError') {
            disconnectSSE();
          }
        });
      }

      read();
    }).catch(err => {
      if (err.name !== 'AbortError') {
        status.textContent = 'Failed: ' + err.message;
        status.className = 'sse-status disconnected';
        connectBtn.classList.remove('hidden');
        disconnectBtn.classList.add('hidden');
      }
    });
  }

  function disconnectSSE() {
    if (_eventSource) {
      _eventSource.abort();
      _eventSource = null;
    }

    const connectBtn = document.getElementById('inapp-connect-btn');
    const disconnectBtn = document.getElementById('inapp-disconnect-btn');
    const status = document.getElementById('inapp-status');

    if (connectBtn) connectBtn.classList.remove('hidden');
    if (disconnectBtn) disconnectBtn.classList.add('hidden');
    if (status) {
      status.textContent = 'Disconnected';
      status.className = 'sse-status disconnected';
    }
  }

  function appendNotification(msg) {
    const feedItems = document.getElementById('inapp-feed-items');
    const item = document.createElement('div');
    item.className = 'sse-item';

    const subject = msg.subject || '(no subject)';
    const body = msg.body || '';
    const time = msg.timestamp ? new Date(msg.timestamp).toLocaleTimeString() : '';

    item.innerHTML =
      '<div class="sse-item-subject">' + escHtml(subject) + '</div>' +
      (body ? '<div class="sse-item-body">' + escHtml(body) + '</div>' : '') +
      (time ? '<div class="sse-item-time">' + time + '</div>' : '');

    feedItems.prepend(item);
  }

  function escHtml(s) {
    const el = document.createElement('span');
    el.textContent = s;
    return el.innerHTML;
  }

  function addEmailProviderBlock(container, data) {
    const d = data || {};
    const provider = d.provider || 'smtp';
    const block = document.createElement('div');
    block.className = 'email-provider-block';
    block.innerHTML =
      '<div class="ep-header">' +
        '<select class="ep-field ep-provider-select" data-key="provider" style="padding:10px 12px;border:1px solid #e2e8f0;border-radius:8px;font-size:14px;font-family:inherit;background:#fff;cursor:pointer;outline:none">' +
          '<option value="smtp"' + (provider === 'smtp' ? ' selected' : '') + '>SMTP</option>' +
          '<option value="sendgrid"' + (provider === 'sendgrid' ? ' selected' : '') + '>SendGrid</option>' +
        '</select>' +
        '<button type="button" class="ep-remove">&times;</button>' +
      '</div>' +
      '<div class="ep-smtp-fields' + (provider === 'sendgrid' ? ' hidden' : '') + '">' +
        '<div class="form-group"><label>SMTP Host</label>' +
          '<input type="text" class="ep-field" data-key="host" placeholder="smtp.example.com" value="' + escAttr(d.host) + '"></div>' +
        '<div class="form-group"><label>Port</label>' +
          '<input type="text" class="ep-field" data-key="port" placeholder="587" value="' + escAttr(d.port) + '"></div>' +
        '<div class="form-group"><label>Username</label>' +
          '<input type="text" class="ep-field" data-key="username" placeholder="user@example.com" value="' + escAttr(d.username) + '"></div>' +
        '<div class="form-group"><label>Password</label>' +
          '<input type="password" class="ep-field" data-key="password" placeholder="password" value="' + escAttr(d.password) + '"></div>' +
        '<div class="toggle-row">' +
          '<input type="checkbox" class="ep-field-bool" data-key="tls"' + (d.tls ? ' checked' : '') + '>' +
          '<label>TLS</label></div>' +
      '</div>' +
      '<div class="ep-sendgrid-fields' + (provider !== 'sendgrid' ? ' hidden' : '') + '">' +
        '<div class="form-group"><label>API Key</label>' +
          '<input type="password" class="ep-field" data-key="api_key" placeholder="SG.xxxx..." value="' + escAttr(d.api_key) + '"></div>' +
      '</div>' +
      '<div class="form-group"><label>From Address</label>' +
        '<input type="text" class="ep-field" data-key="from" placeholder="noreply@example.com" value="' + escAttr(d.from) + '"></div>';

    // Provider toggle within this block
    block.querySelector('.ep-provider-select').addEventListener('change', (e) => {
      const smtp = block.querySelector('.ep-smtp-fields');
      const sg = block.querySelector('.ep-sendgrid-fields');
      if (e.target.value === 'sendgrid') {
        smtp.classList.add('hidden');
        sg.classList.remove('hidden');
      } else {
        smtp.classList.remove('hidden');
        sg.classList.add('hidden');
      }
    });

    // Remove button
    block.querySelector('.ep-remove').addEventListener('click', () => {
      block.remove();
    });

    container.appendChild(block);
  }

  function escAttr(val) {
    if (!val) return '';
    return String(val).replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;');
  }

  function addSlackChannelRow(container, id, name) {
    const row = document.createElement('div');
    row.className = 'kv-row slack-ch-row';
    row.innerHTML = `
      <input type="text" class="slack-ch-id" placeholder="Channel ID" value="${id || ''}">
      <input type="text" class="slack-ch-name" placeholder="Channel Name" value="${name || ''}">
      <button type="button" class="btn-remove">&times;</button>
    `;
    container.appendChild(row);
  }

  function showMsg(el, type, text) {
    el.className = 'msg ' + type;
    el.textContent = text;
    el.style.display = 'block';
    clearTimeout(el._t);
    el._t = setTimeout(() => { el.style.display = 'none'; }, 5000);
  }

  function getConnectionId() {
    return document.getElementById('inapp-connection-id').value || '';
  }

  return { init, load, getConnectionId };
})();
