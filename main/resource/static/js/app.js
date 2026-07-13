/* ============================================
   V2rayPool - Main Application Logic
   Pure native JS (replaces layui main.js)
   ============================================ */

(function() {
  'use strict';

  /* ===== Global State ===== */
  let VPTestUrl = window.__VP_TEST_URL__ || 'https://www.google.com/';
  let VPTestedDomainList = window.__VP_TESTED_DOMAINS__ || [];
  const VP_DEFAULT_MAX_DELAY = 230;

  /* ===== Delay Threshold Click Handler ===== */
  function setupDelayClick(el) {
    if (!el) return;
    el.style.cursor = 'pointer';
    el.style.textDecoration = 'underline dashed';
    el.style.fontSize = '16px';
    el.style.fontWeight = 'bold';
    el.title = '点击修改延迟阈值';
    el.onclick = async function() {
      const oldVal = parseInt(el.textContent) || VP_DEFAULT_MAX_DELAY;
      const val = await VPUI.prompt('修改隧道延迟阈值(ms)', oldVal);
      if (val === null) return;
      const n = parseInt(val);
      if (isNaN(n) || n < 10) { VPUI.toast.warning('阈值必须为≥10的整数'); return; }
      const btn = this;
      API.updateSetting({ VP_TUNNEL_MAX_DELAY: n + '' }, dt2 => {
        VPUI.toast.success('阈值已更新为 ' + n + 'ms');
        refreshStatusBar();
      }, dt2 => VPUI.toast.error(dt2.msg || '更新失败'));
    };
  }

  /* ===== Tab System ===== */
  function initTabs() {
    const header = document.getElementById('tabsHeader');
    const contents = document.querySelectorAll('.tab-content');
    if (!header) return;

    header.addEventListener('click', e => {
      const btn = e.target.closest('.tab-btn');
      if (!btn) return;
      const idx = Array.from(header.children).indexOf(btn);
      header.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      contents.forEach((c, i) => c.classList.toggle('active', i === idx));
      if (idx === 1) refreshV2rayTable();
    });
  }

  /* ===== Status Bar ===== */
  function refreshStatusBar() {
    // SysProxy status
    API.sysProxyStatus(dt => {
      const names = {0: '关闭', 1: '单节点', 2: '隧道'};
      const el = document.getElementById('statusSysProxy');
      const btn = document.getElementById('sysProxyBtn');
      if (dt.data) {
        el.textContent = names[dt.data.type] || '未知';
        el.style.color = dt.data.type === 0 ? '#999' : '#5FB878';
        if (btn) {
          btn.textContent = '系统代理: ' + (names[dt.data.type] || '未知');
          btn.className = dt.data.type === 0 ? 'btn btn-sm' : 'btn btn-sm btn-cyan';
        }
      }
    }, function() {
      // 静默失败 — 后台轮询不影响用户
      document.getElementById('statusSysProxy') && (document.getElementById('statusSysProxy').textContent = '查询失败');
    });

    // Nodes count
    API.getNodes('', dt => {
      const el = document.getElementById('statusRunCount');
      if (dt.data) {
        let run = 0, total = dt.data.length;
        dt.data.forEach(n => { if (n.is_running) run++; });
        el.textContent = run + '/' + total;
      }
    }, function() {
      document.getElementById('statusRunCount') && (document.getElementById('statusRunCount').textContent = '失败');
    });

    // Tunnel status
    API.tunnelStatus(dt => {
      const btn = document.getElementById('tunnelBtn');
      const stateEl = document.getElementById('statusTunnelState');
      const nodesEl = document.getElementById('statusTunnelNodes');
      const portEl = document.getElementById('statusTunnelPort');
      const delayEl = document.getElementById('statusTunnelMaxDelay');

      // 延迟阈值始终可点击修改，不管隧道是否运行
      setupDelayClick(delayEl);

      if (dt.data && dt.data.running) {
        stateEl.textContent = '运行中'; stateEl.style.color = '#5FB878';
        nodesEl.textContent = dt.data.node_count || 0;
        portEl.textContent = dt.data.port || '';
        delayEl.textContent = dt.data.max_delay_ms || VP_DEFAULT_MAX_DELAY;
        if (btn) { btn.textContent = '隧道池: 开'; btn.className = 'btn btn-sm btn-orange'; }
      } else {
        stateEl.textContent = '未启动'; stateEl.style.color = '#999';
        nodesEl.textContent = '0';
        portEl.textContent = '-';
        delayEl.textContent = VP_DEFAULT_MAX_DELAY + '';
        if (btn) { btn.textContent = '隧道池: 关'; btn.className = 'btn btn-sm btn-success'; }
      }
    }, function() {
      const stateEl = document.getElementById('statusTunnelState');
      if (stateEl) { stateEl.textContent = '状态未知'; stateEl.style.color = '#FF5722'; }
    });
  }

  /* ===== Nodes Table ===== */
  function renderNodesTable(domain) {
    const container = document.getElementById('nodesTableBody');
    if (!container) return;
    container.innerHTML = '<div class="table-loading"><span class="spinner"></span>加载中...</div>';

    API.getNodes(domain, function(dt) {
      const data = dt.data;
      if (!data || data.length === 0) {
        container.innerHTML = '<div class="table-empty">' +
          '<div style="font-size:40px;margin-bottom:8px;opacity:0.3">📡</div>' +
          '<div style="font-size:14px;font-weight:600;margin-bottom:4px">暂无节点</div>' +
          '<div style="font-size:12px;color:#999">请先通过「订阅」或「添加v2ray」添加代理节点</div>' +
          '</div>';
        return;
      }
        let html = '<div class="table-wrap"><table class="vp-table"><thead><tr>' +
          '<th style="width:40px">#</th>' +
          '<th>协议</th>' +
          '<th>本机地址</th>' +
          '<th>速度/秒</th>' +
          '<th>标题</th>' +
          '<th>节点地址</th>' +
          '<th>状态</th>' +
          '<th>测速时间</th>' +
          '<th style="width:120px">操作</th>' +
          '</tr></thead><tbody>';

        data.forEach((row, idx) => {
          const tested = row.test_at && row.test_at !== '0001-01-01 00:00';
          const speedDisplay = tested
            ? '<span class="badge badge-green" onclick="VPUI.showTips(\'' + (row.test_url || '').replace(/'/g, "\\'") + '\', this)">' + (row.speed >= 5 ? 100 : (row.speed || '-')) + ' s</span>'
            : '<span class="badge">- s</span>';
          const statusDisplay = row.is_running
            ? '<span class="badge badge-green">运行中</span>'
            : '<span class="badge">已停止</span>';
          const activeBtn = row.is_active
            ? '<button class="btn btn-sm btn-purple" onclick="handleActive(' + idx + ')">已启用</button>'
            : '<button class="btn btn-sm btn-primary" onclick="handleActive(' + idx + ')">启用</button>';

          html += '<tr>' +
            '<td class="col-num">' + (idx + 1) + '</td>' +
            '<td>' + escHtml(row.protocol || '-') + '</td>' +
            '<td>' + escHtml(row.local_addr || '-') + '</td>' +
            '<td>' + speedDisplay + '</td>' +
            '<td>' + escHtml(row.title || '-') + '</td>' +
            '<td>' + escHtml(row.remote_addr || '-') + '</td>' +
            '<td>' + statusDisplay + '</td>' +
            '<td style="font-size:12px;color:#999">' + (tested ? row.test_at : '-') + '</td>' +
            '<td><span class="vp-table-actions">' + activeBtn +
            '<button class="btn btn-sm btn-danger" onclick="handleDelete(' + idx + ')">删除</button></span></td>' +
            '</tr>';
        });

        html += '</tbody></table></div>';
        html += '<div class="text-muted" style="margin-top:6px;font-size:12px">共 ' + data.length + ' 个节点</div>';
        container.innerHTML = html;
    }, function() {
      container.innerHTML = '<div class="table-empty">' +
        '<div style="font-size:40px;margin-bottom:8px;opacity:0.3">⚠️</div>' +
        '<div>加载失败，请检查服务是否正常运行</div>' +
        '</div>';
    });
  }

  function escHtml(s) {
    if (!s) return '-';
    const d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
  }

  /* ===== Routing Rules: Add item helper ===== */
  window.addRoutingItem = function(btn) {
    const list = btn.parentElement.querySelector('.routing-list');
    if (!list) return;
    const name = list.getAttribute('data-name');
    if (!name) return;
    const item = document.createElement('div');
    item.className = 'routing-item';
    item.innerHTML = '<input type="text" name="' + name + '" class="form-control routing-input" placeholder="输入新值" value="">' +
      '<button class="btn btn-sm btn-danger routing-item-del" onclick="this.parentElement.remove()" tabindex="-1">×</button>';
    list.appendChild(item);
    item.querySelector('input').focus();
  };

  /* ===== Nodes Table Actions (exposed globally) ===== */
  window.handleActive = async function(idx) {
    const dt = await getNodesData();
    if (!dt || !dt.data || !dt.data[idx]) return;
    const node = dt.data[idx];
    if (node.is_active) {
      const ok = await VPUI.confirm('取消[' + node.title + ']节点为系统代理吗？');
      if (!ok) return;
      API.unactiveNode(node.remote_addr, d => {
        VPUI.toast.success('已取消');
        renderNodesTable();
        refreshStatusBar();
      }, d => VPUI.toast.error(d.msg || '操作失败'));
    } else {
      const modal = VPUI.openModal({
        title: '激活[' + node.title + ']节点为系统代理吗？',
        width: '360px',
        content: `
          <div class="form">
            <div class="form-row">
              <label class="form-label-inline" style="width:80px">分流模式</label>
              <div class="radio-group">
                <label><input type="radio" name="active_global_proxy" value="1" checked> 全局代理</label>
                <label><input type="radio" name="active_global_proxy" value="0"> 路由分流</label>
              </div>
            </div>
          </div>
        `,
        buttons: [
          { text: '取消', action: () => {} },
          { text: '启用', primary: true, action: (m) => {
            const val = m.querySelector('input[name="active_global_proxy"]:checked');
            const globalProxy = val ? val.value !== '0' : true;
            API.activeNode(node.remote_addr, globalProxy, d => {
              VPUI.toast.success('启用成功');
              setTimeout(() => { renderNodesTable(); refreshStatusBar(); }, 500);
            }, d => VPUI.toast.error(d.msg || '启用失败'));
          }}
        ]
      });
    }
  };

  window.handleDelete = async function(idx) {
    const dt = await getNodesData();
    if (!dt || !dt.data || !dt.data[idx]) return;
    const node = dt.data[idx];
    const ok = await VPUI.confirm('确定删除节点【' + node.title + '】？');
    if (!ok) return;
    API.deleteNode(idx, d => {
      VPUI.toast.success('已删除');
      renderNodesTable();
      refreshStatusBar();
    }, d => VPUI.toast.error(d.msg || '删除失败'));
  };

  function getNodesData() {
    return fetch('/api/nodes').then(r => r.json()).catch(() => null);
  }

  /* ===== Toolbar Events ===== */
  function initToolbar() {
    document.getElementById('toolbar')?.addEventListener('click', e => {
      const btn = e.target.closest('[data-action]');
      if (!btn) return;
      const action = btn.dataset.action;
      if (typeof toolbarActions[action] === 'function') toolbarActions[action](btn);
    });
  }

  const toolbarActions = {
    testNodes: function(btn) {
      let html = '<div class="form">';
      html += '<div class="form-group"><label class="form-label">测速地址</label><input type="text" id="testUrlInput" class="form-control" value="' + VPTestUrl.replace(/"/g, '&quot;') + '"></div>';
      if (VPTestedDomainList.length > 0) {
        html += '<div class="form-group"><label class="form-label">历史测速</label><div class="flex flex-wrap gap-8">';
        VPTestedDomainList.forEach(d => {
          html += '<span class="badge badge-blue badge-clickable" onclick="var v=this.innerText;document.getElementById(\'testUrlInput\').value=v.indexOf(\'http\')===0?v:\'https://\'+v">' + escHtml(d) + '</span>';
        });
        html += '</div></div>';
      }
      html += '</div>';

      VPUI.openModal({
        title: '节点测速',
        width: '500px',
        content: html,
        buttons: [
          { text: '取消', action: () => {} },
          { text: '开始测速', primary: true, action: (m) => {
            const input = m.querySelector('#testUrlInput');
            if (!input || !input.value) { VPUI.toast.warning('请输入测速地址'); return false; }
            API.testNodes(input.value, d => {
              VPUI.toast.success(d.msg || '测速已开始');
              // 轮询刷新表格显示测速结果
              let count = 12;
              const iv = setInterval(() => {
                renderNodesTable();
                refreshStatusBar();
                count--;
                if (count <= 0) clearInterval(iv);
              }, 2000);
            }, d => VPUI.toast.error(d.msg || '测速失败'));
          }}
        ]
      });
    },

    startNodes: function(btn) {
      VPUI.setLoading(btn, true);
      API.startNodes(d => {
        VPUI.toast.success(d.msg || '启动成功');
        setTimeout(() => { VPUI.setLoading(btn, false); renderNodesTable(); refreshStatusBar(); }, 2000);
      }, d => {
        VPUI.setLoading(btn, false);
        VPUI.toast.error(d.msg || '启动失败');
      });
    },

    subscribe: function() {
      const modal = VPUI.openModal({
        title: '更新订阅',
        width: '60%',
        maxWidth: '600px',
        content: document.getElementById('subscribeFormContent')?.innerHTML || '加载中...',
        buttons: [
          { text: '取消', action: () => {} },
          { text: '更新订阅', primary: true, action: (m) => {
            const proxyVal = m.querySelector('#subscribeProxyInput')?.value || '';
            const btn = m.querySelector('.btn-primary');
            API.subscribeNodes(proxyVal, d => {
              VPUI.toast.success(d.msg || '更新成功');
              modal.close();
              renderNodesTable();
              refreshStatusBar();
            }, d => { VPUI.toast.error(d.msg || '更新失败'); });
          }}
        ],
        onOpen: (m, body) => {
          // Bind enter key
          const input = m.querySelector('#subscribeProxyInput');
          if (input) input.addEventListener('keydown', e => { if (e.key === 'Enter') m.querySelector('.btn-primary')?.click(); });
        }
      });
    },

    routingrules: function() {
      const modal = VPUI.openModal({
        title: '路由分流规则 (配置文件: default.env -> .env)',
        width: '80%',
        maxWidth: '800px',
        content: document.getElementById('routingRulesFormContent')?.innerHTML || '加载中...',
        buttons: [
          { text: '取消', action: () => {} },
          { text: '提交', primary: true, action: (m) => {
            const data = {};
            // 路由规则：收集每个列表下的所有输入框值
            ['direct_domain_list', 'direct_ip_list', 'proxy_domain_list', 'proxy_ip_list'].forEach(name => {
              const inputs = m.querySelectorAll(`[name="${name}"]`);
              data[name] = [];
              inputs.forEach(el => {
                const v = el.value.trim();
                if (v) data[name].push(v);
              });
            });
            API.updateRoutingRules(data, d => {
              VPUI.toast.success(d.msg || '更新成功');
              modal.close();
              renderNodesTable();
              refreshStatusBar();
            }, d => VPUI.toast.error(d.msg || '更新失败'));
          }}
        ]
      });
    },

    tunnelToggle: function() {
      API.tunnelStatus(dt => {
        if (dt.data && dt.data.running) {
          API.tunnelStop(d => { VPUI.toast.success('隧道代理池已关闭'); refreshStatusBar(); }, d => VPUI.toast.error(d.msg || '关闭失败'));
        } else {
          API.tunnelStart(d => {
            VPUI.toast.success('隧道代理池已启动');
            refreshStatusBar();
            setTimeout(refreshStatusBar, 5000);
          }, d => VPUI.toast.error(d.msg || '启动失败'));
        }
      }, d => VPUI.toast.error('获取隧道状态失败'));
    },

    sysProxyOpen: function() {
      API.sysProxyStatus(dt => {
        const cur = (dt.data && dt.data.type) || 0;
        const next = cur === 0 ? 2 : 0;
        API.sysProxySwitch(next, -1, d => {
          VPUI.toast.success(d.msg || '切换成功');
          refreshStatusBar();
        }, d => VPUI.toast.error(d.msg || '切换失败'));
      }, d => VPUI.toast.error('获取代理状态失败'));
    },

    settinglayer: function() {
      const modal = VPUI.openModal({
        title: '系统设置(' + (window.__VP_ENV_FILE__ || '.env') + '文件)',
        width: '50%',
        maxWidth: '600px',
        content: document.getElementById('settingFormContent')?.innerHTML || '加载中...',
        buttons: [
          { text: '取消', action: () => {} },
          { text: '提交', primary: true, action: (m) => {
            const data = {};
            m.querySelectorAll('[name]').forEach(el => { data[el.name] = el.value; });
            const btn = m.querySelector('.btn-primary');
            API.updateSetting(data, d => {
              VPUI.toast.success(d.msg || '保存成功');
              VPUI.alert('设置已保存，3秒后刷新页面', {title: '提示'}).then(() => { window.location.reload(); });
              modal.close();
            }, d => { VPUI.toast.error(d.msg || '保存失败'); });
          }}
        ]
      });
    },

    clearlayer: async function() {
      const ok = await VPUI.confirm('删除runtime目录内所有文件：包括测速缓存等数据。', '确定清除应用缓存？');
      if (!ok) return;
      API.clearCache(d => {
        VPUI.toast.success('已清除');
        setTimeout(() => { window.location.reload(); }, 1000);
      }, d => VPUI.toast.error(d.msg || '清除失败'));
    }
  };

  /* ===== Domain Select ===== */
  function initDomainSelect() {
    const sel = document.getElementById('domainSelect');
    if (!sel) return;
    // Populate from server data
    VPTestedDomainList.forEach(d => {
      const opt = document.createElement('option');
      opt.value = d;
      opt.textContent = d;
      if (VPTestUrl.indexOf(d) !== -1) opt.selected = true;
      sel.appendChild(opt);
    });
    sel.addEventListener('change', function() {
      const domain = this.value;
      if (!domain) { renderNodesTable(); return; }
      const list = VPTestedDomainList || [];
      if (list.indexOf(domain) === -1) { VPUI.toast.warning('该域名未测试过，请先测速'); return; }
      const testUrl = 'https://' + domain + '/';
      VPTestUrl = testUrl;
      API.setTestUrl(testUrl, d => {
        VPUI.toast.success('测速地址已切换到: ' + domain);
        document.getElementById('statusTestUrl').textContent = testUrl;
        renderNodesTable(domain);
      }, d => VPUI.toast.error(d.msg || '切换失败'));
    });
  }

  /* ===== V2ray Table (Tab 2) ===== */
  let v2rayDataCache = null;

  function refreshV2rayTable() {
    const container = document.getElementById('v2rayTableBody');
    if (!container) return;
    container.innerHTML = '<div class="table-loading"><span class="spinner"></span>加载中...</div>';

    API.getV2rayList(dt => {
      v2rayDataCache = dt;
      if (!dt.data || dt.data.length === 0) {
        container.innerHTML = '<div class="table-empty">' +
          '<div style="font-size:40px;margin-bottom:8px;opacity:0.3">🖥</div>' +
          '<div style="font-size:14px;font-weight:600;margin-bottom:4px">暂无 v2ray 服务</div>' +
          '<div style="font-size:12px;color:#999">点击上方「添加v2ray服务」按钮启动新实例</div>' +
          '</div>';
        return;
      }
      let html = '<div class="table-wrap"><table class="vp-table"><thead><tr>' +
        '<th style="width:40px">#</th>' +
        '<th>PID</th>' +
        '<th>类型</th>' +
        '<th>本地端口</th>' +
        '<th>配置文件</th>' +
        '<th style="width:160px">操作</th>' +
        '</tr></thead><tbody>';

      dt.data.forEach((row, idx) => {
        const isCustom = row.run_mode === '个性配置';
        html += '<tr>' +
          '<td class="col-num">' + (idx + 1) + '</td>' +
          '<td>' + escHtml(row.pid || '-') + '</td>' +
          '<td><span class="badge ' + (isCustom ? 'badge-blue' : 'badge-green') + '">' + escHtml(row.run_mode || '-') + '</span></td>' +
          '<td>' + escHtml(row.local_ports || '-') + '</td>' +
          '<td style="font-size:12px">' + escHtml(row.config_file || '-') + '</td>' +
          '<td><span class="vp-table-actions">' +
            '<button class="btn btn-sm" onclick="handleViewConf(' + idx + ')" title="查看配置">&#9881;</button>' +
            (isCustom
              ? '<button class="btn btn-sm" onclick="handleCopyRun(' + idx + ')" title="复制并运行">&#128206;</button>' +
                '<button class="btn btn-sm" onclick="handleRestart(' + idx + ')" title="重启">&#8635;</button>' +
                '<button class="btn btn-sm btn-danger" onclick="handleDeleteV2ray(' + idx + ')" title="删除">&#10005;</button>'
              : '<button class="btn btn-sm" disabled title="系统进程不可操作">&#9881;</button>' +
                '<button class="btn btn-sm" disabled title="系统进程不可复制">&#128206;</button>' +
                '<button class="btn btn-sm" disabled title="系统进程不可重启">&#8635;</button>' +
                '<button class="btn btn-sm" disabled title="系统进程不可删除">&#10005;</button>') +
          '</span></td>' +
          '</tr>';
      });

      html += '</tbody></table></div>';
      container.innerHTML = html;
    }, function() {
      container.innerHTML = '<div class="table-empty">' +
        '<div style="font-size:40px;margin-bottom:8px;opacity:0.3">⚠️</div>' +
        '<div>加载失败</div>' +
        '</div>';
    });
  }

  /* V2ray Table Actions */
  window.handleViewConf = function(idx) {
    if (!v2rayDataCache || !v2rayDataCache.data || !v2rayDataCache.data[idx]) return;
    const row = v2rayDataCache.data[idx];
    VPUI.openModal({
      title: '配置内容',
      width: '60%',
      maxWidth: '700px',
      content: '<pre style="background:#f5f5f5;padding:12px;border-radius:4px;overflow:auto;max-height:400px;font-size:12px;font-family:var(--font-mono);white-space:pre-wrap;word-break:break-all;">' + escHtml(row.config_json || '') + '</pre>',
      buttons: [{ text: '关闭', primary: true, action: () => {} }]
    });
  };

  window.handleCopyRun = function(idx) {
    if (!v2rayDataCache || !v2rayDataCache.data || !v2rayDataCache.data[idx]) return;
    const row = v2rayDataCache.data[idx];
    const oldConfigFile = (row.config_file || '').replace('routing.rules.json -> ', '').trim();

    VPUI.openModal({
      title: '复制配置文件并运行',
      width: '50%',
      maxWidth: '600px',
      content: `
        <div class="form">
          <div class="form-group">
            <label class="form-label">原配置文件</label>
            <input type="text" class="form-control" id="copyOldConfig" disabled value="${escHtml(oldConfigFile)}">
          </div>
          <div class="form-group">
            <label class="form-label">新配置文件</label>
            <input type="text" class="form-control" id="copyNewConfig" placeholder="例: v2ray.config.json" value="">
          </div>
          <div class="form-group">
            <label class="form-label">全局代理</label>
            <div class="toggle-wrap">
              <span class="toggle" id="copyGlobalToggle"></span>
              <span class="toggle-label" id="copyGlobalLabel">关闭</span>
            </div>
          </div>
          <div class="form-group">
            <label class="form-label">新入站协议</label>
            <select class="form-control" id="copyProtocol">
              <option value="http">http</option>
              <option value="socks">socks(socks5)</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">新入站端口</label>
            <input type="text" class="form-control" id="copyPort" placeholder="填写新的本地端口号" value="">
          </div>
        </div>
      `,
      buttons: [
        { text: '取消', action: () => {} },
        { text: '启动', primary: true, action: (m) => {
          const globalToggle = m.querySelector('#copyGlobalToggle');
          const data = {
            old_config_file: m.querySelector('#copyOldConfig').value,
            config_file: m.querySelector('#copyNewConfig').value,
            inbound_protocol: m.querySelector('#copyProtocol').value,
            local_port: parseInt(m.querySelector('#copyPort').value) || 0,
            global_proxy: globalToggle.classList.contains('active') ? 'on' : 'off'
          };
          if (!data.config_file) { VPUI.toast.warning('请输入新配置文件名称'); return false; }
          API.copyRunV2ray(data, d => {
            VPUI.toast.success(d.msg || '启动成功');
            setTimeout(() => refreshV2rayTable(), 1000);
          }, d => VPUI.toast.error(d.msg || '启动失败'));
        }}
      ],
      onOpen: (m) => {
        const toggle = m.querySelector('#copyGlobalToggle');
        const label = m.querySelector('#copyGlobalLabel');
        toggle.addEventListener('click', () => {
          toggle.classList.toggle('active');
          label.textContent = toggle.classList.contains('active') ? '开启' : '关闭';
        });
      }
    });
  };

  window.handleRestart = function(idx) {
    if (!v2rayDataCache || !v2rayDataCache.data || !v2rayDataCache.data[idx]) return;
    const row = v2rayDataCache.data[idx];
    API.restartV2ray({ pid: row.pid, config_file: row.config_file }, d => {
      VPUI.toast.success(d.msg || '重启成功');
      refreshV2rayTable();
    }, d => VPUI.toast.error(d.msg || '重启失败'));
  };

  window.handleDeleteV2ray = async function(idx) {
    if (!v2rayDataCache || !v2rayDataCache.data || !v2rayDataCache.data[idx]) return;
    const row = v2rayDataCache.data[idx];
    const ok = await VPUI.confirm('确定删除此 v2ray 进程？');
    if (!ok) return;
    API.deleteV2ray(row.pid, d => {
      VPUI.toast.success(d.msg || '已删除');
      refreshV2rayTable();
    }, d => VPUI.toast.error(d.msg || '删除失败'));
  };

  /* ===== V2ray Add ===== */
  window.handleAddV2ray = function() {
    VPUI.openModal({
      title: '添加v2ray服务',
      width: '50%',
      maxWidth: '600px',
      content: document.getElementById('addV2rayFormContent')?.innerHTML || '加载中...',
      buttons: [
        { text: '取消', action: () => {} },
        { text: '启动', primary: true, action: (m) => {
          const configFile = m.querySelector('#addConfigFile')?.value;
          if (!configFile) { VPUI.toast.warning('请输入配置文件路径'); return false; }
          API.runV2ray(configFile, d => {
            VPUI.toast.success(d.msg || '启动成功');
            setTimeout(() => refreshV2rayTable(), 500);
          }, d => VPUI.toast.error(d.msg || '启动失败'));
        }}
      ],
      onOpen: function(modal, bodyEl) {
        // 初始化添加v2ray的标签切换
        var header = bodyEl.querySelector('.vp-tabs-mini > div:first-child');
        if (!header) return;
        header.addEventListener('click', function(e) {
          var btn = e.target.closest('[data-tab]');
          if (!btn) return;
          var tabId = btn.getAttribute('data-tab');
          header.querySelectorAll('[data-tab]').forEach(function(b) {
            b.style.color = 'var(--text-secondary)';
            b.style.fontWeight = 'normal';
            b.style.borderBottom = '2px solid transparent';
          });
          btn.style.color = 'var(--primary)';
          btn.style.fontWeight = '600';
          btn.style.borderBottom = '2px solid var(--primary)';
          bodyEl.querySelectorAll('.mini-tab-panel').forEach(function(p) { p.style.display = 'none'; });
          var panel = bodyEl.querySelector('#' + tabId);
          if (panel) panel.style.display = '';
        });
      }
    });
  };


  /* ===== INIT ===== */
  document.addEventListener('DOMContentLoaded', function() {
    initTabs();
    initToolbar();
    initDomainSelect();
    renderNodesTable();
    refreshStatusBar();

    // Auto-refresh status bar every 30s
    setInterval(refreshStatusBar, 30000);
  });

})();
