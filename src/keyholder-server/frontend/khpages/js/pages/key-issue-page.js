$(function () {
    //Masked Input ============================================================================================================================
    var $demoMaskedInput = $('.masked-input');

    //Date
    $demoMaskedInput.find('.dateinput').inputmask('dd/mm/yyyy', { placeholder: '__/__/____' });

    //Time
    $demoMaskedInput.find('.time12').inputmask('hh:mm t', { placeholder: '__:__ _m', alias: 'time12', hourFormat: '12' });
    $demoMaskedInput.find('.time24').inputmask('hh:mm', { placeholder: '__:__ _m', alias: 'time24', hourFormat: '24' });

    //Date Time
    $demoMaskedInput.find('.datetime').inputmask('d/m/y h:s', { placeholder: '__/__/____ __:__', alias: "datetime", hourFormat: '24' });

    //Mobile Phone Number
    $demoMaskedInput.find('.mobile-phone-number').inputmask('+99 (999) 999-99-99', { placeholder: '+__ (___) ___-__-__' });
    //Phone Number
    $demoMaskedInput.find('.phone-number').inputmask('+99 (999) 999-99-99', { placeholder: '+__ (___) ___-__-__' });

    //Dollar Money
    $demoMaskedInput.find('.money-dollar').inputmask('99,99 $', { placeholder: '__,__ $' });
    //Euro Money
    $demoMaskedInput.find('.money-euro').inputmask('99,99 €', { placeholder: '__,__ €' });

    //IP Address
    $demoMaskedInput.find('.ip').inputmask('999.999.999.999', { placeholder: '___.___.___.___' });

    //Credit Card
    $demoMaskedInput.find('.credit-card').inputmask('9999 9999 9999 9999', { placeholder: '____ ____ ____ ____' });

    //Email
    $demoMaskedInput.find('.email').inputmask({ alias: "email" });

    //Serial Key
    $demoMaskedInput.find('.key').inputmask('****-****-****-****', { placeholder: '____-____-____-____' });
    //===========================================================================================================================================

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

    // Form Validation
    $('#key_issue_form').validate({
        rules: {
            'date': {
                customdate: true
            },
            'key-num': {
                keynum: true
            }
        },
        highlight: function (input) {
            $(input).parents('.form-line').addClass('error');
        },
        unhighlight: function (input) {
            $(input).parents('.form-line').removeClass('error');
        },
        errorPlacement: function (error, element) {
            $(element).parents('.form-group').append(error);
        },
        // 提交函数
        submitHandler: function() {
            function getDBDateString() {
                tokens = document.getElementById("date").value.split('/');
                return tokens[2] + "-" + tokens[1] + "-" + tokens[0] + "T00:00:00Z";
            }
            var data = {
                name: document.getElementById("name").value,
                pass: document.getElementById("password").value,
                num: document.getElementById("key-num").value,
                capacity: parseInt(sliderBasic.noUiSlider.get()),
                date: getDBDateString()
            }

            datastr = '["INSERT INTO KEYS(KEY_NUM, KEY_CAPACITY, KEY_RESIDUE, USERNAME, PASSWORD, EXPIRE_DATE, LAST_LOGIN) VALUES(\'' + data.num + '\', '+ data.capacity + ', ' + data.capacity + ', \'' + data.name + '\', \'' + data.pass + '\', \'' + data.date + '\', ' + Date.parse(new Date()) + ')"]'

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
                                        showNotification("bg-red", "This key already exists. Try another one.", "top", "center");
                                        break;
                                    default:
                                        showNotification("bg-red", "Unknown error from rqlite: " + ret["error"], "top", "center");
                                        break;
                                }
                            } else {
                                showNotification("bg-green", "Operation Succeeded.", "top", "center");
                            }
                        } else {
                            showNotification("bg-red", "Weird Response" + JSON.stringify(result), "top", "center");
                        }
                    },
                    error : function(a, errorInfo) {
                        showNotification("bg-red", "Failed to access the internal database server", "top", "center");
                    }
                });
            return false
        }
    });

    // Date
    $.validator.addMethod('customdate', function (value) {
        return value.match(/^\d\d?\/\d\d?\/\d\d\d\d$/);
    },
        'Please enter a date in the format dd/MM/yyyy'
    );

    // key
    $.validator.addMethod('keynum', function (value) {
        return value.match(/^[A-Za-z0-9@#$]+$/);
    },
        'Please enter digits, letters, @, # or $ only, and less than 8 characters'
    );
});

// Get noUISlider Value and write on
function getNoUISliderValue(slider, percentage) {
    slider.noUiSlider.on('update', function () {
        var val = slider.noUiSlider.get();
        $(slider).parent().find('span.js-nouislider-value').text(parseInt(val));
    });
}

// 显示一个alert
function showNotification(colorName, text, placementFrom, placementAlign, animateEnter, animateExit) {
    if (colorName === null || colorName === '') { colorName = 'bg-black'; }
    if (text === null || text === '') { text = 'Turning standard Bootstrap alerts'; }
    if (animateEnter === null || animateEnter === '') { animateEnter = 'animated fadeInDown'; }
    if (animateExit === null || animateExit === '') { animateExit = 'animated fadeOutUp'; }
    var allowDismiss = true;

    notify = $.notify({
        message: text
    },
        {
            type: colorName,
            allow_dismiss: allowDismiss,
            newest_on_top: true,
            timer: 1000,
            placement: {
                from: placementFrom,
                align: placementAlign
            },
            animate: {
                enter: animateEnter,
                exit: animateExit
            },
            template: '<div data-notify="container" class="bootstrap-notify-container alert alert-dismissible {0} ' + (allowDismiss ? "p-r-35" : "") + '" role="alert">' +
            '<button type="button" aria-hidden="true" class="close" data-notify="dismiss">×</button>' +
            '<span data-notify="icon"></span> ' +
            '<span data-notify="title">{1}</span> ' +
            '<span data-notify="message">{2}</span>' +
            '<div class="progress" data-notify="progressbar">' +
            '<div class="progress-bar progress-bar-{0}" role="progressbar" aria-valuenow="0" aria-valuemin="0" aria-valuemax="100" style="width: 0%;"></div>' +
            '</div>' +
            '<a href="{3}" target="{4}" data-notify="url"></a>' +
            '</div>'
        });
}