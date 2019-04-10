// qif2csv.go

// Tool that extracts account data from a QIF file and saves the totals as a CSV file.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var line, buffer string
var lines int	// total processed lines
var reader *bufio.Reader
var accounts map[string]string	// accounts and categories
var amounts map[string]float64	// totals for each account/category
var width int // number of columns to reserve for the account/category name

// Saves the totals to a CSV file
func save() {
	output := os.Args[2]
	fmt.Printf("Saving to %v\n", output)
	f, err := os.Create(output)
	if err != nil {
		fmt.Println(err)
	} else {
		defer f.Close()
		csv := bufio.NewWriter(f)
		defer csv.Flush()
		// header row
		fmt.Fprint(csv, "Envelope")
		for i := 0; i < width; i++ {
			fmt.Fprint(csv, ",")
		}
		fmt.Fprintln(csv, "Amount")
		// each account/category on its own row
		for account := range amounts {
			fmt.Fprint(csv, strings.ReplaceAll(account, ":", ","))	// one subcategory per column
			for i := len(strings.Split(account, ":")); i < width; i++ {
				fmt.Fprintf(csv, ",")
			}
			fmt.Fprintf(csv, ",%v\n", amounts[account])
		}
	}
}

// Unreads one line from the input file
func unread() {
	buffer = line
}

func getline() string {
	var err error
	if buffer != "" {	// returns an unread line if available
		line = buffer
		buffer = ""
		return line
	}
	line, err = reader.ReadString('\n')	// reads next line from the file
	if err == nil {
		if n := len(line); n > 0 {
			line = line[:n-1] // remove trailing EOL
		}
		lines++
		return line
	}
	// end of file: save CSV
	save()
	fmt.Printf("Processed %v lines.\n", lines)
	os.Exit(0)
	return "" // never gets executed; just to avoid lint errors
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %v <input (QIF) file> <output (CSV) file>", os.Args[0])
	} else {
		input := os.Args[1]
		f, err := os.Open(input)
		if err != nil {
			fmt.Printf("File %v does not exist.", input)
		} else {
			defer f.Close()
			fmt.Printf("Reading from %v\n", input)
			reader = bufio.NewReader(f)
			accounts = make(map[string]string)
			amounts = make(map[string]float64)
			width = 0
			for {
				if getline() == "!Account" {	// get accounts and categories
					accountName := ""
					accountType := ""
					for {
						if getline()[0] == 'N' {
							accountName = line[1:]
						} else if line[0] == 'T' {
							accountType = line[1:]
						} else if line[0] == '^' {
							accounts[accountName] = accountType
							break
						}
					}
				} else if len(line) > 6 && line[:6] == "!Type:" {	// gets amounts for each account/category
					if code := line[6:]; code != "Class" && code != "Cat" {
						amount := float64(0)
						account := ""
						for {
							if getline()[0] == '!' {
								unread()
								break	// end of this account/category; start a new one
							} else {
								if line[0] == 'L' || line[0] == 'S' {	// account name
									account = line[1:]
									if account[0] == '[' {
										account = account[1 : len(account)-1] // remove enclosing brackets from account name
									}
									w := len(strings.Split(account, ":"))
									if w > width {
										width = w
									}
								} else if line[0] == 'T' || line[0] == '$' {	// transaction amount
									amount, _ = strconv.ParseFloat(line[1:], 64)
									if line[0] == '$' {	// split transaction
										amounts[account] = amounts[account] + amount
										amount = 0
									}
								} else if line[0] == '^' {
									amounts[account] = amounts[account] + amount
								}
							}
						}
					}
				}
			}
		}
	}
}
