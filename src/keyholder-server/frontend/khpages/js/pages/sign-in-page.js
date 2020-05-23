$(function () {
    $('#sign_in').validate({
        highlight: function (input) {
            console.log(input);
            $(input).parents('.form-line').addClass('error');
        },
        unhighlight: function (input) {
            $(input).parents('.form-line').removeClass('error');
        },
        errorPlacement: function (error, element) {
            $(element).parents('.input-group').append(error);
        },

        // 提交函数
        submitHandler: function() {
            data = {
                name: document.getElementById("name").value,
                pass: document.getElementById("pass").value
            }
            if (data.name == "admin" && data.pass == "admin") {
                swal({
                    title: "Success",
                    text: "Waiting for redirect...",
                    timer: 2000,
                    showConfirmButton: false
                });
                window.open('index.html','_self');
            } else {
                swal("Oops", "Your username and/or password seems to be mistaken.", "error");
            }
        }
    });

    
});