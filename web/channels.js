const Channels = (() => {
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

  return { init, load };
})();
