package models

type SerializedQuoteResponse struct {
	SwapTransaction           string `json:"swapTransaction"`
	LastValidBlockHeight      int    `json:"lastValidBlockHeight"`
	PrioritizationFeeLamports int    `json:"prioritizationFeeLamports"`
	ComputeUnitLimit          int    `json:"computeUnitLimit"`
	PrioritizationType        struct {
		ComputeBudget map[string]interface{} `json:"computeBudget"`
	} `json:"prioritizationType"`
	SimulationSlot        int `json:"simulationSlot"`
	DynamicSlippageReport struct {
		SlippageBps                  int    `json:"slippageBps"`
		OtherAmount                  int    `json:"otherAmount"`
		SimulatedIncurredSlippageBps int    `json:"simulatedIncurredSlippageBps"`
		AmplificationRatio           string `json:"amplificationRatio"`
		CategoryName                 string `json:"categoryName"`
		HeuristicMaxSlippageBps      int    `json:"heuristicMaxSlippageBps"`
	} `json:"dynamicSlippageReport"`
	SimulationError string `json:"simulationError"`
}
