/* ============================================
   V2rayPool - Main Application Logic
   Pure native JS (replaces layui main.js)
   ============================================ */

(function() {
  'use strict';

  /* ===== Global State ===== */
  let VPTestUrl = window.__VP_TEST_URL__ || 'https://www.google.com/';
  let VPTestedDomainList = window.__VP_TESTED_DOMAINS__ || [];
  let showExtraCols = true; // 列显示切换

  /* ===== 隧道卡片可点击数值的公共设置 ===== */
  function setupTunnelClickable(elId, value, displayUnit, configKey, title, minVal) {
    var el = document.getElementById(elId);
    if (!el) return;
    if (value === '-') {
      el.className = 'card-value';
      el.style.cursor = 'default';
      el.style.textDecoration = 'none';
      el.title = '';
      el.innerHTML = '-';
      el.onclick = null;
    } else {
      el.className = 'card-value card-value-clickable';
      el.innerHTML = value + ' <span class="unit">' + displayUnit + '</span>';
      el.title = '点击修改' + title;
      var promptTitle = '修改' + title + (displayUnit ? '(' + displayUnit + ')' : '');
      el.onclick = function() {
        VPUI.prompt(promptTitle, value).then(function(val) {
          if (val === null) return;
          var n = parseInt(val);
          if (isNaN(n) || n < minVal) {
            VPUI.toast.warning('值必须为≥' + minVal + '的整数');
            return;
          }
          var data = {};
          data[configKey] = String(n);
          API.updateSetting(data, function() {
            VPUI.toast.success(configKey + ' 已更新为 ' + n + displayUnit);
            refreshStatusBar();
          }, function(dt2) { VPUI.toast.error(dt2.msg || '更新失败'); });
        });
      };
    }
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
      if (idx === 0) adjustTableHeight();
      if (idx === 1) { refreshV2rayTable(); adjustTableHeight(); };
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

    // Tunnel status — 更新隧道卡片
    API.tunnelStatus(dt => {
      var stateEl = document.getElementById('tunnelState');
      var nodesEl = document.getElementById('tunnelNodes');
      var portEl = document.getElementById('tunnelPort');
      var delayEl = document.getElementById('tunnelMaxDelay');
      var intervalEl = document.getElementById('tunnelRefreshInterval');
      var btn = document.getElementById('tunnelBtn');

      if (stateEl) {
        var running = dt.data && dt.data.running;
        stateEl.textContent = running ? '运行中' : '未启动';
        stateEl.style.color = running ? '#5FB878' : '#999';
      }
      if (nodesEl) nodesEl.textContent = (dt.data && dt.data.node_count) || 0;
      if (portEl) {
        var portVal = running ? (dt.data.port || '-') : '-';
        setupTunnelClickable('tunnelPort', String(portVal), '', 'VP_TUNNEL_PORT', '端口', 1);
      }
      if (delayEl) {
        var delayVal = running ? (dt.data.max_delay_ms || 230) : '-';
        setupTunnelClickable('tunnelMaxDelay', String(delayVal), 'ms', 'VP_TUNNEL_MAX_DELAY', '延迟阈值', 10);
      }
      if (intervalEl) {
        var intervalVal = running ? ((dt.data && dt.data.refresh_interval) || 1200) : '-';
        setupTunnelClickable('tunnelRefreshInterval', String(intervalVal), 's', 'VP_TUNNEL_REFRESH_INTERVAL', '测速间隔', 10);
      }
      if (btn) {
        btn.textContent = running ? '🕳 隧道池: 开' : '🕳 隧道池: 关';
        btn.className = running ? 'btn btn-sm btn-orange' : 'btn btn-sm btn-success';
      }
    }, function() {
      var el = document.getElementById('tunnelState');
      if (el) { el.textContent = '状态未知'; el.style.color = '#FF5722'; }
    });
  }

  /* ===== Nodes Table ===== */
  function renderNodesTable(domain) {
    const container = document.getElementById('nodeTableInner');
    if (!container) return;
    container.innerHTML = '<div class="table-loading"><span class="spinner"></span>加载中...</div>';

    API.getNodes(domain, function(dt) {
      nodesDataCache = dt;
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
          '<th data-sort="speed" style="cursor:pointer">速度/秒</th>' +
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
            ? '<button class="btn btn-sm btn-purple" onclick="handleActive(' + row.index + ')">已启用</button>'
            : '<button class="btn btn-sm btn-primary" onclick="handleActive(' + row.index + ')">启用</button>';

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
            '<button class="btn btn-sm btn-danger" onclick="handleDelete(' + row.index + ')">删除</button></span></td>' +
            '</tr>';
        });

        html += '</tbody></table></div>';
        html += '<div class="text-muted" style="margin-top:6px;font-size:12px">共 ' + data.length + ' 个节点</div>';
        container.innerHTML = html;
        adjustTableHeight();
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
  window.handleToolbarAction = function(action) {
    if (typeof toolbarActions[action] === 'function') toolbarActions[action]();
  };

  window.handleActive = async function(serverIdx) {
    const dt = await getNodesData();
    if (!dt || !dt.data) return;
    const node = dt.data.find(function(n) { return n.index === serverIdx; });
    if (!node) { VPUI.toast.error('节点不存在'); return; }
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

  window.handleDelete = async function(serverIdx) {
    const dt = await getNodesData();
    if (!dt || !dt.data) return;
    const node = dt.data.find(function(n) { return n.index === serverIdx; });
    if (!node) { VPUI.toast.error('节点不存在'); return; }
    const ok = await VPUI.confirm('确定删除节点【' + node.title + '】？');
    if (!ok) return;
    API.deleteNode(serverIdx, d => {
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
    document.addEventListener('click', function(e) {
      var btn = e.target.closest('[data-action]');
      if (!btn) return;
      var action = btn.dataset.action;
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
              // 在工具栏显示测速进度
              var testBtn = document.querySelector('[data-action="testNodes"]');
              var origText = testBtn ? testBtn.textContent : '';
              if (testBtn) { testBtn.disabled = true; }
              var count = 12;
              if (testBtn) { testBtn.textContent = '⏱ 测速中 ' + count + 's'; }
              var iv = setInterval(function() {
                renderNodesTable();
                refreshStatusBar();
                count--;
                if (testBtn) { testBtn.textContent = '⏱ 测速中 ' + count + 's'; }
                if (count <= 0) {
                  clearInterval(iv);
                  if (testBtn) { testBtn.disabled = false; testBtn.textContent = origText || '⏱ 测速'; }
                }
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
        onOpen: function(m) {
          // 检查配置，高亮缺失字段
          API.checkConfig(function(dt) {
            if (!dt.data || dt.data.length === 0) return;
            dt.data.forEach(function(item) {
              var input = m.querySelector('[name="' + item.field + '"]');
              if (!input) return;
              input.style.borderColor = '#ff4d4f';
              input.style.boxShadow = '0 0 0 2px rgba(255,77,79,0.2)';
              // 在输入框下方添加错误提示
              var errEl = document.createElement('div');
              errEl.style.color = '#cf1322';
              errEl.style.fontSize = '12px';
              errEl.style.marginTop = '2px';
              errEl.textContent = '⚠ ' + item.message;
              input.parentNode.appendChild(errEl);
            });
          });
        },
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
    },

    toggleCols: function() {
      showExtraCols = !showExtraCols;
      var container = document.getElementById('nodesTableBody');
      if (container) {
        container.classList.toggle('hide-cols', !showExtraCols);
      }
      VPUI.toast.info(showExtraCols ? '已显示全部列' : '已隐藏节点地址/测速时间列');
    }
  };

  /* ===== Config Check ===== */
  /* ===== Config Check ===== */
  function checkConfigAndShowBanner() {
    var banner = document.getElementById('configAlertBanner');
    if (!banner) return;

    var issues = [];
    var popupShown = false;

    function renderAll() {
      if (issues.length === 0) {
        banner.style.display = 'none';
        return;
      }
      var html = '<strong>⚠ 待处理事项：</strong>';
      issues.forEach(function(item, i) {
        html += '<div style="margin-top:' + (i > 0 ? '4px' : '4px') + '">' + item.icon + ' <strong>' + item.label + '</strong>：' + item.msg + '</div>';
      });
      html += '<div style="margin-top:6px"><button class="btn btn-sm" onclick="handleToolbarAction(\'settinglayer\')" style="background:#cf1322;color:#fff;border:none;padding:2px 12px;border-radius:3px;cursor:pointer">⚙ 立即设置</button></div>';
      banner.innerHTML = html;
      banner.style.display = 'block';

      if (!popupShown) {
        popupShown = true;
        // 构建纯文本提示（给原生 alert 用）
        var popupLines = [];
        issues.forEach(function(item) {
          popupLines.push(item.icon + ' ' + item.label + '：' + item.msg);
        });
        popupLines.push('请先配置必要参数，然后点击「▶ 启动」按钮，程序会自动初始化并启动节点。');
        var popupText = popupLines.join('\n');

        // 先记录到控制台方便调试
        console.log('--- 配置检查发现 ' + issues.length + ' 个问题 ---');
        console.log(popupText);

        // 1) 原生 alert 确保一定能弹出来
        window.alert('⚠ 配置检查\n\n' + popupText + '\n\n（点击确定后，顶部横幅会显示详细待办事项）');

        // 2) 再显示 VPUI 风格的弹层（更美观）
        var popupHtml = '<div style="color:#cf1322;font-size:14px;line-height:1.8">';
        issues.forEach(function(item) {
          popupHtml += '<div>' + item.icon + ' <strong>' + item.label + '</strong>：' + item.msg + '</div>';
        });
        popupHtml += '</div><div style="margin-top:12px;font-size:13px;color:#666">请先配置必要参数，然后点击「启动」按钮，程序会自动初始化并启动节点。</div>';
        VPUI.alert(popupHtml, {title: '⚠ 配置检查'});
      }
    }

    // 直接调用 API 检查配置，不依赖服务端模板注入
    API.checkConfig(function(dt) {
      if (dt.data && dt.data.length > 0) {
        dt.data.forEach(function(item) {
          var icon = '🟡';
          if (item.status === 'missing') icon = '🔴';
          else if (item.status === 'error') icon = '🟠';
          issues.push({ icon: icon, label: item.label, msg: item.message });
        });
        renderAll();
      }
    }, function() {
      // API 请求失败（如后端未运行），静默处理
      console.warn('配置检查 API 请求失败，跳过弹窗');
    });
  }

  /* ===== Node Table Sort ===== */
  let nodesDataCache = null;
  let nodesSortDir = { speed: 0 }; // 0=none, 1=asc, 2=desc

  function initTableSort() {
    var container = document.getElementById('nodeTableInner');
    if (!container) return;
    container.addEventListener('click', function(e) {
      var th = e.target.closest('th[data-sort]');
      if (!th) return;
      var field = th.getAttribute('data-sort');
      if (!field) return;
      nodesSortDir[field] = ((nodesSortDir[field] || 0) % 3) + 1;
      var dir = nodesSortDir[field]; // 1=asc, 2=desc
      if (!nodesDataCache || !nodesDataCache.data) return;
      var sorted = nodesDataCache.data.slice();
      if (dir === 1) {
        sorted.sort(function(a, b) {
          var va = parseFloat(a[field]) || 0;
          var vb = parseFloat(b[field]) || 0;
          if (va === 0 && a.test_at === '0001-01-01 00:00') va = Infinity;
          if (vb === 0 && b.test_at === '0001-01-01 00:00') vb = Infinity;
          return va - vb;
        });
      } else if (dir === 2) {
        sorted.sort(function(a, b) {
          var va = parseFloat(a[field]) || 0;
          var vb = parseFloat(b[field]) || 0;
          if (va === 0 && a.test_at === '0001-01-01 00:00') va = Infinity;
          if (vb === 0 && b.test_at === '0001-01-01 00:00') vb = Infinity;
          return vb - va;
        });
      } else {
        sorted = nodesDataCache.data;
      }
      nodesSortDir[field] = dir;
      renderSortedTable(sorted);
    });
  }

  function renderSortedTable(data) {
    var container = document.getElementById('nodeTableInner');
    if (!container) return;
    var dir = nodesSortDir.speed || 0;
    var arrow = dir === 0 ? '' : (dir === 1 ? ' ▲' : ' ▼');
    var html = '<div class="table-wrap"><table class="vp-table"><thead><tr>' +
      '<th style="width:40px">#</th>' +
      '<th>协议</th>' +
      '<th>本机地址</th>' +
      '<th data-sort="speed" style="cursor:pointer">速度/秒' + arrow + '</th>' +
      '<th>标题</th>' +
      '<th>节点地址</th>' +
      '<th>状态</th>' +
      '<th>测速时间</th>' +
      '<th style="width:120px">操作</th>' +
      '</tr></thead><tbody>';
    data.forEach(function(row) {
      var tested = row.test_at && row.test_at !== '0001-01-01 00:00';
      var speedDisplay = tested
        ? '<span class="badge badge-green" onclick="VPUI.showTips(\'' + (row.test_url || '').replace(/'/g, "\\'") + '\', this)">' + (row.speed >= 5 ? 100 : (row.speed || '-')) + ' s</span>'
        : '<span class="badge">- s</span>';
      var statusDisplay = row.is_running
        ? '<span class="badge badge-green">运行中</span>'
        : '<span class="badge">已停止</span>';
      var activeBtn = row.is_active
        ? '<button class="btn btn-sm btn-purple" onclick="handleActive(' + row.index + ')">已启用</button>'
        : '<button class="btn btn-sm btn-primary" onclick="handleActive(' + row.index + ')">启用</button>';
      html += '<tr>' +
        '<td class="col-num">' + (data.indexOf(row) + 1) + '</td>' +
        '<td>' + escHtml(row.protocol || '-') + '</td>' +
        '<td>' + escHtml(row.local_addr || '-') + '</td>' +
        '<td>' + speedDisplay + '</td>' +
        '<td>' + escHtml(row.title || '-') + '</td>' +
        '<td>' + escHtml(row.remote_addr || '-') + '</td>' +
        '<td>' + statusDisplay + '</td>' +
        '<td style="font-size:12px;color:#999">' + (tested ? row.test_at : '-') + '</td>' +
        '<td><span class="vp-table-actions">' + activeBtn +
        '<button class="btn btn-sm btn-danger" onclick="handleDelete(' + row.index + ')">删除</button></span></td>' +
        '</tr>';
    });
    html += '</tbody></table></div>';
    html += '<div class="text-muted" style="margin-top:6px;font-size:12px">共 ' + data.length + ' 个节点</div>';
    container.innerHTML = html;
    adjustTableHeight();
  }

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


  /* ===== Table Auto-Height ===== */
  function adjustTableHeight() {
    var container = document.getElementById('nodesTableBody');
    if (!container || container.classList.contains('active') === false) return;
    // 计算剩余高度：视口高度 - 容器顶部偏移 - 底部留白
    var rect = container.getBoundingClientRect();
    var available = window.innerHeight - rect.top - 20;
    if (available > 300) {
      container.style.minHeight = available + 'px';
      // 让 table-wrap 占满
      var wrap = container.querySelector('.table-wrap');
      if (wrap) {
        wrap.style.maxHeight = (available - 60) + 'px';
        wrap.style.overflowY = 'auto';
      }
    }
  }

  /* ===== INIT ===== */
  document.addEventListener('DOMContentLoaded', function() {
    console.log('V2rayPool WebUI loaded');
    initTabs();
    initToolbar();
    initDomainSelect();
    initTableSort();
    renderNodesTable();
    refreshStatusBar();
    checkConfigAndShowBanner();

    // Auto-refresh status bar every 30s
    setInterval(refreshStatusBar, 30000);
    // 浏览器 resize 时更新表格高度
    window.addEventListener('resize', adjustTableHeight);
  });

})();
