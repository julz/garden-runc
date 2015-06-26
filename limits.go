package gardenrunc

import "github.com/cloudfoundry-incubator/garden"

type LimitsHandler struct {
}

func (c *LimitsHandler) LimitBandwidth(limits garden.BandwidthLimits) error {
	return nil
}

func (c *LimitsHandler) CurrentBandwidthLimits() (garden.BandwidthLimits, error) {
	return garden.BandwidthLimits{}, nil
}

func (c *LimitsHandler) LimitCPU(limits garden.CPULimits) error {
	return nil
}

func (c *LimitsHandler) CurrentCPULimits() (garden.CPULimits, error) {
	return garden.CPULimits{}, nil
}

func (c *LimitsHandler) LimitDisk(limits garden.DiskLimits) error {
	return nil
}
func (c *LimitsHandler) CurrentDiskLimits() (garden.DiskLimits, error) {
	return garden.DiskLimits{}, nil
}

func (c *LimitsHandler) LimitMemory(limits garden.MemoryLimits) error {
	return nil
}

func (c *LimitsHandler) CurrentMemoryLimits() (garden.MemoryLimits, error) {
	return garden.MemoryLimits{}, nil
}
