# ton smart contracts

## Standard:
    TRX, TRC10, TRC20, TRC721?????

## How to deploy:
    1. Compile smart contracts into `counter.cell` binary file 
    2. Replace `mnemonic` in deploy.ts with the owner's seed phrase

## Interface:
    recv_internal(int msg_value, cell msg, slice in_msg_body)

    msg_value   - int    // bid amount in range 1 to 5 ton
    msg         - cell   // ???
    in_msg_body - slice  // actual parameters for the function (*)


    (*) in_msg_body slice parameters:
        1) Join the game:
            - signature 512 bytes // owner of the contract signs the body, the contract would verify the only the owner can send transactions
            - operation 32 bytes  // operation = 0 if a user joins the game
            - game_id 64 bytes    // unique game identification number
            - user_id 32 bytes    // unique user's id (telegram_id)
            - address             // user's address
        
        2) Receive the results:
            - signature 512 bytes // owner of the contract signs the body, the contract would verify the only the owner can send transactions
            - operation 32 bytes  // operation = 1 if a user wants to see the results
            - comission 16 bytes  // fee that a user pays for participating in the game. The owner chooses the amount of the fee
            - game_id   64 bytes  // unique game identification number. WARNING - not implemented
    

## Client example:
    1) Join the game:    
``` go
client := liteclient.NewConnectionPool()

configUrl := "https://ton-blockchain.github.io/testnet-global.config.json" // testnet configuration
err := client.AddConnectionsFromConfigUrl(context.Background(), configUrl)
if err != nil {
    ...
}
api := ton.NewAPIClient(client)

userSeedPhrase := strings.Split(<user's seed phrase>, " ") // Replace with the actual user's seed phrase
ownerSeedPhrase := strings.Split(<owner's seed phrase>, " ") // Replace with the actual user's seed phrase
userWallet, err := wallet.FromSeed(api, userSeedPhrase, wallet.V3)
if err != nil {
    ...
}
ownerWallet, err := wallet.FromSeed(api, wordsOwner, wallet.V3)
if err != nil {
    ...
}

block, err := api.CurrentMasterchainInfo(context.Background())
if err != nil {
    ...
}

userBalance, err := userWallet.GetBalance(context.Background(), block)
if err != nil {
    ...
}
var game_id uint64 = 1
var user_id uint64 = 1
    
if userBalance.NanoTON().Uint64() >= 3000000 { // 1 ton = 1.000.000.000 nano ton
	signature := cell.BeginCell().
		MustStoreUInt(0, 32).       // operation (0 because the user joins the game)
		MustStoreUInt(game_id, 64). // game id
		MustStoreUInt(user_id, 32). // user_id
		MustStoreAddr(w.Address()). // user's address
        EndCell().Sign(wOwner.PrivateKey())

	body := cell.BeginCell().
		MustStoreSlice(signature, 512). // signature on the rest of the body (in the previous step)
		MustStoreUInt(0, 32).           // operation
		MustStoreUInt(game_id, 64).     // game id
		MustStoreUInt(user_id, 32).     // user id
		MustStoreAddr(w.Address()).     // user's address
		EndCell()

	log.Println("sending transaction and waiting for confirmation...")

	err = w.Send(context.Background(), &wallet.Message{
		Mode: 1, // pay fees separately (from balance, not from amount)
		InternalMessage: &tlb.InternalMessage{
			Bounce:  true, // return amount in case of processing error
			DstAddr: address.MustParseAddr(<smart contract address>), // is the same as the owner's address
			Amount:  tlb.MustFromTON(<bid amount>),
			Body:    body,
		},
	}, true)
	if err != nil {
        ...
	}
}
```