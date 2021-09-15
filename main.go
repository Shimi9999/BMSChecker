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

func main() {
	varsion := "1.3.0"
	language := "en"

	core.QCoreApplication_SetApplicationName("BMSChecker")
	core.QCoreApplication_SetOrganizationName("Shimi9999")
	core.QCoreApplication_SetAttribute(core.Qt__AA_EnableHighDpiScaling, true)

	app := widgets.NewQApplication(len(os.Args), os.Args)

	window := widgets.NewQMainWindow(nil, 0)
	window.SetWindowTitle(fmt.Sprintf("BMSChecker %s", varsion))
	window.Resize2(700, 450)
	window.SetMinimumSize2(300, 250)
	window.SetAcceptDrops(true)

	base := widgets.NewQWidget(nil, 0)
	base.SetLayout(widgets.NewQVBoxLayout())
	window.SetCentralWidget(base)

	setting := window.MenuBar().AddMenu2("Setting")
	languageMenu := setting.AddMenu2("Language")
	langActGroup := widgets.NewQActionGroup(nil)
	langActGroup.SetExclusive(true)
	enAct := langActGroup.AddAction2("English")
	enAct.SetCheckable(true)
	if language == "en" {
		enAct.SetChecked(true)
	}
	jaAct := langActGroup.AddAction2("日本語")
	jaAct.SetCheckable(true)
	if language == "ja" {
		enAct.SetChecked(true)
	}
	var changeLanguage func(lang string)
	langActGroup.ConnectTriggered(func(action *widgets.QAction) {
		switch action.Pointer() {
		case enAct.Pointer():
			language = "en"
		case jaAct.Pointer():
			language = "ja"
		}
		changeLanguage(language)
	})
	languageMenu.AddActions(langActGroup.Actions())

	menu := widgets.NewQWidget(nil, 0)
	menu.SetLayout(widgets.NewQHBoxLayout())
	menu.Layout().SetContentsMargins(0, 0, 0, 0)
	base.Layout().AddWidget(menu)

	logText := widgets.NewQTextEdit(nil)
	logText.SetFontPointSize(11.0)
	logText.SetReadOnly(true)
	logText.SetLineWrapMode(widgets.QTextEdit__NoWrap)
	logText.SetHorizontalScrollBar(widgets.NewQScrollBar2(core.Qt__Horizontal, nil))
	logText.SetPlaceholderText("Drag and drop bms file/folder!")
	base.Layout().AddWidget(logText)

	progressSnake := widgets.NewQLabel(nil, 0)
	progressSnake.SetMovie(gui.NewQMovie3(":/qml/images/snake.gif", core.NewQByteArray(), nil))
	progressSnake.Movie().Start()
	progressSnake.Hide()
	base.Layout().AddWidget(progressSnake)

	openIcon := app.Style().StandardIcon(widgets.QStyle__SP_DialogOpenButton, nil, nil)
	openButton := widgets.NewQPushButton3(openIcon, "", nil)
	menu.Layout().AddWidget(openButton)

	pathInput := widgets.NewQLineEdit(nil)
	pathInput.SetPlaceholderText("bms file/folder path")
	menu.Layout().AddWidget(pathInput)

	diffCheck := widgets.NewQCheckBox2("diff", nil)
	menu.Layout().AddWidget(diffCheck)

	checkIcon := app.Style().StandardIcon(widgets.QStyle__SP_DialogApplyButton, nil, nil)
	checkButton := widgets.NewQPushButton3(checkIcon, "Check", nil)
	menu.Layout().AddWidget(checkButton)

	setLogText := false
	execCheck := func() {
		progressSnake.Show()
		base.SetEnabled(false)
		go func() {
			log, err := checkBmsLog(pathInput.Text(), language, diffCheck.IsChecked())
			if err != nil {
				logText.SetText(err.Error())
			} else {
				logText.SetText(log)
				setLogText = true
			}
			progressSnake.Hide()
			base.SetEnabled(true)
		}()
	}

	execDiffBmsDir := func(path1, path2 string) {
		progressSnake.Show()
		base.SetEnabled(false)
		go func() {
			log, err := diffBmsDirLog(path1, path2)
			if err != nil {
				logText.SetText(err.Error())
			} else {
				logText.SetText(log)
				setLogText = true
			}
			progressSnake.Hide()
			base.SetEnabled(true)
		}()
	}

	logText.ConnectTextChanged(func() {
		if setLogText {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(logText.ToHtml()))
			if err != nil {
				logText.SetText(err.Error())
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
					if language == "ja" {
						level = lc.Level_ja
					}
					if strings.HasPrefix(s.Text(), level+": ") {
						s.SetHtml(`<span style="color: ` + lc.Color + `">` + level + `</span>` + s.Text()[len(level):])
						break
					}
				}
			})
			setLogText = false

			_html, _ := doc.Html()
			logText.SetHtml(_html)
		}
	})

	window.ConnectDragEnterEvent(func(e *gui.QDragEnterEvent) {
		if e.MimeData().HasUrls() {
			e.AcceptProposedAction()
		}
	})
	window.ConnectDropEvent(func(e *gui.QDropEvent) {
		if e.MimeData().HasUrls() {
			switch len(e.MimeData().Urls()) {
			case 1:
				pathInput.SetText(e.MimeData().Urls()[0].ToLocalFile())
				execCheck()
			case 2:
				path1, path2 := e.MimeData().Urls()[0].ToLocalFile(), e.MimeData().Urls()[1].ToLocalFile()
				execDiffBmsDir(path1, path2)
			}
		}
	})

	//fileDialog := widgets.NewQFileDialog2(nil, "Open bms file/folder", pathInput.Text(), "bms files (*.bms *.bme *.bml *.pms)")
	fileDialog := widgets.NewQFileDialog2(nil, "Open bms folder", "", "bms folder")
	fileDialog.SetFileMode(widgets.QFileDialog__Directory)
	//fileDialog.SetOption(widgets.QFileDialog__DontUseNativeDialog, true)
	openButton.ConnectClicked(func(bool) {
		fileDialog.SetDirectory(pathInput.Text())
		if fileDialog.Exec() == int(widgets.QDialog__Accepted) {
			pathInput.SetText(fileDialog.SelectedFiles()[0])
			execCheck()
		}
	})

	checkButton.ConnectClicked(func(bool) {
		execCheck()
	})

	changeLanguage = func(lang string) {
		switch lang {
		case "en":
			setting.SetTitle("Setting")
			languageMenu.SetTitle("Language")
			logText.SetPlaceholderText("Drag and drop bms file/folder!")
			pathInput.SetPlaceholderText("bms file/folder path")
			diffCheck.SetText("diff")
			checkButton.SetText("Check")
			fileDialog.SetWindowTitle("Open bms folder")
		case "ja":
			setting.SetTitle("設定")
			languageMenu.SetTitle("言語")
			logText.SetPlaceholderText("BMSファイル/フォルダをドラッグ&ドロップ!")
			pathInput.SetPlaceholderText("BMSファイル/フォルダ パス")
			diffCheck.SetText("差分")
			checkButton.SetText("チェック")
			fileDialog.SetWindowTitle("BMSフォルダを開く")
		}
	}

	window.Show()

	widgets.QApplication_Exec()
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
