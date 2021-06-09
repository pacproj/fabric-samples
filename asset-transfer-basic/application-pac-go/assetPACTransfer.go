/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	mspproviders "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

const (
	org  = "Org3"
	user = "User1"
	caId = "ca.org3.example.com"
)

func main() {

	log.Println("============ PAC sample application-golang starts ============")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	fabricConnectionProfile := filepath.Join(
		"..",
		"..",
		"..",
		"connectionprofile.yaml",
	)
	channel1 := "mychannel1"
	channel2 := "mychannel2"

	client1 := NewFabricClient(fabricConnectionProfile, channel1)
	client2 := NewFabricClient(fabricConnectionProfile, channel2)
	ch1 := client1.channelClient()
	ch2 := client2.channelClient()

	//map contains dependency list of the pac-transaction
	tmap := make(map[string][]byte)
	tmap["pac"] = []byte("Test attempt")
	tmap["pacpart1"] = []byte("mychannel1")
	tmap["pacpart2"] = []byte("mychannel2")

	var stubHashPair *common.HashPair = nil

	//Init ledger in channel1
	PrintRequest(channel1)
	resultCh1, err := ch1.Execute(channel.Request{
		ChaincodeID: "cc1",
		Fcn:         "InitLedger",
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	PrintResult(resultCh1, stubHashPair, "InitLedger", "", channel1)
	//QUERY GetAllAssets
	readChannelData(ch1, channel1, stubHashPair)

	//Init ledger in channel2
	PrintRequest(channel2)
	resultCh2, err := ch2.Execute(channel.Request{
		ChaincodeID: "cc2",
		Fcn:         "InitLedger",
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	PrintResult(resultCh2, stubHashPair, "InitLedger", "", channel2)
	//QUERY GetAllAccounts
	readChannelData(ch2, channel2, stubHashPair)

	//Send PAC request to mychannel1
	PrintPACRequest(channel1, "INITIAL_REQUEST", "transfer asset4 from Max to Ivan")
	resp1Ch1, err := ch1.Execute(channel.Request{
		ChaincodeID:  "cc1",
		Fcn:          "TransferAsset",
		Args:         [][]byte{[]byte("asset4"), []byte("Ivan")},
		TransientMap: tmap,
		PACClientData: fab.ClientData{
			RequestedTransaction: fab.PACRequest,
		},
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}

	//TODO: check if Status == 200, then data Message is OK for PAC.
	mychannel1HashPair := getShardHashPair(resp1Ch1.Responses[0].ProposalResponse.Response.Message)
	mychannel1HPBytes, err := base64.StdEncoding.DecodeString(resp1Ch1.Responses[0].ProposalResponse.Response.Message)
	if err != nil {
		log.Fatalf("Failed to decode based64 response message: %v", err)
	}

	PrintResult(resp1Ch1, &mychannel1HashPair, "INITIAL_REQUEST", "transfer asset4 from Max to Ivan", channel1)
	//QUERY GetAllAssets
	readChannelData(ch1, channel1, &mychannel1HashPair)

	//Send PAC request to mychannel2
	PrintPACRequest(channel2, "INITIAL_REQUEST", "update account4 (+600)")
	resp1Ch2, err := ch2.Execute(channel.Request{
		ChaincodeID:  "cc2",
		Fcn:          "UpdateAccount",
		Args:         [][]byte{[]byte("account4"), []byte("Max"), []byte("666.192")},
		TransientMap: tmap,
		PACClientData: fab.ClientData{
			RequestedTransaction: fab.PACRequest,
		},
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	//TODO: check if payload is 200, then the message is OK for PAC
	mychannel2HashPair := getShardHashPair(resp1Ch2.Responses[0].ProposalResponse.Response.Message)
	mychannel2HPBytes, err := base64.StdEncoding.DecodeString(resp1Ch2.Responses[0].ProposalResponse.Response.Message)
	if err != nil {
		log.Fatalf("Failed to decode based64 response message: %v", err)
	}
	PrintResult(resp1Ch2, &mychannel2HashPair, "INITIAL_REQUEST", "update account4 (+600)", channel2)
	//QUERY GetAllAssets
	readChannelData(ch2, channel2, &mychannel2HashPair)

	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//TODO: implement hashes handling from other shards!!!
	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	//add hash pairs into AC temp map
	tmap["pacHP"] = []byte("")
	tmap["pacpart1HP"] = []byte(mychannel1HPBytes)
	tmap["pacpart2HP"] = []byte(mychannel2HPBytes)
	//TODO: is it safe to send txids in the open way?
	//we need to send txid for every channel to update
	//its corresponding local file with dependency list which name consists of txid
	tmap["pactxid"] = []byte(resp1Ch1.Proposal.PACClientData.BCS.ResponseProposal.TxnID)

	//The hash pairs spreading to mychannel1:
	PrintPACRequest(channel1, "SPREAD_HASHPAIRS", "spread HashPairs")
	bcs1 := resp1Ch1.Proposal.PACClientData.BCS
	resp2Ch1, err := ch1.Execute(channel.Request{
		ChaincodeID:  "cc1",
		Fcn:          "TransferAsset",
		Args:         [][]byte{[]byte("asset4"), []byte("Ivan")},
		TransientMap: tmap,
		PACClientData: fab.ClientData{
			RequestedTransaction: fab.PACRequest,
			ValidationData:       []byte(""), //TODO add validation data!
			BCS:                  bcs1,
		},
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	PrintResult(resp2Ch1, &mychannel1HashPair, "SPREAD_HASHPAIRS", "spread HashPairs", channel1)

	//The hash pairs spreading to mychannel1:
	tmap["pactxid"] = []byte(resp1Ch2.Proposal.PACClientData.BCS.ResponseProposal.TxnID)
	PrintPACRequest(channel2, "SPREAD_HASHPAIRS", "spread HashPairs")
	bcs2 := resp1Ch2.Proposal.PACClientData.BCS
	resp2Ch2, err := ch2.Execute(channel.Request{
		ChaincodeID:  "cc2",
		Fcn:          "UpdateAccount",
		Args:         [][]byte{[]byte("account4"), []byte("Max"), []byte("666.192")},
		TransientMap: tmap,
		PACClientData: fab.ClientData{
			RequestedTransaction: fab.PACRequest,
			ValidationData:       []byte(""), //TODO add validation data!
			BCS:                  bcs2,
		},
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	PrintResult(resp2Ch2, &mychannel2HashPair, "SPREAD_HASHPAIRS", "spread HashPairs", channel2)

	PrintPACRequest(channel1, "PREPARE_TX", "transfer asset4 from Max to Ivan")
	//The PrepareTx creation for shard mychannel1:
	resp3Ch1, err := ch1.Execute(channel.Request{
		ChaincodeID:  "cc1",
		Fcn:          "TransferAsset",
		Args:         [][]byte{[]byte("asset4"), []byte("Ivan")},
		TransientMap: tmap,
		PACClientData: fab.ClientData{
			RequestedTransaction: fab.PrepareTxRequest,
			ValidationData:       []byte(""), //TODO add validation data!
			BCS:                  bcs1,
		},
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	PrintResult(resp3Ch1, &mychannel1HashPair, "PREPARE_TX", "transfer asset4 from Max to Ivan", channel1)

	//Test sending the same PrepareTx after previous sending
	//Wrong REPEATED PrepareTx creation for shard mychannel1:
	//WRONG REPEATED PREPARE_TX!
	PrintPACRequest(channel1, "WRONG REPEATED PREPARE_TX", "transfer asset4 from Max to Ivan")
	wrongResp3Ch1, err := ch1.Execute(channel.Request{
		ChaincodeID:  "cc1",
		Fcn:          "TransferAsset",
		Args:         [][]byte{[]byte("asset4"), []byte("Ivan")},
		TransientMap: tmap,
		PACClientData: fab.ClientData{
			RequestedTransaction: fab.PrepareTxRequest,
			ValidationData:       []byte(""),
			BCS:                  bcs1,
		}, //WRONG REPEATED PREPARE_TX!
	})
	//WRONG REPEATED PREPARE_TX!
	if err != nil {
		log.Printf("Failed to Submit transaction: %v", err)
	}
	PrintResult(wrongResp3Ch1, &mychannel1HashPair, "WRONG REPEATED PREPARE_TX", "transfer asset4 from Max to Ivan", channel1)

	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//TODO: add test here to check if WSet keys has become locked
	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	//The DecideTx creation for shard mychannel1:
	PrintPACRequest(channel1, "DECIDE_TX", "transfer asset4 from Max to Ivan")
	resp4Ch1, err := ch1.Execute(channel.Request{
		ChaincodeID:  "cc1",
		Fcn:          "TransferAsset",
		Args:         [][]byte{[]byte("asset4"), []byte("Ivan")},
		TransientMap: tmap,
		PACClientData: fab.ClientData{
			RequestedTransaction: fab.DecideTxRequest,
			ValidationData:       []byte(""),
			BCS:                  bcs1,
		},
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	PrintResult(resp4Ch1, &mychannel1HashPair, "DECIDE_TX", "transfer asset4 from Max to Ivan", channel1)

	//QUERY GetAllAssets
	readChannelData(ch1, channel1, &mychannel1HashPair)

	//=======================================
	//======TEST ABORT_TRANSACTION===========
	//=======================================
	//checkAbortTransaction(channel1, ch1, tmap)

	client1.Close()
	client2.Close()

	log.Println("============ application-golang ends ============")
}

func readChannelData(ch *channel.Client, channelName string, chanHashPair *common.HashPair) {
	var ccName, ccCommand string
	if channelName == "mychannel1" {
		ccName = "cc1"
		ccCommand = "GetAllAssets"
	} else {
		ccName = "cc2"
		ccCommand = "GetAllAccounts"
	}
	resultCh, err := ch.Query(channel.Request{
		ChaincodeID: ccName,
		Fcn:         ccCommand,
	})
	if err != nil {
		log.Fatalf("Failed to Query transaction: %v", err)
	}
	PrintResult(resultCh, chanHashPair, ccCommand, "", channelName)
}

func getShardHashPair(message string) common.HashPair {
	mes, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		log.Fatalf("Failed to decode based64 response message: %v", err)
	}
	HashPair := common.HashPair{}
	err = proto.Unmarshal(mes, &HashPair)
	if err != nil {
		log.Fatal(err)
	}
	return HashPair
}

func PrintPACRequest(channel string, txType string, requestDescription string) {
	log.Printf("\n\n\n\n\n\n\n\n\n\n\n\n")
	log.Println("===============================================================")
	log.Printf("=======================~~~%s~~~========================", channel)
	log.Println("=================~~~PAC MESSAGE START~~~=======================")
	log.Printf("===================~~~%s~~~====================", txType)
	log.Printf("Description: %s", requestDescription)
	log.Println("===============================================================")
}

func PrintRequest(channel string) {
	log.Printf("\n\n\n\n\n\n\n\n\n\n\n\n")
	log.Println("===============================================================")
	log.Printf("=======================~~~%s~~~========================", channel)
	log.Println("=================~~~Request start~~~===========================")
	log.Println("===============================================================")
	log.Println("===============================================================")
}

func PrintResult(response channel.Response, chanHashPair *common.HashPair, CCFunc string, CCDescription string, channel string) {
	gotMessage := response.Responses[0].ProposalResponse.Response.Message

	requestTypeWas := ""
	if response.Payload != nil {
		requestTypeWas = "QUERY"
	} else {
		requestTypeWas = "INVOKE"
	}

	log.Println("===============================================================")
	log.Println("===============================================================")
	log.Println("===============================================================")
	log.Printf("~~~~~~~~~~~~~~~~~~~  %s:   %s   ~~~~~~~~~~~~~~~~~~~~~", requestTypeWas, CCFunc)
	log.Printf("DESCRIPTION: %s", CCDescription)
	log.Printf("TX_VALIDATION_CODE %d: %s", response.TxValidationCode, response.TxValidationCode.String())
	log.Println("===============================================================")
	log.Printf("=======================~~~%s~~~========================", channel)
	log.Println("=======================~~~Request end~~~=======================")

	var prettyPayload bytes.Buffer
	err := json.Indent(&prettyPayload, response.Payload, "", "\t")
	if err != nil {
		log.Println("Error getting pretty payload:", err)
		log.Printf("\n\n NOT PRETTY result.Payload is:\n %s", response.Payload)
	}
	log.Printf("\n\nresult.Payload is:\n %s", prettyPayload)
	log.Printf("--> Channel %s was requested Transaction for CC func: %s, ", channel, CCFunc)
	log.Printf("Response.Message is:\n%s\n", gotMessage)
	log.Printf("Full response struct: [%+v]", response.Responses[0].ProposalResponse.Response)
	log.Println("Vsr, Nsr struct:")
	if chanHashPair != nil {
		log.Printf("Struct:[%v], where\n Vsr: [%s], Nsr: [%s]", chanHashPair, chanHashPair.HashedVsr, chanHashPair.HashedNsr)
	} else {
		log.Printf("u didn't use PAC request")
	}
}

type fabricClient struct {
	sdk         *fabsdk.FabricSDK
	channelName string
	sid         mspproviders.SigningIdentity
}

type MSPIdentity struct {
	key  []byte
	cert []byte
}

func NewFabricClient(profile string, channelName string) *fabricClient {
	connectionProfile := config.FromFile(profile)

	sdk, err := fabsdk.New(connectionProfile)
	if err != nil {
		panic(err)
	}
	// Create signing identity based on certificate and private key
	// Create msp client
	c, err := msp.New(sdk.Context(), msp.WithOrg(org), msp.WithCAInstance(caId))
	if err != nil {
		log.Fatalf("failed to create msp client\n")
	}
	identity := getCertAndKey()

	sid, err := c.CreateSigningIdentity(mspproviders.WithCert(identity.cert), mspproviders.WithPrivateKey(identity.key))
	if err != nil {
		log.Fatalf("failed when creating identity based on certificate and private key: %s\n", err)
	}

	return &fabricClient{
		sdk:         sdk,
		channelName: channelName,
		sid:         sid,
	}
}

func (f *fabricClient) channelClient() *channel.Client {
	//f.sdk.provider.IdentityManager(opts.orgName)
	channelCtx := f.sdk.ChannelContext(f.channelName,
		fabsdk.WithUser(user),
		fabsdk.WithOrg(org),
		fabsdk.WithIdentity(f.sid))
	ch, err := channel.New(channelCtx)
	if err != nil {
		panic(err)
	}
	return ch
}

func (f *fabricClient) Close() {
	f.sdk.Close()
}

func getCertAndKey() *MSPIdentity {
	credPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org3.example.com",
		"users",
		"User1@org3.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "User1@org3.example.com-cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		log.Fatalf("failed to get certificate: %s\n", err)
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	if len(files) != 1 {
		log.Fatalf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	return &MSPIdentity{
		key:  key,
		cert: cert,
	}
}

/*func checkAbortTransaction(channel1 string, ch1 *channel.Client, tmap map[string][]byte) {

	//TODO: add validation data
	var stubHashPair *common.HashPair = nil
	txDescription := "transfer asset1 from Tomoko to John"
	args := [][]byte{[]byte("asset1"), []byte("John")}

	//Send PAC request to mychannel1 for AbortTx
	PrintPACRequest(channel1, "INITIAL_REQUEST", txDescription)
	resp1Ch1, err := ch1.Execute(channel.Request{
		ChaincodeID:  "cc1",
		Fcn:          "TransferAsset",
		Args:         args,
		TransientMap: tmap,
		PACClientData: fab.ClientData{
			RequestedTransaction: fab.PACRequest,
		},
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}

	//TODO: check if Status == 200, then data Message is OK for PAC.
	message1, err := base64.StdEncoding.DecodeString(resp1Ch1.Responses[0].ProposalResponse.Response.Message)
	if err != nil {
		log.Fatalf("Failed to decode based64 response message: %v", err)
	}
	mychannel1HashPair := common.HashPair{}
	err = proto.Unmarshal(message1, &mychannel1HashPair)
	if err != nil {
		log.Fatal(err)
	}
	PrintResult(resp1Ch1, &mychannel1HashPair, "INITIAL_REQUEST", txDescription, channel1)
	//QUERY GetAllAssets
	readChannelData(ch1, channel1, &mychannel1HashPair)

	PrintPACRequest(channel1, "PREPARE_TX", txDescription)
	//The PrepareTx creation for shard mychannel1:
	bcs1 := resp1Ch1.Proposal.PACClientData.BCS
	resp2Ch1, err := ch1.Execute(channel.Request{
		ChaincodeID:  "cc1",
		Fcn:          "TransferAsset",
		Args:         args,
		TransientMap: tmap,
		PACClientData: fab.ClientData{
			RequestedTransaction: fab.PrepareTxRequest,
			ValidationData:       []byte(""), //TODO add validation data!
			BCS:                  bcs1,
		},
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	PrintResult(resp2Ch1, &mychannel1HashPair, "PREPARE_TX", txDescription, channel1)

	//The AbortTx creation for shard mychannel1:
	PrintPACRequest(channel1, "ABORT_TX", txDescription)
	resp3Ch1, err := ch1.Execute(channel.Request{
		ChaincodeID:  "cc1",
		Fcn:          "TransferAsset",
		Args:         args,
		TransientMap: tmap,
		PACClientData: fab.ClientData{
			RequestedTransaction: fab.AbortTxRequest,
			ValidationData:       []byte(""),
			BCS:                  bcs1,
		},
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	PrintResult(resp3Ch1, &mychannel1HashPair, "DECIDE_TX", "transfer asset4 from Max to Ivan", channel1)

	//QUERY GetAllAssets
	readChannelData(ch1, channel1, stubHashPair)

	//Check that key asset1 is unlocked
	PrintPACRequest(channel1, "ENDORSEMENT_TX", "!!!transfer asset1 from Tomoko to Pet")
	//The PrepareTx creation for shard mychannel1:
	respCh1, err := ch1.Execute(channel.Request{
		ChaincodeID: "cc1",
		Fcn:         "TransferAsset",
		Args:        [][]byte{[]byte("asset1"), []byte("Pet")},
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	PrintResult(respCh1, &mychannel1HashPair, "ENDORSEMENT_TX", "transfer asset1 from Tomoko to Pet", channel1)

	//QUERY GetAllAssets
	readChannelData(ch1, channel1, stubHashPair)
}*/
