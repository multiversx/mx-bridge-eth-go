package dtos

type SharedOwner struct {
	Shared struct {
		InitialSharedVersion uint64 `json:"initial_shared_version"`
	} `json:"Shared"`
}
