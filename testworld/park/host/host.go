//go:build testworld

package host

import (
	"fmt"
)

type Name string

const (
	Alice   Name = "Alice"
	Bob     Name = "Bob"
	Charlie Name = "Charlie"
)

type Host struct {
	mainAccount *Account

	controlUnit *ControlUnit
}

func NewHost(
	mainAccount *Account,
	controlUnit *ControlUnit,
) *Host {
	return &Host{
		mainAccount,
		controlUnit,
	}
}

func (h *Host) GetMainAccount() *Account {
	return h.mainAccount
}

func (h *Host) Start() error {
	return h.controlUnit.Start()
}

func (h *Host) Stop() error {
	return h.controlUnit.Stop()
}

func (h *Host) GetAPIURL() string {
	return fmt.Sprintf("http://localhost:%d", h.controlUnit.GetPodCfg().GetServerPort())
}

func (h *Host) GetServiceCtx() map[string]any {
	return h.controlUnit.GetServiceCtx()
}
