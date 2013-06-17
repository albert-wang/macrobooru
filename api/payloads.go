package api

import (
	"sync"
)

var (
	operationRegistry     map[string]Operation
	operationRegistryLock sync.RWMutex
)

func RegisterOperation(operationType Operation) {
	operationRegistryLock.Lock()
	defer operationRegistryLock.Unlock()

	if operationRegistry == nil {
		operationRegistry = make(map[string]Operation)
	}

	operationRegistry[operationType.Name()] = operationType
}

func OperationByName(name string) Operation {
	operationRegistryLock.RLock()
	defer operationRegistryLock.RUnlock()

	if operation, ok := operationRegistry[name]; ok {
		return operation
	}

	return nil
}

func (req *RequestWrapper) Parse() error {
	operationRegistryLock.RLock()
	defer operationRegistryLock.RUnlock()

	operation, ok := operationRegistry[req.Operation]
	if !ok {
		return ErrorInvalidInputFormat("invalid operation")
	}

	payload, er := operation.Parse(req)
	if er != nil {
		return er
	}

	req.Data = payload
	return nil
}
