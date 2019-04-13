package pages

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/decred/dcrd/dcrutil"
	"github.com/gdamore/tcell"
	"github.com/raedahgroup/godcr/app/walletcore"
	"github.com/raedahgroup/godcr/terminal/helpers"
	"github.com/raedahgroup/godcr/terminal/primitives"
	"github.com/rivo/tview"
)

var displayedTxHashes []string

// var _24hours, _ = time.ParseDuration("24h")

func historyPage(wallet walletcore.Wallet, hintTextView *primitives.TextView, setFocus func(p tview.Primitive) *tview.Application, clearFocus func()) tview.Primitive {
	// parent flexbox layout container to hold other primitives
	body := tview.NewFlex().SetDirection(tview.FlexRow)

	// page title and tip
	titleTextView := primitives.NewLeftAlignedTextView("History")
	body.AddItem(titleTextView, 2, 0, false)

	historyTable := tview.NewTable().
		SetBorders(false).
		SetFixed(1, 0).
		SetSelectable(true, false)

	transactionDetailsTable := tview.NewTable().SetBorders(false)

	displayHistoryTable := func() {
		body.RemoveItem(transactionDetailsTable)

		titleTextView.SetText("History")
		hintTextView.SetText("TIP: Use ARROW UP/DOWN to select txn, ENTER to view details, ESC to return to navigation menu")

		body.AddItem(historyTable, 0, 1, true)
		setFocus(historyTable)
	}

	errorTextView := primitives.WordWrappedTextView("")
	errorTextView.SetTextColor(helpers.DecredOrangeColor)

	displayError := func(errorMessage string) {
		body.RemoveItem(errorTextView)
		errorTextView.SetText(errorMessage)
		body.AddItem(errorTextView, 2, 0, false)
	}

	historyTable.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			clearFocus()
		}
	})

	displayedTxHashes = []string{}

	// method for getting transaction details when a tx is selected from the history table
	historyTable.SetSelectedFunc(func(row, column int) {
		body.RemoveItem(historyTable)
		txHash := displayedTxHashes[row-1]

		titleTextView.SetText("Transaction Details")
		hintTextView.SetText("TIP: Use ARROW UP/DOWN to scroll, BACKSPACE to view History page, ESC to return to navigation menu")

		transactionDetailsTable.Clear()
		body.AddItem(transactionDetailsTable, 0, 1, true)

		setFocus(transactionDetailsTable)

		fetchTransactionDetail(txHash, wallet, displayError, transactionDetailsTable)
	})

	// handler for returning back to history table
	transactionDetailsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			clearFocus()
			return nil
		}
		if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
			displayHistoryTable()
			return nil
		}

		return event
	})

	tableHeaderCell := func(text string) *tview.TableCell {
		return tview.NewTableCell(text).SetAlign(tview.AlignCenter).SetSelectable(false)
	}

	// history table header
	historyTable.SetCell(0, 0, tableHeaderCell("Date (UTC)"))
	historyTable.SetCell(0, 1, tableHeaderCell("Direction"))
	historyTable.SetCell(0, 2, tableHeaderCell("Amount"))
	historyTable.SetCell(0, 3, tableHeaderCell("Status"))
	historyTable.SetCell(0, 4, tableHeaderCell("Type"))

	displayHistoryTable()

	fetchAndDisplayTransactions(-1, wallet, historyTable, displayError)

	hintTextView.SetText("TIP: Use ARROW UP/DOWN to select txn, ENTER to view details, ESC to return to navigation menu")

	setFocus(body)

	return body
}

func fetchAndDisplayTransactions(startBlockHeight int32, wallet walletcore.Wallet, historyTable *tview.Table, displayError func(errorMessage string)) {
	txns, endBlockHeight, err := wallet.TransactionHistory(context.Background(), startBlockHeight, walletcore.TransactionHistoryCountPerPage)
	if err != nil {
		displayError(err.Error())
		return
	}

	loc, _ := time.LoadLocation("UTC")
	currentDate := time.Now().In(loc)

	// var confirmations int32
	// confirmations := walletcore.DefaultRequiredConfirmations
	for _, tx := range txns {
		row := historyTable.GetRowCount()
		displayedTxHashes = append(displayedTxHashes, tx.Hash)

		txns, err := wallet.GetTransaction(tx.Hash)
		if err != nil {
			displayError(err.Error())
			return
		}

		transactionDate := time.Unix(tx.Timestamp, 0).In(loc)
		transactionAge := currentDate.Sub(transactionDate)

		txDateTime := strings.Split(tx.FormattedTime, " ")

		if transactionAge > _24hours {
			historyTable.SetCell(row, 0, tview.NewTableCell(txDateTime[0]).SetAlign(tview.AlignCenter))
		} else {
			historyTable.SetCell(row, 0, tview.NewTableCell(txDateTime[1]).SetAlign(tview.AlignCenter))
		}

		if txns.Confirmations >= walletcore.DefaultRequiredConfirmations {
			historyTable.SetCell(row, 3, tview.NewTableCell("Confirmed").SetAlign(tview.AlignCenter))
		} else {
			historyTable.SetCell(row, 3, tview.NewTableCell("Unconfirmed").SetAlign(tview.AlignCenter))
		}
		historyTable.SetCell(row, 1, tview.NewTableCell(tx.Direction.String()).SetAlign(tview.AlignCenter))
		historyTable.SetCell(row, 2, tview.NewTableCell(tx.Amount).SetAlign(tview.AlignCenter))
		historyTable.SetCell(row, 4, tview.NewTableCell(tx.Type).SetAlign(tview.AlignCenter))
	}

	if endBlockHeight > 0 {
		// set or reset selection changed listener to load more data when the table is almost scrolled to the end
		historyTable.SetSelectionChangedFunc(func(row, column int) {
			if row >= historyTable.GetRowCount()-10 {
				historyTable.SetSelectionChangedFunc(nil) // unset selection change listener until table is populated
				go fetchAndDisplayTransactions(endBlockHeight-1, wallet, historyTable, displayError)
			}
		})
	}

	return
}

