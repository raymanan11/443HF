/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	"strconv"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type serverConfig struct {
	CCID    string
	Address string
}

// SmartContract provides functions for managing an asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type Asset struct {
	Name   string `json:"name"`
	IsDefect  bool `json:"isDefect"`
	SerialNumber string `json:"serialNumber"`
	Owner  string `json:"owner"`
}


// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResult struct {
	Record    *Asset    `json:"record"`
	TxId     string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string `json:"Key"`
	Record *Asset
}

// InitLedger adds a base set of cars to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{Name: "Flight Controls", IsDefect: false, SerialNumber: "E96GJE93D", Owner: "United Airlines"},
		{Name: "Landing Gear", IsDefect: false, SerialNumber: "U46834HJ3", Owner: "American Airlines"},
		{Name: "Fuselage", IsDefect: true, SerialNumber: "FOIE463U2", Owner: "Delta"},
		{Name: "Rudder Pedals", IsDefect: false, SerialNumber: "DFU9436OB", Owner: "Spirit"},
		{Name: "Instrument Panels", IsDefect: false, SerialNumber: "FJE582KFD3", Owner: "Frontier"},
		{Name: "Engine", IsDefect: true, SerialNumber: "DFJRO895D", Owner: "Alaska Airlines"},
		{Name: "Wings", IsDefect: false, SerialNumber: "RID5569D2", Owner: "Southwest Airlines"},
		{Name: "Rudders", IsDefect: false, SerialNumber: "TOIE835D3", Owner: "JetBlue"},
		{Name: "Vertical Stabalizer", IsDefect: false, SerialNumber: "TI45GMD32W", Owner: "Hawaiian Airlines"},
		{Name: "Overhead Panel", IsDefect: true, SerialNumber: "EKLF8534H", Owner: "Allegiant Air"},
	}

	for i, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState("PART"+strconv.Itoa(i), assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, airlinePartNumber string, name string, isDefect bool, serialNumber string, owner string) error {
	exists, err := s.AssetExists(ctx, airlinePartNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", airlinePartNumber)
	}
	asset := Asset{
		Name:   name,
		IsDefect:  isDefect,
		SerialNumber: serialNumber,
		Owner:  owner,
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(airlinePartNumber, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, airlinePartNumber string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(airlinePartNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", airlinePartNumber)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, airlinePartNumber string, name string, isDefect bool, serialNumber string, owner string) error {
	exists, err := s.AssetExists(ctx, airlinePartNumber)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", airlinePartNumber)
	}

	// overwritting original asset with new asset
	asset := Asset{
		Name:   name,
		IsDefect:  isDefect,
		SerialNumber: serialNumber,
		Owner:  owner,
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(airlinePartNumber, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, airlinePartNumber string) error {
	exists, err := s.AssetExists(ctx, airlinePartNumber)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", airlinePartNumber)
	}

	return ctx.GetStub().DelState(airlinePartNumber)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, airlinePartNumber string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(airlinePartNumber)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	return assetJSON != nil, nil
}

// TransferAsset updates the owner field of asset with given id in world state.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, airlinePartNumber string, newOwner string) error {
	asset, err := s.ReadAsset(ctx, airlinePartNumber)
	if err != nil {
		return err
	}

	asset.Owner = newOwner
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(airlinePartNumber, assetJSON)
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	// range query with empty string for startKey and endKey does an open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var results []QueryResult

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}

		queryResult := QueryResult{Key: queryResponse.Key, Record: &asset}
		results = append(results, queryResult)
	}

	return results, nil
}

// GetAssetHistory returns the chain of custody for an asset since issuance.
func (t *SmartContract) GetAssetHistory(ctx contractapi.TransactionContextInterface, airlinePartNumber string) ([]HistoryQueryResult, error) {
	log.Printf("GetAssetHistory: PartNumber %v", airlinePartNumber)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(airlinePartNumber)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, err
			}
		} else {
			asset = Asset{
				
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &asset,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}


func main() {
	// See chaincode.env.example
	config := serverConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	chaincode, err := contractapi.NewChaincode(&SmartContract{})

	if err != nil {
		log.Panicf("error create asset-transfer-basic chaincode: %s", err)
	}

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      chaincode,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	if err := server.Start(); err != nil {
		log.Panicf("error starting asset-transfer-basic chaincode: %s", err)
	}
}
