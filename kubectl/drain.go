package kubectl

func Drain(node string) error {
	return kubectl("drain", "--delete-local-data", "--ignore-daemonsets", "--force", node)
}
