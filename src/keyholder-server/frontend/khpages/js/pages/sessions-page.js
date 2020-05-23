$(function () {
    getSessions(completion);
})

// 点击删除Session之后要做的事情
function deleteSession() {
    swal({
        title: "Delete",
        text: "Type Session ID whose record is to be remove",
        type: "input",
        showCancelButton: true,
        closeOnConfirm: false,
        animation: "slide-from-top",
        inputPlaceholder: "KHID"
    }, function (inputValue) {
        if (inputValue === false) return false;
        if (inputValue === "") {
            swal.showInputError("The Session ID field may not be nil"); return false
        }
        performDelete(inputValue);
    });
}

// 删除Session ID的数据
function performDelete(sesID) {
    datastr = '["DELETE FROM SESSIONS WHERE SESSION_ID=' + sesID + '"]';
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
                            text: "Record of Session: " + sesID + " was removed",
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
                        swal("Error", "No such Session: " + sesID + " so nothing changed", "error");
                    }
                } else {
                    // 失败
                    swal("Error", "No such Session: " + sesID + " or internal server error, so nothing changed", "error");
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

// GET的封装
function getSessions(callback) {
    // 获取文档内容
    var httpRequest = new XMLHttpRequest();//第一步：建立所需的对象
    httpRequest.open('GET', '/db/query?q=SELECT%20*%20FROM%20SESSIONS', true);//第二步：打开连接  将请求参数写在url中  ps:"./Ptest.php?name=test&nameone=testone"
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

// get之后需要做的事情
function completion(obj) {
    // 表身
    var tbody = document.createElement("tbody");
    // 拿到查询到的所有session信息，并渲染到表中
    sessions = obj["results"][0]["values"];
    // 表尾统计 （PART 1）
    tfoot = document.createElement("tfoot");
    tfootrow = document.createElement("tr");
    tfootTotalLabel = document.createElement("th");
    tfootTotalLabel.innerHTML = "<strong>TOTAL</strong>";
    // 仅当有数据的时候才印入表格
    if (sessions != undefined) {
        sessions.forEach(element => {
            var tr = document.createElement("tr");
            // 现将日期全转换为可读的
            unixTimeStamp = new Date(element[2] * 1000);
            element[2] = unixTimeStamp.toLocaleString();
            unixTimeStamp = new Date(element[3] * 1000);
            element[3] = unixTimeStamp.toLocaleString();
            element.forEach(col => {
                var td = document.createElement("td");
                td.innerHTML = col.toString();
                tr.appendChild(td);
            })
            tbody.appendChild(tr)
            tfootTotalLabel.innerHTML = "<strong>TOTAL: " + sessions.length + " Records</strong>";
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
