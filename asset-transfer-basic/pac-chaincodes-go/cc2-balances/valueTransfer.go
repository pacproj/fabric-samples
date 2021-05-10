/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-samples/asset-transfer-basic/pac-chaincodes-go/cc2-balances/cc2"
)

func main() {
	assetChaincode, err := contractapi.NewChaincode(&cc2.SmartContract{})
	if err != nil {
		log.Panicf("Error creating cc1: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting cc1: %v", err)
	}
}
