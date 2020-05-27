package ui

import "C"
import (
    "fmt"
    kij "github.com/chrishanli/libkij"
    "key-client/client"
)

const logo = "    __ __           __  __      __    __         \n   / //_/__  __  __/ / / /___  / /___/ /__  _____\n  / ,< / _ \\/ / / / /_/ / __ \\/ / __  / _ \\/ ___/\n / /| /  __/ /_/ / __  / /_/ / / /_/ /  __/ /    \n/_/ |_\\___/\\__, /_/ /_/\\____/_/\\__,_/\\___/_/     \n          /____/                                 \n"
var errWin kij.ErrorWin
var infoWin kij.InfoWin

func Start(port int) {
    errWin.ErrorWinTitle = "Error"
    errWin.Button = []string {
        "OK",
    }
    infoWin.InfoWinTitle = "Information"
    infoWin.Button = []string {
        "OK",
    }

    // 加载
    cl := client.NewClient("KeyHolder", 5000, "127.0.0.1", port)
    if cl == nil {
        errWin.Error = fmt.Sprintf(`The client server (at localhost:%d) was not 
opened or refuses to accept our initialise request.`, port)
        _ = kij.NewErrorWin(&errWin)
        return
    }
    defer cl.CloseClient()

    // 假装加载
    var init kij.InitWin
    init.Logo = logo
    init.Prompt = "Contacting Server, please wait..."
    init.ShowPeriod = 1000
    init.NeedProgBar = true
    kij.NewInitWindow(&init)
    init.ShowPeriod = 600
    init.Prompt = "Please Wait..."
    init.NeedProgBar = false

    // 打开主菜单
    var tbl kij.SelectWin
    tbl.MainTitle = "Keyholder Client"
    tbl.ChoicePadTitle = "Please Select"
    tbl.ChoicePadDesc = "I would like to..."
    tbl.Choices = []string {
        "Start This Programme With My KHID",
        "Buy a New KHID",
    }
    tbl.ChoicePadFootnote = ""
    tbl.StatusBarContent = "Keyholder Client 1.0"
    for {
        sel := kij.NewSelectWin(&tbl)
        if sel == 255 { // 退出
            return
        } else if sel == 1 { // 想购买
            infoWin.Info =
                `You could buy your new KHID at
https://apply.keys.han-li.cn:2333`
            _ = kij.NewInfoWin(&infoWin)
        } else { // 想开始
            var inp kij.InputWin
            inp.MainTitle = "Keyholder Client"
            inp.InputWinName = "Before Start"
            inp.InputBoxNames = []string {
                "KHID",
            }
            inp.Button = []string {
                "OK", "Cancel",
            }
            selected, txt := kij.NewInputWin(&inp)
            if selected == 1 {
                continue
            }
            cl.KHID = txt[0]
            // 输入了KHID，假装在验证
            kij.NewInitWindow(&init)

            // 输入用户名及密码
            var aut kij.AuthWin
            aut.Logo = logo
            aut.Prompt = "Please enter your username and \npassword concerning this KHID"
            aut.StatusBarText = "Keyholder Client 1.0"
            usr, pas := kij.NewAuthWindow(&aut)

            // 验证用户名及密码
            code := 0
            go func() {
                code = cl.Request(usr, pas)
            }()
            // 等待超时秒钟数，等待服务器联系验证
            init.ShowPeriod = cl.Timeout
            kij.NewInitWindow(&init)

            // 按照协议给出提示，如果需要回主页则回去
            if judgeCode(code) {
                continue
            }
            break
        }
    } // 再也不显示主页了

    // 提示成功
    infoWin.InfoWinTitle = "Success"
    infoWin.Info = fmt.Sprintf(`Your application was approved. 
Your session id is %d`, cl.SesID)
    kij.NewInfoWin(&infoWin)

    // （更换主页措辞）
    tbl.Choices[0] = "Return your license"
    for {
        sel := kij.NewSelectWin(&tbl)
        if sel == 255 { // 禁止退出，必须先还回再退出
            errWin.Error = "You must return your license first."
            kij.NewErrorWin(&errWin)
        } else if sel == 1 { // 想购买
            infoWin.Info =
                `You could buy your new KHID at
https://apply.keys.han-li.cn:2333`
            _ = kij.NewInfoWin(&infoWin)
        } else { // 想退出
            // 确认是否要退出
            infoWin.InfoWinTitle = "Confirm"
            infoWin.Info = "Do you really wan to return your license?"
            infoWin.Button = []string {
                "OK", "Cancel",
            }
            sel = kij.NewInfoWin(&infoWin)
            if sel == 1 {
                continue
            }
            // 尝试退出
            code := 0
            go func() {
                code = cl.Return()
            }()
            // 等待超时秒钟数，等待服务器联系验证
            init.ShowPeriod = cl.Timeout
            kij.NewInitWindow(&init)

            // 按照协议给出提示，如果需要回主页则回去
            if judgeCode(code) {
                continue
            }
            break
        }
    } // 再也不显示主页了
	
	// 退出
    infoWin.InfoWinTitle = "Goodbye"
    infoWin.Info = "We are awaiting your next coming"
    infoWin.Button = []string {
        "OK",
    }
    kij.NewInfoWin(&infoWin)

    // 退出程式
    return
}


func judgeCode(code int) (backHome bool) {
    backHome = true
    switch code {
    case 0:
        backHome = false
        break
    case 1:
        errWin.Error =
            `E-001: Sorry, this KHID is lack of 
available licenses. Please try another.`
        kij.NewErrorWin(&errWin)
        break
    case 2:
        errWin.Error =
            `E-002: Sorry, this KHID is invalid. Please try another.`
        kij.NewErrorWin(&errWin)
        break
    case 3:
        errWin.Error =
            `E-003: There's another user with the same id as you, sorry.'`
        kij.NewErrorWin(&errWin)
        break
    case 4:
        errWin.Error =
            `E-004: Unknown error from remote server`
        kij.NewErrorWin(&errWin)
        break
    case 5:
        errWin.Error =
            `E-005: We feel sorry that the server is  
too busy right now. Please Try again later`
        kij.NewErrorWin(&errWin)
        break
    case 6:
        errWin.Error =
            `E-006: This KHID was not applied by you! Perhaps mistyped?`
        kij.NewErrorWin(&errWin)
        break
    case 7:
        errWin.Error =
            `E-007: Invalid username or password. Please try again.`
        kij.NewErrorWin(&errWin)
        break
    case 8:
        errWin.Error =
            `E-008: The returning KHID is not corresponding 
to your session ID. Perhaps mistyped?`
        kij.NewErrorWin(&errWin)
        break
    case 11:
        errWin.Error =
            `E-011: Sending request to client server failed.`
        kij.NewErrorWin(&errWin)
        break
    case 12:
        errWin.Error =
            `E-012: Setting timeout failed.`
        kij.NewErrorWin(&errWin)
        break
    case 13:
        errWin.Error =
            `E-013: Operation timed out, please try again
or check if the client server is available.`
        kij.NewErrorWin(&errWin)
        break
    case 14:
        errWin.Error =
            `E-014: Reply from client server is too short.`
        kij.NewErrorWin(&errWin)
        break
    case 15:
        errWin.Error =
            `E-015: Reply from client server was too weird to be understood.`
        kij.NewErrorWin(&errWin)
        break
    default:
        errWin.Error =
            `Returned weird code`
        kij.NewErrorWin(&errWin)
        break
    }
    return
}