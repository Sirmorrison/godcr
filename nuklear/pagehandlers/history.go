package pagehandlers

import (
	"fmt"

	"github.com/aarzilli/nucular"
	"github.com/decred/dcrd/dcrutil"
	"github.com/raedahgroup/dcrlibwallet/txindex"
	"github.com/raedahgroup/godcr/app/config"
	"github.com/raedahgroup/godcr/app/walletcore"
	"github.com/raedahgroup/godcr/nuklear/styles"
	"github.com/raedahgroup/godcr/nuklear/widgets"
)

type HistoryHandler struct {
	wallet               walletcore.Wallet
	refreshWindowDisplay func()

	filterSelectorWidget *widgets.FilterSelector
	filterSelectorErr    error
	currentFilterText    string

	txCountForCurrentFilter int
	currentPage             int
	txPerPage               int
	transactions            []*walletcore.Transaction
	fetchHistoryError       error
	isFetchingTransactions  bool

	selectedTxHash      string
	selectedTxDetails   *walletcore.Transaction
	isFetchingTxDetails bool
	fetchTxDetailsError error
}

func (handler *HistoryHandler) BeforeRender(wallet walletcore.Wallet, settings *config.Settings, refreshWindowDisplay func()) bool {
	handler.wallet = wallet
	handler.refreshWindowDisplay = refreshWindowDisplay

	handler.currentPage = 1
	handler.txPerPage = walletcore.TransactionHistoryCountPerPage
	handler.transactions = nil
	handler.isFetchingTransactions = false

	handler.clearTxDetails()

	// fetch initial table data
	handler.currentFilterText = "All"
	handler.txCountForCurrentFilter, handler.fetchHistoryError = wallet.TransactionCount(nil)
	if handler.fetchHistoryError != nil {
		// no need to fetch txs or setup filter widget if there was an error getting total tx count.
		return true
	}

	go handler.fetchTransactions(nil)

	// set up the filter widget
	handler.filterSelectorWidget, handler.filterSelectorErr = widgets.FilterSelectorWidget(wallet, func() {
		selectedFilterText, txCountForSelectedFilter := handler.filterSelectorWidget.GetSelectedFilter()
		if selectedFilterText != handler.currentFilterText {
			handler.currentFilterText = selectedFilterText
			handler.txCountForCurrentFilter = txCountForSelectedFilter
			handler.transactions = nil
			go handler.fetchTransactions(walletcore.BuildTransactionFilter(handler.currentFilterText))
		}
	})

	return true
}

func (handler *HistoryHandler) fetchTransactions(filter *txindex.ReadFilter) {
	handler.isFetchingTransactions = true
	handler.refreshWindowDisplay() // refresh display to show loading indicator

	txHistoryOffset := 0
	if handler.transactions != nil {
		txHistoryOffset = len(handler.transactions)
	}

	transactions, err := handler.wallet.TransactionHistory(int32(txHistoryOffset), int32(handler.txPerPage), filter)
	handler.fetchHistoryError = err
	handler.transactions = append(handler.transactions, transactions...)

	handler.isFetchingTransactions = false
	handler.refreshWindowDisplay()
}

func (handler *HistoryHandler) Render(window *nucular.Window) {
	if handler.selectedTxHash == "" {
		handler.renderHistoryPage(window)
		return
	}
	handler.renderTransactionDetailsPage(window)
}

func (handler *HistoryHandler) renderHistoryPage(window *nucular.Window) {
	widgets.PageContentWindowDefaultPadding("History", window, func(contentWindow *widgets.Window) {
		if handler.filterSelectorWidget != nil {
			handler.filterSelectorWidget.Render(contentWindow)
		}

		if handler.filterSelectorErr != nil {
			contentWindow.DisplayErrorMessage("Error with filter selector", handler.filterSelectorErr)
		}

		if len(handler.transactions) == 0 {
			contentWindow.AddWrappedLabel("No transactions to display yet", widgets.CenterAlign)
		} else {
			handler.displayTransactions(contentWindow)
		}

		// show fetch error if any
		if handler.fetchHistoryError != nil {
			contentWindow.DisplayErrorMessage("Error loading history", handler.fetchHistoryError)
		}
		// show loading indicator if tx fetching is progress
		if handler.isFetchingTransactions {
			contentWindow.DisplayIsLoadingMessage()
		}
	})
}

func (handler *HistoryHandler) displayTransactions(contentWindow *widgets.Window) {
	historyTable := widgets.NewTable()

	// render table header with nav font
	historyTable.AddRowWithFont(styles.NavFont,
		widgets.NewLabelTableCell("#", "LC"),
		widgets.NewLabelTableCell("Date", "LC"),
		widgets.NewLabelTableCell("Direction", "LC"),
		widgets.NewLabelTableCell("Amount", "LC"),
		widgets.NewLabelTableCell("Fee", "LC"),
		widgets.NewLabelTableCell("Type", "LC"),
		widgets.NewLabelTableCell("Hash", "LC"),
	)

	pageTxOffset := (handler.currentPage - 1) * handler.txPerPage
	maxTxIndexForCurrentPage := pageTxOffset + handler.txPerPage
	for currentTxIndex, tx := range handler.transactions {
		if currentTxIndex < pageTxOffset {
			continue // skip txs not belonging to this page
		}
		if currentTxIndex >= maxTxIndexForCurrentPage {
			break // max number of tx displayed for this page
		}

		historyTable.AddRow(
			widgets.NewLabelTableCell(fmt.Sprintf("%d", currentTxIndex+1), "LC"),
			widgets.NewLabelTableCell(tx.ShortTime, "LC"),
			widgets.NewLabelTableCell(tx.Direction.String(), "LC"),
			widgets.NewLabelTableCell(dcrutil.Amount(tx.Amount).String(), "RC"),
			widgets.NewLabelTableCell(dcrutil.Amount(tx.Fee).String(), "RC"),
			widgets.NewLabelTableCell(tx.Type, "LC"),
			widgets.NewLinkTableCell(tx.Hash, "Click to see transaction details", handler.gotoTransactionDetails),
		)
	}
	historyTable.Render(contentWindow)

	if !handler.isFetchingTransactions {
		contentWindow.Row(40).Static(110, 110)

		// show previous button only if current page is greater than 1
		if handler.currentPage > 1 {
			contentWindow.AddButtonToCurrentRow("Previous", func() {
				handler.loadPreviousPage(contentWindow)
			})
		}

		// show next button only if there are more txs to be loaded
		if handler.txCountForCurrentFilter > maxTxIndexForCurrentPage {
			contentWindow.AddButtonToCurrentRow("Next", func() {
				handler.loadNextPage(contentWindow)
			})
		}
	}
}

func (handler *HistoryHandler) loadPreviousPage(window *widgets.Window) {
	handler.currentPage--
	window.Master().Changed()
}

func (handler *HistoryHandler) loadNextPage(window *widgets.Window) {
	nextPage := handler.currentPage + 1
	handler.currentPage = nextPage

	nextPageTxOffset := (nextPage - 1) * handler.txPerPage
	if nextPageTxOffset >= len(handler.transactions) {
		// we've not loaded txs for this page
		go handler.fetchTransactions(walletcore.BuildTransactionFilter(handler.currentFilterText))
	}

	handler.refreshWindowDisplay()
}
