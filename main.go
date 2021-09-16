package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Shimi9999/checkbms"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

type BmsCheckerWindow struct {
	widgets.QMainWindow

	language string

	base *widgets.QWidget

	setting      *widgets.QMenu
	languageMenu *widgets.QMenu
	actionEn     *widgets.QAction
	actionJa     *widgets.QAction

	menu        *widgets.QWidget
	openButton  *widgets.QPushButton
	pathInput   *widgets.QLineEdit
	diffCheck   *widgets.QCheckBox
	checkButton *widgets.QPushButton
	fileDialog  *widgets.QFileDialog

	logText    *widgets.QTextEdit
	setLogText bool

	progressSnake *widgets.QLabel
}

var app *widgets.QApplication

func main() {
	varsion := "1.3.0"
	defaultLanguage := "en" // TODO Jsonの設定ファイルを生成し、そこから言語設定を取得する

	core.QCoreApplication_SetApplicationName("BMSChecker")
	core.QCoreApplication_SetOrganizationName("Shimi9999")
	core.QCoreApplication_SetAttribute(core.Qt__AA_EnableHighDpiScaling, true)

	app = widgets.NewQApplication(len(os.Args), os.Args)

	window := NewBmsCheckerWindow(nil, 0)
	window.SetWindowTitle(fmt.Sprintf("BMSChecker %s", varsion))
	window.Resize2(700, 450)
	window.SetMinimumSize2(300, 250)
	window.SetAcceptDrops(true)

	window.menuBar()

	window.base = widgets.NewQWidget(nil, 0)
	window.base.SetLayout(widgets.NewQVBoxLayout())
	window.SetCentralWidget(window.base)

	window.mainMenu()

	window.logTextArea()

	window.progressSnake = widgets.NewQLabel(nil, 0)
	window.progressSnake.SetMovie(gui.NewQMovie3(":/qml/images/snake.gif", core.NewQByteArray(), nil))
	window.progressSnake.Movie().Start()
	window.progressSnake.Hide()
	window.base.Layout().AddWidget(window.progressSnake)

	window.setLanguage(defaultLanguage)

	window.ConnectDragEnterEvent(func(e *gui.QDragEnterEvent) {
		if e.MimeData().HasUrls() {
			e.AcceptProposedAction()
		}
	})
	window.ConnectDropEvent(func(e *gui.QDropEvent) {
		if e.MimeData().HasUrls() {
			switch len(e.MimeData().Urls()) {
			case 1:
				window.pathInput.SetText(e.MimeData().Urls()[0].ToLocalFile())
				window.execCheck()
			case 2:
				path1, path2 := e.MimeData().Urls()[0].ToLocalFile(), e.MimeData().Urls()[1].ToLocalFile()
				window.execDiffBmsDir(path1, path2)
			}
		}
	})

	window.Show()

	widgets.QApplication_Exec()
}

func (w *BmsCheckerWindow) menuBar() {
	w.setting = w.MenuBar().AddMenu2("Setting")
	w.languageMenu = w.setting.AddMenu2("Language")
	langActGroup := widgets.NewQActionGroup(nil)
	langActGroup.SetExclusive(true)
	w.actionEn = langActGroup.AddAction2("English")
	w.actionEn.SetCheckable(true)
	w.actionJa = langActGroup.AddAction2("日本語")
	w.actionJa.SetCheckable(true)
	langActGroup.ConnectTriggered(func(action *widgets.QAction) {
		switch action.Pointer() {
		case w.actionEn.Pointer():
			w.setLanguage("en")
		case w.actionJa.Pointer():
			w.setLanguage("ja")
		}
	})
	w.languageMenu.AddActions(langActGroup.Actions())
}

func (w *BmsCheckerWindow) mainMenu() {
	w.menu = widgets.NewQWidget(nil, 0)
	w.menu.SetLayout(widgets.NewQHBoxLayout())
	w.menu.Layout().SetContentsMargins(0, 0, 0, 0)
	w.base.Layout().AddWidget(w.menu)

	openIcon := app.Style().StandardIcon(widgets.QStyle__SP_DialogOpenButton, nil, nil)
	w.openButton = widgets.NewQPushButton3(openIcon, "", nil)
	w.menu.Layout().AddWidget(w.openButton)

	w.pathInput = widgets.NewQLineEdit(nil)
	w.pathInput.SetPlaceholderText("bms file/folder path")
	w.menu.Layout().AddWidget(w.pathInput)

	w.diffCheck = widgets.NewQCheckBox2("diff", nil)
	w.menu.Layout().AddWidget(w.diffCheck)

	checkIcon := app.Style().StandardIcon(widgets.QStyle__SP_DialogApplyButton, nil, nil)
	w.checkButton = widgets.NewQPushButton3(checkIcon, "Check", nil)
	w.menu.Layout().AddWidget(w.checkButton)

	//fileDialog := widgets.NewQFileDialog2(nil, "Open bms file/folder", pathInput.Text(), "bms files (*.bms *.bme *.bml *.pms)")
	w.fileDialog = widgets.NewQFileDialog2(nil, "Open bms folder", "", "bms folder")
	w.fileDialog.SetFileMode(widgets.QFileDialog__Directory)
	//fileDialog.SetOption(widgets.QFileDialog__DontUseNativeDialog, true)

	w.openButton.ConnectClicked(func(bool) {
		w.fileDialog.SetDirectory(w.pathInput.Text())
		if w.fileDialog.Exec() == int(widgets.QDialog__Accepted) {
			w.pathInput.SetText(w.fileDialog.SelectedFiles()[0])
			w.execCheck()
		}
	})

	w.checkButton.ConnectClicked(func(bool) {
		w.execCheck()
	})
}

