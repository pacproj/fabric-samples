module asset-transfer-basic

go 1.14

require (
	github.com/hyperledger/fabric-sdk-go v1.0.0-rc1
	golang.org/x/tools v0.1.0 // indirect
)

replace github.com/hyperledger/fabric-sdk-go => github.com/pacproj/fabric-sdk-go v1.0.0-beta3.0.20210413180833-ad9bc7d746bd

replace github.com/hyperledger/fabric-protos-go => github.com/pacproj/fabric-protos-go v0.0.0-20210413120152-f3382d191db5
