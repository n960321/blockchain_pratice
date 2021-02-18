help:
	go run main.go

getbalance:
	go run main.go getbalance -address $(address)

createblockchain:
	go run main.go createblockchain -address $(address)

printchain:
	go run main.go printchain

createwallet:
	go run main.go createwwallet

listaddresses:
	go run main.go listaddresses

send: 
	go run main.go send -from $(send) -to $(to) -amount $(amount)

