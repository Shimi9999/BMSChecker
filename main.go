package main

import (
	"os"
	//"fmt"

	"github.com/Shimi9999/checkbms"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func main() {
	core.QCoreApplication_SetApplicationName("BMSChecker")
	core.QCoreApplication_SetOrganizationName("Shimi9999")
	core.QCoreApplication_SetAttribute(core.Qt__AA_EnableHighDpiScaling, true)

	app := widgets.NewQApplication(len(os.Args), os.Args)

	window := widgets.NewQMainWindow(nil, 0)
	window.SetMinimumSize2(700, 450)
	window.SetWindowTitle("BMSChecker")
	window.SetAcceptDrops(true)

	base := widgets.NewQWidget(nil, 0)
	base.SetLayout(widgets.NewQVBoxLayout())
	window.SetCentralWidget(base)

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

	diffCheck := widgets.NewQRadioButton2("diff", nil)
	menu.Layout().AddWidget(diffCheck)

	checkIcon := app.Style().StandardIcon(widgets.QStyle__SP_DialogApplyButton, nil, nil)
	checkButton := widgets.NewQPushButton3(checkIcon, "Check", nil)
	menu.Layout().AddWidget(checkButton)

	execCheck := func() {
		progressSnake.Show()
		base.SetEnabled(false)
		go func() {
			log, err := checkBmsLog(pathInput.Text(), diffCheck.IsChecked())
			if err != nil {
				logText.SetText(err.Error())
			} else {
				logText.SetText(log)
			}
			progressSnake.Hide()
			base.SetEnabled(true)
		}()
	}

	window.ConnectDragEnterEvent(func(e *gui.QDragEnterEvent) {
		if e.MimeData().HasUrls() {
			e.AcceptProposedAction()
		}
	})
	window.ConnectDropEvent(func(e *gui.QDropEvent) {
		if e.MimeData().HasUrls() {
			pathInput.SetText(e.MimeData().Urls()[0].ToLocalFile())
			execCheck()
		}
	})

	openButton.ConnectClicked(func(bool) {
		//fileDialog := widgets.NewQFileDialog2(nil, "Open bms file/folder", pathInput.Text(), "bms files (*.bms *.bme *.bml *.pms)")
		fileDialog := widgets.NewQFileDialog2(nil, "Open bms folder", pathInput.Text(), "bms folder")
		fileDialog.SetFileMode(widgets.QFileDialog__Directory)
		//fileDialog.SetOption(widgets.QFileDialog__DontUseNativeDialog, true)
		if fileDialog.Exec() == int(widgets.QDialog__Accepted) {
			pathInput.SetText(fileDialog.SelectedFiles()[0])
			execCheck()
		}
	})

	checkButton.ConnectClicked(func(bool) {
		execCheck()
	})

	window.Show()

	widgets.QApplication_Exec()
}

func checkBmsLog(path string, diff bool) (log string, _ error) {
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
					logs = append(logs, bmsFile.LogString(true))
				}
			}
			if len(dir.Logs) > 0 {
				logs = append(logs, dir.LogString(true))
			}
		}
		for i, l := range logs {
			if i != 0 {
				log += "\n"
			}
			log += l + "\n"
		}
	} else if checkbms.IsBmsFile(path) {
		bmsFile, err := checkbms.ScanBmsFile(path)
		if err != nil {
			return "", err
		}
		checkbms.CheckBmsFile(bmsFile)
		if len(bmsFile.Logs) > 0 {
			log = bmsFile.LogString(true)
		}
	} else {
		log = "Error: Not bms file/folder"
	}

	if log == "" {
		log = "All OK"
	}

	return log, nil
}
