//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/pkg/errors"
	"sync"

	"github.com/yutopp/go-rtmp/message"
)

type transaction struct {
	decoder  message.BodyDecoderFunc
	callback func(v interface{}, err error)
}

type transactions struct {
	transactions map[int64]transaction
	m            sync.RWMutex
}

func newTransactions() *transactions {
	return &transactions{
		transactions: make(map[int64]transaction),
	}
}

func (ts *transactions) Create(transactionID int64, t transaction) error {
	ts.m.Lock()
	defer ts.m.Unlock()

	_, ok := ts.transactions[transactionID]
	if ok {
		return errors.Errorf("Transaction already exists: TransactionID = %d", transactionID)
	}

	ts.transactions[transactionID] = t

	return nil
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

func (ts *transactions) At(transactionID int64) (transaction, error) {
	t, ok := ts.transactions[transactionID]
	if !ok {
		return transaction{}, errors.Errorf("Transaction is not found: TransactionID = %d", transactionID)
	}

	return t, nil
}
