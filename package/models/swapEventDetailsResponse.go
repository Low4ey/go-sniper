package models

type SwapEventDetailsResponse struct {
	ProgramInfo struct {
		Source          string `json:"source"`
		Account         string `json:"account"`
		ProgramName     string `json:"programName"`
		InstructionName string `json:"instructionName"`
	} `json:"programInfo"`
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
	Fee         int    `json:"fee"`
	Slot        int    `json:"slot"`
	Timestamp   int    `json:"timestamp"`
	Description string `json:"description"`
}
