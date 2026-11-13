package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const appTitle = "Observatório Limeira"

var (
	user32            = windows.NewLazySystemDLL("user32.dll")
	procMessageBoxW   = user32.NewProc("MessageBoxW")
	procGetDpiForSys  = user32.NewProc("GetDpiForSystem")
	procGetSysMetrics = user32.NewProc("GetSystemMetrics")
)

func msgBox(text string, flags uintptr) {
	t, _ := windows.UTF16PtrFromString(text)
	c, _ := windows.UTF16PtrFromString(appTitle)
	_, _, _ = procMessageBoxW.Call(0, uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(c)), flags)
}

func openURL(u string) {
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Start()
}

// O WebView2 (motor do Edge) já vem no Windows 11 e na maioria dos Windows 10.
func hasWebView2() bool {
	const guid = `{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`
	type loc struct {
		k    registry.Key
		path string
	}
	for _, l := range []loc{
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\` + guid},
		{registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\EdgeUpdate\Clients\` + guid},
		{registry.CURRENT_USER, `Software\Microsoft\EdgeUpdate\Clients\` + guid},
	} {
		k, err := registry.OpenKey(l.k, l.path, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		v, _, err := k.GetStringValue("pv")
		k.Close()
		if err == nil && v != "" && v != "0.0.0.0" {
			return true
		}
	}
	return false
}

func main() {
	// uma única janela por vez, para dois programas não gravarem o mesmo arquivo
	name, _ := windows.UTF16PtrFromString("ObservatorioLimeira-instancia-unica")
	if _, err := windows.CreateMutex(nil, false, name); err == windows.ERROR_ALREADY_EXISTS {
		msgBox("O Observatório Limeira já está aberto.", 0x40)
		return
	}

	dir := filepath.Join(os.Getenv("APPDATA"), "ObservatorioLimeira")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		msgBox("Não foi possível criar a pasta de dados:\n"+dir, 0x10)
		return
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		msgBox("Não foi possível iniciar o aplicativo:\n"+err.Error(), 0x10)
		return
	}
	a := &app{dir: dir, token: newToken(), openURL: openURL}
	go func() { _ = http.Serve(ln, a) }()
	url := fmt.Sprintf("http://127.0.0.1:%d/?t=%s", ln.Addr().(*net.TCPAddr).Port, a.token)

	fallback := func() {
		openURL(url)
		msgBox("Este computador não tem o componente Microsoft Edge WebView2, então o aplicativo foi aberto no navegador padrão.\n\n"+
			"Mantenha esta janela aberta enquanto usa o aplicativo e clique em OK somente para encerrar.", 0x40)
	}
	if !hasWebView2() {
		fallback()
		return
	}

	// tamanho da janela proporcional à escala do Windows, limitado à altura da tela
	scale := 1.0
	if procGetDpiForSys.Find() == nil {
		if dpi, _, _ := procGetDpiForSys.Call(); dpi >= 96 {
			scale = float64(dpi) / 96
		}
	}
	width, height := uint(560*scale), uint(900*scale)
	if workH, _, _ := procGetSysMetrics.Call(17); workH > 200 && height > uint(workH) { // SM_CYFULLSCREEN
		height = uint(workH)
	}

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		AutoFocus: true,
		DataPath:  filepath.Join(dir, "WebView2"),
		WindowOptions: webview2.WindowOptions{
			Title:  appTitle,
			Width:  width,
			Height: height,
			IconId: 2,
			Center: true,
		},
	})
	if w == nil {
		fallback()
		return
	}
	defer w.Destroy()
	w.SetSize(int(360*scale), int(520*scale), webview2.HintMin)
	w.Navigate(url)
	w.Run()
}
