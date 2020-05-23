$(function () {

    //noUISlider
    var sliderBasic = document.getElementById('capacity_slider');
    noUiSlider.create(sliderBasic, {
        start: [30],
        connect: 'lower',
        step: 1,
        range: {
            'min': [1],
            'max': [500]
        }
    });
    getNoUISliderValue(sliderBasic, false);

    $('#sign_up').validate({
        rules: {
            'terms': {
                required: true
            },
            'confirm': {
                equalTo: '[name="password"]'
            }
        },
        highlight: function (input) {
            console.log(input);
            $(input).parents('.form-line').addClass('error');
        },
        unhighlight: function (input) {
            $(input).parents('.form-line').removeClass('error');
        },
        errorPlacement: function (error, element) {
            $(element).parents('.input-group').append(error);
            $(element).parents('.form-group').append(error);
        },

        // 提交函数
        submitHandler: function () {
            data = {
                name: document.getElementById("name").value,
                pass: document.getElementById("pass").value,
                num: genKeyNum(8),
                capacity: parseInt(sliderBasic.noUiSlider.get()),
                expDate: getStdExpireDate()
            }

            // 请求新建密钥
            swal({
                title: "Please Wait...",
                text: "Your purchase is being processed. Please wait...",
                timer: 2000,
                showConfirmButton: false
            }, function() {
                applyKey(data); 
            });
            
        return false
        }

    });
});

// 申请
function applyKey(data) {
    datastr = '["INSERT INTO KEYS(KEY_NUM, KEY_CAPACITY, KEY_RESIDUE, USERNAME, PASSWORD, EXPIRE_DATE, LAST_LOGIN) VALUES(\'' + data.num + '\', '+ data.capacity + ', ' + data.capacity + ', \'' + data.name + '\', \'' + data.pass + '\', \'' + Date.parse(data.expDate) + '\', ' + Date.parse(new Date()) + ')"]'
    // 发送ajax
    try {
        $.ajax({
            type: "POST",
            contentType: "json", //发给服务器的数据类型
            dataType: "json", //预期服务器返回的数据类型
            url: "/db/execute" ,//url
            data: datastr,
            success: function (result) {
                // 检视返回的消息
                if (result.hasOwnProperty("results")) {
                    ret = result["results"][0]
                    if (ret.hasOwnProperty("error")) {
                        switch (ret["error"].substr(0,6)) {
                            case "UNIQUE":
                                swal("Oops", "This key already exists. Try another one.", "error");
                                break;
                            default:
                                swal("Oops", "Unknown error from internal server: " + ret["error"], "error");
                                break;
                        }
                    } else {
                        // 成功
                        swal({
                            title: "Well Done!",
                            text: "Your KHID application has been approved.",
                            type: "success",
                            confirmButtonColor: "#DD6B55",
                            confirmButtonText: "Show your KHID",
                            closeOnConfirm: false
                        }, function () {
                            showAppliedKey(data.num, data.expDate.getDate() + "/" + (data.expDate.getMonth() + 1) + "/" + data.expDate.getFullYear());
                        });
                    }
                } else {
                    swal("Oops", "Weird Response from Internal server: " + JSON.stringify(result), "error");
                }
            },
            error : function(a, errorInfo) {
                swal("Oops", "Error response from Internal server: " + JSON.stringify(result), "error");
            }
        });
    } catch (error) {
        swal("Oops", "Connection to Internal server error", "error");
    }
}

// 显示钥匙号码
function showAppliedKey(key, expire) {
    swal({
        title: "Purchased Info",
        text: "<div>Your New KHID is</div><div><h1>" + key + "</h1></div><div>The key will expire at " + expire + " </div><div>Please remember carefully</div>",
        html: true
    });
}


// 生成一个标准过期日编号（1年后）
function getStdExpireDate() {
    date = new Date();
    date.setFullYear(date.getFullYear() + 1);
    return date
}

// 生成一个凭据编号
function genKeyNum(len) {
    var str = "",
        arr = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '@', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', "#", 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '$'];
    // 随机产生
    for (var i = 0; i < len; i++) {
        pos = Math.round(Math.random() * (arr.length - 1));
        str += arr[pos];
    }
    return str;
}

// Get noUISlider Value and write on
function getNoUISliderValue(slider, percentage) {
    slider.noUiSlider.on('update', function () {
        var val = slider.noUiSlider.get();
        $(slider).parent().find('span.js-nouislider-value').text(parseInt(val));
        $(slider).parent().find('span.js-nouislider-estcost').text("$ " + (val * 999.00).toFixed(2));
        evaluate = document.getElementById("evaluate")
        if (val < 100) {
            evaluate.setAttribute("class", "label label-default")
            evaluate.innerHTML = "Normal Pack";
        } else if (val < 350) {
            evaluate.setAttribute("class", "label label-primary")
            evaluate.innerHTML = "Intermidiate Pack";
        } else {
            evaluate.setAttribute("class", "label label-warning")
            evaluate.innerHTML = "Deluxe Pack";
        }
    });
}

function register() {
    swal({
        title: "Success",
        text: "Waiting for redirect...",
        timer: 2000,
        showConfirmButton: false
    });
}