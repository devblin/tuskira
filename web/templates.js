const Templates = (() => {
  function initForm() {
    const form = document.getElementById('template-form');
    const msg = document.getElementById('template-msg');

    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      msg.className = 'msg';
      msg.style.display = 'none';

      const data = {
        name: form.tmpl_name.value,
        channel: form.tmpl_channel.value,
        subject: form.tmpl_subject.value,
        body: form.tmpl_body.value,
      };

      try {
        await API.createTemplate(data);
        showMsg(msg, 'success', 'Template created!');
        form.reset();
        loadList();
      } catch (err) {
        showMsg(msg, 'error', err.message);
      }
    });
  }

  async function loadList() {
    const container = document.getElementById('template-list');
    container.innerHTML = '<p style="color:var(--text-muted)">Loading...</p>';

    try {
      const list = await API.listTemplates();
      if (!list || list.length === 0) {
        container.innerHTML = '<p style="color:var(--text-muted)">No templates yet.</p>';
        return;
      }
      container.innerHTML = list.map(t => `
        <div class="template-card">
          <h4>${esc(t.name)}</h4>
          <div class="meta">${esc(t.channel)}${t.subject ? ' &middot; ' + esc(t.subject) : ''}</div>
          <div class="body-preview">${esc(t.body)}</div>
          ${t.variables && t.variables.length ? '<div class="meta" style="margin-top:6px">Variables: ' + t.variables.map(v => '<code>' + esc(v) + '</code>').join(', ') + '</div>' : ''}
        </div>
      `).join('');
    } catch (err) {
      container.innerHTML = '<p style="color:var(--danger)">' + esc(err.message) + '</p>';
    }
  }

  async function populateDropdown() {
    const select = document.getElementById('template-select');
    if (select.options.length > 1) return;

    try {
      const list = await API.listTemplates();
      if (!list) return;
      list.forEach(t => {
        const opt = document.createElement('option');
        opt.value = t.ID;
        opt.textContent = t.name + ' (' + t.channel + ')';
        select.appendChild(opt);
      });
    } catch (err) {
      // ignore
    }
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

  return { initForm, loadList, populateDropdown };
})();
