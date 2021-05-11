/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"

	//"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	mspproviders "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const (
	org  = "Org3"
	user = "User1"
	caId = "ca.org3.example.com"
	//caName = "ca-org1"
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

	//Send PAC request to mychannel1
	resultCh1, err := ch1.Execute(channel.Request{
		ChaincodeID: "cc1",
		Fcn:         "InitLedger",
		/*TransientMap: tmap,
		PACClientData: fab.ClientData{
			RequestedTransaction: fab.PACRequest,
		},*/
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}

	//TODO: check if Status == 200, then data Message is OK for PAC.

	var mychannel1HashPair *common.HashPair = nil
	/*message1, err := base64.StdEncoding.DecodeString(gotMessage)
	if err != nil {
		log.Fatalf("Failed to decode based64 response message: %v", err)
	}
	mychannel1HashPair := common.HashPair{}
	err = proto.Unmarshal(message1, &mychannel1HashPair)
	if err != nil {
		log.Fatal(err)
	}*/

	PrintResult(resultCh1, mychannel1HashPair, "InitLedger", channel1)

	//query GetAllAssets
	resultCh1, err = ch1.Query(channel.Request{
		ChaincodeID: "cc1",
		Fcn:         "GetAllAssets",
	}) //QUERY!!!
	if err != nil {
		log.Fatalf("Failed to Query transaction: %v", err)
	}
	//QUERY!!!
	//QUERY!!!
	PrintResult(resultCh1, mychannel1HashPair, "GetAllAssets", channel1)

	//Send PAC request to mychannel2
	resultCh2, err := ch2.Execute(channel.Request{
		ChaincodeID: "cc2",
		Fcn:         "InitLedger",
		/*TransientMap: tmap,
		PACClientData: fab.ClientData{
			RequestedTransaction: fab.PACRequest,
		},*/
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}

	//TODO: check if payload is 200, then the message is OK for PAC
	/*message2, err := base64.StdEncoding.DecodeString(gotMessage)
	if err != nil {
		log.Fatalf("Failed to decode based64 response message: %v", err)
	}
	mychannel2HashPair := common.HashPair{}
	err = proto.Unmarshal(message2, &mychannel2HashPair)
	if err != nil {
		log.Fatal(err)
	}*/
	var mychannel2HashPair *common.HashPair = nil

	PrintResult(resultCh2, mychannel2HashPair, "InitLedger", channel2)

	//query GetAllAssets
	resultCh2, err = ch2.Query(channel.Request{
		ChaincodeID: "cc2",
		Fcn:         "GetAllAccounts",
	}) //QUERY!!!
	if err != nil {
		log.Fatalf("Failed to Query transaction: %v", err)
	}
	//QUERY!!!
	//QUERY!!!

	PrintResult(resultCh2, mychannel2HashPair, "GetAllAccounts", channel2)

	//Send PAC request to mychannel1
	resultCh1, err = ch1.Execute(channel.Request{
		ChaincodeID: "cc1",
		Fcn:         "TransferAsset",
		Args:        [][]byte{[]byte("asset4"), []byte("Ivan")},
		/*TransientMap: tmap,
		PACClientData: fab.ClientData{
			RequestedTransaction: fab.PACRequest,
		},*/
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}

	//TODO: check if Status == 200, then data Message is OK for PAC.

	/*message1, err := base64.StdEncoding.DecodeString(gotMessage)
	if err != nil {
		log.Fatalf("Failed to decode based64 response message: %v", err)
	}
	mychannel1HashPair := common.HashPair{}
	err = proto.Unmarshal(message1, &mychannel1HashPair)
	if err != nil {
		log.Fatal(err)
	}*/

	PrintResult(resultCh1, mychannel1HashPair, "Transfer asset4 from Max to Ivan", channel1)

	//query GetAllAssets
	resultCh1, err = ch1.Query(channel.Request{
		ChaincodeID: "cc1",
		Fcn:         "GetAllAssets",
	}) //QUERY!!!
	if err != nil {
		log.Fatalf("Failed to Query transaction: %v", err)
	}
	//QUERY!!!
	//QUERY!!!
	PrintResult(resultCh1, mychannel1HashPair, "GetAllAssets", channel1)

	//Send PAC request to mychannel2
	resultCh2, err = ch2.Execute(channel.Request{
		ChaincodeID: "cc2",
		Fcn:         "UpdateAccount",
		Args:        [][]byte{[]byte("account4"), []byte("Max"), []byte("666.192")},
		/*TransientMap: tmap,
		PACClientData: fab.ClientData{
			RequestedTransaction: fab.PACRequest,
		},*/
	})
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}

	//TODO: check if payload is 200, then the message is OK for PAC
	/*message2, err := base64.StdEncoding.DecodeString(gotMessage)
	if err != nil {
		log.Fatalf("Failed to decode based64 response message: %v", err)
	}
	mychannel2HashPair := common.HashPair{}
	err = proto.Unmarshal(message2, &mychannel2HashPair)
	if err != nil {
		log.Fatal(err)
	}*/
	PrintResult(resultCh2, mychannel2HashPair, "Update account4 (+600)", channel2)

	//query GetAllAssets
	resultCh2, err = ch2.Query(channel.Request{
		ChaincodeID: "cc2",
		Fcn:         "GetAllAccounts",
	}) //QUERY!!!
	if err != nil {
		log.Fatalf("Failed to Query transaction: %v", err)
	}
	//QUERY!!!
	//QUERY!!!

	PrintResult(resultCh2, mychannel2HashPair, "GetAllAccounts", channel2)

	client1.Close()
	client2.Close()

	log.Println("============ application-golang ends ============")
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

func PrintResult(response channel.Response, chanHashPair *common.HashPair, CCFunc string, channel string) {
	gotMessage := response.Responses[0].ProposalResponse.Response.Message

	log.Printf("\n\n\n\n\n\n\n\n\n\n\n\n")
	log.Println("===============================================================")
	log.Printf("=======================~~~%s~~~========================", channel)
	log.Println("===============================================================")
	log.Println("\n\n\nresult.Payload here: ", string(response.Payload))
	log.Printf("--> Channel %s was requested Transaction for CC func: %s, ", channel, CCFunc)
	log.Printf("Response.Message is:\n%s\n", gotMessage)
	log.Printf("Full response struct: [%+v]", response.Responses[0].ProposalResponse.Response)
	log.Println("Transaction status:", response.TxValidationCode.String())
	log.Println("Vsr, Nsr struct:")
	if chanHashPair != nil {
		log.Printf("Struct:[%v], where\n Vsr: [%s], Nsr: [%s]", chanHashPair, chanHashPair.HashedVsr, chanHashPair.HashedNsr)
	} else {
		log.Printf("u didn't use PAC request")
	}
}
