package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	rpcURL      = "http://127.0.0.1:18443"
	rpcUser     = "bootcamp"
	rpcPassword = "bootcamp123"
)

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func rpc(method string, params []any, wallet string, out any) error {
	url := rpcURL
	if wallet != "" {
		url += "/wallet/" + wallet
	}
	body, _ := json.Marshal(rpcRequest{
		JSONRPC: "1.0", ID: "explorer",
		Method: method, Params: params,
	})
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.SetBasicAuth(rpcUser, rpcPassword)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var parsed rpcResponse
	json.NewDecoder(resp.Body).Decode(&parsed)
	if parsed.Error != nil {
		return fmt.Errorf("RPC error: %s", parsed.Error.Message)
	}
	return json.Unmarshal(parsed.Result, out)
}

// Challenge 1: Blockchain Info
func showBlockchainInfo() error {
	var info struct {
		Chain      string  `json:"chain"`
		Blocks     int     `json:"blocks"`
		Difficulty float64 `json:"difficulty"`
	}
	if err := rpc("getblockchaininfo", nil, "", &info); err != nil {
		return err
	}
	fmt.Println("=== Blockchain Info ===")
	fmt.Printf("Chain:      %s\n", info.Chain)
	fmt.Printf("Blocks:     %d\n", info.Blocks)
	fmt.Printf("Difficulty: %v\n", info.Difficulty)
	return nil
}

// Challenge 2: Wallet Balance
func showWalletBalance(wallet string) error {
	_ = rpc("loadwallet", []any{wallet}, "", nil)
	var balance float64
	if err := rpc("getbalance", nil, wallet, &balance); err != nil {
		return err
	}
	fmt.Printf("=== Wallet: %s ===\n", wallet)
	fmt.Printf("Balance: %v BTC\n", balance)
	return nil
}

// Challenge 3: List Transactions
func listTransactions(wallet string, count int) error {
	_ = rpc("loadwallet", []any{wallet}, "", nil)
	var txs []struct {
		Category      string  `json:"category"`
		Amount        float64 `json:"amount"`
		TxID          string  `json:"txid"`
		Confirmations int     `json:"confirmations"`
	}
	if err := rpc("listtransactions", []any{"*", count}, wallet, &txs); err != nil {
		return err
	}
	fmt.Printf("=== Transactions: %s ===\n", wallet)
	for _, tx := range txs {
		dir := "OUT"
		switch tx.Category {
		case "receive", "generate", "immature":
			dir = "IN "
		}
		fmt.Printf("%s %+.8f BTC | %d confs\n", dir, tx.Amount, tx.Confirmations)
		fmt.Printf("    TXID: %s\n", tx.TxID)
	}
	return nil
}

// Challenge 4: Decode Transaction
func decodeTransaction(txid string) error {
	var tx struct {
		Vin []struct {
			Coinbase string `json:"coinbase"`
			TxID     string `json:"txid"`
			Vout     int    `json:"vout"`
		} `json:"vin"`
		Vout []struct {
			Value        float64 `json:"value"`
			ScriptPubKey struct {
				Address string `json:"address"`
			} `json:"scriptPubKey"`
		} `json:"vout"`
	}
	if err := rpc("getrawtransaction", []any{txid, true}, "", &tx); err != nil {
		return err
	}
	fmt.Printf("=== Transaction: %s... ===\n", txid[:20])
	for _, vin := range tx.Vin {
		if vin.Coinbase != "" {
			fmt.Println("  COINBASE (mining reward)")
		} else {
			fmt.Printf("  From: %s...\n", vin.TxID[:20])
		}
	}
	for _, vout := range tx.Vout {
		fmt.Printf("  %.8f BTC -> %s\n", vout.Value, vout.ScriptPubKey.Address)
	}
	return nil
}

// Challenge 5: Block Details
func showBlock(blockhash string) error {
	if blockhash == "" {
		rpc("getbestblockhash", nil, "", &blockhash)
	}
	var block struct {
		Height int      `json:"height"`
		Hash   string   `json:"hash"`
		Time   int64    `json:"time"`
		NTx    int      `json:"nTx"`
		Tx     []string `json:"tx"`
	}
	if err := rpc("getblock", []any{blockhash, 1}, "", &block); err != nil {
		return err
	}
	fmt.Printf("=== Block #%d ===\n", block.Height)
	fmt.Printf("Hash:         %s...\n", block.Hash[:32])
	fmt.Printf("Time:         %d\n", block.Time)
	fmt.Printf("Transactions: %d\n", block.NTx)
	return nil
}

func main() {
	fmt.Println("Bitcoin Explorer")
	fmt.Println("================")

	if err := showBlockchainInfo(); err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println()

	if err := showWalletBalance("alice"); err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println()

	if err := listTransactions("alice", 5); err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println()

	if err := decodeTransaction("3517d32581c3160ab2ef2bd99ede54bd78a4f9ce28d2a510e7398fed0cbfdf64"); err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println()

	if err := showBlock(""); err != nil {
		fmt.Println("Error:", err)
	}
}