func fetchTransactionDetail(txHash string, wallet walletcore.Wallet, displayError func(errorMessage string), transactionDetailsTable *tview.Table) {
	tx, err := wallet.GetTransaction(txHash)
	if err != nil {
		displayError(err.Error())
		return
	}

	transactionDetailsTable.SetCellSimple(0, 0, "Hash ")
	transactionDetailsTable.SetCellSimple(1, 0, "Confirmations")
	transactionDetailsTable.SetCellSimple(2, 0, "Included in block")
	transactionDetailsTable.SetCellSimple(3, 0, "Type")
	transactionDetailsTable.SetCellSimple(4, 0, "Amount received")
	transactionDetailsTable.SetCellSimple(5, 0, "Date")
	transactionDetailsTable.SetCellSimple(6, 0, "Direction")
	transactionDetailsTable.SetCellSimple(7, 0, "Fee")
	transactionDetailsTable.SetCellSimple(8, 0, "Fee Rate")

	transactionDetailsTable.SetCellSimple(0, 1, tx.Hash)
	transactionDetailsTable.SetCellSimple(1, 1, strconv.Itoa(int(tx.Confirmations)))
	transactionDetailsTable.SetCellSimple(2, 1, strconv.Itoa(int(tx.BlockHeight)))
	transactionDetailsTable.SetCellSimple(3, 1, tx.Type)
	transactionDetailsTable.SetCellSimple(4, 1, tx.Amount)
	transactionDetailsTable.SetCellSimple(5, 1, tx.FormattedTime)
	transactionDetailsTable.SetCellSimple(6, 1, tx.Direction.String())
	transactionDetailsTable.SetCellSimple(7, 1, tx.Fee)
	transactionDetailsTable.SetCellSimple(8, 1, fmt.Sprintf("%s/kB", tx.FeeRate))

	decimalPlaces := func(n float64) string {
		decimalPlaces := fmt.Sprintf("%f", n-math.Floor(n))
		decimalPlaces = strings.Replace(decimalPlaces, "0", "", -1)
		decimalPlaces = strings.Replace(decimalPlaces, ".", "", -1)
		return decimalPlaces
	}

	// calculate max number of digits after decimal point for inputs and outputs
	maxDecimalPlaces := 0
	for _, txIn := range tx.Inputs {
		decimalPlaces := decimalPlaces(dcrutil.Amount(txIn.AmountIn).ToCoin())
		nDecimalPlaces := len(decimalPlaces)
		if nDecimalPlaces > maxDecimalPlaces {
			maxDecimalPlaces = nDecimalPlaces
		}
	}
	for _, txOut := range tx.Outputs {
		decimalPlaces := decimalPlaces(dcrutil.Amount(txOut.Value).ToCoin())
		nDecimalPlaces := len(decimalPlaces)
		if nDecimalPlaces > maxDecimalPlaces {
			maxDecimalPlaces = nDecimalPlaces
		}
	}

	formatAmount := func(amount int64) string {
		dcrAmount := dcrutil.Amount(amount).ToCoin()
		wholeNumber := int(math.Floor(dcrAmount))
		decimalPlaces := decimalPlaces(dcrAmount)

		if len(decimalPlaces) == 0 {
			//decimalPlaces = "0"
			return fmt.Sprintf("%d%-*s DCR", wholeNumber, maxDecimalPlaces+1, decimalPlaces)
		}

		return fmt.Sprintf("%d.%-*s DCR", wholeNumber, maxDecimalPlaces, decimalPlaces)
	}

	transactionDetailsTable.SetCellSimple(9, 0, "-Inputs-")
	for _, txIn := range tx.Inputs {
		row := transactionDetailsTable.GetRowCount()
		transactionDetailsTable.SetCell(row, 0, tview.NewTableCell(formatAmount(txIn.AmountIn)).SetAlign(tview.AlignRight))
		transactionDetailsTable.SetCellSimple(row, 1, txIn.PreviousOutpoint)
	}

	row := transactionDetailsTable.GetRowCount()
	transactionDetailsTable.SetCellSimple(row, 0, "-Outputs-")
	for _, txOut := range tx.Outputs {
		row++
		if len(txOut.Addresses) == 0 {
			transactionDetailsTable.SetCellSimple(row, 0, fmt.Sprintf("  %s (no address)", dcrutil.Amount(txOut.Value).String()))
			continue
		}

		outputAmount := formatAmount(txOut.Value)
		for _, address := range txOut.Addresses {
			accountName := address.AccountName
			if !address.IsMine {
				accountName = "external"
			}

			transactionDetailsTable.SetCell(row, 0, tview.NewTableCell(outputAmount).SetAlign(tview.AlignRight))
			transactionDetailsTable.SetCellSimple(row, 1, fmt.Sprintf("%s (%s)", address.Address, accountName))
		}
	}
}
