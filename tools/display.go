package tools

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
	"log"
	"os"
)

const (
	FILE_LOG = true
)

var Filename string

func PrintColorTable(header []string, dataList [][]string, title_opt ...string) {
	if len(title_opt) == 1 {
		log.Printf("[%s] %s", aurora.Magenta("---"), aurora.Cyan("[ "+title_opt[0]+" ]"))
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	for _, data := range dataList {
		var dataColor []tablewriter.Colors
		for _, dataValue := range data {
			if dataValue == "OK" {
				dataColor = append(dataColor, tablewriter.Colors{
					tablewriter.Bold,
					tablewriter.FgGreenColor,
				})
			} else if dataValue == "NOK" {
				dataColor = append(dataColor, tablewriter.Colors{
					tablewriter.Bold,
					tablewriter.FgRedColor,
				})
			}
		}
		table.Rich(data, dataColor)
	}
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.Render()
}

func PrintColorTableNoBorder(header []string, dataList [][]string, title_opt ...string) {
	if len(title_opt) == 1 {
		log.Printf("%s %s", aurora.Blue(">>>"), aurora.Cyan("[ "+title_opt[0]+" ]"))
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	for _, data := range dataList {
		var dataColor []tablewriter.Colors
		for _, dataValue := range data {
			if dataValue == "OK" || dataValue == "YES" || dataValue == "+" {
				dataColor = append(dataColor, tablewriter.Colors{
					tablewriter.Bold,
					tablewriter.FgGreenColor,
				})
			} else if dataValue == "FAIL" || dataValue == "NO" || dataValue == "-" {
				dataColor = append(dataColor, tablewriter.Colors{
					tablewriter.Bold,
					tablewriter.FgRedColor,
				})
			} else {
				dataColor = append(dataColor, tablewriter.Colors{})
			}
		}
		table.Rich(data, dataColor)
	}
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)
	table.Render()
}

func _randomFilename(n int) string {
	if Filename == "" {
		bytes := make([]byte, n)
		if _, err := rand.Read(bytes); err != nil {
			panic(err)
		}
		Filename = hex.EncodeToString(bytes) + ".log"
	}
	return "./log/" + Filename
}
func _printFileLog(logMsg string) {
	if FILE_LOG {
		f, err := os.OpenFile(_randomFilename(10), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			return
		}
		defer f.Close()
		if _, err := f.WriteString(logMsg + "\n"); err != nil {
			log.Println(err)
		}
	}
}
func Title(title string) {
	log.Printf("[%s]", aurora.Magenta(title))
}

func Info(message string, tab ...bool) {
	isTab := ""
	if len(tab) > 0 {
		if tab[0] {
			isTab = "\t"
		}
	}
	logMsg := fmt.Sprintf("%s[%s] [%s]", isTab, aurora.Blue(">>>"), aurora.Cyan(message))
	log.Println(logMsg)
	_printFileLog(logMsg)
}

func Success(message string, tab ...bool) {
	isTab := ""
	if len(tab) > 0 {
		if tab[0] {
			isTab = "\t"
		}
	}

	logMsg := fmt.Sprintf("%s[%s] [%s] [%s]", isTab, aurora.Green("+"), aurora.Yellow(message), aurora.Green("OK"))
	log.Println(logMsg)
	_printFileLog(logMsg)
}

func Warning(message string, tab ...bool) {
	isTab := ""
	if len(tab) > 0 {
		if tab[0] {
			isTab = "\t"
		}
	}
	logMsg := fmt.Sprintf("%s[%s] [%s]", isTab, aurora.Yellow("***"), aurora.White(message).Bold())
	log.Println(logMsg)
	_printFileLog(logMsg)
}

func Fail(message string, tab ...bool) {
	isTab := ""
	if len(tab) > 0 {
		if tab[0] {
			isTab = "\t"
		}
	}
	logMsg := fmt.Sprintf("%s[%s] [%s] [%s]", isTab, aurora.Red("-"), aurora.Yellow(message), aurora.Red("FAIL"))
	log.Println(logMsg)
	_printFileLog(logMsg)
}

func Error(message string) {
	logMsg := fmt.Sprintf("\t[%s] [%s]", aurora.Red("X"), aurora.Red(message))
	log.Println(logMsg)
	_printFileLog(logMsg)
}

func Log(message string, logData string, tab ...bool) {
	isTab := ""
	if len(tab) > 0 {
		if tab[0] {
			isTab = "\t"
		}
	}
	logMsg := fmt.Sprintf("%s[%s] [%s] (%s)", isTab, aurora.Yellow("LOG"), aurora.Yellow(message), aurora.Yellow(logData))
	log.Println(logMsg)
	_printFileLog(logMsg)
}
