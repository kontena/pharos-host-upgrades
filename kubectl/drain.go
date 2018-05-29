package kubectl

func Drain(node string) error {
	return kubectl("drain", "--ignore-daemonsets", "--force", node)
}
