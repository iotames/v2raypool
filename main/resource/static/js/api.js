/* ============================================
   V2rayPool - API Request Module
   Native fetch-based (replaces layui request.js)
   所有函数接受 (success, fail) 两个回调，确保任何错误都能反馈到前端
   ============================================ */

const API = (() => {
  'use strict';

  function getJson(url, success, fail) {
    fetch(url)
      .then(res => res.json())
      .then(dt => {
        if (dt.code === 0 || dt.code === 200) {
          success && success(dt);
        } else {
          fail && fail(dt);
        }
      })
      .catch(err => {
        fail && fail({ code: 0, msg: err.message || '网络错误' });
      });
  }

  function postJson(url, data, success, fail, btn) {
    if (btn) btn.disabled = true;

    fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    })
      .then(res => res.json())
      .then(dt => {
        if (btn) btn.disabled = false;
        if (dt.code === 200) {
          success && success(dt);
        } else {
          fail && fail(dt);
        }
      })
      .catch(err => {
        if (btn) btn.disabled = false;
        fail && fail({ code: 0, msg: err.message || '网络错误' });
      });
  }

  /* ===== Node APIs ===== */
  function getNodes(domain, success, fail) {
    const url = domain ? '/api/nodes?domain=' + encodeURIComponent(domain) : '/api/nodes';
    getJson(url, success, fail);
  }

  function testNodes(testUrl, success, fail) {
    postJson('/api/nodes/test', { TestUrl: testUrl }, success, fail);
  }

  function setTestUrl(testUrl, success, fail) {
    postJson('/api/nodes/test-url', { TestUrl: testUrl }, success, fail);
  }

  function startNodes(success, fail) {
    postJson('/api/nodes/start', {}, success, fail);
  }

  function subscribeNodes(subscribeByProxy, success, fail) {
    postJson('/api/nodes/subscribe', { subscribe_by_proxy: subscribeByProxy }, success, fail);
  }

  function activeNode(remoteAddr, globalProxy, success, fail) {
    postJson('/api/node/active', { remote_addr: remoteAddr, global_proxy: !!globalProxy }, success, fail);
  }

  function unactiveNode(remoteAddr, success, fail) {
    postJson('/api/node/unactive', { remote_addr: remoteAddr }, success, fail);
  }

  function deleteNode(index, success, fail) {
    postJson('/api/node/delete', { index: index }, success, fail);
  }

  /* ===== V2ray APIs ===== */
  function getV2rayList(success, fail) {
    getJson('/api/v2ray/list', success, fail);
  }

  function runV2ray(configFile, success, fail) {
    postJson('/api/v2ray/run', { config_file: configFile }, success, fail);
  }

  function copyRunV2ray(data, success, fail) {
    postJson('/api/v2ray/copyrun', data, success, fail);
  }

  function restartV2ray(data, success, fail) {
    postJson('/api/v2ray/restart', data, success, fail);
  }

  function deleteV2ray(pid, success, fail) {
    postJson('/api/v2ray/delete', { pid: pid }, success, fail);
  }

  function updateRoutingRules(data, success, fail) {
    postJson('/api/v2ray/routing-rules/update', data, success, fail);
  }

  /* ===== Tunnel APIs ===== */
  function tunnelStart(success, fail) {
    postJson('/api/tunnel/start', {}, success, fail);
  }

  function tunnelStop(success, fail) {
    postJson('/api/tunnel/stop', {}, success, fail);
  }

  function tunnelStatus(success, fail) {
    getJson('/api/tunnel/status', success, fail);
  }

  /* ===== SysProxy APIs ===== */
  function sysProxyStatus(success, fail) {
    getJson('/api/sysproxy/status', success, fail);
  }

  function sysProxySwitch(typeVal, nodeIdx, success, fail) {
    postJson('/api/sysproxy/switch', { type: typeVal, node_idx: nodeIdx || -1 }, success, fail);
  }

  /* ===== Settings APIs ===== */
  function updateSetting(data, success, fail) {
    postJson('/api/setting/update', data, success, fail);
  }

  function clearCache(success, fail) {
    postJson('/api/setting/clearcache', {}, success, fail);
  }

  return {
    getJson, postJson,
    getNodes, testNodes, setTestUrl, startNodes,
    subscribeNodes, activeNode, unactiveNode, deleteNode,
    getV2rayList, runV2ray, copyRunV2ray, restartV2ray, deleteV2ray,
    updateRoutingRules,
    tunnelStart, tunnelStop, tunnelStatus,
    sysProxyStatus, sysProxySwitch,
    updateSetting, clearCache
  };
})();
