package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/Shimi9999/checkbms"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

type BmsCheckerWindow struct {
	widgets.QMainWindow

	setting *Setting

	base *widgets.QWidget

	settingMenu  *widgets.QMenu
	languageMenu *widgets.QMenu
	actionEn     *widgets.QAction
	actionJa     *widgets.QAction

	menu        *widgets.QWidget
	openButton  *widgets.QPushButton
	pathInput   *widgets.QLineEdit
	diffCheck   *widgets.QCheckBox
	checkButton *widgets.QPushButton
	fileDialog  *widgets.QFileDialog

	logText      *widgets.QTextEdit
	isLogTextSet bool
	logSource    interface{}

	progressSnake *widgets.QLabel
}

type multiLangString struct {
	en string
	ja string
}

func (m multiLangString) string(lang string) string {
	switch lang {
	case "en":
		return m.en
	case "ja":
		return m.ja
	}
	return ""
}

var app *widgets.QApplication

func main() {
	varsion := "1.3.0"

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

	window.setting, _ = ReadSetting()

	window.setLanguage(window.setting.Language)

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
				window.execCheckBmsOrDirectory()
			case 2:
				path1, path2 := e.MimeData().Urls()[0].ToLocalFile(), e.MimeData().Urls()[1].ToLocalFile()
				window.execDiffBmsDir(path1, path2)
			default:
				tooMamyFileMessage := multiLangString{
					en: "ERROR: Too many file. Please drop one bms file/folder or two bms folders.",
					ja: "エラー: ファイルが多すぎます。1個のBMSファイル/フォルダか、2個のBMSフォルダをドロップしてください。",
				}
				window.setLogText(&tooMamyFileMessage)
			}
		}
	})

	window.Show()

	widgets.QApplication_Exec()
}

func (w *BmsCheckerWindow) menuBar() {
	w.settingMenu = w.MenuBar().AddMenu2("Setting")
	w.languageMenu = w.settingMenu.AddMenu2("Language")
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
			w.execCheckBmsOrDirectory()
		}
	})

	w.checkButton.ConnectClicked(func(bool) {
		w.execCheckBmsOrDirectory()
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

	w.isLogTextSet = false

	w.logText.ConnectTextChanged(func() {
		if w.isLogTextSet {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(w.logText.ToHtml()))
			if err != nil {
				w.setLogText(err.Error())
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
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				doc.Find("p span").Each(func(i int, s *goquery.Selection) {
					for _, lc := range levelColors {
						level := lc.Level_en
						if w.setting.Language == "ja" {
							level = lc.Level_ja
						}
						if strings.HasPrefix(s.Text(), level+": ") {
							s.SetHtml(`<span style="color: ` + lc.Color + `">` + level + `</span>` + s.Text()[len(level):])
							break
						}
					}
				})
			}()
			wg.Wait()
			w.isLogTextSet = false

			_html, _ := doc.Html()
			w.logText.SetHtml(_html)
		}
	})
}

func (w *BmsCheckerWindow) execCheckFunction(execFunc func() (interface{}, error)) {
	w.progressSnake.Show()
	w.SetEnabled(false)
	go func() {
		logSource, err := execFunc()
		if err != nil {
			if logSource != nil {
				w.setLogText(logSource)
			} else {
				w.setLogText(err.Error())
			}
		} else {
			w.setLogText(logSource)
		}
		w.progressSnake.Hide()
		w.SetEnabled(true)
	}()
}

func (w *BmsCheckerWindow) execCheckBmsOrDirectory() {
	checkFunc := func() (interface{}, error) {
		return checkBmsOrDirectory(w.pathInput.Text(), w.setting.Language, w.diffCheck.IsChecked())
	}
	w.execCheckFunction(checkFunc)
}

func (w *BmsCheckerWindow) execDiffBmsDir(path1, path2 string) {
	diffFunc := func() (interface{}, error) {
		notBmsDirPaths := []string{}
		for _, path := range [2]string{path1, path2} {
			if !checkbms.IsBmsDirectory(path) {
				notBmsDirPaths = append(notBmsDirPaths, path)
			}
		}
		if len(notBmsDirPaths) > 0 {
			var notBmsFolderMessages multiLangString
			for i, notBmsPath := range notBmsDirPaths {
				newline := ""
				if i > 0 {
					newline = "\n"
				}
				notBmsFolderMessages.en += newline + "ERROR: Not bms folder: " + notBmsPath
				notBmsFolderMessages.ja += newline + "エラー: BMSフォルダではありません: " + notBmsPath
			}
			return &notBmsFolderMessages, fmt.Errorf(notBmsFolderMessages.string(w.setting.Language))
		}
		return checkbms.DiffBmsDirectories(path1, path2)
	}
	w.execCheckFunction(diffFunc)
}

