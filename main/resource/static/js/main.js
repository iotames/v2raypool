var showTips = function (msg, ele) {
    layer.tips(msg, ele);
}

layui.use(['table', 'dropdown', 'form'], function () {
    var table = layui.table;
    var dropdown = layui.dropdown;
    var layer = layui.layer;
    var util = layui.util;
    var form = layui.form;
    var $ = layui.$;
    var element = layui.element;

    element.on('tab(v2raytab)', function (data) {
        if (data.index == 1) table.reload('v2raylist');
    });

    form.on('select(test_domain)', function (data) {
        var domain = data.value;
        table.reloadData('nodeslist', { where: { domain: domain } });
        if (domain) {
            // 下拉框数据来自已测速域列表，安全防御检查
            var list = (typeof VPTestedDomainList !== 'undefined') ? VPTestedDomainList : [];
            if (list.indexOf(domain) === -1) {
                layer.msg("该域名未测试过，请先测速", {icon: 2});
                return;
            }
            var testUrl = 'https://' + domain + '/';
            VPTestUrl = testUrl;
            postjson("/api/nodes/test-url", { TestUrl: testUrl }, function(dt) {
                layer.msg("测速地址已切换到: " + domain, {time: 1500});
                document.getElementById("statusTestUrl").innerText = testUrl;
            }, function(dt) { layer.alert(dt.msg || "切换失败", {icon:2, title:"错误"}); });
        }
    });

    util.on('lay-on', {
        'routingrules': function () {
            var idx = layer.open({
                title: '<strong>路由分流规则</strong>(配置文件: default.env -> .env )',
                type: 1, area: ['80%', '80%'],
                content: $('#routingruleslayer'),
                success: function() {
                    form.render();
                    form.on('submit(routingrules)', function (data) {
                        var submitbtn = document.getElementById("routingrulesbtn");
                        postjson("/api/v2ray/routing-rules/update", data.field, function (dt) {
                            layer.close(idx); layer.msg(dt.msg);
                            table.reload('nodeslist'); refreshStatusBar();
                        }, function (dt) { layer.alert(dt.msg, { icon: 2, title: dt.code + "错误" }); }, submitbtn);
                    });
                }
            });
        },
        'subscribe': function () {
            var idx = layer.open({
                title: "更新订阅", type: 1, area: ['60%', '230px'],
                content: $('#subscribelayer'),
                success: function() {
                    form.render();
                    form.on('submit(updatesubscribe)', function (data) {
                        var submitbtn = document.getElementById("updatesubscribe");
                        postjson("/api/nodes/subscribe", data.field, function (dt) {
                            layer.close(idx); layer.msg(dt.msg);
                            table.reload('nodeslist'); refreshStatusBar();
                        }, function (dt) { layer.alert(dt.msg, { icon: 2, title: dt.code + "错误" }); }, submitbtn);
                    });
                }
            });
        }
    });

    table.render({
        elem: '#nodeslist', url: '/api/nodes',
        toolbar: '#toolbarDemo',
        defaultToolbar: ['filter', {title: '基本设置', layEvent: 'settinglayer', icon: 'layui-icon-set'}, {title: '清除缓存', layEvent:'clearlayer', icon:'layui-icon-delete'}],
        escape: false, height: 'full-35',
        cellMinWidth: 80,
        cols: [[
            { type: "numbers", fixed: 'left' },
            { field: 'index', width: 100, title: 'Index', sort: true, hide: true },
            { field: 'protocol', title: '协议', width: 60 },
            { field: 'local_port', title: '本地端口', width: 110, sort: true, hide: true },
            { field: 'local_addr', title: '本机地址', width: 190 },
            { field: 'speed', title: '速度/秒', templet: '<span class="layui-badge {{= d.test_at != "0001-01-01 00:00" ? "layui-bg-green" : ""}}" onclick="showTips(\'{{= d.test_url }}\', this)">{{= d.speed >= 5 ? 100 : d.speed }} s</span>', width: 100, sort: true },
            { field: 'title', title: '标题', width: 200 },
            { field: 'remote_addr', title: '节点地址', width: 180 },
            { field: 'is_running', title: '状态', templet: '<span class="layui-badge {{= d.is_running ? "layui-bg-green" : ""}}">{{= d.is_running ? "运行中" : "已停止"}}</span>', width: 60 },
            { field: 'test_at', title: '测速时间', width: 180 },
            { fixed: 'right', title: '操作', width: 80, minWidth: 115, toolbar: '#barDemo' }
        ]],
        error: function (res, msg) { console.log(res, msg); }
    });

    table.on('toolbar(nodeslist)', function (obj) {
        switch (obj.event) {
            case 'sysProxyOpen':
                getjson("/api/sysproxy/status", function(dt) {
                    var cur = (dt.data && dt.data.type) || 0;
                    var next = cur === 0 ? 2 : 0; // 关闭↔隧道
                    postjson("/api/sysproxy/switch", {type: next, node_idx: -1}, function(dt2) {
                        layer.msg(dt2.msg);
                        refreshStatusBar();
                    }, function(dt2) {
                        layer.alert(dt2.msg, {icon:2, title:"切换失败"});
                    });
                }, function(dt) {
                    layer.alert("获取代理状态失败", {icon:2, title:"错误"});
                });
                break;
            case 'tunnelToggle':
                getjson("/api/tunnel/status", function(dt) {
                    if (dt.data && dt.data.running) {
                        postjson("/api/tunnel/stop", {}, function() { layer.msg("隧道代理池已关闭"); refreshStatusBar(); }, function(dt2) { layer.alert(dt2.msg, {icon:2, title:"错误"}); });
                    } else {
                        postjson("/api/tunnel/start", {}, function() { layer.msg("隧道代理池已启动"); refreshStatusBar(); setTimeout(refreshStatusBar, 5000); }, function(dt2) { layer.alert(dt2.msg, {icon:2, title:"启动失败"}); });
                    }
                }, function(dt) {
                    layer.alert("获取隧道状态失败", {icon:2, title:"错误"});
                });
                break;
            case 'testNodes':
                var testedDomains = (typeof VPTestedDomainList !== 'undefined') ? VPTestedDomainList : [];
                var defUrl = (typeof VPTestUrl !== 'undefined' && VPTestUrl) ? VPTestUrl : 'https://www.google.com/';
                var html = '<div style="padding:16px;">';
                html += '<form class="layui-form">';
                html += '<div class="layui-form-item"><label class="layui-form-label">测速地址</label><div class="layui-input-block">';
                html += '<input type="text" id="testUrlInput" value="' + defUrl.replace(/"/g, '&quot;') + '" class="layui-input">';
                html += '</div></div>';
                if (testedDomains.length > 0) {
                    html += '<div class="layui-form-item"><label class="layui-form-label">历史测速</label><div class="layui-input-block">';
                    for (var j = 0; j < testedDomains.length; j++) {
                        html += '<span class="layui-badge layui-bg-blue" style="cursor:pointer;margin:2px;" onclick="var v=this.innerText;document.getElementById(\'testUrlInput\').value=v.indexOf(\'http\')===0?v:\'https://\'+v">' + testedDomains[j] + '</span>';
                    }
                    html += '</div></div>';
                }
                html += '</form>';
                html += '</div>';
                layer.open({
                    title: '节点测速', type: 1, area: ['500px', '280px'],
                    content: html, btn: ['开始测速', '取消'],
                    yes: function(index) {
                        var btn = document.getElementById("testproxynodes");
                        var testUrl = document.getElementById("testUrlInput").value;
                        if (!testUrl) { layer.msg("请输入测速地址"); return; }
                        postjson("/api/nodes/test", { TestUrl: testUrl }, function() {
                            layer.msg("测速已开始"); layer.close(index);
                            btn.disabled = true; btn.innerText = "测速中...";
                            var times = 8, interval = setInterval(function () {
                                if (times > 0) { table.reload('nodeslist'); btn.innerText = ((times < 10 ? "0" + times : times) + "s"); times--; }
                                else { clearInterval(interval); btn.disabled = false; btn.innerText = "测速"; window.location.reload(); }
                            }, 2000);
                        }, function(dt2) { layer.alert(dt2.msg, {icon:2, title:"错误"}); });
                    }
                });
                break;
            case 'startNodes':
                var btn = document.getElementById("startproxynodes");
                postjson('/api/nodes/start', {}, function () {
                    layer.msg("启动成功");
                    btn.disabled = true; btn.innerText = "启动中...";
                    setTimeout(function() { btn.disabled = false; btn.innerText = "启动"; table.reload('nodeslist'); refreshStatusBar(); }, 2000);
                }, function (dt) { layer.alert(dt.msg, { icon: 2, title: dt.code + "错误" }); }, btn);
                break;
            case 'clearlayer':
                layer.confirm("删除runtime目录内所有文件：包括测速缓存等数据。",{icon:3, title:"确定清除应用缓存?"}, function(index){
                    postjson("/api/setting/clearcache", {}, function () {
                        layer.close(index); layer.msg("已清除");
                        setTimeout(function() { window.location.reload(); }, 1000);
                    }, function (dt) { layer.alert(dt.msg, { icon: 2, title: dt.code + "错误" }); });
                });
                break;
            case 'settinglayer':
                var idx = layer.open({
                    title: '系统设置(' + VPEnvFile + '文件)',
                    type: 1, area: ['50%', 'auto'],
                    content: $('#settinglayer'),
                    success: function() {
                        form.render();
                        form.on('submit(setsubmit)', function (data) {
                            var submitbtn = document.getElementById("setupdatebtn");
                            postjson("/api/setting/update", data.field, function (dt) {
                                layer.close(idx); layer.msg(dt.msg);
                                setTimeout(function() { window.location.reload(); }, 3000);
                            }, function (dt) { layer.alert(dt.msg, { icon: 2, title: dt.code + "错误" }); }, submitbtn);
                            return false;
                        });
                    }
                });
                break;
        }
    });

    table.on('tool(nodeslist)', function (obj) {
        var data = obj.data;
        switch (obj.event) {
            case 'active':
                var posturl = "/api/node/active";
                var hintmsg = '激活[' + data.title + ']节点为系统代理吗？';
                if (data.is_active) {
                    hintmsg = '取消[' + data.title + ']节点为系统代理吗？';
                    posturl = "/api/node/unactive";
                    layer.confirm(hintmsg, function (index) {
                        postjson(posturl, { remote_addr: data.remote_addr }, function () {
                            layer.msg("已取消"); table.reload('nodeslist'); refreshStatusBar();
                        }, function (dt) { layer.alert(dt.msg, { icon: 2, title: dt.code + "错误" }); });
                        layer.close(index);
                    });
                } else {
                    var idx = layer.open({
                        title: hintmsg, type: 1, area: ['30%', 'auto'],
                        content: $('#activenodelayer'),
                        success: function() {
                            form.render();
                            form.on('submit(activenode)', function (fdt) {
                                var submitbtn = document.getElementById("activenode");
                                var postdata = { remote_addr: data.remote_addr, global_proxy: fdt.field.global_proxy != "0" };
                                postjson(posturl, postdata, function () {
                                    layer.close(idx); layer.msg("启用成功");
                                    setTimeout(function() { table.reload('nodeslist'); refreshStatusBar(); }, 500);
                                }, function (dt) { layer.alert(dt.msg, { icon: 2, title: dt.code + "错误" }); }, submitbtn);
                                return false;
                            });
                        }
                    });
                }
                break;
            case 'delete':
                layer.confirm("确定删除节点【" + data.title + "】？", function (index) {
                    postjson("/api/node/delete", {index: data.index}, function () {
                        layer.msg("已删除"); table.reload('nodeslist'); refreshStatusBar();
                    }, function (dt) { layer.alert(dt.msg, { icon: 2, title: dt.code + "错误" }); });
                    layer.close(index);
                });
                break;
        }
    });

    function refreshStatusBar() {
        getjson("/api/sysproxy/status", function(dt) {
            var names = {0:"关闭", 1:"单节点", 2:"隧道"};
            var btn = document.getElementById("sysProxyBtn");
            if (dt.data) {
                document.getElementById("statusSysProxy").innerText = names[dt.data.type] || "未知";
                document.getElementById("statusSysProxy").style.color = dt.data.type === 0 ? "#999" : "#5FB878";
                btn.innerText = "系统代理: " + (names[dt.data.type] || "未知");
                btn.className = dt.data.type === 0 ? "layui-btn layui-btn-sm" : "layui-btn layui-btn-sm layui-bg-cyan";
            }
        });
        getjson("/api/nodes", function(dt) {
            var runCount = 0, total = 0;
            if (dt.data) { for (var i = 0; i < dt.data.length; i++) { total++; if (dt.data[i].is_running) runCount++; } }
            document.getElementById("statusRunCount").innerText = runCount + "/" + total;
        });
        getjson("/api/tunnel/status", function(dt) {
            var btn = document.getElementById("tunnelBtn");
            if (dt.data && dt.data.running) {
                document.getElementById("statusTunnelState").innerText = "运行中";
                document.getElementById("statusTunnelState").style.color = "#5FB878";
                document.getElementById("statusTunnelNodes").innerText = dt.data.node_count || 0;
                document.getElementById("statusTunnelPort").innerText = dt.data.port + "";
                document.getElementById("statusTunnelMaxDelay").innerText = dt.data.max_delay_ms || 230;
                var el = document.getElementById("statusTunnelMaxDelay");
                el.style.cursor = "pointer";
                el.style.textDecoration = "underline";
                el.style.textDecorationStyle = "dashed";
                el.style.fontSize = "16px";
                el.style.fontWeight = "bold";
                el.title = "点击修改延迟阈值";
                el.onclick = function() {
                    var oldVal = parseInt(this.innerText) || 230;
                    layer.prompt({title: "修改隧道延迟阈值(ms)", value: oldVal, formType: 0}, function(val, idx) {
                        var n = parseInt(val);
                        if (!n || n < 10) { layer.msg("阈值必须≥10ms"); return; }
                        postjson("/api/setting/update", {VP_TUNNEL_MAX_DELAY: n + ""}, function(dt2) {
                            layer.close(idx); layer.msg("阈值已更新为 " + n + "ms");
                            refreshStatusBar();
                        }, function(dt2) { layer.alert(dt2.msg, {icon:2, title:"更新失败"}); });
                    });
                };
                btn.innerText = "隧道池: 开"; btn.className = "layui-btn layui-btn-sm layui-bg-orange";
            } else {
                document.getElementById("statusTunnelState").innerText = "未启动";
                document.getElementById("statusTunnelState").style.color = "#999";
                document.getElementById("statusTunnelNodes").innerText = "0";
                document.getElementById("statusTunnelPort").innerText = "-";
                document.getElementById("statusTunnelMaxDelay").innerText = "-";
                btn.innerText = "隧道池: 关"; btn.className = "layui-btn layui-btn-sm layui-bg-green";
            }
        });
    }

    refreshStatusBar();
});
