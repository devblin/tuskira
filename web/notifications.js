const Notifications = (() => {
  function initSendForm() {
    const form = document.getElementById('send-form');
    const msg = document.getElementById('send-msg');
    const useTemplate = document.getElementById('use-template');
    const templateSection = document.getElementById('template-section');
    const scheduleToggle = document.getElementById('schedule-toggle');
    const scheduleSection = document.getElementById('schedule-section');

    useTemplate.addEventListener('change', () => {
      templateSection.classList.toggle('hidden', !useTemplate.checked);
      if (useTemplate.checked) Templates.populateDropdown();
    });

    scheduleToggle.addEventListener('change', () => {
      scheduleSection.classList.toggle('hidden', !scheduleToggle.checked);
    });

    // Auto-fill recipient for in-app channel
    const channelSelect = document.getElementById('channel');
    const recipientInput = document.getElementById('recipient');

    channelSelect.addEventListener('change', async () => {
      if (channelSelect.value === 'inapp') {
        // Try to get connection_id from already-loaded channels card first
        let connId = Channels.getConnectionId();
        if (!connId) {
          try {
            const cfg = await API.getChannelConfig('inapp');
            if (cfg && cfg.config && cfg.config.connection_id) {
              connId = cfg.config.connection_id;
            }
          } catch (_) {}
        }
        if (connId) {
          recipientInput.value = connId;
          recipientInput.readOnly = true;
        }
      } else {
        if (recipientInput.readOnly) {
          recipientInput.value = '';
          recipientInput.readOnly = false;
        }
      }
    });

    document.getElementById('template-select').addEventListener('change', (e) => {
      const container = document.getElementById('kv-container');
      container.innerHTML = '';
      const tmplId = parseInt(e.target.value);
      if (!tmplId) return;
      const tmpl = Templates.getTemplateById(tmplId);
      if (tmpl && tmpl.variables && tmpl.variables.length) {
        tmpl.variables.forEach(v => addKVRow(v));
      }
    });

    document.getElementById('add-kv').addEventListener('click', () => {
      addKVRow();
    });

    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      msg.className = 'msg';
      msg.style.display = 'none';

      const data = {
        recipient: form.recipient.value,
        channel: form.channel.value,
        subject: form.subject.value,
        body: form.body.value,
      };

      if (useTemplate.checked) {
        const tmplId = parseInt(document.getElementById('template-select').value);
        if (tmplId) data.template_id = tmplId;

        const kvData = {};
        document.querySelectorAll('#kv-container .kv-row').forEach(row => {
          const key = row.querySelector('.kv-key').value.trim();
          const val = row.querySelector('.kv-val').value.trim();
          if (key) kvData[key] = val;
        });
        if (Object.keys(kvData).length > 0) data.template_data = kvData;
      }

      if (scheduleToggle.checked) {
        const dt = document.getElementById('schedule-at').value;
        if (dt) data.schedule_at = new Date(dt).toISOString();
      }

      try {
        await API.sendNotification(data);
        showMsg(msg, 'success', scheduleToggle.checked ? 'Notification scheduled!' : 'Notification sent!');
        form.reset();
        templateSection.classList.add('hidden');
        scheduleSection.classList.add('hidden');
      } catch (err) {
        showMsg(msg, 'error', err.message);
      }
    });
  }

  function addKVRow(key) {
    const container = document.getElementById('kv-container');
    const row = document.createElement('div');
    row.className = 'kv-row';
    row.innerHTML = `
      <input type="text" class="kv-key" placeholder="Key" value="${key || ''}">
      <input type="text" class="kv-val" placeholder="Value">
      <button type="button" class="btn-remove" onclick="this.parentElement.remove()">&times;</button>
    `;
    container.appendChild(row);
  }

  async function loadScheduled() {
    const tbody = document.getElementById('scheduled-tbody');
    const msg = document.getElementById('scheduled-msg');
    tbody.innerHTML = '<tr><td colspan="7">Loading...</td></tr>';
    msg.className = 'msg';
    msg.style.display = 'none';

    try {
      const list = await API.listScheduled();
      if (!list || list.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" style="text-align:center;color:var(--text-muted)">No scheduled notifications</td></tr>';
        return;
      }
      tbody.innerHTML = list.map(n => `
        <tr data-id="${n.ID}">
          <td>${n.ID}</td>
          <td>${esc(n.recipient)}</td>
          <td>${esc(n.channel)}</td>
          <td>${esc(n.subject || '-')}</td>
          <td class="schedule-cell">${formatDate(n.schedule_at)}</td>
          <td><span class="badge badge-${n.status}">${n.status}</span></td>
          <td class="actions">
            <button class="btn btn-sm btn-success act-send">Send Now</button>
            <button class="btn btn-sm btn-warning act-reschedule">Reschedule</button>
            <button class="btn btn-sm btn-danger act-cancel">Cancel</button>
          </td>
        </tr>
      `).join('');
    } catch (err) {
      tbody.innerHTML = '';
      showMsg(msg, 'error', err.message);
    }
  }

  function initScheduledTable() {
    const table = document.getElementById('scheduled-table');
    const msg = document.getElementById('scheduled-msg');

    table.addEventListener('click', async (e) => {
      const btn = e.target.closest('button');
      if (!btn) return;
      const row = btn.closest('tr');
      const id = row.dataset.id;

      if (btn.classList.contains('act-send')) {
        try {
          await API.triggerSend(id);
          showMsg(msg, 'success', 'Notification sent!');
          loadScheduled();
        } catch (err) {
          showMsg(msg, 'error', err.message);
        }
      } else if (btn.classList.contains('act-cancel')) {
        try {
          await API.cancelNotification(id);
          showMsg(msg, 'success', 'Notification cancelled.');
          loadScheduled();
        } catch (err) {
          showMsg(msg, 'error', err.message);
        }
      } else if (btn.classList.contains('act-reschedule')) {
        const cell = row.querySelector('.schedule-cell');
        cell.innerHTML = `
          <div class="reschedule-inline">
            <input type="datetime-local" class="reschedule-input">
            <button class="btn btn-sm btn-primary reschedule-save">Save</button>
          </div>
        `;
      } else if (btn.classList.contains('reschedule-save')) {
        const input = row.querySelector('.reschedule-input');
        if (!input.value) return;
        try {
          await API.reschedule(id, new Date(input.value).toISOString());
          showMsg(msg, 'success', 'Notification rescheduled.');
          loadScheduled();
        } catch (err) {
          showMsg(msg, 'error', err.message);
        }
      }
    });
  }

  function formatDate(d) {
    if (!d) return '-';
    return new Date(d).toLocaleString();
  }

  function esc(s) {
    const el = document.createElement('span');
    el.textContent = s;
    return el.innerHTML;
  }

  function showMsg(el, type, text) {
    el.className = 'msg ' + type;
    el.textContent = text;
    el.style.display = 'block';
    clearTimeout(el._t);
    el._t = setTimeout(() => { el.style.display = 'none'; }, 5000);
  }

  async function loadSent() {
    const tbody = document.getElementById('sent-tbody');
    const msg = document.getElementById('sent-msg');
    msg.className = 'msg';
    msg.style.display = 'none';
    tbody.innerHTML = '<tr><td colspan="6">Loading...</td></tr>';

    try {
      const list = await API.listSent();
      if (!list || list.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;color:var(--text-muted)">No sent/failed notifications</td></tr>';
        return;
      }
      tbody.innerHTML = list.map(n => `
        <tr>
          <td>${n.ID}</td>
          <td>${esc(n.recipient)}</td>
          <td>${esc(n.channel)}</td>
          <td>${esc(n.subject || '-')}</td>
          <td><span class="badge badge-${n.status}">${n.status}</span></td>
          <td>${formatDate(n.sent_at)}</td>
        </tr>
      `).join('');
    } catch (err) {
      tbody.innerHTML = '';
      showMsg(msg, 'error', err.message);
    }
  }

  return { initSendForm, loadScheduled, initScheduledTable, loadSent };
})();
