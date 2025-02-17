package transactions

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"your_project/tracker/db" // import your DB functions

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/low4ey/sniper/internal/config"
	"github.com/low4ey/sniper/package/models"
	"github.com/mr-tron/base58"
)

func newHTTPClient(timeoutMs int) *http.Client {
	return &http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
	}
}

// ---------- Function: FetchTransactionDetails ----------

func FetchTransactionDetails(signature string) (*models.MintsDataReponse, error) {
	txUrl := os.Getenv("HELIUS_HTTPS_URI_TX")
	maxRetries := config.ConfigVal.Tx.FetchTxMaxRetries
	initialDelay := time.Duration(config.ConfigVal.Tx.FetchTxInitialDelay) * time.Millisecond

	log.Printf("Waiting %v seconds for transaction to be confirmed...", initialDelay.Seconds())
	time.Sleep(initialDelay)

	client := newHTTPClient(config.ConfigVal.Tx.GetTimeout)
	retryCount := 0

	for retryCount < maxRetries {
		log.Printf("Attempt %d of %d to fetch transaction details...", retryCount+1, maxRetries)
		payload := map[string]interface{}{
			"transactions": []string{signature},
			"commitment":   "finalized",
			"encoding":     "jsonParsed",
		}
		payloadBytes, _ := json.Marshal(payload)
		req, err := http.NewRequest("POST", txUrl, bytes.NewReader(payloadBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Attempt %d failed: %v", retryCount+1, err)
			retryCount++
			delay := time.Duration(min(4000*pow(1.5, float64(retryCount)), 15000)) * time.Millisecond
			log.Printf("Waiting %v seconds before next attempt...", delay.Seconds())
			time.Sleep(delay)
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var transactions []models.TransactionDetailResponse
		if err := json.Unmarshal(body, &transactions); err != nil {
			log.Printf("Attempt %d failed: %v", retryCount+1, err)
			retryCount++
			delay := time.Duration(min(4000*pow(1.5, float64(retryCount)), 15000)) * time.Millisecond
			log.Printf("Waiting %v seconds before next attempt...", delay.Seconds())
			time.Sleep(delay)
			continue
		}

		if len(transactions) == 0 || len(transactions[0].Instructions) == 0 {
			log.Printf("Attempt %d failed: transaction or instructions not found", retryCount+1)
			retryCount++
			delay := time.Duration(min(4000*pow(1.5, float64(retryCount)), 15000)) * time.Millisecond
			log.Printf("Waiting %v seconds before next attempt...", delay.Seconds())
			time.Sleep(delay)
			continue
		}

		// Find instruction for the liquidity pool program.
		targetInsts := transactions[0].Instructions
		var targetInst *models.Instructions
		for _, inst := range targetInsts {
			if inst.ProgramId == config.ConfigVal.LiquidityPool.RadiyumProgramID {
				targetInst = &inst
				break
			}
		}
		if targetInst == nil || len(targetInst.Accounts) < 10 {
			log.Printf("Attempt %d failed: no valid market maker instruction found", retryCount+1)
			retryCount++
			delay := time.Duration(min(4000*pow(1.5, float64(retryCount)), 15000)) * time.Millisecond
			log.Printf("Waiting %v seconds before next attempt...", delay.Seconds())
			time.Sleep(delay)
			continue
		}

		// Extract accounts 8 and 9.
		accountOne := targetInst.Accounts[8]
		accountTwo := targetInst.Accounts[9]
		if accountOne == "" || accountTwo == "" {
			return nil, fmt.Errorf("required accounts not found")
		}

		var solTokenAccount, newTokenAccount string
		if accountOne == config.ConfigVal.LiquidityPool.WsolPcMint {
			solTokenAccount = accountOne
			newTokenAccount = accountTwo
		} else {
			solTokenAccount = accountTwo
			newTokenAccount = accountOne
		}

		log.Printf("Successfully fetched transaction details!")
		log.Printf("SOL Token Account: %s", solTokenAccount)
		log.Printf("New Token Account: %s", newTokenAccount)

		return &models.MintsDataReponse{
			TokenMint: newTokenAccount,
			SolMint:   solTokenAccount,
		}, nil
	}

	log.Printf("All attempts to fetch transaction details failed")
	return nil, fmt.Errorf("failed to fetch transaction details")
}

// ---------- Function: CreateSwapTransaction ----------

func CreateSwapTransaction(solMint, tokenMint string) (string, error) {
	quoteUrl := os.Getenv("JUP_HTTPS_QUOTE_URI")
	swapUrl := os.Getenv("JUP_HTTPS_SWAP_URI")
	rpcUrl := os.Getenv("HELIUS_HTTPS_URI")

	client := newHTTPClient(config.ConfigVal.Tx.GetTimeout)
	var quoteResponseData *models.QuoteResponse

	// Create a Solana RPC client.
	rpcClient := rpc.New(rpcUrl)

	// Create wallet from secret key.
	privKeyStr := os.Getenv("PRIV_KEY_WALLET")
	walletPubKey, err := solana.PublicKeyFromBase58(privKeyStr)
	if err != nil {
		return "", fmt.Errorf("failed to create keypair: %v", err)
	}
	// --- Get Swap Quote ---
	retryCount := 0
	maxRetries := config.ConfigVal.Swap.TokenNotTradable400ErrorRetries
	for retryCount < maxRetries {
		reqURL := fmt.Sprintf("%s?inputMint=%s&outputMint=%s&amount=%s&slippageBps=%s", quoteUrl, solMint, tokenMint, config.ConfigVal.Swap.Amount, config.ConfigVal.Swap.SlippageBps)
		resp, err := client.Get(reqURL)
		if err != nil {
			log.Printf("Swap quote attempt %d failed: %v", retryCount+1, err)
			retryCount++
			time.Sleep(time.Duration(config.ConfigVal.Swap.TokenNotTradable400ErrorDelay) * time.Millisecond)
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", err
		}
		if err := json.Unmarshal(body, &quoteResponseData); err != nil {
			// Check if error indicates TOKEN_NOT_TRADABLE (this logic may vary based on actual response)
			if resp.StatusCode == 400 {
				retryCount++
				time.Sleep(time.Duration(config.ConfigVal.Swap.TokenNotTradable400ErrorDelay) * time.Millisecond)
				continue
			}
			return "", err
		}
		log.Printf("✅ Swap quote received.")
		break
	}
	if quoteResponseData == nil {
		return "", fmt.Errorf("failed to get swap quote")
	}

	// --- Serialize the Quote into a Swap Transaction ---
	swapPayload := map[string]interface{}{
		"quoteResponse":    quoteResponseData,
		"userPublicKey":    walletPubKey.String(),
		"wrapAndUnwrapSol": true,
		"dynamicSlippage": map[string]interface{}{
			"maxBps": 300,
		},
		"prioritizationFeeLamports": map[string]interface{}{
			"priorityLevelWithMaxLamports": map[string]interface{}{
				"maxLamports":   config.ConfigVal.Swap.PrioFeeMaxLamports,
				"priorityLevel": config.ConfigVal.Swap.PrioLevel,
			},
		},
	}
	swapPayloadBytes, _ := json.Marshal(swapPayload)
	req, err := http.NewRequest("POST", swapUrl, bytes.NewReader(swapPayloadBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}
	var serializedQuoteResponse models.SerializedQuoteResponse
	if err := json.Unmarshal(body, &serializedQuoteResponse); err != nil {
		return "", err
	}
	log.Printf("✅ Swap quote serialized.")

	//-------------------------------------------------------
	//   Work From Here 									|
	//------------------------------------------------------|

	// --- Deserialize, Sign and Send Transaction ---
	swapTxBytes, err := base58.Decode(serializedQuoteResponse.SwapTransaction)
	if err != nil {
		// If the transaction is base64 encoded instead, use:
		swapTxBytes, err = decodeBase64(serializedQuoteResponse.SwapTransaction)
		if err != nil {
			return "", fmt.Errorf("failed to decode swap transaction: %v", err)
		}
	}
	// Deserialize versioned transaction (solana-go supports legacy transactions; for versioned ones, additional work may be needed)
	tx, err := solana.TransactionFromDecoder()
	if err != nil {
		return "", fmt.Errorf("failed to deserialize transaction: %v", err)
	}
	// Sign the transaction
	tx.Signatures = append(tx.Signatures, solana.Signature{}, solana.Signature{}) // placeholder; proper signing logic depends on tx type
	// (In practice, use the appropriate Sign method provided by the SDK.)
	// Here we assume tx.SignPartial(keypair) exists.
	// tx.SignPartial(keypair)

	// Get recent blockhash.
	ctx := context.Background()
	recent, err := rpcClient.GetLatestBlockhash(ctx)
	if err != nil {
		return "", err
	}
	// Set the recent blockhash in tx (if needed).
	// tx.Message.RecentBlockhash = recent.Blockhash

	// Serialize and send transaction.
	rawTx, err := tx.Serialize()
	if err != nil {
		return "", err
	}
	txid, err := rpcClient.SendRawTransaction(ctx, rawTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %v", err)
	}
	log.Printf("✅ Raw transaction id received: %s", txid)

	// Confirm transaction.
	confirmCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	_, err = rpcClient.ConfirmTransaction(confirmCtx, txid)
	if err != nil {
		return "", fmt.Errorf("transaction confirmation failed: %v", err)
	}
	log.Printf("Transaction confirmed.")
	return txid, nil
}

// ---------- Function: GetRugCheckConfirmed ----------

func GetRugCheckConfirmed(tokenMint string) (bool, error) {
	rugUrl := fmt.Sprintf("https://api.rugcheck.xyz/v1/tokens/%s/report", tokenMint)
	client := newHTTPClient(config.ConfigVal.Tx.GetTimeout)
	resp, err := client.Get(rugUrl)
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return false, err
	}
	var tokenReport RugResponseExtended
	if err := json.Unmarshal(body, &tokenReport); err != nil {
		return false, err
	}
	if config.ConfigVal.RugCheck.VerboseLog {
		log.Printf("%+v", tokenReport)
	}

	// Extract fields
	tokenCreator := tokenReport.Creator
	if tokenCreator == "" {
		tokenCreator = tokenMint
	}
	mintAuthority := tokenReport.Token.MintAuthority
	freezeAuthority := tokenReport.Token.FreezeAuthority
	isInitialized := tokenReport.Token.IsInitialized
	tokenMutable := tokenReport.TokenMeta.Mutable
	topHolders := tokenReport.TopHolders
	marketsLength := 0
	if tokenReport.Markets != nil {
		marketsLength = len(tokenReport.Markets)
	}
	totalLPProviders := tokenReport.TotalLPProviders
	totalMarketLiquidity := tokenReport.TotalMarketLiquidity
	isRugged := tokenReport.Rugged
	rugScore := tokenReport.Score
	rugRisks := tokenReport.Risks
	rugCheckLegacy := config.ConfigVal.RugCheck.LegacyNotAllowed

	// Exclude liquidity pools from top holders if configured.
	if config.ConfigVal.RugCheck.ExcludeLPFromTopholders && tokenReport.Markets != nil {
		var liquidityAddresses []string
		for _, market := range tokenReport.Markets {
			if market.LiquidityA != "" {
				liquidityAddresses = append(liquidityAddresses, market.LiquidityA)
			}
			if market.LiquidityB != "" {
				liquidityAddresses = append(liquidityAddresses, market.LiquidityB)
			}
		}
		filtered := make([]Holder, 0, len(topHolders))
		for _, holder := range topHolders {
			exclude := false
			for _, addr := range liquidityAddresses {
				if holder.Address == addr {
					exclude = true
					break
				}
			}
			if !exclude {
				filtered = append(filtered, holder)
			}
		}
		topHolders = filtered
	}

	// Conditions to check.
	conditions := []struct {
		Check   bool
		Message string
	}{
		{!config.ConfigVal.RugCheck.AllowMintAuthority && mintAuthority != nil, "🚫 Mint authority should be null"},
		{!config.ConfigVal.RugCheck.AllowNotInitialized && !isInitialized, "🚫 Token is not initialized"},
		{!config.ConfigVal.RugCheck.AllowFreezeAuthority && freezeAuthority != nil, "🚫 Freeze authority should be null"},
		{!config.ConfigVal.RugCheck.AllowMutable && tokenMutable, "🚫 Mutable should be false"},
		{!config.ConfigVal.RugCheck.AllowInsiderTopholders && containsInsider(topHolders), "🚫 Insider accounts should not be part of the top holders"},
		{anyHolderExceeds(topHolders, config.ConfigVal.RugCheck.MaxAlowedPctTopholders), "🚫 A top holder exceeds the allowed percentage"},
		{totalLPProviders < config.ConfigVal.RugCheck.MinTotalLPProviders, "🚫 Not enough LP Providers."},
		{marketsLength < config.ConfigVal.RugCheck.MinTotalMarkets, "🚫 Not enough Markets."},
		{totalMarketLiquidity < config.ConfigVal.RugCheck.MinTotalMarketLiquidity, "🚫 Not enough Market Liquidity."},
		{!config.ConfigVal.RugCheck.AllowRugged && isRugged, "🚫 Token is rugged"},
		{contains(config.ConfigVal.RugCheck.BlockSymbols, tokenReport.TokenMeta.Symbol), "🚫 Symbol is blocked"},
		{contains(config.ConfigVal.RugCheck.BlockNames, tokenReport.TokenMeta.Name), "🚫 Name is blocked"},
		{rugScore > config.ConfigVal.RugCheck.MaxScore && config.ConfigVal.RugCheck.MaxScore != 0, "🚫 Rug score too high."},
		{anyRiskInLegacy(rugRisks, rugCheckLegacy), "🚫 Token has legacy risks that are not allowed."},
	}

	// If tracking duplicate tokens is enabled, query DB.
	if config.ConfigVal.RugCheck.BlockReturningTokenNames || config.ConfigVal.RugCheck.BlockReturningTokenCreators {
		duplicates, err := db.SelectTokenByNameAndCreator(tokenReport.TokenMeta.Name, tokenCreator)
		if err == nil && len(duplicates) > 0 {
			for _, token := range duplicates {
				if config.ConfigVal.RugCheck.BlockReturningTokenNames && token.Name == tokenReport.TokenMeta.Name {
					log.Println("🚫 Token with this name was already created")
					return false, nil
				}
				if config.ConfigVal.RugCheck.BlockReturningTokenCreators && token.Creator == tokenCreator {
					log.Println("🚫 Token from this creator was already created")
					return false, nil
				}
			}
		}
	}

	// Create new token record.
	newToken := NewTokenRecord{
		Time:    time.Now().UnixMilli(),
		Mint:    tokenMint,
		Name:    tokenReport.TokenMeta.Name,
		Creator: tokenCreator,
	}
	if err := db.InsertNewToken(newToken); err != nil {
		log.Printf("⛔ Unable to store new token: %v", err)
	}

	// Evaluate conditions.
	for _, cond := range conditions {
		if cond.Check {
			log.Println(cond.Message)
			return false, nil
		}
	}
	return true, nil
}

// ---------- Function: FetchAndSaveSwapDetails ----------

func FetchAndSaveSwapDetails(tx string) (bool, error) {
	txUrl := os.Getenv("HELIUS_HTTPS_URI_TX")
	priceUrl := os.Getenv("JUP_HTTPS_PRICE_URI")
	rpcUrl := os.Getenv("HELIUS_HTTPS_URI")
	client := newHTTPClient(10000) // hardcoded timeout; adjust as needed

	// POST to get transaction details.
	payload := map[string]interface{}{
		"transactions": []string{tx},
	}
	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", txUrl, bytes.NewReader(payloadBytes))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("⛔ Could not fetch swap details: %v", err)
		return false, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return false, err
	}
	var transactions []TransactionDetails
	if err := json.Unmarshal(body, &transactions); err != nil || len(transactions) == 0 {
		log.Println("⛔ Could not fetch swap details: invalid response")
		return false, fmt.Errorf("invalid response")
	}
	// Assume the first transaction holds our swap details.
	swapEvent := transactions[0]
	innerSwap := swapEvent.Events.Swap.InnerSwaps[0]
	swapData := struct {
		ProgramInfo  *ProgramInfo  `json:"programInfo"`
		TokenInputs  []TokenAmount `json:"tokenInputs"`
		TokenOutputs []TokenAmount `json:"tokenOutputs"`
		Fee          int           `json:"fee"`
		Slot         int           `json:"slot"`
		Timestamp    int64         `json:"timestamp"`
		Description  string        `json:"description"`
	}{
		ProgramInfo:  innerSwap.ProgramInfo,
		TokenInputs:  innerSwap.TokenInputs,
		TokenOutputs: innerSwap.TokenOutputs,
		Fee:          swapEvent.Fee,
		Slot:         swapEvent.Slot,
		Timestamp:    swapEvent.Timestamp,
		Description:  swapEvent.Description,
	}

	// Get latest SOL price from priceUrl.
	reqPrice, err := http.NewRequest("GET", priceUrl, nil)
	if err != nil {
		return false, err
	}
	q := reqPrice.URL.Query()
	q.Add("ids", config.ConfigVal.LiquidityPool.WsolPcMint)
	reqPrice.URL.RawQuery = q.Encode()
	respPrice, err := client.Do(reqPrice)
	if err != nil {
		return false, err
	}
	priceBody, err := ioutil.ReadAll(respPrice.Body)
	respPrice.Body.Close()
	if err != nil {
		return false, err
	}
	var priceResp map[string]map[string]struct {
		Price float64 `json:"price"`
	}
	if err := json.Unmarshal(priceBody, &priceResp); err != nil {
		return false, err
	}
	priceData, ok := priceResp["data"][config.ConfigVal.LiquidityPool.WsolPcMint]
	if !ok || priceData.Price == 0 {
		return false, fmt.Errorf("price not found")
	}
	solPrice := priceData.Price

	// Calculate estimated prices.
	solPaidUSDC := swapData.TokenInputs[0].TokenAmount * solPrice
	solFeePaidUSDC := (float64(swapData.Fee) / 1_000_000_000) * solPrice
	perTokenUSDC := solPaidUSDC / swapData.TokenOutputs[0].TokenAmount

	// Get token meta data from DB.
	tokenName := "N/A"
	tokens, err := db.SelectTokenByMint(swapData.TokenOutputs[0].Mint)
	if err == nil && len(tokens) > 0 {
		tokenName = tokens[0].Name
	}

	newHolding := HoldingRecord{
		Time:             swapData.Timestamp,
		Token:            swapData.TokenOutputs[0].Mint,
		TokenName:        tokenName,
		Balance:          swapData.TokenOutputs[0].TokenAmount,
		SolPaid:          swapData.TokenInputs[0].TokenAmount,
		SolFeePaid:       swapData.Fee,
		SolPaidUSDC:      solPaidUSDC,
		SolFeePaidUSDC:   solFeePaidUSDC,
		PerTokenPaidUSDC: perTokenUSDC,
		Slot:             swapData.Slot,
	}
	if err := db.InsertHolding(newHolding); err != nil {
		log.Printf("⛔ Database Error: %v", err)
		return false, err
	}
	return true, nil
}

// ---------- Function: CreateSellTransaction ----------

func CreateSellTransaction(solMint, tokenMint, amount string) (*CreateSellTransactionResponse, error) {
	quoteUrl := os.Getenv("JUP_HTTPS_QUOTE_URI")
	swapUrl := os.Getenv("JUP_HTTPS_SWAP_URI")
	rpcUrl := os.Getenv("HELIUS_HTTPS_URI")

	// Create Solana RPC client.
	rpcClient := rpc.New(rpcUrl)
	privKeyStr := os.Getenv("PRIV_KEY_WALLET")
	privKeyBytes, err := base58.Decode(privKeyStr)
	if err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	keypair, err := solana.PrivateKeyFromBytes(privKeyBytes)
	if err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	walletPubKey := keypair.PublicKey()

	// Check token balance.
	ctx := context.Background()
	parsedMint, err := solana.PublicKeyFromString(tokenMint)
	if err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	tokenAccounts, err := rpcClient.GetTokenAccountsByOwner(ctx, walletPubKey, rpc.GetTokenAccountsConfig{
		Mint: parsedMint.String(),
	})
	if err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	totalBalance := uint64(0)
	for _, acc := range tokenAccounts.Value {
		// Assume the parsed JSON has "tokenAmount" as a string; adjust as necessary.
		amountStr := acc.Account.Data.Parsed.Info.TokenAmount.Amount
		balance, _ := strconv.ParseUint(amountStr, 10, 64)
		totalBalance += balance
	}
	amountUint, _ := strconv.ParseUint(amount, 10, 64)
	if totalBalance == 0 || totalBalance != amountUint {
		// Remove holding from DB.
		_ = db.RemoveHolding(tokenMint)
		return &CreateSellTransactionResponse{
			Success: false,
			Msg:     "Token balance mismatch or zero balance; sell manually.",
			Tx:      "",
		}, fmt.Errorf("balance error")
	}

	// Request a sell quote.
	client := newHTTPClient(config.ConfigVal.Tx.GetTimeout)
	reqURL := fmt.Sprintf("%s?inputMint=%s&outputMint=%s&amount=%s&slippageBps=%s", quoteUrl, tokenMint, solMint, amount, config.ConfigVal.Sell.SlippageBps)
	resp, err := client.Get(reqURL)
	if err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	var quoteResp QuoteResponse
	if err := json.Unmarshal(body, &quoteResp); err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	if quoteResp == nil {
		return &CreateSellTransactionResponse{Success: false, Msg: "No valid quote received"}, nil
	}

	// Serialize the quote into a swap transaction.
	swapPayload := map[string]interface{}{
		"quoteResponse":    quoteResp,
		"userPublicKey":    walletPubKey.String(),
		"wrapAndUnwrapSol": true,
		"dynamicSlippage": map[string]interface{}{
			"maxBps": 300,
		},
		"prioritizationFeeLamports": map[string]interface{}{
			"priorityLevelWithMaxLamports": map[string]interface{}{
				"maxLamports":   config.ConfigVal.Sell.PrioFeeMaxLamports,
				"priorityLevel": config.ConfigVal.Sell.PrioLevel,
			},
		},
	}
	swapPayloadBytes, _ := json.Marshal(swapPayload)
	reqSwap, err := http.NewRequest("POST", swapUrl, bytes.NewReader(swapPayloadBytes))
	if err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	reqSwap.Header.Set("Content-Type", "application/json")
	respSwap, err := client.Do(reqSwap)
	if err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	swapBody, err := ioutil.ReadAll(respSwap.Body)
	respSwap.Body.Close()
	if err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	var serializedSwap SerializedQuoteResponse
	if err := json.Unmarshal(swapBody, &serializedSwap); err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	// Deserialize transaction.
	swapTxBytes, err := base58.Decode(serializedSwap.SwapTransaction)
	if err != nil {
		swapTxBytes, err = decodeBase64(serializedSwap.SwapTransaction)
		if err != nil {
			return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
		}
	}
	tx, err := solana.TransactionDeserialize(swapTxBytes)
	if err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	// Sign the transaction.
	// (Assuming appropriate signing method exists.)
	// tx.SignPartial(keypair)

	// Send transaction.
	rawTx, err := tx.Serialize()
	if err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	txid, err := rpcClient.SendRawTransaction(ctx, rawTx)
	if err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	// Confirm transaction.
	recent, err := rpcClient.GetLatestBlockhash(ctx)
	if err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: err.Error()}, err
	}
	_, err = rpcClient.ConfirmTransaction(ctx, txid)
	if err != nil {
		return &CreateSellTransactionResponse{Success: false, Msg: "Transaction confirmation failed"}, err
	}
	// Remove holding.
	_ = db.RemoveHolding(tokenMint)
	return &CreateSellTransactionResponse{
		Success: true,
		Msg:     "",
		Tx:      txid,
	}, nil
}

// ---------- Helper Functions ----------

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func pow(a, b float64) float64 {
	return math.Pow(a, b)
}

func decodeBase64(s string) ([]byte, error) {
	return ioutil.ReadAll(base64.NewDecoder(base64.StdEncoding, strings.NewReader(s)))
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsInsider(holders []Holder) bool {
	for _, h := range holders {
		if h.Insider {
			return true
		}
	}
	return false
}

func anyHolderExceeds(holders []Holder, maxPct int) bool {
	for _, h := range holders {
		if h.Pct > float64(maxPct) {
			return true
		}
	}
	return false
}

func anyRiskInLegacy(risks []Risk, legacy []string) bool {
	for _, risk := range risks {
		if contains(legacy, risk.Name) {
			return true
		}
	}
	return false
}
