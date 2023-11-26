package kypo_test

type AllocationRequest struct {
	Id               int      `json:"id"`
	AllocationUnitId int      `json:"allocation_unit_id"`
	Created          string   `json:"created"`
	Stages           []string `json:"stages"`
}
