package main

const sharedDarkCSS = `
:root {
  --bg: #f0f2f5;
  --card-bg: #fff;
  --header-bg: #fff;
  --text: #1a1a1a;
  --text-secondary: #666;
  --border: #e8e8e8;
  --hover-bg: #f0f2f5;
  --table-header-bg: #fafafa;
  --input-border: #d9d9d9;
  --input-bg: #fff;
  --mask-bg: rgba(0,0,0,0.45);
  --nav-text: #333;
  --nav-active-text: #fff;
}
body.dark {
  --bg: #141414;
  --card-bg: #1f1f1f;
  --header-bg: #1f1f1f;
  --text: rgba(255,255,255,0.85);
  --text-secondary: #888;
  --border: #333;
  --hover-bg: #2a2a2a;
  --table-header-bg: #2a2a2a;
  --input-border: #444;
  --input-bg: #2a2a2a;
  --mask-bg: rgba(0,0,0,0.6);
  --nav-text: rgba(255,255,255,0.85);
  --nav-active-text: #fff;
}
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;background:var(--bg);color:var(--text);transition:background .3s,color .3s}
.header{background:var(--header-bg);border-bottom:1px solid var(--border);padding:0 24px;height:64px;display:grid;grid-template-columns:1fr auto 1fr;align-items:center;position:sticky;top:0;z-index:10;transition:background .3s,border-color .3s}
.header h1{font-size:18px;font-weight:600;justify-self:start}
.nav{display:flex;gap:4px;justify-self:center}
.nav a{color:var(--nav-text);text-decoration:none;font-size:14px;padding:8px 16px;border-radius:6px;transition:all .2s}
.nav a:hover{background:var(--hover-bg)}
.nav a.active{background:#1677ff;color:var(--nav-active-text)}
.header-right{display:flex;align-items:center;gap:12px;justify-self:end}
.btn-theme{padding:6px 12px;background:transparent;border:1px solid var(--input-border);border-radius:6px;font-size:14px;cursor:pointer;color:var(--text);transition:all .2s}
.btn-theme:hover{border-color:#1677ff;color:#1677ff}
.header a.logout{color:var(--text-secondary);text-decoration:none;font-size:14px;padding:6px 16px;border:1px solid var(--input-border);border-radius:6px;transition:all .2s}
.header a.logout:hover{color:#1677ff;border-color:#1677ff}
.container{max-width:1200px;margin:0 auto;padding:24px}
.toolbar{display:flex;justify-content:space-between;align-items:center;margin-bottom:16px}
.toolbar h2{font-size:18px;font-weight:600}
.btn-refresh{padding:6px 16px;background:var(--card-bg);color:var(--text);border:1px solid var(--input-border);border-radius:6px;font-size:14px;cursor:pointer;transition:all .2s}
.btn-refresh:hover{border-color:#1677ff;color:#1677ff}
.btn-refresh:disabled{opacity:.5;cursor:not-allowed}
`

const loginHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>邮件服务后台 - 登录</title>
<style>
` + sharedDarkCSS + `
body{min-height:100vh;display:flex;justify-content:center;align-items:center}
.login-card{background:var(--card-bg);padding:40px;border-radius:12px;box-shadow:0 4px 20px rgba(0,0,0,0.08);width:100%;max-width:400px;transition:background .3s}
@media(max-width:480px){body{align-items:center;padding:40px 0;background:var(--bg)}.login-card{padding:28px 24px;margin:0 16px;max-width:none;width:auto;box-shadow:0 4px 20px rgba(0,0,0,0.08);border-radius:16px}.login-card h1{font-size:22px;margin-bottom:4px}.login-card p{margin-bottom:24px}.form-group{margin-bottom:18px}.form-group label{font-size:15px}.form-group input{font-size:16px;padding:14px 16px;-webkit-appearance:none;appearance:none}.btn-login{padding:16px;font-size:16px}}
.login-card h1{text-align:center;margin-bottom:8px;font-size:24px;color:var(--text)}
.login-card p{text-align:center;color:var(--text-secondary);margin-bottom:32px;font-size:14px}
.form-group{margin-bottom:20px}
.form-group label{display:block;margin-bottom:6px;font-size:14px;color:var(--text);font-weight:500}
.form-group input{width:100%;padding:12px 16px;border:1px solid var(--input-border);border-radius:8px;font-size:14px;background:var(--input-bg);color:var(--text);transition:border-color .2s}
.form-group input:focus{outline:none;border-color:#1677ff}
.btn-login{width:100%;padding:12px;background:#1677ff;color:#fff;border:none;border-radius:8px;font-size:15px;font-weight:500;cursor:pointer;transition:background .2s}
.btn-login:hover{background:#0958d9}
.btn-login:disabled{background:#91caff;cursor:not-allowed}
.error-msg{color:#ff4d4f;font-size:13px;margin-top:8px;text-align:center;display:none}
.ban-msg{color:#faad14;font-size:13px;margin-top:8px;text-align:center;display:none;background:#fffbe6;padding:8px;border-radius:4px;border:1px solid #ffe58f}
body.dark .ban-msg{background:#2a2105;border-color:#594214}
</style>
</head>
<body>
<div class="login-card">
  <h1>邮件服务后台</h1>
  <p>管理员登录</p>
  <form id="loginForm">
    <div class="form-group">
      <label>用户名</label>
      <input type="text" name="username" required placeholder="请输入用户名" autofocus>
    </div>
    <div class="form-group">
      <label>密码</label>
      <input type="password" name="password" required placeholder="请输入密码">
    </div>
    <button type="submit" class="btn-login" id="btn">登录</button>
    <div class="error-msg" id="err"></div>
    <div class="ban-msg" id="ban"></div>
  </form>
</div>
<script>
document.getElementById('loginForm').addEventListener('submit', async function(e) {
  e.preventDefault();
  const btn = document.getElementById('btn');
  const err = document.getElementById('err');
  const ban = document.getElementById('ban');
  err.style.display = 'none'; ban.style.display = 'none';
  btn.disabled = true; btn.textContent = '登录中...';
  const fd = new FormData(this);
  try {
    const res = await fetch('/admin/login', {method:'POST', body: fd});
    const data = await res.json();
    if (data.success) { location.href = data.redirect; }
    else { throw new Error(data.error || '登录失败'); }
  } catch(e) {
    if (e.message.includes('封禁')) { ban.style.display = 'block'; ban.textContent = e.message; }
    else { err.style.display = 'block'; err.textContent = e.message; }
    btn.disabled = false; btn.textContent = '登录';
  }
});
(function(){
  const theme = localStorage.getItem('admin-theme');
  if (theme === 'dark') document.body.classList.add('dark');
})();
</script>
</body>
</html>`

const adminHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>邮件服务后台</title>
<script src="https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js"></script>
<style>
` + sharedDarkCSS + `
.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(240px,1fr));gap:16px;margin-bottom:24px}
.card{background:var(--card-bg);border-radius:12px;padding:24px;box-shadow:0 2px 8px rgba(0,0,0,0.04);transition:background .3s}
.card .label{font-size:14px;color:var(--text-secondary);margin-bottom:8px}
.card .value{font-size:32px;font-weight:600;color:var(--text)}
.card .value.success{color:#52c41a}
.card .value.failed{color:#ff4d4f}
.card .value.rate{color:#1677ff}
.chart-row{display:grid;grid-template-columns:2fr 1fr;gap:16px}
@media(max-width:768px){.chart-row{grid-template-columns:1fr}}
@media(max-width:768px){.chart{height:260px}}
.chart-box{background:var(--card-bg);border-radius:12px;padding:20px;box-shadow:0 2px 8px rgba(0,0,0,0.04);transition:background .3s}
.chart-box h3{font-size:15px;font-weight:600;margin-bottom:16px;color:var(--text)}
.chart{height:320px}
.btn-primary{padding:8px 20px;background:#1677ff;color:#fff;border:none;border-radius:6px;font-size:14px;cursor:pointer;transition:background .2s}
.btn-primary:hover{background:#0958d9}
.table-box{background:var(--card-bg);border-radius:12px;padding:20px;box-shadow:0 2px 8px rgba(0,0,0,0.04);overflow-x:auto;transition:background .3s}
table{width:100%;border-collapse:collapse;font-size:14px}
th{text-align:left;padding:12px 16px;background:var(--table-header-bg);color:var(--text-secondary);font-weight:500;border-bottom:1px solid var(--border);transition:background .3s}
td{padding:12px 16px;border-bottom:1px solid var(--border);color:var(--text)}
tr:hover td{background:var(--hover-bg)}
.actions{display:flex;gap:8px}
.btn-small{padding:4px 12px;border-radius:4px;font-size:13px;cursor:pointer;border:none}
.btn-edit{background:#e6f4ff;color:#1677ff}
.btn-edit:hover{background:#bae0ff}
.btn-delete{background:#fff2f0;color:#ff4d4f}
.btn-delete:hover{background:#ffccc7}
.mask{position:fixed;top:0;left:0;right:0;bottom:0;background:var(--mask-bg);display:none;z-index:100;justify-content:center;align-items:center}
.modal{background:var(--card-bg);border-radius:12px;width:100%;max-width:480px;padding:24px;box-shadow:0 8px 32px rgba(0,0,0,0.15);transition:background .3s}
.modal h3{font-size:16px;font-weight:600;margin-bottom:20px;color:var(--text)}
.modal .form-group{margin-bottom:16px}
.modal label{display:block;margin-bottom:6px;font-size:14px;color:var(--text);font-weight:500}
.modal input{width:100%;padding:10px 14px;border:1px solid var(--input-border);border-radius:6px;font-size:14px;background:var(--input-bg);color:var(--text)}
.modal input:focus{outline:none;border-color:#1677ff}
.modal-actions{display:flex;justify-content:flex-end;gap:8px;margin-top:24px}
.modal-actions button{padding:8px 20px;border-radius:6px;font-size:14px;cursor:pointer;border:none}
.btn-cancel{background:var(--hover-bg);color:var(--text)}
.btn-cancel:hover{background:var(--border)}
.btn-save{background:#1677ff;color:#fff}
.btn-save:hover{background:#0958d9}
.log-table{font-size:13px}
.log-table th{white-space:nowrap}
.log-table td{padding:10px 16px;vertical-align:top}
.col-time{width:150px;white-space:nowrap}
.col-user{width:180px}
.col-subject{min-width:200px;max-width:300px}
.col-mailid{width:70px;text-align:center}
.col-ip{width:120px;white-space:nowrap}
.col-result{min-width:200px}
.col-duration{width:80px;text-align:center;white-space:nowrap}
.col-body{max-width:400px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:12px;color:var(--text-secondary)}
textarea{width:100%;padding:10px 14px;border:1px solid var(--input-border);border-radius:6px;font-size:14px;background:var(--input-bg);color:var(--text);font-family:inherit;resize:vertical;min-height:120px}
textarea:focus{outline:none;border-color:#1677ff}
.status-tag{display:inline-block;padding:2px 8px;border-radius:4px;font-size:12px;font-weight:500}
.search-bar{display:flex;gap:8px;align-items:center}
.search-bar input{padding:6px 12px;border:1px solid var(--input-border);border-radius:6px;font-size:14px;background:var(--input-bg);color:var(--text);width:260px;transition:border-color .2s;-webkit-appearance:none;appearance:none}
.search-bar input:focus{outline:none;border-color:#1677ff}
.search-bar select{padding:6px 12px;border:1px solid var(--input-border);border-radius:6px;font-size:14px;background:var(--input-bg);color:var(--text);cursor:pointer;transition:border-color .2s}
.search-bar select:focus{outline:none;border-color:#1677ff}
.status-success{background:#f6ffed;color:#52c41a;border:1px solid #b7eb8f}
.status-fail{background:#fff2f0;color:#ff4d4f;border:1px solid #ffccc7}
.toast{position:fixed;top:20px;right:20px;left:20px;padding:12px 20px;border-radius:8px;color:#fff;font-size:14px;z-index:200;box-shadow:0 4px 12px rgba(0,0,0,0.15);display:none;text-align:center}
.toast.success{background:#52c41a}
.toast.error{background:#ff4d4f}
.empty{text-align:center;padding:60px;color:var(--text-secondary)}
.view{display:none}
.view.active{display:block}
.footer{text-align:center;padding:16px;font-size:12px;color:var(--text-secondary);border-top:1px solid var(--border);margin-top:24px}
.pagination{display:flex;justify-content:center;align-items:center;gap:12px;margin-top:16px;font-size:14px}
.pagination button{padding:6px 14px;border:1px solid var(--input-border);border-radius:6px;background:var(--card-bg);color:var(--text);cursor:pointer;transition:all .2s}
.pagination button:hover:not(:disabled){border-color:#1677ff;color:#1677ff}
.pagination button:disabled{opacity:.4;cursor:not-allowed}
.pagination span{color:var(--text-secondary)}
.page-size-label{font-weight:500}
.page-size-group{display:flex;gap:2px;border:1px solid var(--input-border);border-radius:6px;overflow:hidden}
.page-size-btn{padding:5px 12px;font-size:13px;cursor:pointer;border-right:1px solid var(--input-border);color:var(--text-secondary);transition:all .2s;background:var(--card-bg)}
.page-size-btn:last-child{border-right:none}
.page-size-btn:hover{background:var(--hover-bg);color:var(--text)}
.page-size-btn.active{background:#1677ff;color:#fff}
.page-btn{padding:6px 14px;border:1px solid var(--input-border);border-radius:6px;background:var(--card-bg);color:var(--text);cursor:pointer;font-size:14px;transition:all .2s}
.page-btn:hover:not(:disabled){border-color:#1677ff;color:#1677ff}
.page-btn:disabled{opacity:.4;cursor:not-allowed}
@media(max-width:768px){
.container{padding:16px}
.header{grid-template-columns:1fr auto;padding:0 8px;height:52px;gap:0}
.header h1{display:none}
.nav{gap:0;justify-self:start;overflow-x:auto;flex-wrap:nowrap;-webkit-overflow-scrolling:touch;scrollbar-width:none}
.nav::-webkit-scrollbar{display:none}
.nav a{font-size:13px;padding:6px 8px;white-space:nowrap;flex-shrink:0}
.header-right{gap:6px}
.btn-theme{font-size:13px;padding:4px 8px}
.header a.logout{font-size:12px;padding:4px 8px}
.toolbar{flex-direction:column;align-items:flex-start;gap:8px}
.card{padding:16px}
.card .value{font-size:24px}
.table-box{padding:12px;overflow-x:auto}
.log-table{font-size:12px}
.log-table td{padding:8px 10px}
.col-time{width:auto}
.col-user{width:auto}
.col-subject{max-width:none}
.col-ip{width:auto}
.col-duration{width:auto}
.col-result{min-width:auto}
.col-body{max-width:none;white-space:normal}
.search-bar{width:100%;flex-wrap:wrap}
.search-bar input{width:100%;flex:1;min-width:100%;font-size:16px}
.search-bar select{flex:1;font-size:16px;min-width:80px}
.search-bar .btn-refresh{flex-shrink:0}
.modal{margin:0 16px;padding:20px;max-width:calc(100vw - 32px)}
#previewMask .modal{max-width:calc(100vw - 32px)}
.modal input{font-size:16px}
textarea{font-size:16px}
.modal-actions{flex-wrap:wrap;gap:8px}
.modal-actions button{flex:1;min-width:80px}
.toast{left:16px;right:16px;text-align:center}
.footer{padding:12px;margin-top:16px}
.pagination{font-size:13px;flex-wrap:wrap}
.pagination button,.pagination .page-btn{padding:5px 12px}
.page-size-btn{padding:4px 10px;font-size:12px}
.login-card input{font-size:16px}
}
</style>
</head>
<body>
<div class="header">
  <h1>邮件服务后台</h1>
  <nav class="nav">
    <a href="/admin/dashboard" data-view="dashboard">仪表盘</a>
    <a href="/admin/mails" data-view="mails">邮箱池</a>
    <a href="/admin/logs" data-view="logs">发信日志</a>
    <a href="/admin/templates" data-view="templates">模板管理</a>
    <a href="/admin/audit" data-view="audit">操作审计</a>
  </nav>
  <div class="header-right">
    <button class="btn-theme" id="themeBtn" onclick="toggleTheme()">🌙</button>
    <a href="/admin/logout" class="logout">退出登录</a>
  </div>
</div>

<div id="view-dashboard" class="view">
  <div class="container">
    <div class="toolbar">
      <h2>数据概览</h2>
      <div class="search-bar">
        <select id="daysSelect" onchange="changeDays()">
          <option value="1">近1天</option>
          <option value="7" selected>近7天</option>
          <option value="30">近30天</option>
        </select>
        <button class="btn-refresh" id="refreshBtn" onclick="refreshStats()">刷新</button>
      </div>
    </div>
    <div class="cards">
      <div class="card"><div class="label"><span id="totalLabel">7天总发信量</span></div><div class="value" id="total">-</div></div>
      <div class="card"><div class="label">发送成功</div><div class="value success" id="success">-</div></div>
      <div class="card"><div class="label">发送失败</div><div class="value failed" id="failed">-</div></div>
      <div class="card"><div class="label">成功率</div><div class="value rate" id="rate">-</div></div>
    </div>
    <div class="chart-row">
      <div class="chart-box">
        <h3>近7天发信趋势</h3>
        <div class="chart" id="trendChart"></div>
      </div>
      <div class="chart-box">
        <h3>邮箱后缀占比</h3>
        <div class="chart" id="pieChart"></div>
      </div>
    </div>
    <div class="chart-box" style="margin-top:16px">
      <h3>邮箱账号统计</h3>
      <table id="mailStatTable">
        <thead><tr><th>ID</th><th>发件人</th><th>域名</th><th>总发信</th><th>成功</th><th>失败</th><th>成功率</th></tr></thead>
        <tbody id="mailStatBody"><tr><td colspan="7" class="empty">加载中...</td></tr></tbody>
      </table>
    </div>
  </div>
</div>

<div id="view-mails" class="view">
  <div class="container">
    <div class="toolbar">
      <h2>邮箱池管理</h2>
      <button class="btn-primary" onclick="openModal()">+ 新增邮箱</button>
    </div>
    <div class="table-box">
      <table id="mailTable">
        <thead><tr><th>ID</th><th>Domain</th><th>Port</th><th>Sender</th><th>Password</th><th>操作</th></tr></thead>
        <tbody id="mailBody"><tr><td colspan="6" class="empty">加载中...</td></tr></tbody>
      </table>
    </div>
  </div>
  <div class="mask" id="mask" onclick="if(event.target===this)closeModal()">
    <div class="modal">
      <h3 id="modalTitle">新增邮箱</h3>
      <div class="form-group">
        <label>SMTP 域名</label>
        <input type="text" id="domain" placeholder="例如: smtp.gmail.com">
      </div>
      <div class="form-group">
        <label>端口</label>
        <input type="number" id="port" value="465" placeholder="例如: 465">
      </div>
      <div class="form-group">
        <label>发件人邮箱</label>
        <input type="text" id="sender" placeholder="例如: noreply@gmail.com">
      </div>
      <div class="form-group">
        <label>密码 / 授权码</label>
        <input type="text" id="password" placeholder="邮箱密码或授权码">
      </div>
      <div class="modal-actions">
        <button class="btn-cancel" onclick="closeModal()">取消</button>
        <button class="btn-small btn-edit" onclick="testMail()" id="testMailBtn" style="margin-right:auto">测试发送</button>
        <button class="btn-save" onclick="saveMail()">保存</button>
      </div>
    </div>
  </div>
</div>

<div id="view-logs" class="view">
  <div class="container">
    <div class="toolbar">
      <h2 id="logTitle">发信日志</h2>
      <div class="search-bar">
        <input type="text" id="logSearch" placeholder="搜索收件人、主题、IP、结果..." onkeydown="if(event.key==='Enter')debouncedRefreshLogs()">
        <select id="logStatus" onchange="refreshLogs()">
          <option value="">全部状态</option>
          <option value="success">成功</option>
          <option value="failed">失败</option>
        </select>
        <button class="btn-refresh" onclick="exportLogs()">导出 CSV</button>
        <button class="btn-refresh" id="refreshLogsBtn" onclick="refreshLogs()">刷新</button>
      </div>
    </div>
    <div class="table-box">
      <table class="log-table">
        <thead>
          <tr>
            <th class="col-time">时间</th>
            <th class="col-user">收件人</th>
            <th class="col-subject">主题</th>
            <th class="col-mailid">邮箱ID</th>
            <th class="col-ip">来源IP</th>
            <th class="col-duration">耗时</th>
            <th class="col-result">结果</th>
          </tr>
        </thead>
        <tbody id="logBody"><tr><td colspan="7" class="empty">加载中...</td></tr></tbody>
      </table>
    </div>
    <div class="pagination" id="logPagination" style="display:none">
      <span class="page-size-label">每页</span>
      <span class="page-size-group" id="logLimitGroup">
        <span class="page-size-btn" onclick="setLimit(10)">10</span>
        <span class="page-size-btn active" onclick="setLimit(20)">20</span>
        <span class="page-size-btn" onclick="setLimit(50)">50</span>
        <span class="page-size-btn" onclick="setLimit(100)">100</span>
      </span>
      <span id="logPageInfo"></span>
      <button class="page-btn" onclick="prevPage()" id="prevBtn">上一页</button>
      <button class="page-btn" onclick="nextPage()" id="nextBtn">下一页</button>
    </div>
  </div>
</div>

<div id="view-templates" class="view">
  <div class="container">
    <div class="toolbar">
      <h2>邮件模板管理</h2>
      <button class="btn-primary" onclick="openTemplateModal()">+ 新增模板</button>
    </div>
    <div class="table-box">
      <table id="templateTable">
        <thead><tr><th>ID</th><th>名称</th><th>主题</th><th class="col-body">正文预览</th><th>操作</th></tr></thead>
        <tbody id="templateBody"><tr><td colspan="5" class="empty">加载中...</td></tr></tbody>
      </table>
    </div>
  </div>
  <div class="mask" id="templateMask" onclick="if(event.target===this)closeTemplateModal()">
    <div class="modal" style="max-width:640px">
      <h3 id="templateModalTitle">新增模板</h3>
      <div class="form-group">
        <label>模板名称</label>
        <input type="text" id="tplName" placeholder="例如：验证码邮件">
      </div>
      <div class="form-group">
        <label>邮件主题</label>
        <input type="text" id="tplSubject" placeholder="例如：您的验证码是 {{code}}">
      </div>
      <div class="form-group">
        <label>邮件正文（HTML）</label>
        <textarea id="tplBody" rows="8" placeholder="支持变量占位符，如 {{code}} {{username}}"></textarea>
      </div>
      <div class="modal-actions">
        <button class="btn-cancel" onclick="closeTemplateModal()">取消</button>
        <button class="btn-save" onclick="saveTemplate()">保存</button>
      </div>
    </div>
  </div>
</div>

<div class="mask" id="previewMask" onclick="closePreview()">
  <div class="modal" style="max-width:800px;max-height:80vh;overflow:auto">
    <h3 id="previewTitle">模板预览</h3>
    <div id="previewContent" style="border:1px solid var(--border);border-radius:8px;padding:16px;background:#fff"></div>
    <div class="modal-actions"><button class="btn-cancel" onclick="closePreview()">关闭</button></div>
  </div>
</div>

<div class="toast" id="toast"></div>

<script>
let trendChart, pieChart, currentView = '';
let currentDays = 7;
let dashboardLoaded = false, mailsLoaded = false, logsLoaded = false, templatesLoaded = false, auditLoaded = false;
let auditOffset = 0;
const auditLimit = 20;
const viewTitles = {dashboard: '仪表盘', mails: '邮箱池', logs: '发信日志', templates: '模板管理', audit: '操作审计'};
const isDark = () => document.body.classList.contains('dark');
const chartTheme = () => isDark() ? 'dark' : undefined;

function toggleTheme() {
  const body = document.body;
  const btn = document.getElementById('themeBtn');
  if (body.classList.contains('dark')) {
    body.classList.remove('dark');
    localStorage.setItem('admin-theme', 'light');
    btn.textContent = '🌙';
  } else {
    body.classList.add('dark');
    localStorage.setItem('admin-theme', 'dark');
    btn.textContent = '☀️';
  }
  if (trendChart) { trendChart.dispose(); trendChart = null; }
  if (pieChart) { pieChart.dispose(); pieChart = null; }
  if (currentView === 'dashboard') {
    initDashboardCharts();
    const cached = sessionStorage.getItem('dashboardStats');
    if (cached) { renderStats(JSON.parse(cached)); loadMailStats(); }
  }
}

function initTheme() {
  const theme = localStorage.getItem('admin-theme');
  if (theme === 'dark') {
    document.body.classList.add('dark');
    document.getElementById('themeBtn').textContent = '☀️';
  }
}

function initNav() {
  document.querySelectorAll('.nav a').forEach(a => {
    a.addEventListener('click', e => {
      e.preventDefault();
      const view = a.dataset.view;
      if (view === currentView) return;
      showView(view);
    });
  });
}

function getViewFromPath() {
  const path = location.pathname;
  if (path === '/admin/mails') return 'mails';
  if (path === '/admin/logs') return 'logs';
  if (path === '/admin/templates') return 'templates';
  if (path === '/admin/audit') return 'audit';
  return 'dashboard';
}

function showView(name, pushState) {
  pushState = pushState !== false;
  document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
  document.getElementById('view-' + name).classList.add('active');
  document.querySelectorAll('.nav a').forEach(a => {
    a.classList.toggle('active', a.dataset.view === name);
  });
  if (pushState) history.pushState({view: name}, '', '/admin/' + name);
  document.title = '邮件服务后台 - ' + viewTitles[name];
  currentView = name;

  if (name === 'dashboard') {
    if (!trendChart) initDashboardCharts();
    else { setTimeout(() => { if (trendChart) trendChart.resize(); if (pieChart) pieChart.resize(); }, 50); }
    const cached = sessionStorage.getItem('dashboardStats');
    if (cached) { renderStats(JSON.parse(cached)); loadMailStats(); }
    else if (!dashboardLoaded) loadStats();
    dashboardLoaded = true;
    initSSE();
  } else if (name === 'mails') {
    if (!mailsLoaded) { loadMails(); mailsLoaded = true; }
  } else if (name === 'logs') {
    if (!logsLoaded) { loadLogs(); logsLoaded = true; }
    initSSE();
  } else if (name === 'templates') {
    if (!templatesLoaded) { loadTemplates(); templatesLoaded = true; }
  } else if (name === 'audit') {
    if (!auditLoaded) { loadAuditLogs(); auditLoaded = true; }
  }
}

let eventSource = null;

function initSSE() {
  if (eventSource || !window.EventSource) return;
  eventSource = new EventSource('/admin/api/events');
  eventSource.onmessage = function(e) {
    const data = JSON.parse(e.data);
    if (currentView === 'dashboard') {
      debouncedLoadStats();
    } else if (currentView === 'logs' && logOffset === 0) {
      prependLog(data);
    }
  };
  eventSource.onerror = function() {
    // 浏览器会自动重连，无需处理
  };
}

function prependLog(l) {
  const tbody = document.getElementById('logBody');
  if (!tbody) return;
  const emptyRow = tbody.querySelector('.empty');
  if (emptyRow) {
    tbody.innerHTML = '';
  }
  const row = document.createElement('tr');
  row.innerHTML = '<td class="col-time">' + formatTime(l.time_unix) + '</td>' +
    '<td class="col-user">' + escapeHtml(l.user || '-') + '</td>' +
    '<td class="col-subject">' + escapeHtml(l.subject || '-') + '</td>' +
    '<td class="col-mailid">' + l.mail_id + '</td>' +
    '<td class="col-ip">' + escapeHtml(l.ip || '-') + '</td>' +
    '<td class="col-duration">' + (l.duration_ms != null ? l.duration_ms + 'ms' : '-') + '</td>' +
    '<td class="col-result">' +
      '<span class="status-tag ' + getStatusClass(l.result) + '">' + getStatusText(l.result) + '</span> ' +
      '<span style="color:var(--text-secondary)">' + escapeHtml(l.result) + '</span>' +
    '</td>';
  tbody.insertBefore(row, tbody.firstChild);
  while (tbody.children.length > logLimit) {
    tbody.removeChild(tbody.lastChild);
  }
}

window.addEventListener('popstate', (e) => {
  const view = (e.state && e.state.view) || getViewFromPath();
  if (view && view !== currentView) showView(view, false);
});

function initDashboardCharts() {
  trendChart = echarts.init(document.getElementById('trendChart'), chartTheme());
  pieChart = echarts.init(document.getElementById('pieChart'), chartTheme());
  window.addEventListener('resize', () => {
    if (trendChart) trendChart.resize();
    if (pieChart) pieChart.resize();
  });
}

function renderStats(data) {
  document.getElementById('totalLabel').textContent = currentDays + '天总发信量';
  document.getElementById('total').textContent = data.total;
  document.getElementById('success').textContent = data.success;
  document.getElementById('failed').textContent = data.failed;
  document.getElementById('rate').textContent = data.successRate + '%';

  const dates = data.dailyStats.map(d => d.date.slice(5));
  const totals = data.dailyStats.map(d => d.total);
  const successes = data.dailyStats.map(d => d.success);
  const failures = data.dailyStats.map(d => d.failed);

  const textColor = isDark() ? '#ccc' : '#333';
  const axisColor = isDark() ? '#555' : '#ccc';

  trendChart.setOption({
    tooltip: { trigger: 'axis', backgroundColor: isDark()?'#1f1f1f':'#fff', borderColor: isDark()?'#333':'#ddd', textStyle:{color:textColor} },
    legend: { data: ['总发信','成功','失败'], bottom: 0, textStyle:{color:textColor} },
    grid: { left: '3%', right: '4%', bottom: '15%', containLabel: true },
    xAxis: { type: 'category', data: dates.reverse(), boundaryGap: false, axisLine:{lineStyle:{color:axisColor}}, axisLabel:{color:textColor} },
    yAxis: { type: 'value', minInterval: 1, axisLine:{lineStyle:{color:axisColor}}, axisLabel:{color:textColor}, splitLine:{lineStyle:{color:isDark()?'#333':'#eee'}} },
    series: [
      { name:'总发信', type:'line', data:totals.reverse(), smooth:true, areaStyle:{opacity:.1}, lineStyle:{color:'#1677ff'}, itemStyle:{color:'#1677ff'} },
      { name:'成功', type:'line', data:successes.reverse(), smooth:true, lineStyle:{color:'#52c41a'}, itemStyle:{color:'#52c41a'} },
      { name:'失败', type:'line', data:failures.reverse(), smooth:true, lineStyle:{color:'#ff4d4f'}, itemStyle:{color:'#ff4d4f'} }
    ]
  }, true);

  document.querySelector('.chart-box h3').textContent = '近' + currentDays + '天发信趋势';

  const isMobile = window.innerWidth <= 768;

  const pieData = data.domainStats.map(d => ({ value: d.count, name: d.domain }));
  pieChart.setOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)', backgroundColor: isDark()?'#1f1f1f':'#fff', borderColor: isDark()?'#333':'#ddd', textStyle:{color:textColor} },
    legend: { type: 'scroll', orient: isMobile ? 'horizontal' : 'vertical', right: isMobile ? 'center' : 10, top: isMobile ? 'bottom' : 'center', left: isMobile ? 0 : 'auto', textStyle:{color:textColor} },
    series: [{
      type: 'pie', radius: isMobile ? ['30%', '45%'] : ['35%', '55%'], center: isMobile ? ['50%', '42%'] : ['38%', '50%'],
      avoidLabelOverlap: true, itemStyle: { borderRadius: 6, borderColor: isDark()?'#1f1f1f':'#fff', borderWidth: 2 },
      label: { show: false },
      data: pieData
    }]
  });
}

async function loadStats() {
  try {
    const res = await fetch('/admin/api/stats?days=' + currentDays);
    if (!res.ok) throw new Error('加载失败');
    const data = await res.json();
    sessionStorage.setItem('dashboardStats', JSON.stringify(data));
    renderStats(data);
    loadMailStats();
  } catch(e) {
    console.error(e);
  }
}

async function loadMailStats() {
  try {
    const res = await fetch('/admin/api/mail-stats?days=' + currentDays);
    if (!res.ok) throw new Error('加载失败');
    const stats = await res.json();
    const tbody = document.getElementById('mailStatBody');
    if (stats.length === 0) {
      tbody.innerHTML = '<tr><td colspan="7" class="empty">暂无邮箱配置</td></tr>';
      return;
    }
    tbody.innerHTML = stats.map(function(s) {
      var rateClass = s.rate === '-' ? '' : (parseFloat(s.rate) >= 90 ? 'color:#52c41a' : (parseFloat(s.rate) < 50 ? 'color:#ff4d4f' : ''));
      return '<tr>' +
        '<td>' + s.id + '</td>' +
        '<td>' + escapeHtml(s.sender) + '</td>' +
        '<td>' + escapeHtml(s.domain) + '</td>' +
        '<td>' + s.total + '</td>' +
        '<td style="color:#52c41a">' + s.success + '</td>' +
        '<td style="color:#ff4d4f">' + s.failed + '</td>' +
        '<td style="font-weight:600;' + rateClass + '">' + s.rate + '</td>' +
      '</tr>';
    }).join('');
  } catch(e) {
    document.getElementById('mailStatBody').innerHTML = '<tr><td colspan="7" class="empty">加载失败</td></tr>';
  }
}

async function refreshStats() {
  const btn = document.getElementById('refreshBtn');
  btn.disabled = true; btn.textContent = '刷新中...';
  await loadStats();
  btn.disabled = false; btn.textContent = '刷新';
}

const debouncedLoadStats = debounce(loadStats, 800);

let editingId = null;
async function loadMails() {
  try {
    const res = await fetch('/admin/api/mails');
    if (!res.ok) throw new Error('加载失败');
    const mails = await res.json();
    const tbody = document.getElementById('mailBody');
    if (mails.length === 0) {
      tbody.innerHTML = '<tr><td colspan="6" class="empty">暂无邮箱配置</td></tr>';
      return;
    }
    tbody.innerHTML = mails.map(function(m) {
      return '<tr>' +
        '<td>' + m.id + '</td>' +
        '<td>' + escapeHtml(m.domain) + '</td>' +
        '<td>' + m.port + '</td>' +
        '<td>' + escapeHtml(m.sender) + '</td>' +
        '<td>' + escapeHtml(m.password) + '</td>' +
        '<td class="actions">' +
          '<button class="btn-small btn-edit" onclick="editMail(' + m.id + ',\'' + escapeHtml(m.domain).replace(/'/g, "\\'") + '\',' + m.port + ',\'' + escapeHtml(m.sender).replace(/'/g, "\\'") + '\',\'' + escapeHtml(m.password).replace(/'/g, "\\'") + '\')">编辑</button>' +
          '<button class="btn-small btn-delete" onclick="deleteMail(' + m.id + ')">删除</button>' +
        '</td>' +
      '</tr>';
    }).join('');
  } catch(e) { showToast(e.message, 'error'); }
}

function openModal() {
  editingId = null;
  document.getElementById('modalTitle').textContent = '新增邮箱';
  document.getElementById('domain').value = '';
  document.getElementById('port').value = '465';
  document.getElementById('sender').value = '';
  document.getElementById('password').value = '';
  document.getElementById('mask').style.display = 'flex';
}

function editMail(id, domain, port, sender, password) {
  editingId = id;
  document.getElementById('modalTitle').textContent = '编辑邮箱';
  document.getElementById('domain').value = domain;
  document.getElementById('port').value = port;
  document.getElementById('sender').value = sender;
  document.getElementById('password').value = password;
  document.getElementById('mask').style.display = 'flex';
}

function closeModal() {
  document.getElementById('mask').style.display = 'none';
}

async function testMail() {
  if (!editingId) { showToast('请先保存邮箱配置再测试', 'error'); return; }
  const btn = document.getElementById('testMailBtn');
  btn.disabled = true; btn.textContent = '发送中...';
  try {
    const res = await fetch('/admin/api/mails/' + editingId + '/test', {method: 'POST'});
    const data = await res.json();
    if (data.success) { showToast(data.message, 'success'); }
    else { throw new Error(data.message || '测试失败'); }
  } catch(e) { showToast(e.message, 'error'); }
  btn.disabled = false; btn.textContent = '测试发送';
}

async function saveMail() {
  const data = {
    domain: document.getElementById('domain').value.trim(),
    port: parseInt(document.getElementById('port').value) || 465,
    sender: document.getElementById('sender').value.trim(),
    password: document.getElementById('password').value.trim()
  };
  if (!data.domain || !data.sender || !data.password) {
    showToast('请填写完整信息', 'error'); return;
  }
  try {
    const url = editingId ? '/admin/api/mails/' + editingId : '/admin/api/mails';
    const method = editingId ? 'PUT' : 'POST';
    const res = await fetch(url, {method, headers:{'Content-Type':'application/json'}, body: JSON.stringify(data)});
    if (!res.ok) { const err = await res.json(); throw new Error(err.error || '保存失败'); }
    showToast(editingId ? '修改成功' : '添加成功', 'success');
    closeModal();
    loadMails();
  } catch(e) { showToast(e.message, 'error'); }
}

async function deleteMail(id) {
  if (!confirm('确定要删除该邮箱配置吗？')) return;
  try {
    const res = await fetch('/admin/api/mails/' + id, {method: 'DELETE'});
    if (!res.ok) throw new Error('删除失败');
    showToast('删除成功', 'success');
    loadMails();
  } catch(e) { showToast(e.message, 'error'); }
}

function escapeHtml(text) {
  if (text == null) return '';
  return String(text).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;');
}

function formatTime(ts) {
  const d = new Date(ts * 1000);
  const pad = n => String(n).padStart(2, '0');
  return d.getFullYear() + '-' + pad(d.getMonth()+1) + '-' + pad(d.getDate()) + ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds());
}

function getStatusClass(result) {
  return result.includes('Message has been sent') ? 'status-success' : 'status-fail';
}

function getStatusText(result) {
  return result.includes('Message has been sent') ? '成功' : '失败';
}

let logOffset = 0;
let logLimit = 20;

function setLimit(limit) {
  logLimit = limit;
  document.querySelectorAll('.page-size-btn').forEach(btn => {
    btn.classList.toggle('active', btn.textContent.trim() === String(limit));
  });
  logOffset = 0;
  loadLogs();
}

function initLogLimit() {
  document.querySelectorAll('.page-size-btn').forEach(btn => {
    btn.classList.toggle('active', btn.textContent.trim() === String(logLimit));
  });
}

function debounce(fn, ms) {
  let timer;
  return function() {
    clearTimeout(timer);
    timer = setTimeout(fn, ms);
  };
}
const debouncedRefreshLogs = debounce(refreshLogs, 300);

async function refreshLogs() {
  logOffset = 0;
  const btn = document.getElementById('refreshLogsBtn');
  btn.disabled = true; btn.textContent = '刷新中...';
  await loadLogs();
  btn.disabled = false; btn.textContent = '刷新';
}

async function loadLogs() {
  try {
    const keyword = document.getElementById('logSearch').value.trim();
    const status = document.getElementById('logStatus').value;
    let url = '/admin/api/logs?limit=' + logLimit + '&offset=' + logOffset;
    if (keyword) url += '&keyword=' + encodeURIComponent(keyword);
    if (status) url += '&status=' + encodeURIComponent(status);
    const res = await fetch(url);
    if (!res.ok) throw new Error('加载失败');
    const logs = await res.json();

    // 获取总数
    let countUrl = '/admin/api/logs-count?';
    if (keyword) countUrl += 'keyword=' + encodeURIComponent(keyword) + '&';
    if (status) countUrl += 'status=' + encodeURIComponent(status);
    const countRes = await fetch(countUrl);
    const countData = countRes.ok ? await countRes.json() : {total: 0};
    const total = countData.total || 0;
    const totalPages = Math.max(1, Math.ceil(total / logLimit));
    const currentPage = logOffset / logLimit + 1;

    const tbody = document.getElementById('logBody');
    const pagination = document.getElementById('logPagination');
    if (logs.length === 0) {
      if (logOffset === 0) {
        tbody.innerHTML = '<tr><td colspan="7" class="empty">暂无日志记录</td></tr>';
      } else {
        tbody.innerHTML = '<tr><td colspan="7" class="empty">没有更多记录了</td></tr>';
      }
      pagination.style.display = 'none';
      return;
    }
    tbody.innerHTML = logs.map(function(l) {
      return '<tr>' +
        '<td class="col-time">' + formatTime(l.time_unix) + '</td>' +
        '<td class="col-user">' + escapeHtml(l.user || '-') + '</td>' +
        '<td class="col-subject">' + escapeHtml(l.subject || '-') + '</td>' +
        '<td class="col-mailid">' + l.mail_id + '</td>' +
        '<td class="col-ip">' + escapeHtml(l.ip || '-') + '</td>' +
        '<td class="col-duration">' + (l.duration_ms != null ? l.duration_ms + 'ms' : '-') + '</td>' +
        '<td class="col-result">' +
          '<span class="status-tag ' + getStatusClass(l.result) + '">' + getStatusText(l.result) + '</span> ' +
          '<span style="color:var(--text-secondary)">' + escapeHtml(l.result) + '</span>' +
        '</td>' +
      '</tr>';
    }).join('');
    document.getElementById('logTitle').textContent = '发信日志';
    document.getElementById('logPageInfo').textContent = '共 ' + total + ' 条记录，' + currentPage + '/' + totalPages + ' 页';
    document.getElementById('prevBtn').disabled = logOffset === 0;
    document.getElementById('nextBtn').disabled = logs.length < logLimit;
    pagination.style.display = 'flex';
  } catch(e) {
    showToast(e.message, 'error');
    document.getElementById('logBody').innerHTML = '<tr><td colspan="7" class="empty">加载失败</td></tr>';
    document.getElementById('logPagination').style.display = 'none';
  }
}

function prevPage() {
  if (logOffset >= logLimit) {
    logOffset -= logLimit;
    loadLogs();
  }
}

function nextPage() {
  logOffset += logLimit;
  loadLogs();
}

function showToast(msg, type) {
  const t = document.getElementById('toast');
  t.textContent = msg; t.className = 'toast ' + type; t.style.display = 'block';
  setTimeout(() => t.style.display = 'none', 3000);
}

async function refreshAuditLogs() {
  auditOffset = 0;
  const btn = document.getElementById('refreshAuditBtn');
  btn.disabled = true; btn.textContent = '刷新中...';
  await loadAuditLogs();
  btn.disabled = false; btn.textContent = '刷新';
}

let editingTemplateId = null;

async function loadTemplates() {
  try {
    const res = await fetch('/admin/api/templates');
    if (!res.ok) throw new Error('加载失败');
    const templates = await res.json();
    const tbody = document.getElementById('templateBody');
    if (templates.length === 0) {
      tbody.innerHTML = '<tr><td colspan="5" class="empty">暂无模板</td></tr>';
      return;
    }
    tbody.innerHTML = templates.map(function(t) {
      var preview = t.body ? t.body.replace(/<[^>]+>/g, '').substring(0, 80) + (t.body.length > 80 ? '...' : '') : '-';
      return '<tr>' +
        '<td>' + t.id + '</td>' +
        '<td>' + escapeHtml(t.name) + '</td>' +
        '<td>' + escapeHtml(t.subject) + '</td>' +
        '<td class="col-body" title="' + escapeHtml(t.body).replace(/"/g, '&quot;') + '">' + escapeHtml(preview) + '</td>' +
        '<td class="actions">' +
          '<button class="btn-small btn-edit" onclick="previewTemplate(' + t.id + ')">预览</button> ' +
          '<button class="btn-small btn-edit" onclick="editTemplate(' + t.id + ')">编辑</button>' +
          '<button class="btn-small btn-delete" onclick="deleteTemplate(' + t.id + ')">删除</button>' +
        '</td>' +
      '</tr>';
    }).join('');
  } catch(e) { showToast(e.message, 'error'); }
}

function openTemplateModal() {
  editingTemplateId = null;
  document.getElementById('templateModalTitle').textContent = '新增模板';
  document.getElementById('tplName').value = '';
  document.getElementById('tplSubject').value = '';
  document.getElementById('tplBody').value = '';
  document.getElementById('templateMask').style.display = 'flex';
}

async function editTemplate(id) {
  try {
    const res = await fetch('/admin/api/templates/' + id);
    if (!res.ok) throw new Error('加载失败');
    const tpl = await res.json();
    editingTemplateId = tpl.id;
    document.getElementById('templateModalTitle').textContent = '编辑模板';
    document.getElementById('tplName').value = tpl.name;
    document.getElementById('tplSubject').value = tpl.subject;
    document.getElementById('tplBody').value = tpl.body;
    document.getElementById('templateMask').style.display = 'flex';
  } catch(e) { showToast(e.message, 'error'); }
}

function closeTemplateModal() {
  document.getElementById('templateMask').style.display = 'none';
}

async function saveTemplate() {
  const data = {
    name: document.getElementById('tplName').value.trim(),
    subject: document.getElementById('tplSubject').value.trim(),
    body: document.getElementById('tplBody').value
  };
  if (!data.name || !data.subject || !data.body) {
    showToast('请填写完整信息', 'error'); return;
  }
  try {
    const url = editingTemplateId ? '/admin/api/templates/' + editingTemplateId : '/admin/api/templates';
    const method = editingTemplateId ? 'PUT' : 'POST';
    const res = await fetch(url, {method, headers:{'Content-Type':'application/json'}, body: JSON.stringify(data)});
    if (!res.ok) { const err = await res.json(); throw new Error(err.error || '保存失败'); }
    showToast(editingTemplateId ? '修改成功' : '添加成功', 'success');
    closeTemplateModal();
    loadTemplates();
  } catch(e) { showToast(e.message, 'error'); }
}

async function deleteTemplate(id) {
  if (!confirm('确定要删除该模板吗？')) return;
  try {
    const res = await fetch('/admin/api/templates/' + id, {method: 'DELETE'});
    if (!res.ok) throw new Error('删除失败');
    showToast('删除成功', 'success');
    loadTemplates();
  } catch(e) { showToast(e.message, 'error'); }
}

function changeDays() {
  currentDays = parseInt(document.getElementById('daysSelect').value);
  sessionStorage.removeItem('dashboardStats');
  dashboardLoaded = false;
  if (currentView === 'dashboard') loadStats();
  document.getElementById('mailStatBody').innerHTML = '<tr><td colspan="7" class="empty">加载中...</td></tr>';
}

async function previewTemplate(id) {
  try {
    const res = await fetch('/admin/api/templates');
    if (!res.ok) throw new Error('加载失败');
    const templates = await res.json();
    var tpl = null;
    for (var i = 0; i < templates.length; i++) {
      if (templates[i].id === id) { tpl = templates[i]; break; }
    }
    if (!tpl) { showToast('模板不存在', 'error'); return; }
    document.getElementById('previewTitle').textContent = '模板预览 - ' + tpl.name;
    document.getElementById('previewContent').innerHTML = tpl.body;
    document.getElementById('previewMask').style.display = 'flex';
  } catch(e) { showToast(e.message, 'error'); }
}

function closePreview() {
  document.getElementById('previewMask').style.display = 'none';
  document.getElementById('previewContent').innerHTML = '';
}

function exportLogs() {
  const kw = document.getElementById('logSearch').value.trim();
  const st = document.getElementById('logStatus').value;
  let url = '/admin/api/export-logs?';
  if (kw) url += 'keyword=' + encodeURIComponent(kw) + '&';
  if (st) url += 'status=' + encodeURIComponent(st);
  window.open(url, '_blank');
}

async function loadAuditLogs() {
  try {
    let url = '/admin/api/audit-logs?limit=' + auditLimit + '&offset=' + auditOffset;
    const res = await fetch(url);
    if (!res.ok) throw new Error('加载失败');
    const logs = await res.json();
    const tbody = document.getElementById('auditBody');
    const pagination = document.getElementById('auditPagination');
    if (logs.length === 0) {
      if (auditOffset === 0) {
        tbody.innerHTML = '<tr><td colspan="7" class="empty">暂无审计记录</td></tr>';
      } else {
        tbody.innerHTML = '<tr><td colspan="7" class="empty">没有更多记录了</td></tr>';
      }
      pagination.style.display = 'none';
      return;
    }
    tbody.innerHTML = logs.map(function(l) {
      return '<tr>' +
        '<td class="col-time">' + formatTime(l.time_unix) + '</td>' +
        '<td>' + escapeHtml(l.username || '-') + '</td>' +
        '<td>' + escapeHtml(l.ip || '-') + '</td>' +
        '<td>' + escapeHtml(l.action) + '</td>' +
        '<td>' + escapeHtml(l.target || '-') + '</td>' +
        '<td>' + escapeHtml(l.detail || '-') + '</td>' +
        '<td><span class="status-tag ' + (l.result === '成功' ? 'status-success' : (l.result === '失败' || l.result === '封禁' ? 'status-fail' : 'status-success')) + '">' + escapeHtml(l.result) + '</span></td>' +
      '</tr>';
    }).join('');
    document.getElementById('auditTitle').textContent = '操作审计';
    document.getElementById('auditPageInfo').textContent = '第 ' + (auditOffset / auditLimit + 1) + ' 页';
    document.getElementById('auditPrevBtn').disabled = auditOffset === 0;
    document.getElementById('auditNextBtn').disabled = logs.length < auditLimit;
    pagination.style.display = 'flex';
  } catch(e) {
    showToast(e.message, 'error');
    document.getElementById('auditBody').innerHTML = '<tr><td colspan="7" class="empty">加载失败</td></tr>';
    document.getElementById('auditPagination').style.display = 'none';
  }
}

function prevAuditPage() {
  if (auditOffset >= auditLimit) {
    auditOffset -= auditLimit;
    loadAuditLogs();
  }
}

function nextAuditPage() {
  auditOffset += auditLimit;
  loadAuditLogs();
}

async function loadVersion() {
  try {
    const res = await fetch('/admin/api/version');
    const data = await res.json();
    document.getElementById('version').textContent = data.version || '-';
  } catch(e) {}
}

(function init() {
  initTheme();
  initNav();
  initLogLimit();
  const view = getViewFromPath();
  showView(view);
  loadVersion();
})();
</script>
<div id="view-audit" class="view">
  <div class="container">
    <div class="toolbar">
      <h2 id="auditTitle">操作审计</h2>
      <button class="btn-refresh" id="refreshAuditBtn" onclick="refreshAuditLogs()">刷新</button>
    </div>
    <div class="table-box">
      <table class="log-table">
        <thead>
          <tr>
            <th class="col-time">时间</th>
            <th>操作人</th>
            <th>来源IP</th>
            <th>操作</th>
            <th>目标</th>
            <th>详情</th>
            <th>结果</th>
          </tr>
        </thead>
        <tbody id="auditBody"><tr><td colspan="7" class="empty">加载中...</td></tr></tbody>
      </table>
    </div>
    <div class="pagination" id="auditPagination" style="display:none">
      <button class="page-btn" onclick="prevAuditPage()" id="auditPrevBtn">上一页</button>
      <span id="auditPageInfo">第 1 页</span>
      <button class="page-btn" onclick="nextAuditPage()" id="auditNextBtn">下一页</button>
    </div>
  </div>
</div>

<div class="footer">版本: <span id="version">-</span></div>
</body>
</html>`
