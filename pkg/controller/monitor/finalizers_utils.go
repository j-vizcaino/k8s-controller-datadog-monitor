package monitor

const (
	FinalizerName = "monitors.datadoghq.com"
)

func addFinalizer(finalizers []string) []string {
	for _, f := range finalizers {
		if f == FinalizerName {
			return finalizers
		}
	}
	return append(finalizers, FinalizerName)
}

func removeFinalizer(finalizers []string) ([]string, bool) {
	for idx, f := range finalizers {
		if f == FinalizerName {
			return append(finalizers[:idx], finalizers[idx+1:]...), true
		}
	}
	return finalizers, false
}
