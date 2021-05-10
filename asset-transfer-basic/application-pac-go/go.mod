module asset-transfer-basic

go 1.14

require (
	github.com/golang/protobuf v1.3.3
	github.com/hyperledger/fabric-protos-go v0.0.0-20200707132912-fee30f3ccd23
	github.com/hyperledger/fabric-sdk-go v1.0.0-rc1
	golang.org/x/tools v0.1.0 // indirect
)

replace github.com/hyperledger/fabric-sdk-go => github.com/pacproj/fabric-sdk-go v1.0.0-beta3.0.20210420170914-fd141536631b

replace github.com/hyperledger/fabric-protos-go => github.com/pacproj/fabric-protos-go v0.0.0-20210413120152-f3382d191db5
