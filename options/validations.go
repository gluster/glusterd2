package options

func clusterOptValidate(key string, value interface{}) bool {
	status := true
	for _, fn := range clusterDefaultOptions[key].validationFuncs {
		if !fn(value) {
			status = false
			break
		}
	}
	return status
}

func volumeOptValidate(key string, value interface{}) bool {
	status := true
	for _, fn := range volumeDefaultOptions[key].validationFuncs {
		if !fn(value) {
			status = false
			break
		}
	}
	return status
}

func boolValidation(val interface{}) bool {
	_, ok := val.(bool)
	return ok
}

func stringValidation(val interface{}) bool {
	_, ok := val.(string)
	return ok
}

func onOffValidation(val interface{}) bool {
	return (val == "on" || val == "off")
}
