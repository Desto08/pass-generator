package main

import (
	"database/sql"
	"math/rand"
	"strconv"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

const (
	dbname = "mijndb"
	dbuser = "david"
	dbpass = "geheim"
)

const letters string = "abcdfghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const numbers string = "0123456789"
const symbols string = "!@#$%&*+_-="

func MaakWachtwoord(length int, heeftnummers bool, heeftsymbolen bool) string {
	chars := letters
	if heeftnummers {
		chars += numbers
	}
	if heeftsymbolen {
		chars += symbols
	}

	return genereerWachtwoord(length, chars)
}

func genereerWachtwoord(length int, chars string) string {
	wachtwoord := ""
	for i := 0; i < length; i++ {
		wachtwoord += string([]rune(chars)[rand.Intn(len(chars))])
	}
	return wachtwoord
}

func hashWachtwoord(wachtwoord string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(wachtwoord), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	myApp := app.New()
	myWindow := myApp.NewWindow("Wachtwoordgenerator")

	lengthEntry := widget.NewEntry()
	lengthEntry.SetPlaceHolder("Lengte van het wachtwoord")
	numbersCheck := widget.NewCheck("Nummers toestaan", func(checked bool) {})
	symbolsCheck := widget.NewCheck("Symbolen toestaan", func(checked bool) {})

	wachtwoordLabel := widget.NewLabel("")

	generateButton := widget.NewButtonWithIcon("Genereer wachtwoord", theme.ContentAddIcon(), func() {
		length, err := strconv.Atoi(lengthEntry.Text)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		db, err := sql.Open("postgres", "dbname="+dbname+" user="+dbuser+" password="+dbpass+" sslmode=disable")
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		defer db.Close()

		var wachtwoord string
		loopActief := true

		for loopActief {
			wachtwoord = MaakWachtwoord(length, numbersCheck.Checked, symbolsCheck.Checked)
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM gebruikers WHERE wachtwoord = $1", wachtwoord).Scan(&count)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			if count == 0 {
				loopActief = false
			}
		}

		hashedWachtwoord, err := hashWachtwoord(wachtwoord)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		_, err = db.Exec("INSERT INTO gebruikers (gebruikersnaam, wachtwoord) VALUES ($1, $2)", "David", hashedWachtwoord)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		wachtwoordLabel.SetText("Gegenereerd wachtwoord: " + wachtwoord)
		dialog.ShowInformation("Wachtwoord opgeslagen", "Wachtwoord is gegenereerd en opgeslagen in de database", myWindow)
	})

	content := container.NewVBox(
		layout.NewSpacer(),
		widget.NewForm(
			widget.NewFormItem("Lengte:", lengthEntry),
			widget.NewFormItem("", numbersCheck),
			widget.NewFormItem("", symbolsCheck),
		),
		generateButton,
		wachtwoordLabel,
		layout.NewSpacer(),
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
