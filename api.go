package roi

// APIResponse는 /api/ 하위 사이트로 사용자가 질의했을 때 json 응답을 위해 사용한다.
type APIResponse struct {
	Msg string `json:"msg"`
	Err string `json:"err"`
}
