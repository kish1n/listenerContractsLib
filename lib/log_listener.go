package lib

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	vaultapi "github.com/hashicorp/vault/api"
	"gitlab.com/distributed_lab/dig"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type LogListener interface {
	ExtractPrivateKey()
	ExtractSignerOpts()
	ExtractChainID(ctx context.Context)
	Run(ctx context.Context)
}

type logListener struct {
	RPC            *ethclient.Client
	AuditContracts []EventListener

	PrivateKey     *ecdsa.PrivateKey
	VaultAddress   string
	VaultMountPath string
	ChainID        *big.Int
	SignerOpts     *bind.TransactOpts
	TimeOut        time.Duration
}

func (l *logListener) ExtractPrivateKey() {
	conf := vaultapi.DefaultConfig()
	conf.Address = l.VaultAddress

	vaultClient, err := vaultapi.NewClient(conf)
	if err != nil {
		panic(errors.Wrap(err, "failed to initialize new client"))
	}

	token := struct {
		Token string `dig:"VAULT_TOKEN,clear"`
	}{}

	err = dig.Out(&token).Now()
	if err != nil {
		panic(errors.Wrap(err, "failed to dig out token"))
	}

	vaultClient.SetToken(token.Token)

	secret, err := vaultClient.KVv2(l.VaultMountPath).Get(context.Background(), "relayer")
	if err != nil {
		panic(errors.Wrap(err, "failed to get secret"))
	}

	vaultRelayerConf := struct {
		PrivateKey *ecdsa.PrivateKey
	}{}

	if err := figure.
		Out(&vaultRelayerConf).
		With(figure.EthereumHooks).
		From(secret.Data).
		Please(); err != nil {
		panic(errors.Wrap(err, "failed to figure out private_key"))
	}

	l.PrivateKey = vaultRelayerConf.PrivateKey
}

func (l *logListener) ExtractChainID(ctx context.Context, cancel context.CancelFunc) {
	chainID, err := l.RPC.ChainID(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			cancel()
			fmt.Println("Request to RPC timed out")
		} else {
			cancel()
			panic(fmt.Errorf("failed to get chain id: %w", err))
		}
		return
	}
	l.ChainID = chainID
}

func (l *logListener) Run(ctx context.Context) {
	for _, event := range l.AuditContracts {
		event.Build(ctx)
	}
}
