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
      card.querySelectorAll('.ch-field').forEach(input => {
        input.value = config[input.dataset.key] || '';
      });
      const tlsCheck = card.querySelector('.ch-field-bool[data-key="tls"]');
      if (tlsCheck) tlsCheck.checked = !!config.tls;
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
      const config = {};
      card.querySelectorAll('.ch-field').forEach(input => {
        config[input.dataset.key] = input.value.trim();
      });
      const tlsCheck = card.querySelector('.ch-field-bool[data-key="tls"]');
      if (tlsCheck) config.tls = tlsCheck.checked;
      return config;
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
    card.querySelectorAll('.ch-field').forEach(input => { input.value = ''; });
    const tlsCheck = card.querySelector('.ch-field-bool[data-key="tls"]');
    if (tlsCheck) tlsCheck.checked = false;
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
