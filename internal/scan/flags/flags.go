package flags

type Flag struct {
	Label     string `json:"label"`
	InputType string `json:"inputType"`
	Value     any    `json:"value"`
	Required  bool   `json:"required"`
}
type FlagSet []Flag
