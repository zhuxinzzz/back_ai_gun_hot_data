package remote

// AdminRankingResponse admin ranking API响应结构
type AdminRankingResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
