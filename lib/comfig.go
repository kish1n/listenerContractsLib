package lib

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type Market interface {
	NewLogListener(RPC *ethclient.Client, AuditContracts []EventListener, PrivateKey *ecdsa.PrivateKey,
		VaultAddress string, VaultMountPath string, TimeOut time.Duration, SignerOpts *bind.TransactOpts) *logListener
	ConfigureListener()
}

type market struct {
	once   comfig.Once
	getter kv.Getter
}

func NewMarket(getter kv.Getter) Market {
	return &market{
		getter: getter,
	}
}

func (c *market) NewLogListener(RPC *ethclient.Client, AuditContracts []EventListener, PrivateKey *ecdsa.PrivateKey,
	VaultAddress string, VaultMountPath string, TimeOut time.Duration, SignerOpts *bind.TransactOpts) *logListener {
	return c.once.Do(func() interface{} {
		ctx := context.Background()
		var listener logListener

		listener.RPC = RPC
		listener.AuditContracts = AuditContracts
		listener.PrivateKey = PrivateKey
		listener.VaultAddress = VaultAddress
		listener.VaultMountPath = VaultMountPath
		listener.TimeOut = TimeOut
		listener.SignerOpts = SignerOpts

		privateKey := listener.PrivateKey
		if privateKey == nil {
			listener.ExtractPrivateKey()
		}

		listener.ExtractChainID(context.WithTimeout(ctx, listener.TimeOut))

		return listener
	}).(*logListener)
}

func (c *market) ConfigureListener() {
	err := figure.Out(c).
		From(kv.MustGetStringMap(c.getter, "market")).
		With(figure.EthereumHooks).
		Please()
	if err != nil {
		panic(fmt.Errorf("failed to figure market market: %w", err))
	}
}