func (w *BmsCheckerWindow) setLanguage(lang string) {
	w.setting.Language = lang
	switch w.setting.Language {
	case "en":
		w.settingMenu.SetTitle("Setting")
		w.languageMenu.SetTitle("Language")
		w.logText.SetPlaceholderText("Drag and drop bms file/folder!")
		w.pathInput.SetPlaceholderText("bms file/folder path")
		w.diffCheck.SetText("diff")
		w.checkButton.SetText("Check")
		w.fileDialog.SetWindowTitle("Open bms folder")

		w.actionEn.SetChecked(true)
	case "ja":
		w.settingMenu.SetTitle("設定")
		w.languageMenu.SetTitle("言語")
		w.logText.SetPlaceholderText("BMSファイル/フォルダをドラッグ&ドロップ!")
		w.pathInput.SetPlaceholderText("BMSファイル/フォルダ パス")
		w.diffCheck.SetText("差分")
		w.checkButton.SetText("チェック")
		w.fileDialog.SetWindowTitle("BMSフォルダを開く")

		w.actionJa.SetChecked(true)
	}
	w.updateLogText()
	WriteSetting(w.setting)
}

func (w *BmsCheckerWindow) setLogText(logSource interface{}) {
	w.logSource = logSource
	w.updateLogText()
}

func (w *BmsCheckerWindow) updateLogText() {
	var updatedLog string

	switch w.logSource.(type) {
	case string:
		updatedLog = w.logSource.(string)
	case *multiLangString:
		str := w.logSource.(*multiLangString)
		updatedLog = str.string(w.setting.Language)
	case *checkbms.Directory:
		bmsDir := w.logSource.(*checkbms.Directory)
		updatedLog = bmsDirectoryLog(bmsDir, w.setting.Language)
	case *checkbms.BmsFile:
		bmsFile := w.logSource.(*checkbms.BmsFile)
		updatedLog = bmsFileLog(bmsFile, w.setting.Language)
	case *checkbms.BmsonFile:
		bmsonFile := w.logSource.(*checkbms.BmsonFile)
		updatedLog = bmsonFileLog(bmsonFile, w.setting.Language)
	case *checkbms.DiffBmsDirResult:
		result := w.logSource.(*checkbms.DiffBmsDirResult)
		updatedLog = diffBmsDirResultLog(result, w.setting.Language)
	}

	if updatedLog != "" {
		w.isLogTextSet = true
		w.logText.SetText(updatedLog)
	}
}

func checkBmsOrDirectory(path, lang string, diff bool) (logSource interface{}, _ error) {
	if checkbms.IsBmsDirectory(path) {
		bmsDir, err := checkbms.ScanBmsDirectory(path, true, true)
		if err != nil {
			return nil, err
		}
		checkbms.CheckBmsDirectory(bmsDir, diff)
		return bmsDir, nil
	} else if checkbms.IsBmsFile(path) {
		bmsFileBase, err := checkbms.ReadBmsFileBase(path)
		if err != nil {
			return nil, err
		}
		if checkbms.IsBmsonFile(path) {
			bmsonFile := checkbms.NewBmsonFile(bmsFileBase)
			if err = bmsonFile.ScanBmsonFile(); err != nil {
				return nil, err
			}
			checkbms.CheckBmsonFile(bmsonFile)
			return bmsonFile, nil
		} else {
			bmsFile := checkbms.NewBmsFile(bmsFileBase)
			if err := bmsFile.ScanBmsFile(); err != nil {
				return nil, err
			}
			checkbms.CheckBmsFile(bmsFile)
			return bmsFile, nil
		}
	}
	notBmsMessage := multiLangString{
		en: "ERROR: Not bms file/folder",
		ja: "エラー: BMSファイル/フォルダではありません",
	}
	return &notBmsMessage, fmt.Errorf(notBmsMessage.string(lang))
}

func bmsDirectoryLog(bmsDir *checkbms.Directory, lang string) (log string) {
	logs := []string{}
	for _, bmsFile := range bmsDir.BmsFiles {
		if len(bmsFile.Logs) > 0 {
			logs = append(logs, bmsFile.LogStringWithLang(true, lang))
		}
	}
	for _, bmsonFile := range bmsDir.BmsonFiles {
		if len(bmsonFile.Logs) > 0 {
			logs = append(logs, bmsonFile.LogStringWithLang(true, lang))
		}
	}
	if len(bmsDir.Logs) > 0 {
		logs = append(logs, bmsDir.LogStringWithLang(true, lang))
	}
	for i, l := range logs {
		if i != 0 {
			log += "\n"
		}
		log += l + "\n"
	}
	if log == "" {
		log = "All OK"
	}
	return log
}

func bmsFileLog(bmsFile *checkbms.BmsFile, lang string) (log string) {
	if len(bmsFile.Logs) > 0 {
		log = bmsFile.LogStringWithLang(true, lang)
	}
	if log == "" {
		log = "All OK"
	}
	return log
}

func bmsonFileLog(bmsonFile *checkbms.BmsonFile, lang string) (log string) {
	if len(bmsonFile.Logs) > 0 {
		log = bmsonFile.LogStringWithLang(true, lang)
	}
	if log == "" {
		log = "All OK"
	}
	return log
}

func diffBmsDirResultLog(result *checkbms.DiffBmsDirResult, lang string) (log string) {
	log = result.LogStringWithLang(lang)
	if log == "" {
		noDifferenceMessage := multiLangString{
			en: fmt.Sprintf("OK, no difference:\n  %s\n  %s", result.DirPath1, result.DirPath2),
			ja: fmt.Sprintf("OK、違いはありません:\n  %s\n  %s", result.DirPath1, result.DirPath2),
		}
		log = noDifferenceMessage.string(lang)
	}
	return log
}
