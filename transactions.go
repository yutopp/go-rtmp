//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bytes"
	"io"
	"sync"

	"github.com/pkg/errors"

	"github.com/edgeware/go-rtmp/message"
)

type transaction struct {
	commandName string
	encoding    message.EncodingType
	body        *bytes.Buffer
	lastErr     error
	doneCh      chan struct{}
}

func (t *transaction) Reply(commandName string, encoding message.EncodingType, body io.Reader) {
	t.commandName = commandName
	t.encoding = encoding
	t.body = new(bytes.Buffer)
	_, err := io.Copy(t.body, body)
	t.lastErr = err
	close(t.doneCh)
}

type transactions struct {
	transactions map[int64]*transaction
	m            sync.RWMutex
}

func newTransactions() *transactions {
	return &transactions{
		transactions: make(map[int64]*transaction),
	}
}

func (ts *transactions) Create(transactionID int64) (*transaction, error) {
	ts.m.Lock()
	defer ts.m.Unlock()

	_, ok := ts.transactions[transactionID]
	if ok {
		return nil, errors.Errorf("Transaction already exists: TransactionID = %d", transactionID)
	}

	ts.transactions[transactionID] = &transaction{
		doneCh: make(chan struct{}),
	}

	return ts.transactions[transactionID], nil
}

func (ts *transactions) Delete(transactionID int64) error {
	ts.m.Lock()
	defer ts.m.Unlock()

	_, ok := ts.transactions[transactionID]
	if !ok {
		return errors.Errorf("Transaction not exists: TransactionID = %d", transactionID)
	}

	delete(ts.transactions, transactionID)

	return nil
}

func (ts *transactions) At(transactionID int64) (*transaction, error) {
	t, ok := ts.transactions[transactionID]
	if !ok {
		return nil, errors.Errorf("Transaction is not found: TransactionID = %d", transactionID)
	}

	return t, nil
}
