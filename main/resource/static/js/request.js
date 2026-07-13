var postjson = function (posturl, postdata, success, fail, btn) {
    var times = 5;
    var interval;
    if (btn != undefined) {
      btn.disabled = true;
    }
    var xhr = new XMLHttpRequest();
    xhr.open('post', posturl, true);
    xhr.setRequestHeader('Content-Type', 'application/json');
    xhr.send(JSON.stringify(postdata));
    xhr.onreadystatechange = function () {
      // btn.innerText = "提交onreadystatechange"
      if (xhr.readyState === 4) {
        if (xhr.status === 200) {
          dt = JSON.parse(xhr.responseText)
          if (dt.code == 200) {
            interval = setInterval(function () {
              if (times > 0) {
                if (btn != undefined) {
                  btn.disabled = true;
                }
                times--;
              } else {
                clearInterval(interval);
                if (btn != undefined) {
                  btn.disabled = false;
                }
              }
            }, 1000);
            success(dt)
            // layer.msg(dt.msg);
          } else {
            if (btn != undefined) {
              btn.disabled = false;
            }
            fail(dt)
          }
        } else {
          if (btn != undefined) {
            btn.disabled = false;
          }
        }
      }
    };
  }

var getjson = function (geturl, success, fail) {
    var xhr = new XMLHttpRequest();
    xhr.open('get', geturl, true);
    xhr.send();
    xhr.onreadystatechange = function () {
      if (xhr.readyState === 4) {
        if (xhr.status === 200) {
          var dt = JSON.parse(xhr.responseText);
          if (dt.code == 0 || dt.code == 200) {
            success(dt);
          } else {
            if (fail) fail(dt);
          }
        } else {
          if (fail) fail({code: xhr.status, msg: xhr.statusText});
        }
      }
    };
  }
