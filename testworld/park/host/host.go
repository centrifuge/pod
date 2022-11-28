//go:build testworld

package host

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
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

	originKrp signature.KeyringPair
}

func NewHost(
	mainAccount *Account,
	controlUnit *ControlUnit,
	originKrp signature.KeyringPair,
) *Host {
	return &Host{
		mainAccount,
		controlUnit,
		originKrp,
	}
}

func (h *Host) GetOriginKeyringPair() signature.KeyringPair {
	return h.originKrp
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
