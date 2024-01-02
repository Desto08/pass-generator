package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
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

// Database configuratiegegevens
const (
	dbname = "mijndb"
	dbuser = "david"
	dbpass = "geheim"
)

// letters bevat alle kleine en hoofdletters
const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// numbers bevat alle cijfers van 0 tot 9
const numbers = "0123456789"

// symbols bevat een reeks speciale tekens
const symbols = "!@#$%&*+_-="

const (
	// geeft het minimale aantal vereiste letters aan voor het wachtwoord
	minLetters = 8
	// geeft het minimale aantal vereiste nummers aan voor het wachtwoord
	minNumbers = 2
	// geeft het minimale aantal vereiste symbolen aan voor het wachtwoord
	minSymbols = 1
)

// Genereert een wachtwoord op basis van opgegeven lengte en vereisten voor letters, cijfers en symbolen.
// Als het opgegeven wachtwoord niet voldoet aan de vereisten, wordt een fout geretourneerd.
func MaakWachtwoord(length, numLetters, numNumbers, numSymbols int) (string, error) {
	requiredLetters := minLetters
	requiredNumbers := minNumbers
	requiredSymbols := minSymbols

	if length != numLetters+numNumbers+numSymbols || numLetters < requiredLetters || numNumbers < requiredNumbers || numSymbols < requiredSymbols {
		return "", fmt.Errorf("wachtwoord voldoet niet aan de vereisten")
	}

	var chars strings.Builder
	chars.WriteString(letters)
	chars.WriteString(numbers)
	chars.WriteString(symbols)

	wachtwoord := generatePassword(length, numLetters, numNumbers, numSymbols, chars.String())

	return wachtwoord, nil
}

// genereert een wachtwoord van de opgegeven lengte met het aangegeven aantal letters, cijfers en symbolen
// gebaseerd op de beschikbare tekens in de 'chars'-string.
// De resulterende wachtwoordstring wordt willekeurig geschud voordat deze wordt geretourneerd.
func generatePassword(length, numLetters, numNumbers, numSymbols int, chars string) string {
	var result strings.Builder

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < numLetters; i++ {
		result.WriteByte(chars[rng.Intn(len(letters))])
	}

	for i := 0; i < numNumbers; i++ {
		result.WriteByte(chars[len(letters)+rng.Intn(len(numbers))])
	}

	for i := 0; i < numSymbols; i++ {
		result.WriteByte(chars[len(letters)+len(numbers)+rng.Intn(len(symbols))])
	}

	return shuffleString(result.String())
}

// herschikt de tekens willekeurig in de opgegeven strings
func shuffleString(s string) string {
	r := []rune(s)
	rand.Shuffle(len(r), func(i, j int) {
		r[i], r[j] = r[j], r[i]
	})
	return string(r)
}

// genereert een unieke hash op basis van de combinatie van gebruikersnaam en wachtwoord.
func hashUnique(username, password string) (string, error) {
	combined := username + password
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(combined), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Wachtwoordgenerator")

	lengthEntry := widget.NewEntry()
	lengthEntry.SetPlaceHolder("Lengte van het wachtwoord")

	lettersEntry := widget.NewEntry()
	lettersEntry.SetPlaceHolder("Aantal letters")

	numbersEntry := widget.NewEntry()
	numbersEntry.SetPlaceHolder("Aantal cijfers")

	symbolsEntry := widget.NewEntry()
	symbolsEntry.SetPlaceHolder("Aantal symbolen")

	wachtwoordLabel := widget.NewLabel("")

	content := container.NewVBox(
		layout.NewSpacer(),
		widget.NewForm(
			widget.NewFormItem("Lengte:", lengthEntry),
			widget.NewFormItem("Aantal letters:", lettersEntry),
			widget.NewFormItem("Aantal cijfers:", numbersEntry),
			widget.NewFormItem("Aantal symbolen:", symbolsEntry),
		),
		widget.NewButtonWithIcon("Genereer wachtwoord", theme.ContentAddIcon(), func() {
			length, err := strconv.Atoi(lengthEntry.Text)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}

			numLetters, err := strconv.Atoi(lettersEntry.Text)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}

			numNumbers, err := strconv.Atoi(numbersEntry.Text)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}

			numSymbols, err := strconv.Atoi(symbolsEntry.Text)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}

			if numLetters+numNumbers+numSymbols != length {
				dialog.ShowError(errors.New("de som van opgegeven aantallen komt niet overeen met de totale lengte"), myWindow)
				return
			}

			wachtwoord, err := MaakWachtwoord(length, numLetters, numNumbers, numSymbols)
			if err != nil {
				dialog.ShowError(errors.New("wachtwoord voldoet niet aan de minimale vereisten"), myWindow)
				return
			}

			db, err := sql.Open("postgres", "dbname="+dbname+" user="+dbuser+" password="+dbpass+" sslmode=disable")
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			defer db.Close()

			// Controleerd of de wachtwoord al bestaat voor de specifiek genoemde gebruikersnaam
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM gebruikers WHERE gebruikersnaam = $1 AND wachtwoord = $2", "David2", wachtwoord).Scan(&count)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			if count != 0 {
				dialog.ShowInformation("Wachtwoord bestaat al", "Dit wachtwoord bestaat al in de database voor deze gebruiker", myWindow)
				return
			}

			hashedUnique, err := hashUnique("David2", wachtwoord)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}

			_, err = db.Exec("INSERT INTO gebruikers (gebruikersnaam, wachtwoord) VALUES ($1, $2)", "David2", hashedUnique)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			wachtwoordLabel.SetText("Gegenereerd wachtwoord: " + wachtwoord)
			dialog.ShowInformation("Wachtwoord opgeslagen", "Wachtwoord is gegenereerd en opgeslagen in de database", myWindow)
		}),
		wachtwoordLabel,
		layout.NewSpacer(),
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
