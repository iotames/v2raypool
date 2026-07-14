/* ============================================
   V2rayPool - Native UI Components
   Web Components for Modal / Toast / Confirm / Table
   ============================================ */

const VPUI = (() => {
  'use strict';

  /* ===== Toast ===== */
  let toastContainer = null;
  function ensureToastContainer() {
    if (!toastContainer) {
      toastContainer = document.createElement('div');
      toastContainer.className = 'toast-container';
      document.body.appendChild(toastContainer);
    }
    return toastContainer;
  }

  function toast(msg, type, duration) {
    type = type || 'info';
    duration = duration || 2500;
    const el = document.createElement('div');
    el.className = 'toast ' + type;
    el.textContent = msg;
    const c = ensureToastContainer();
    c.appendChild(el);
    setTimeout(() => {
      el.style.opacity = '0';
      el.style.transform = 'translateY(-10px)';
      el.style.transition = 'all 0.3s ease';
      setTimeout(() => el.remove(), 300);
    }, duration);
  }

  toast.info = (m, d) => toast(m, 'info', d);
  toast.success = (m, d) => toast(m, 'success', d);
  toast.error = (m, d) => toast(m, 'error', d);
  toast.warning = (m, d) => toast(m, 'warning', d);

  /** Set button loading state */
  function setLoading(btn, loading) {
    if (!btn) return;
    if (loading) {
      btn._origText = btn.textContent;
      btn.disabled = true;
      btn.innerHTML = '<span class="spinner" style="width:14px;height:14px;border-width:2px;margin:0"></span> ' + (btn._origText || '');
    } else {
      btn.disabled = false;
      btn.innerHTML = btn._origText || '';
    }
  }

  /* ===== Close all overlays helper ===== */
  function onKeyEscape(e) {
    if (e.key === 'Escape') {
      const overlays = document.querySelectorAll('.confirm-overlay, .prompt-overlay, .modal-overlay');
      if (overlays.length > 0) {
        const last = overlays[overlays.length - 1];
        last.remove();
      }
    }
  }

  /* ===== Alert ===== */
  function alert(msg, opts) {
    return new Promise(resolve => {
      opts = opts || {};
      const overlay = document.createElement('div');
      overlay.className = 'confirm-overlay';
      overlay.innerHTML = `
        <div class="confirm-box">
          <div class="confirm-title">${opts.title || '提示'}</div>
          <div class="confirm-body">${msg}</div>
          <div class="confirm-actions">
            <button class="btn btn-primary alert-ok-btn">确定</button>
          </div>
        </div>
      `;
      document.body.appendChild(overlay);
      const okBtn = overlay.querySelector('.alert-ok-btn');
      okBtn.focus();
      function close() { overlay.remove(); document.removeEventListener('keydown', onKeyEscape); resolve(); }
      okBtn.addEventListener('click', close);
      overlay.addEventListener('mousedown', e => { if (e.target === overlay) close(); });
      document.addEventListener('keydown', function handler(e) { if (e.key === 'Escape') { document.removeEventListener('keydown', handler); close(); } });
    });
  }

  /* ===== Confirm ===== */
  function confirm(msg, title) {
    return new Promise(resolve => {
      const overlay = document.createElement('div');
      overlay.className = 'confirm-overlay';
      overlay.innerHTML = `
        <div class="confirm-box">
          <div class="confirm-title">${title || '确认'}</div>
          <div class="confirm-body">${msg}</div>
          <div class="confirm-actions">
            <button class="btn confirm-no-btn">取消</button>
            <button class="btn btn-primary confirm-yes-btn">确定</button>
          </div>
        </div>
      `;
      document.body.appendChild(overlay);
      const yesBtn = overlay.querySelector('.confirm-yes-btn');
      const noBtn = overlay.querySelector('.confirm-no-btn');
      yesBtn.focus();
      function close(result) { overlay.remove(); resolve(result); }
      yesBtn.addEventListener('click', () => close(true));
      noBtn.addEventListener('click', () => close(false));
      overlay.addEventListener('mousedown', e => { if (e.target === overlay) close(false); });
      document.addEventListener('keydown', function handler(e) {
        if (e.key === 'Escape') { document.removeEventListener('keydown', handler); close(false); }
        if (e.key === 'Enter') { document.removeEventListener('keydown', handler); close(true); }
      });
    });
  }

  /* ===== Prompt ===== */
  function prompt(title, defaultValue) {
    return new Promise(resolve => {
      const overlay = document.createElement('div');
      overlay.className = 'prompt-overlay';
      overlay.innerHTML = `
        <div class="prompt-box">
          <div class="prompt-title">${title}</div>
          <input type="text" class="form-control prompt-input" value="${(defaultValue || '').replace(/"/g, '&quot;')}" placeholder="请输入...">
          <div class="prompt-actions">
            <button class="btn prompt-cancel-btn">取消</button>
            <button class="btn btn-primary prompt-ok-btn">确定</button>
          </div>
        </div>
      `;
      document.body.appendChild(overlay);
      const input = overlay.querySelector('.prompt-input');
      const okBtn = overlay.querySelector('.prompt-ok-btn');
      const cancelBtn = overlay.querySelector('.prompt-cancel-btn');
      input.focus();
      input.select();
      function close(result) { overlay.remove(); resolve(result); }
      okBtn.addEventListener('click', () => close(input.value));
      cancelBtn.addEventListener('click', () => close(null));
      overlay.addEventListener('mousedown', e => { if (e.target === overlay) close(null); });
      input.addEventListener('keydown', function handler(e) {
        if (e.key === 'Enter') { close(input.value); }
        if (e.key === 'Escape') { close(null); }
      });
    });
  }

  /* ===== Modal ===== */
  function openModal(opts) {
    const overlay = document.createElement('div');
    overlay.className = 'modal-overlay';
    const modal = document.createElement('div');
    modal.className = 'modal';
    if (opts.width) modal.style.width = opts.width;
    if (opts.maxWidth) modal.style.maxWidth = opts.maxWidth;

    modal.innerHTML = `
      <div class="modal-header">
        <span>${opts.title || ''}</span>
        <button class="modal-close">&times;</button>
      </div>
      <div class="modal-body"></div>
      ${opts.footer !== false ? '<div class="modal-footer"></div>' : ''}
    `;

    overlay.appendChild(modal);
    document.body.appendChild(overlay);

    const bodyEl = modal.querySelector('.modal-body');
    const footerEl = modal.querySelector('.modal-footer');
    const closeBtn = modal.querySelector('.modal-close');

    if (typeof opts.content === 'string') {
      bodyEl.innerHTML = opts.content;
    } else if (opts.content instanceof HTMLElement) {
      bodyEl.appendChild(opts.content);
    }

    if (opts.buttons && footerEl) {
      opts.buttons.forEach(b => {
        const btn = document.createElement('button');
        btn.className = 'btn' + (b.primary ? ' btn-primary' : '');
        btn.textContent = b.text || '按钮';
        btn.addEventListener('click', () => {
          const ret = b.action && b.action(modal);
          if (ret !== false) close();
        });
        footerEl.appendChild(btn);
      });
    }

    function close() { overlay.remove(); opts.onClose && opts.onClose(); }
    closeBtn.addEventListener('click', close);
    overlay.addEventListener('mousedown', e => { if (e.target === overlay) close(); });
    document.addEventListener('keydown', function handler(e) { if (e.key === 'Escape') { document.removeEventListener('keydown', handler); close(); } });

    if (opts.onOpen) opts.onOpen(modal, bodyEl, footerEl);

    return { overlay, modal, bodyEl, footerEl, close };
  }

  /* ===== Native Table Renderer (unused currently, kept for future use) ===== */
  function renderTable(container, options) {
    const { url, columns, onRowClick } = options;
    const wrap = document.createElement('div');
    wrap.className = 'table-wrap';
    const table = document.createElement('table');
    table.className = 'vp-table';
    wrap.appendChild(table);
    container.appendChild(wrap);

    const loadingMsg = document.createElement('div');
    loadingMsg.className = 'table-loading';
    loadingMsg.innerHTML = '<span class="spinner"></span>加载中...';
    container.appendChild(loadingMsg);

    function buildHeader() {
      const thead = document.createElement('thead');
      const tr = document.createElement('tr');
      columns.forEach(col => {
        const th = document.createElement('th');
        th.textContent = col.title || '';
        if (col.width) th.style.width = col.width;
        tr.appendChild(th);
      });
      thead.appendChild(tr);
      return thead;
    }

    function render(data) {
      loadingMsg.style.display = 'none';
      table.innerHTML = '';
      table.appendChild(buildHeader());
      const tbody = document.createElement('tbody');
      if (!data || data.length === 0) {
        const tr = document.createElement('tr');
        const td = document.createElement('td');
        td.colSpan = columns.length;
        td.className = 'table-empty';
        td.textContent = '暂无数据';
        tr.appendChild(td);
        tbody.appendChild(tr);
      } else {
        data.forEach((row, idx) => {
          const tr = document.createElement('tr');
          columns.forEach(col => {
            const td = document.createElement('td');
            if (col.type === 'numbers') {
              td.className = 'col-num';
              td.textContent = idx + 1;
            } else if (col.template && typeof col.template === 'function') {
              td.innerHTML = col.template(row);
            } else if (col.field) {
              td.textContent = row[col.field] !== undefined && row[col.field] !== null ? row[col.field] : '-';
            }
            tr.appendChild(td);
          });
          if (onRowClick) tr.addEventListener('click', () => onRowClick(row));
          tbody.appendChild(tr);
        });
      }
      table.appendChild(tbody);
    }

    if (url) {
      fetch(url)
        .then(res => res.json())
        .then(dt => {
          render(dt.data && Array.isArray(dt.data) ? dt.data : (dt.data && dt.data.list ? dt.data.list : []));
        })
        .catch(() => {
          loadingMsg.style.display = 'none';
          loadingMsg.innerHTML = '加载失败';
        });
    }
    return { render, table, wrap };
  }

  /* ===== Tooltip ===== */
  function showTips(msg, el) {
    const tip = document.createElement('div');
    tip.style.cssText = 'position:fixed;background:#333;color:#fff;padding:6px 12px;border-radius:4px;font-size:12px;z-index:99999;pointer-events:none;white-space:nowrap;max-width:300px;box-shadow:0 2px 8px rgba(0,0,0,0.2)';
    tip.textContent = msg;
    document.body.appendChild(tip);
    const rect = el.getBoundingClientRect();
    let left = rect.left + rect.width / 2 - tip.offsetWidth / 2;
    let top = rect.top - tip.offsetHeight - 8;
    if (left < 4) left = 4;
    if (top < 4) top = rect.bottom + 8;
    tip.style.left = left + 'px';
    tip.style.top = top + 'px';
    setTimeout(() => { tip.style.opacity = '0'; tip.style.transition = 'opacity 0.2s'; setTimeout(() => tip.remove(), 200); }, 2000);
  }

  return {
    toast, alert, confirm, prompt,
    openModal, renderTable, showTips,
    setLoading
  };
})();
