package cmd

func formatBoolYesNo(value bool) string {
	if value == true {
		return "yes"
	}
	return "no"
}
