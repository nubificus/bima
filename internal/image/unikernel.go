package image

func RequiredUnikernelAnnotations() []string {
	return []string{
		"com.urunc.unikernel.unikernelType",
		"com.urunc.unikernel.hypervisor",
		"com.urunc.unikernel.binary",
		"com.urunc.unikernel.cmdline",
	}
}

func cmdAnnotation() string {
	return "com.urunc.unikernel.binary"
}
