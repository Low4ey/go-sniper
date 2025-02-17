package models

type TransactionDetailResponse struct {
	Description    string `json:"description"`
	Type           string `json:"type"`
	Source         string `json:"source"`
	Fee            int    `json:"fee"`
	FeePayer       string `json:"feePayer"`
	Signature      string `json:"signature"`
	Slot           int    `json:"slot"`
	Timestamp      int    `json:"timestamp"`
	TokenTransfers []struct {
		FromTokenAccount string `json:"fromTokenAccount"`
		ToTokenAccount   string `json:"toTokenAccount"`
		FromUserAccount  string `json:"fromUserAccount"`
		ToUserAccount    string `json:"toUserAccount"`
		TokenAmount      string `json:"tokenAmount"`
		Mint             string `json:"mint"`
		TokenStandard    string `json:"tokenStandard"`
	} `json:"tokenTransfers"`
	NativeTransfers []struct {
		FromUserAccount string `json:"fromUserAccount"`
		ToUserAccount   string `json:"toUserAccount"`
		Amount          int    `json:"amount"`
	} `json:"nativeTransfers"`
	AccountData []struct {
		Account             string `json:"account"`
		NativeBalanceChange int    `json:"nativeBalanceChange"`
		TokenBalanceChanges []struct {
			UserAccount    string `json:"userAccount"`
			TokenAccount   string `json:"tokenAccount"`
			RawTokenAmount struct {
				TokenAmount string `json:"tokenAmount"`
				Decimals    int    `json:"decimals"`
			} `json:"rawTokenAmount"`
			Mint string `json:"mint"`
		} `json:"tokenBalanceChanges"`
	} `json:"accountData"`
	TransactionError string `json:"transactionError"`
	Instructions     []struct {
		Accounts          []string `json:"accounts"`
		Data              string   `json:"data"`
		ProgramId         string   `json:"programId"`
		InnerInstructions []struct {
			Accounts  []string `json:"accounts"`
			Data      string   `json:"data"`
			ProgramId string   `json:"programId"`
		} `json:"innerInstructions"`
	} `json:"instructions"`
	Events struct {
		Swap struct {
			NativeInput *struct {
				Account string `json:"account"`
				Amount  string `json:"amount"`
			} `json:"nativeInput"`
			NativeOutput *struct {
				Account string `json:"account"`
				Amount  string `json:"amount"`
			} `json:"nativeOutput"`
			TokenInputs []struct {
				UserAccount    string `json:"userAccount"`
				TokenAccount   string `json:"tokenAccount"`
				RawTokenAmount struct {
					TokenAmount string `json:"tokenAmount"`
					Decimals    int    `json:"decimals"`
				} `json:"rawTokenAmount"`
				Mint string `json:"mint"`
			} `json:"tokenInputs"`
			TokenOutputs []struct {
				UserAccount    string `json:"userAccount"`
				TokenAccount   string `json:"tokenAccount"`
				RawTokenAmount struct {
					TokenAmount string `json:"tokenAmount"`
					Decimals    int    `json:"decimals"`
				} `json:"rawTokenAmount"`
				Mint string `json:"mint"`
			} `json:"tokenOutputs"`
			NativeFees []struct {
				Account string `json:"account"`
				Amount  string `json:"amount"`
			} `json:"nativeFees"`
			TokenFees []struct {
				UserAccount    string `json:"userAccount"`
				TokenAccount   string `json:"tokenAccount"`
				RawTokenAmount struct {
					TokenAmount string `json:"tokenAmount"`
					Decimals    int    `json:"decimals"`
				} `json:"rawTokenAmount"`
				Mint string `json:"mint"`
			} `json:"tokenFees"`
			InnerSwaps []struct {
				TokenInputs []struct {
					FromTokenAccount string `json:"fromTokenAccount"`
					ToTokenAccount   string `json:"toTokenAccount"`
					FromUserAccount  string `json:"fromUserAccount"`
					ToUserAccount    string `json:"toUserAccount"`
					TokenAmount      int    `json:"tokenAmount"`
					Mint             string `json:"mint"`
					TokenStandard    string `json:"tokenStandard"`
				} `json:"tokenInputs"`
				TokenOutputs []struct {
					FromTokenAccount string `json:"fromTokenAccount"`
					ToTokenAccount   string `json:"toTokenAccount"`
					FromUserAccount  string `json:"fromUserAccount"`
					ToUserAccount    string `json:"toUserAccount"`
					TokenAmount      int    `json:"tokenAmount"`
					Mint             string `json:"mint"`
					TokenStandard    string `json:"tokenStandard"`
				} `json:"tokenOutputs"`
				TokenFees []struct {
					UserAccount    string `json:"userAccount"`
					TokenAccount   string `json:"tokenAccount"`
					RawTokenAmount struct {
						TokenAmount string `json:"tokenAmount"`
						Decimals    int    `json:"decimals"`
					} `json:"rawTokenAmount"`
					Mint string `json:"mint"`
				} `json:"tokenFees"`
				NativeFees []struct {
					Account string `json:"account"`
					Amount  string `json:"amount"`
				} `json:"nativeFees"`
			} `json:"innerSwaps"`
		} `json:"swap"`
	} `json:"events"`
}
