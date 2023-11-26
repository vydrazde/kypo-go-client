package kypo

type User struct {
	Id         int64  `json:"id" tfsdk:"id"`
	Sub        string `json:"sub" tfsdk:"sub"`
	FullName   string `json:"full_name" tfsdk:"full_name"`
	GivenName  string `json:"given_name" tfsdk:"given_name"`
	FamilyName string `json:"family_name" tfsdk:"family_name"`
	Mail       string `json:"mail" tfsdk:"mail"`
}
