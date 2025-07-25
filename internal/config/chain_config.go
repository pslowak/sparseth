package config

import (
	"github.com/ethereum/go-ethereum/params"
	"math/big"
)

var (
	DefaultCancunBlobConfig = &params.BlobConfig{
		Target:         3,
		Max:            6,
		UpdateFraction: 3338477,
	}

	DefaultPragueBlobConfig = &params.BlobConfig{
		Target:         6,
		Max:            9,
		UpdateFraction: 5007716,
	}

	// AnvilChainConfig is the parameters to
	// run a node on the Anvil local network.
	AnvilChainConfig = &params.ChainConfig{
		ChainID:                 big.NewInt(31337),
		HomesteadBlock:          big.NewInt(0),
		DAOForkBlock:            nil,
		DAOForkSupport:          false,
		EIP150Block:             big.NewInt(0),
		EIP155Block:             big.NewInt(0),
		EIP158Block:             big.NewInt(0),
		ByzantiumBlock:          big.NewInt(0),
		ConstantinopleBlock:     big.NewInt(0),
		PetersburgBlock:         big.NewInt(0),
		IstanbulBlock:           big.NewInt(0),
		MuirGlacierBlock:        big.NewInt(0),
		BerlinBlock:             big.NewInt(0),
		LondonBlock:             big.NewInt(0),
		ArrowGlacierBlock:       big.NewInt(0),
		GrayGlacierBlock:        big.NewInt(0),
		MergeNetsplitBlock:      big.NewInt(0),
		ShanghaiTime:            newUint64(0),
		CancunTime:              newUint64(0),
		PragueTime:              newUint64(0),
		OsakaTime:               nil,
		VerkleTime:              nil,
		TerminalTotalDifficulty: big.NewInt(0),
		Ethash:                  new(params.EthashConfig),
		Clique:                  nil,
		BlobScheduleConfig: &params.BlobScheduleConfig{
			Cancun: DefaultCancunBlobConfig,
			Prague: DefaultPragueBlobConfig,
		},
	}

	// MainnetChainConfig is the parameters to
	// run a node on the Mainnet production network.
	MainnetChainConfig = params.MainnetChainConfig

	// SepoliaChainConfig is the parameters to
	// run a node on the Sepolia test network.
	SepoliaChainConfig = params.SepoliaChainConfig
)

func newUint64(val uint64) *uint64 {
	return &val
}
