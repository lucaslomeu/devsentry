package docker

func appendFilterArgs(args []string, key string, values []string) []string {
	for _, value := range values {
		args = append(args, "--filter", key+"="+value)
	}
	return args
}
