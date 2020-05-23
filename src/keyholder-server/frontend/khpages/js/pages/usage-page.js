$(function() {
    getKeys(completion);
})

// GET KEYS
function getKeys(callback) {
    // 获取文档内容
    var httpRequest = new XMLHttpRequest();//第一步：建立所需的对象
    httpRequest.open('GET', '/db/query?q=SELECT%20*%20FROM%20KEYS', true);//第二步：打开连接  将请求参数写在url中  ps:"./Ptest.php?name=test&nameone=testone"
    httpRequest.send();//第三步：发送请求  将请求参数写在URL中
    /* 获取数据后的处理程序 */
    httpRequest.onreadystatechange = () => {
        if (httpRequest.readyState == 4 && httpRequest.status == 200) {
            str = httpRequest.responseText//获取到json字符串，还需解析
            callback(JSON.parse(str))
        } else if (httpRequest.status == 503) {
            swal("Oops", "Internal database server was not connectable ", "error");
        }
    };
}

// 获取密钥之后要做的事情
function completion(obj) {
    // 表身
    var tbody = document.createElement("tbody");
    // 拿到查询到的所有session信息，并渲染到表中
    keys = obj["results"][0]["values"];
    // 表尾统计 （PART 1）
    tfoot = document.createElement("tfoot");
    tfootrow = document.createElement("tr");
    tfootTotalLabel = document.createElement("th");
    tfootTotalLabel.innerHTML = "<strong>TOTAL</strong>";
    // 仅当有数据的时候才印入表格
    if (keys != undefined) {
        keys.forEach(element => {
            var tr = document.createElement("tr");
            element.splice(4, 1)
            element.forEach(col => {
                var td = document.createElement("td");
                td.innerHTML = col;
                tr.appendChild(td);
            })
            tbody.appendChild(tr)
            tfootTotalLabel.innerHTML = "<strong>TOTAL: " + keys.length + " Records</strong>";
        });
        document.getElementById("mainTable")
            .appendChild(tbody);
        // 表尾统计 （PART 2）
    } else {
        tfootTotalLabel.innerHTML = "<strong>TOTAL: 0 Records</strong>";
    }
    // 表尾统计 （PART 3）
    tfootrow.appendChild(tfootTotalLabel);
    tfoot.appendChild(tfootrow);
    document.getElementById("mainTable")
        .appendChild(tfoot);
}

// 点击删除之后要做的事情
function deleteKey() {
    swal({
        title: "Delete",
        text: "Type KHID whose record is to be remove",
        type: "input",
        showCancelButton: true,
        closeOnConfirm: false,
        animation: "slide-from-top",
        inputPlaceholder: "KHID"
    }, function (inputValue) {
        if (inputValue === false) return false;
        if (inputValue === "") {
            swal.showInputError("The KHID field may not be nil"); return false
        }
        // 令用户确认是否需要删除
        swal({
            title: "Confirm",
            text: "Continue to remove this KHID?",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "YES",
            cancelButtonText: "NO",
            closeOnConfirm: false,
            closeOnCancel: false
        }, function (isConfirm) {
            if (isConfirm) {
                performDelete(inputValue);
            }
        });
    });
}

// 删除
function performDelete(KHID) {
    datastr = '["DELETE FROM KEYS WHERE KEY_NUM=\'' + KHID + '\'"]';
    // 发送ajax
    $.ajax({
        type: "POST",
        contentType: "json", //发给服务器的数据类型
        dataType: "json", //预期服务器返回的数据类型
        url: "/db/execute",//url
        data: datastr,
        success: function (result) {
            // 检视返回的消息
            if (result.hasOwnProperty("results")) {
                ret = result["results"][0]
                if (ret.hasOwnProperty("rows_affected")) {
                    if (parseInt(ret["rows_affected"]) != 0) {
                        // 成功
                        swal({
                            title: "Success",
                            text: "Record of KHID: " + KHID + " was removed",
                            type: "success",
                            confirmButtonColor: "#DD6B55",
                            confirmButtonText: "YES",
                            closeOnConfirm: false,
                            closeOnCancel: false
                        }, function () {
                            // 刷新页面
                            location.reload()
                        });
                    } else {
                        // 失败
                        swal("Error", "No such KHID: " + KHID + " so nothing changed", "error");
                    }
                } else {
                    // 失败
                    swal("Error", "No such KHID: " + KHID + " or internal server error, so nothing changed", "error");
                }
            } else {
                // 失败
                swal("Error", "Weird response from internal server", "error");
            }
        },
        error: function (_, errorInfo) {
            // 失败
            swal("Error", "Internal server was unable to manipulate because: " + errorInfo, "error");
        }
    });
}