func (w *BmsCheckerWindow) logTextArea() {
	w.logText = widgets.NewQTextEdit(nil)
	w.logText.SetFontPointSize(11.0)
	w.logText.SetReadOnly(true)
	w.logText.SetLineWrapMode(widgets.QTextEdit__NoWrap)
	w.logText.SetHorizontalScrollBar(widgets.NewQScrollBar2(core.Qt__Horizontal, nil))
	w.logText.SetPlaceholderText("Drag and drop bms file/folder!")
	w.base.Layout().AddWidget(w.logText)

	w.setLogText = false

	w.logText.ConnectTextChanged(func() {
		if w.setLogText {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(w.logText.ToHtml()))
			if err != nil {
				w.logText.SetText(err.Error())
			}

			type levelColor struct {
				Level_en string
				Level_ja string
				Color    string
			}
			levelColors := []levelColor{
				{Level_en: "ERROR", Level_ja: "エラー", Color: "#ff0000"},
				{Level_en: "WARNING", Level_ja: "警告", Color: "#e56b00"},
				{Level_en: "NOTICE", Level_ja: "通知", Color: "#0000da"},
			}
			doc.Find("p span").Each(func(i int, s *goquery.Selection) {
				for _, lc := range levelColors {
					level := lc.Level_en
					if w.language == "ja" {
						level = lc.Level_ja
					}
					if strings.HasPrefix(s.Text(), level+": ") {
						s.SetHtml(`<span style="color: ` + lc.Color + `">` + level + `</span>` + s.Text()[len(level):])
						break
					}
				}
			})
			w.setLogText = false

			_html, _ := doc.Html()
			w.logText.SetHtml(_html)
		}
	})
}

func (w *BmsCheckerWindow) execCheck() {
	w.progressSnake.Show()
	w.base.SetEnabled(false)
	go func() {
		log, err := checkBmsLog(w.pathInput.Text(), w.language, w.diffCheck.IsChecked())
		if err != nil {
			w.logText.SetText(err.Error())
		} else {
			w.logText.SetText(log)
			w.setLogText = true
		}
		w.progressSnake.Hide()
		w.base.SetEnabled(true)
	}()
}

func (w *BmsCheckerWindow) execDiffBmsDir(path1, path2 string) {
	w.progressSnake.Show()
	w.base.SetEnabled(false)
	go func() {
		log, err := diffBmsDirLog(path1, path2)
		if err != nil {
			w.logText.SetText(err.Error())
		} else {
			w.logText.SetText(log)
			w.setLogText = true
		}
		w.progressSnake.Hide()
		w.base.SetEnabled(true)
	}()
}

func (w *BmsCheckerWindow) setLanguage(lang string) {
	w.language = lang
	switch w.language {
	case "en":
		w.setting.SetTitle("Setting")
		w.languageMenu.SetTitle("Language")
		w.logText.SetPlaceholderText("Drag and drop bms file/folder!")
		w.pathInput.SetPlaceholderText("bms file/folder path")
		w.diffCheck.SetText("diff")
		w.checkButton.SetText("Check")
		w.fileDialog.SetWindowTitle("Open bms folder")

		w.actionEn.SetChecked(true)
	case "ja":
		w.setting.SetTitle("設定")
		w.languageMenu.SetTitle("言語")
		w.logText.SetPlaceholderText("BMSファイル/フォルダをドラッグ&ドロップ!")
		w.pathInput.SetPlaceholderText("BMSファイル/フォルダ パス")
		w.diffCheck.SetText("差分")
		w.checkButton.SetText("チェック")
		w.fileDialog.SetWindowTitle("BMSフォルダを開く")

		w.actionJa.SetChecked(true)
	}
}

func checkBmsLog(path, lang string, diff bool) (log string, _ error) {
	if checkbms.IsBmsDirectory(path) {
		bmsDirs, err := checkbms.ScanDirectory(path)
		if err != nil {
			return "", err
		}
		logs := []string{}
		for _, dir := range bmsDirs {
			checkbms.CheckBmsDirectory(&dir, diff)
			for _, bmsFile := range dir.BmsFiles {
				if len(bmsFile.Logs) > 0 {
					logs = append(logs, bmsFile.LogStringWithLang(true, lang))
				}
			}
			if len(dir.Logs) > 0 {
				logs = append(logs, dir.LogStringWithLang(true, lang))
			}
		}
		for i, l := range logs {
			if i != 0 {
				log += "\n"
			}
			log += l + "\n"
		}
	} else if checkbms.IsBmsFile(path) {
		bmsFile, err := checkbms.ReadBmsFile(path)
		if err != nil {
			return "", err
		}
		if err := bmsFile.ScanBmsFile(); err != nil {
			return "", err
		}
		checkbms.CheckBmsFile(bmsFile)
		if len(bmsFile.Logs) > 0 {
			log = bmsFile.LogStringWithLang(true, lang)
		}
	} else {
		log = "Error: Not bms file/folder"
	}

	if log == "" {
		log = "All OK"
	}

	return log, nil
}

func diffBmsDirLog(path1, path2 string) (log string, _ error) {
	difflogs, err := checkbms.DiffBmsDirectories(path1, path2)
	if err != nil {
		return "", err
	}
	for _, difflog := range difflogs {
		log += difflog + "\n"
	}
	if log == "" {
		log = fmt.Sprintf("OK, no difference:\n  %s\n  %s", path1, path2)
	}
	return log, nil
}
