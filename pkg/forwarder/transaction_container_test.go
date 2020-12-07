package forwarder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionContainerAdd(t *testing.T) {
	a := assert.New(t)
	path, clean := createTmpFolder(a)
	defer clean()
	s, err := newTransactionsFileStorage(NewTransactionsSerializer(), path, 1000)
	a.NoError(err)
	container := newTransactionContainer(s, 100, 0.6)

	// When adding the last element `15`, the buffer becomes full and the first 3
	// transactions are flushed to the disk as 10 + 20 + 30 >= 100 * 0.6
	for _, payloadSize := range []int{10, 20, 30, 40, 15} {
		_, err := container.Add(createTransactionWithPayloadSize(payloadSize))
		a.NoError(err)
	}
	a.Equal(40+15, container.GetCurrentMemSizeInBytes())
	a.Equal(2, container.GetTransactionCount())

	assertPayloadSizeFromExtractTransactions(a, container, []int{40, 15})

	_, err = container.Add(createTransactionWithPayloadSize(5))
	a.NoError(err)
	a.Equal(5, container.GetCurrentMemSizeInBytes())
	a.Equal(1, container.GetTransactionCount())

	assertPayloadSizeFromExtractTransactions(a, container, []int{5})
	assertPayloadSizeFromExtractTransactions(a, container, []int{10, 20, 30})
	assertPayloadSizeFromExtractTransactions(a, container, nil)
}

func TestTransactionContainerSeveralFlushToDisk(t *testing.T) {
	a := assert.New(t)
	path, clean := createTmpFolder(a)
	defer clean()
	s, err := newTransactionsFileStorage(NewTransactionsSerializer(), path, 1000)
	a.NoError(err)
	container := newTransactionContainer(s, 50, 0.1)

	// Flush to disk when adding `40`
	for _, payloadSize := range []int{9, 10, 11, 40} {
		container.Add(createTransactionWithPayloadSize(payloadSize))
	}
	a.Equal(40, container.GetCurrentMemSizeInBytes())
	a.Equal(3, s.GetFilesCount())

	assertPayloadSizeFromExtractTransactions(a, container, []int{40})
	assertPayloadSizeFromExtractTransactions(a, container, []int{11})
	assertPayloadSizeFromExtractTransactions(a, container, []int{10})
	assertPayloadSizeFromExtractTransactions(a, container, []int{9})
	a.Equal(0, s.GetFilesCount())
	a.Equal(int64(0), s.GetCurrentSizeInBytes())
}

func TestTransactionContainerNoTransactionStorage(t *testing.T) {
	a := assert.New(t)
	container := newTransactionContainer(nil, 50, 0.1)

	for _, payloadSize := range []int{9, 10, 11} {
		dropCount, err := container.Add(createTransactionWithPayloadSize(payloadSize))
		a.Equal(0, dropCount)
		a.NoError(err)
	}

	// Drop when adding `30`
	dropCount, err := container.Add(createTransactionWithPayloadSize(30))
	a.Equal(2, dropCount)
	a.NoError(err)

	a.Equal(11+30, container.GetCurrentMemSizeInBytes())

	assertPayloadSizeFromExtractTransactions(a, container, []int{11, 30})
}

func createTransactionWithPayloadSize(payloadSize int) *HTTPTransaction {
	tr := NewHTTPTransaction()
	payload := make([]byte, payloadSize)
	tr.Payload = &payload
	return tr
}

func assertPayloadSizeFromExtractTransactions(
	a *assert.Assertions,
	container *transactionContainer,
	expectedPayloadSize []int) {

	transactions, err := container.ExtractTransactions()
	a.NoError(err)
	a.Equal(0, container.GetCurrentMemSizeInBytes())

	var payloadSizes []int
	for _, t := range transactions {
		payloadSizes = append(payloadSizes, t.GetPayloadSize())
	}
	a.EqualValues(expectedPayloadSize, payloadSizes)
}